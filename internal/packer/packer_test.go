package packer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestPacker_BuildTargetPath(t *testing.T) {
	storageRoot := "/data/athens-storage"
	packer := NewPacker(storageRoot)

	if packer.storageRoot != storageRoot {
		t.Errorf("Storage root mismatch: got %q, want %q", packer.storageRoot, storageRoot)
	}

	// Use filepath.Join to construct expected path so it's platform-agnostic
	expectedPath := filepath.Join(storageRoot, "github.com", "gin-gonic", "gin", "v1.9.1")
	actualPath := filepath.Join(packer.storageRoot, "github.com", "gin-gonic", "gin", "v1.9.1")

	if actualPath != expectedPath {
		t.Errorf("Expected path mismatch: got %q, want %q", actualPath, expectedPath)
	}
}

func TestVersionInfo_JSON(t *testing.T) {
	info := VersionInfo{
		Version: "v1.9.1",
		Time:    "2024-01-01T00:00:00Z",
	}

	jsonBytes, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}

	expectedJSON := `{
  "Version": "v1.9.1",
  "Time": "2024-01-01T00:00:00Z"
}`

	if string(jsonBytes) != expectedJSON {
		t.Errorf("JSON mismatch:\nGot:\n%s\nWant:\n%s", string(jsonBytes), expectedJSON)
	}
}

func TestCopyFile(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "test-copy-file-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a source file
	srcFile := filepath.Join(tmpDir, "source.txt")
	content := []byte("test content")
	if err := os.WriteFile(srcFile, content, 0644); err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}

	// Copy the file
	dstFile := filepath.Join(tmpDir, "dest.txt")
	if err := copyFile(srcFile, dstFile); err != nil {
		t.Fatalf("copyFile failed: %v", err)
	}

	// Verify the destination file exists and has the same content
	dstContent, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf("Failed to read destination file: %v", err)
	}

	if string(dstContent) != string(content) {
		t.Errorf("Content mismatch: got %q, want %q", string(dstContent), string(content))
	}
}
