package gomod

import (
	"testing"
)

func TestParseModulesList(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    int
		wantErr bool
	}{
		{
			name: "simple modules with versions",
			content: `github.com/gin-gonic/gin@v1.9.1
github.com/sirupsen/logrus@v1.9.3`,
			want:    2,
			wantErr: false,
		},
		{
			name: "modules with and without versions",
			content: `github.com/gin-gonic/gin@v1.9.1
# Comment line
github.com/golang/protobuf

`,
			want:    2,
			wantErr: false,
		},
		{
			name:    "empty file",
			content: "",
			want:    0,
			wantErr: false,
		},
		{
			name: "only comments and empty lines",
			content: `# This is a comment
# Another comment

`,
			want:    0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseModulesList(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseModulesList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.want {
				t.Errorf("ParseModulesList() got %d modules, want %d", len(got), tt.want)
			}
		})
	}
}

func TestIsValidSemver(t *testing.T) {
	tests := []struct {
		version string
		want    bool
	}{
		{"v1.9.1", true},
		{"v0.0.1", true},
		{"v1.0.0-beta", true},
		{"1.0.0", false},
		{"devel", false},
		{"", false},
		{"v", false},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			got := IsValidSemver(tt.version)
			if got != tt.want {
				t.Errorf("IsValidSemver(%q) = %v, want %v", tt.version, got, tt.want)
			}
		})
	}
}

func TestParseModulesListSpec(t *testing.T) {
	content := `github.com/gin-gonic/gin@v1.9.1
google.golang.org/grpc`

	specs, err := ParseModulesList(content)
	if err != nil {
		t.Fatalf("ParseModulesList failed: %v", err)
	}

	if len(specs) != 2 {
		t.Errorf("Expected 2 specs, got %d", len(specs))
	}

	if specs[0].Path != "github.com/gin-gonic/gin" || specs[0].Version != "v1.9.1" {
		t.Errorf("First spec incorrect: %+v", specs[0])
	}

	if specs[1].Path != "google.golang.org/grpc" || specs[1].Version != "" {
		t.Errorf("Second spec incorrect: %+v", specs[1])
	}
}
