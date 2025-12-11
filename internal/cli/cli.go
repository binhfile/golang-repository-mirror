package cli

import (
	"fmt"
	"os"

	"github.com/example/athens-prefill/internal/gomod"
	"github.com/example/athens-prefill/internal/log"
	"github.com/example/athens-prefill/internal/packer"
	"github.com/example/athens-prefill/internal/resolver"
	"github.com/example/athens-prefill/internal/worker"
	"github.com/spf13/cobra"
)

var (
	modulesFile  string
	storageRoot  string
	workDir      string
	concurrency  int
	logLevel     string
)

var rootCmd = &cobra.Command{
	Use:   "athens-prefill",
	Short: "Prefill Athens proxy cache with Go modules",
	Long: `athens-prefill is a tool to download Go modules and their dependencies,
and package them in Athens disk storage format for offline usage.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPrefill()
	},
}

func init() {
	rootCmd.Flags().StringVarP(&modulesFile, "modules", "m", "", "Path to modules.txt file (required)")
	rootCmd.Flags().StringVarP(&storageRoot, "storage-root", "s", "", "Athens disk storage root directory")
	rootCmd.Flags().StringVarP(&workDir, "work-dir", "w", "", "Temporary work directory")
	rootCmd.Flags().IntVarP(&concurrency, "concurrency", "j", 4, "Number of concurrent workers")
	rootCmd.Flags().StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error)")

	rootCmd.MarkFlagRequired("modules")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runPrefill() error {
	// Setup logger
	log.SetLevelFromString(logLevel)

	// Validate and setup storage root
	if storageRoot == "" {
		storageRoot = os.Getenv("ATHENS_DISK_STORAGE_ROOT")
		if storageRoot == "" {
			return fmt.Errorf("--storage-root is required or set ATHENS_DISK_STORAGE_ROOT env var")
		}
	}

	// Setup work directory
	if workDir == "" {
		var err error
		workDir, err = os.MkdirTemp("", "athens-prefill-")
		if err != nil {
			return fmt.Errorf("failed to create temp directory: %w", err)
		}
		defer os.RemoveAll(workDir)
	} else {
		if err := os.MkdirAll(workDir, 0755); err != nil {
			return fmt.Errorf("failed to create work directory: %w", err)
		}
	}

	log.Info("Starting athens-prefill")
	log.Debug("Storage root: %s", storageRoot)
	log.Debug("Work directory: %s", workDir)
	log.Debug("Concurrency: %d", concurrency)

	// Parse modules list
	modules, err := parseModulesList(modulesFile)
	if err != nil {
		return fmt.Errorf("failed to parse modules list: %w", err)
	}
	log.Info("Loaded %d modules from %s", len(modules), modulesFile)

	// Resolve dependencies
	log.Info("Resolving dependencies...")
	resolver := resolver.NewResolver(workDir)
	resolvedModules, err := resolver.ResolveDependencies(modules)
	if err != nil {
		return fmt.Errorf("failed to resolve dependencies: %w", err)
	}
	log.Info("Resolved %d total modules", len(resolvedModules))

	// Pack modules
	log.Info("Packing modules into Athens format...")
	p := packer.NewPacker(storageRoot)
	pool := worker.NewPool(concurrency)

	successCount := 0
	failureCount := 0
	var failures []string

	for _, mod := range resolvedModules {
		mod := mod // capture loop variable
		pool.Submit(func() {
			if err := p.Pack(mod); err != nil {
				failureCount++
				log.Error("Failed to pack %s@%s: %v", mod.Path, mod.Version, err)
				failures = append(failures, fmt.Sprintf("%s@%s: %v", mod.Path, mod.Version, err))
			} else {
				successCount++
			}
		})
	}

	pool.Wait()

	// Print summary
	log.Info("=====================================")
	log.Info("Summary:")
	log.Info("  Total modules: %d", len(resolvedModules))
	log.Info("  Packed: %d", successCount)
	log.Info("  Failed: %d", failureCount)
	if failureCount > 0 {
		log.Error("Failed modules:")
		for _, f := range failures {
			log.Error("  - %s", f)
		}
	}
	log.Info("=====================================")

	if failureCount > 0 {
		return fmt.Errorf("%d modules failed to pack", failureCount)
	}

	log.Info("Done!")
	return nil
}

func parseModulesList(filepath string) ([]gomod.ModuleSpec, error) {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	return gomod.ParseModulesList(string(content))
}
