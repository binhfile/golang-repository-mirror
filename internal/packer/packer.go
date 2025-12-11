package packer

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/example/athens-prefill/internal/gomod"
	"github.com/example/athens-prefill/internal/log"
)

type Packer struct {
	storageRoot string
}

type VersionInfo struct {
	Version string `json:"Version"`
	Time    string `json:"Time"`
}

func NewPacker(storageRoot string) *Packer {
	return &Packer{storageRoot: storageRoot}
}

func (p *Packer) Pack(module gomod.Module) error {
	// Build target directory path
	targetDir := filepath.Join(p.storageRoot, module.Path, module.Version)

	// Check if already exists (idempotent)
	sourceZipPath := filepath.Join(targetDir, "source.zip")
	if _, err := os.Stat(sourceZipPath); err == nil {
		log.Info("Module already exists: %s@%s, skipping", module.Path, module.Version)
		return nil
	}

	log.Info("Packing module: %s@%s", module.Path, module.Version)

	// Create target directory
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Copy go.mod file
	srcGoMod := filepath.Join(module.Dir, "go.mod")
	dstGoMod := filepath.Join(targetDir, "go.mod")
	if err := copyFile(srcGoMod, dstGoMod); err != nil {
		// go.mod might not exist, log warning but continue
		log.Debug("Warning: go.mod not found in %s: %v", module.Dir, err)
		// Create minimal go.mod
		if err := os.WriteFile(dstGoMod, []byte(fmt.Sprintf("module %s\n\ngo 1.21\n", module.Path)), 0644); err != nil {
			return fmt.Errorf("failed to create go.mod: %w", err)
		}
	}

	// Create .info file
	infoFile := filepath.Join(targetDir, module.Version+".info")
	versionInfo := VersionInfo{
		Version: module.Version,
		Time:    time.Now().UTC().Format(time.RFC3339),
	}
	infoData, err := json.MarshalIndent(versionInfo, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal version info: %w", err)
	}
	if err := os.WriteFile(infoFile, infoData, 0644); err != nil {
		return fmt.Errorf("failed to write version info: %w", err)
	}

	// Create source.zip
	if err := zipDirectory(module.Dir, sourceZipPath, module.Path); err != nil {
		return fmt.Errorf("failed to create source.zip: %w", err)
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

func zipDirectory(sourceDir, zipPath, modulePath string) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return fmt.Errorf("failed to create zip file: %w", err)
	}
	defer zipFile.Close()

	writer := zip.NewWriter(zipFile)
	defer writer.Close()

	err = filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip .git directories and cache files
		if strings.Contains(path, ".git") || strings.Contains(path, ".mod.cache") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		// Add module path prefix to zip entries
		zipEntryPath := filepath.Join(modulePath+"@"+extractVersion(sourceDir), relPath)
		zipEntryPath = filepath.ToSlash(zipEntryPath)

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = zipEntryPath

		writer, err := writer.CreateHeader(header)
		if err != nil {
			return err
		}

		_, err = io.Copy(writer, file)
		return err
	})

	return err
}

func extractVersion(moduleDir string) string {
	// Try to extract version from go.mod
	goModPath := filepath.Join(moduleDir, "go.mod")
	if content, err := os.ReadFile(goModPath); err == nil {
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "module ") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					moduleName := parts[1]
					if idx := strings.LastIndex(moduleName, "@"); idx > 0 {
										return moduleName[idx+1:]
					}
				}
			}
		}
	}
	return "unknown"
}
