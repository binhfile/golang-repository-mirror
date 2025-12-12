package resolver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/example/athens-prefill/internal/gomod"
	"github.com/example/athens-prefill/internal/log"
)

type Resolver struct {
	workDir string
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
	return &Resolver{workDir: workDir}
}

func (r *Resolver) ResolveDependencies(specs []gomod.ModuleSpec) ([]gomod.Module, error) {
	// Create a temporary directory for go mod resolution
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

	// Run go get -d for each module to download (not build)
	for _, spec := range specs {
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
	}

	// Download ALL modules (including transitive dependencies)
	log.Debug("Running 'go mod download all' to download all modules")
	downloadAllCmd := exec.Command("go", "mod", "download", "all")
	downloadAllCmd.Dir = tempDir
	if err := downloadAllCmd.Run(); err != nil {
		log.Debug("go mod download all completed with some warnings")
	}

	// Now get paths using go mod download -json
	log.Debug("Running 'go mod download -json all' to get module paths")
	downloadJsonCmd := exec.Command("go", "mod", "download", "-json", "all")
	downloadJsonCmd.Dir = tempDir

	var dlStdout bytes.Buffer
	downloadJsonCmd.Stdout = &dlStdout

	if err := downloadJsonCmd.Run(); err != nil {
		log.Debug("go mod download -json completed")
	}

	// Build maps of module paths to Dir/Zip/Info/GoMod from download output
	modulePathMap := make(map[string]string)  // Path -> Dir
	moduleZipMap := make(map[string]string)   // Path -> Zip file path
	moduleInfoMap := make(map[string]string)  // Path -> Info file path
	moduleModMap := make(map[string]string)   // Path -> GoMod file path
	dlDecoder := json.NewDecoder(&dlStdout)
	for dlDecoder.More() {
		var dlInfo struct {
			Path  string `json:"Path"`
			Dir   string `json:"Dir"`
			Zip   string `json:"Zip"`
			Info  string `json:"Info"`
			GoMod string `json:"GoMod"`
		}
		if err := dlDecoder.Decode(&dlInfo); err != nil {
			continue
		}
		if dlInfo.Path != "" {
			if dlInfo.Dir != "" {
				modulePathMap[dlInfo.Path] = dlInfo.Dir
				log.Debug("Module download info: %s -> %s", dlInfo.Path, dlInfo.Dir)
			}
			if dlInfo.Zip != "" {
				moduleZipMap[dlInfo.Path] = dlInfo.Zip
				log.Debug("Module zip path: %s -> %s", dlInfo.Path, dlInfo.Zip)
			}
			if dlInfo.Info != "" {
				moduleInfoMap[dlInfo.Path] = dlInfo.Info
				log.Debug("Module info path: %s -> %s", dlInfo.Path, dlInfo.Info)
			}
			if dlInfo.GoMod != "" {
				moduleModMap[dlInfo.Path] = dlInfo.GoMod
				log.Debug("Module mod path: %s -> %s", dlInfo.Path, dlInfo.GoMod)
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
		if moduleDir == "" {
			// For indirect dependencies, try to find them in the download map
			if dir, ok := modulePathMap[info.Path]; ok {
				moduleDir = dir
				log.Debug("Found module Dir: %s@%s -> %s", info.Path, info.Version, moduleDir)
			} else {
				log.Debug("Skipping %s@%s - cannot locate in module cache", info.Path, info.Version)
				continue
			}
		}

		modules = append(modules, gomod.Module{
			Path:     info.Path,
			Version:  info.Version,
			Dir:      moduleDir,
			InfoFile: moduleInfoMap[info.Path],
			ModFile:  moduleModMap[info.Path],
			ZipFile:  moduleZipMap[info.Path],
		})

		log.Debug("Resolved: %s@%s -> %s", info.Path, info.Version, moduleDir)
	}

	return modules, nil
}
