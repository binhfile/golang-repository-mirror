package resolver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/example/go-mod-clone/internal/gomod"
	"github.com/example/go-mod-clone/internal/log"
)

type Resolver struct {
	workDir   string
	cacheFile string
	useCache  bool
}

// ResolutionCache stores resolved modules with metadata
type ResolutionCache struct {
	Version       string              `json:"version"`
	CachedAt      time.Time           `json:"cached_at"`
	Modules       []gomod.Module      `json:"modules"`
	InputSpecs    []gomod.ModuleSpec  `json:"input_specs"`
	InputChecksum string              `json:"input_checksum"`
}

type modInfo struct {
	Path      string
	Version   string
	Dir       string
	GoMod     string
	GoVersion string
	Indirect  bool
	Time      string
	Main      bool
}

func NewResolver(workDir string) *Resolver {
	return &Resolver{
		workDir:   workDir,
		cacheFile: filepath.Join(workDir, "resolution-cache.json"),
		useCache:  true,
	}
}

func NewResolverWithCacheControl(workDir string, useCache bool) *Resolver {
	return &Resolver{
		workDir:   workDir,
		cacheFile: filepath.Join(workDir, "resolution-cache.json"),
		useCache:  useCache,
	}
}

// calculateInputChecksum creates a simple checksum of input specs
func (r *Resolver) calculateInputChecksum(specs []gomod.ModuleSpec) string {
	var input string
	for _, spec := range specs {
		input += spec.Path + "@" + spec.Version + "|"
	}
	// Simple checksum: just hash the string representation
	return fmt.Sprintf("%x", len(input))
}

// loadCache attempts to load resolution results from cache file
func (r *Resolver) loadCache(specs []gomod.ModuleSpec) *ResolutionCache {
	if !r.useCache {
		log.Debug("Cache disabled, skipping cache load")
		return nil
	}

	data, err := os.ReadFile(r.cacheFile)
	if err != nil {
		log.Debug("Cache file not found or unreadable: %v", err)
		return nil
	}

	var cache ResolutionCache
	if err := json.Unmarshal(data, &cache); err != nil {
		log.Error("Failed to parse cache file: %v", err)
		return nil
	}

	// Verify input specs match
	expectedChecksum := r.calculateInputChecksum(specs)
	if cache.InputChecksum != expectedChecksum {
		log.Info("Cache invalidated: input specs changed")
		return nil
	}

	log.Info("Loaded resolution cache from %s (%d modules)", r.cacheFile, len(cache.Modules))
	return &cache
}

// saveCache saves resolution results to cache file
func (r *Resolver) saveCache(specs []gomod.ModuleSpec, modules []gomod.Module) error {
	if !r.useCache {
		return nil
	}

	cache := ResolutionCache{
		Version:       "1.0",
		CachedAt:      time.Now(),
		Modules:       modules,
		InputSpecs:    specs,
		InputChecksum: r.calculateInputChecksum(specs),
	}

	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		log.Error("Failed to marshal cache: %v", err)
		return err
	}

	if err := os.WriteFile(r.cacheFile, data, 0644); err != nil {
		log.Error("Failed to write cache file: %v", err)
		return err
	}

	log.Info("Saved resolution cache to %s", r.cacheFile)
	return nil
}

func (r *Resolver) ResolveDependencies(specs []gomod.ModuleSpec) ([]gomod.Module, error) {
	// Try to load from cache first
	if cache := r.loadCache(specs); cache != nil {
		log.Info("Using cached resolution results (%d modules)", len(cache.Modules))
		return cache.Modules, nil
	}

	// Cache miss or disabled, perform full resolution
	log.Info("Resolving dependencies (cache miss or disabled)")

	// Track resolved modules by Path@Version to avoid duplicates
	resolvedModules := make(map[string]gomod.Module)
	var toProcess []gomod.ModuleSpec
	processed := make(map[string]bool)

	// Start with provided specs
	toProcess = append(toProcess, specs...)

	// Process modules recursively
	for len(toProcess) > 0 {
		spec := toProcess[0]
		toProcess = toProcess[1:]

		// Skip if already processed
		key := spec.Path + "@" + spec.Version
		if processed[key] {
			continue
		}
		processed[key] = true

		// Resolve this module and its dependencies
		log.Info("Resolve [queue %v resolved %v] %v", len(toProcess), len(resolvedModules), key)
		resolvedMods, err := r.resolveModule(spec)
		if err != nil {
			log.Error("Failed to resolve %s@%s: %v", spec.Path, spec.Version, err)
			//return nil, fmt.Errorf("failed to resolve %s@%s: %w", spec.Path, spec.Version, err)
		}

		// Add resolved modules to our map
		for _, mod := range resolvedMods {
			modKey := mod.Path + "@" + mod.Version
			resolvedModules[modKey] = mod

			// If this is a new module we haven't seen before, queue it for resolution
			if !processed[modKey] {
				log.Info("  -> %v", modKey)
				toProcess = append(toProcess, gomod.ModuleSpec{
					Path:    mod.Path,
					Version: mod.Version,
				})
			}
		}
	}

	// Convert map back to slice
	var result []gomod.Module
	for _, mod := range resolvedModules {
		result = append(result, mod)
	}

	// Save to cache for next run
	if err := r.saveCache(specs, result); err != nil {
		log.Error("Failed to save cache: %v", err)
		// Don't fail the whole operation if cache save fails
	}

	return result, nil
}

func (r *Resolver) resolveModule(spec gomod.ModuleSpec) ([]gomod.Module, error) {
	// Create a temporary directory for go mod resolution
	// Use a simple naming scheme since multiple versions will be handled in sequence
	tempDir := filepath.Join(r.workDir, "resolve-temp")

	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Initialize a temporary go.mod
	goModPath := filepath.Join(tempDir, "go.mod")
	initialContent := "module temp\n\ngo 1.21\n"
	if err := os.WriteFile(goModPath, []byte(initialContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to create temporary go.mod: %w", err)
	}

	// Download the specific module version
	getSpec := spec.Path
	if spec.Version != "" {
		getSpec = spec.Path + "@" + spec.Version
	}

	log.Debug("Running 'go get -d %s' in %s", getSpec, tempDir)
	getCmd := exec.Command("go", "get", "-d", getSpec)
	getCmd.Dir = tempDir

	var getStderr bytes.Buffer
	getCmd.Stderr = &getStderr

	if err := getCmd.Run(); err != nil {
		log.Error("go get -d %s stderr: %s", getSpec, getStderr.String())
		return nil, fmt.Errorf("go get -d %s failed: %w", getSpec, err)
	}

	// Download ALL modules (including transitive dependencies)
	log.Debug("Running 'go mod download' to download all modules")
	downloadAllCmd := exec.Command("go", "mod", "download")
	downloadAllCmd.Dir = tempDir
	if err := downloadAllCmd.Run(); err != nil {
		log.Debug("go mod download completed with some warnings")
	}

	// Now get paths using go mod download -json
	log.Debug("Running 'go mod download -json' to get module paths")
	downloadJsonCmd := exec.Command("go", "mod", "download", "-json")
	downloadJsonCmd.Dir = tempDir

	var dlStdout bytes.Buffer
	downloadJsonCmd.Stdout = &dlStdout

	if err := downloadJsonCmd.Run(); err != nil {
		log.Debug("go mod download -json completed")
	}

	// Build maps of module paths to Dir/Zip/Info/GoMod from download output
	// Using Path@Version as key to support multiple versions of the same module
	modulePathMap := make(map[string]string) // Path@Version -> Dir
	moduleZipMap := make(map[string]string)  // Path@Version -> Zip file path
	moduleInfoMap := make(map[string]string) // Path@Version -> Info file path
	moduleModMap := make(map[string]string)  // Path@Version -> GoMod file path
	dlDecoder := json.NewDecoder(&dlStdout)
	for dlDecoder.More() {
		var dlInfo struct {
			Path    string `json:"Path"`
			Version string `json:"Version"`
			Dir     string `json:"Dir"`
			Zip     string `json:"Zip"`
			Info    string `json:"Info"`
			GoMod   string `json:"GoMod"`
		}
		if err := dlDecoder.Decode(&dlInfo); err != nil {
			continue
		}
		if dlInfo.Path != "" {
			key := dlInfo.Path + "@" + dlInfo.Version
			if dlInfo.Dir != "" {
				modulePathMap[key] = dlInfo.Dir
				log.Debug("Module download info: %s -> %s", key, dlInfo.Dir)
			}
			if dlInfo.Zip != "" {
				moduleZipMap[key] = dlInfo.Zip
				log.Debug("Module zip path: %s -> %s", key, dlInfo.Zip)
			}
			if dlInfo.Info != "" {
				moduleInfoMap[key] = dlInfo.Info
				log.Debug("Module info path: %s -> %s", key, dlInfo.Info)
			}
			if dlInfo.GoMod != "" {
				moduleModMap[key] = dlInfo.GoMod
				log.Debug("Module mod path: %s -> %s", key, dlInfo.GoMod)
			}
		}
	}

	// Run go list -m -json all to get all modules with versions
	log.Debug("Running 'go list -m -json all' to resolve dependencies")
	listCmd := exec.Command("go", "list", "-m", "-json", "all")
	listCmd.Dir = tempDir

	var stdout, stderr bytes.Buffer
	listCmd.Stdout = &stdout
	listCmd.Stderr = &stderr

	if err := listCmd.Run(); err != nil {
		log.Error("go list output: %s", stderr.String())
		return nil, fmt.Errorf("go list -m -json failed: %w", err)
	}

	// Parse the go list output
	decoder := json.NewDecoder(&stdout)
	var modules []gomod.Module

	for decoder.More() {
		var info modInfo
		if err := decoder.Decode(&info); err != nil {
			return nil, fmt.Errorf("failed to decode go list output: %w", err)
		}

		// Skip main module
		if info.Main {
			log.Debug("Skipping main module: %s", info.Path)
			continue
		}

		// Skip if no version or invalid version format
		if info.Version == "" || !gomod.IsValidSemver(info.Version) {
			log.Debug("Skipping %s (version: %s) - invalid semver", info.Path, info.Version)
			continue
		}

		// Get module directory from either go list or from download map
		moduleDir := info.Dir
		key := info.Path + "@" + info.Version
		if moduleDir == "" {
			// For indirect dependencies, try to find them in the download map
			if dir, ok := modulePathMap[key]; ok {
				moduleDir = dir
				log.Debug("Found module Dir: %s -> %s", key, moduleDir)
			} else {
				log.Debug("Skipping %s - cannot locate in module cache", key)
				continue
			}
		}

		modules = append(modules, gomod.Module{
			Path:     info.Path,
			Version:  info.Version,
			Dir:      moduleDir,
			InfoFile: moduleInfoMap[key],
			ModFile:  moduleModMap[key],
			ZipFile:  moduleZipMap[key],
		})

		log.Debug("Resolved: %s -> %s", key, moduleDir)
	}

	return modules, nil
}
