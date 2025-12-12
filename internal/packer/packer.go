package packer

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/example/go-mod-clone/internal/gomod"
	"github.com/example/go-mod-clone/internal/log"
)

type Packer struct {
	storageRoot string
}

func NewPacker(storageRoot string) *Packer {
	return &Packer{storageRoot: storageRoot}
}

func (p *Packer) Pack(module gomod.Module) error {
	// Build target @v directory path
	atVDir := filepath.Join(p.storageRoot, module.Path, "@v")

	// Check if already exists (idempotent) - check for .zip file
	targetZip := filepath.Join(atVDir, module.Version+".zip")
	if _, err := os.Stat(targetZip); err == nil {
		log.Info("Module already exists: %s@%s, skipping", module.Path, module.Version)
		return nil
	}

	log.Info("Packing module: %s@%s", module.Path, module.Version)

	// Create @v directory
	if err := os.MkdirAll(atVDir, 0755); err != nil {
		return fmt.Errorf("failed to create @v directory: %w", err)
	}

	// Copy .info file from cache
	if module.InfoFile != "" {
		targetInfo := filepath.Join(atVDir, module.Version+".info")
		if err := copyFile(module.InfoFile, targetInfo); err != nil {
			log.Debug("Warning: failed to copy .info file: %v", err)
		}
	}

	// Copy .mod file from cache
	if module.ModFile != "" {
		targetMod := filepath.Join(atVDir, module.Version+".mod")
		if err := copyFile(module.ModFile, targetMod); err != nil {
			log.Debug("Warning: failed to copy .mod file: %v", err)
		}
	}

	// Copy .zip file from cache
	if module.ZipFile != "" {
		if err := copyFile(module.ZipFile, targetZip); err != nil {
			return fmt.Errorf("failed to copy .zip file: %w", err)
		}
	}

	// Update list file (with file locking for concurrent access)
	if err := p.updateListFile(atVDir, module.Version); err != nil {
		return fmt.Errorf("failed to update list file: %w", err)
	}

	log.Debug("Successfully packed: %s@%s", module.Path, module.Version)
	return nil
}

func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}

func (p *Packer) updateListFile(atVDir, version string) error {
	listPath := filepath.Join(atVDir, "list")

	// Use file locking for concurrent writes
	lockPath := listPath + ".lock"
	lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer lockFile.Close()
	defer os.Remove(lockPath)

	// Read existing versions
	versions := make(map[string]bool)
	if data, err := os.ReadFile(listPath); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if line = strings.TrimSpace(line); line != "" {
				versions[line] = true
			}
		}
	}

	// Add new version if not present
	if !versions[version] {
		versions[version] = true

		// Write all versions sorted
		var versionList []string
		for v := range versions {
			versionList = append(versionList, v)
		}
		sort.Strings(versionList)

		content := strings.Join(versionList, "\n") + "\n"
		if err := os.WriteFile(listPath, []byte(content), 0644); err != nil {
			return err
		}
	}

	return nil
}

