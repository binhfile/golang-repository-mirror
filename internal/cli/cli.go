package cli

import (
	"fmt"
	"os"

	"github.com/example/go-mod-clone/internal/gomod"
	"github.com/example/go-mod-clone/internal/log"
	"github.com/example/go-mod-clone/internal/packer"
	"github.com/example/go-mod-clone/internal/resolver"
	"github.com/example/go-mod-clone/internal/server"
	"github.com/example/go-mod-clone/internal/worker"
	"github.com/spf13/cobra"
)

var (
	modulesFile    string
	storageRoot    string
	workDir        string
	concurrency    int
	logLevel       string
	host           string
	port           int
	useCache       bool
	clearCache     bool
)

var rootCmd = &cobra.Command{
	Use:   "go-mod-clone",
	Short: "Clone Go modules and their dependencies to a local repository",
	Long: `go-mod-clone is a tool to download Go modules and their dependencies,
and organize them in Go module proxy format for offline usage.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPrefill()
	},
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run as a Go module proxy server",
	Long: `Start a file server that serves modules in Go module proxy format.
Configure your Go environment with: export GOPROXY=http://host:port`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runServer()
	},
}

func init() {
	// Root command flags
	rootCmd.Flags().StringVarP(&modulesFile, "modules", "m", "", "Path to modules.txt file (required)")
	rootCmd.Flags().StringVarP(&storageRoot, "storage-root", "s", "", "Athens disk storage root directory")
	rootCmd.Flags().StringVarP(&workDir, "work-dir", "w", "", "Temporary work directory")
	rootCmd.Flags().IntVarP(&concurrency, "concurrency", "j", 4, "Number of concurrent workers")
	rootCmd.Flags().StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	rootCmd.Flags().BoolVar(&useCache, "use-cache", true, "Use resolution cache to speed up subsequent runs")
	rootCmd.Flags().BoolVar(&clearCache, "clear-cache", false, "Clear resolution cache before starting")

	rootCmd.MarkFlagRequired("modules")

	// Server command flags
	serverCmd.Flags().StringVarP(&storageRoot, "storage-root", "s", "", "Module storage root directory (required)")
	serverCmd.Flags().StringVarP(&host, "host", "H", "localhost", "Server host address")
	serverCmd.Flags().IntVarP(&port, "port", "p", 3000, "Server port")
	serverCmd.Flags().StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error)")

	serverCmd.MarkFlagRequired("storage-root")

	// Add subcommand to root
	rootCmd.AddCommand(serverCmd)
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
		workDir, err = os.MkdirTemp("", "go-mod-clone-")
		if err != nil {
			return fmt.Errorf("failed to create temp directory: %w", err)
		}
		defer os.RemoveAll(workDir)
	} else {
		if err := os.MkdirAll(workDir, 0755); err != nil {
			return fmt.Errorf("failed to create work directory: %w", err)
		}
	}

	log.Info("Starting go-mod-clone")
	log.Debug("Storage root: %s", storageRoot)
	log.Debug("Work directory: %s", workDir)
	log.Debug("Concurrency: %d", concurrency)
	log.Debug("Use cache: %v", useCache)

	// Clear cache if requested
	if clearCache {
		cacheFile := fmt.Sprintf("%s/resolution-cache.json", workDir)
		if err := os.Remove(cacheFile); err != nil && !os.IsNotExist(err) {
			log.Warn("Failed to clear cache: %v", err)
		} else {
			log.Info("Cache cleared")
		}
	}

	// Parse modules list
	modules, err := parseModulesList(modulesFile)
	if err != nil {
		return fmt.Errorf("failed to parse modules list: %w", err)
	}
	log.Info("Loaded %d modules from %s", len(modules), modulesFile)

	// Resolve dependencies with cache support
	log.Info("Resolving dependencies...")
	res := resolver.NewResolverWithCacheControl(workDir, useCache)
	resolvedModules, err := res.ResolveDependencies(modules)
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

	modIdx := 1
	for modKey, mod := range resolvedModules {
		mod := mod // capture loop variable
		log.Info("Pack %v/%v %v", modIdx, len(resolvedModules), modKey)
		modIdx = modIdx + 1
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

func runServer() error {
	// Setup logger
	log.SetLevelFromString(logLevel)

	log.Info("Starting go-mod-clone server")
	log.Info("Storage root: %s", storageRoot)
	log.Info("Host: %s", host)
	log.Info("Port: %d", port)

	// Create and start server
	srv := server.NewServer(storageRoot, host, port)
	return srv.Start()
}
