package generator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsValidGoIdent(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		// Valid identifiers
		{"simple lowercase", "user", true},
		{"simple uppercase", "User", true},
		{"with digits", "user123", true},
		{"with underscore", "user_name", true},
		{"starts with underscore", "_private", true},
		{"single letter", "x", true},
		{"camelCase", "userName", true},
		{"PascalCase", "UserName", true},
		{"all caps", "API", true},
		{"mixed", "myAPI_v2", true},
		{"unicode letter", "événement", true},
		{"chinese", "用户", true},

		// Invalid identifiers
		{"empty string", "", false},
		{"starts with digit", "123user", false},
		{"contains hyphen", "user-name", false},
		{"contains space", "user name", false},
		{"contains dot", "user.name", false},
		{"starts with digit underscore", "1_user", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidGoIdent(tt.input)
			if got != tt.want {
				t.Errorf("isValidGoIdent(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestConvertToTitleCaseNoSpaces(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty string", "", ""},
		{"single word lowercase", "user", "User"},
		{"single word uppercase", "USER", "User"},
		{"already title case", "User", "User"},
		{"camelCase preserves transitions", "userName", "UserName"},
		{"PascalCase preserved", "UserName", "UserName"},
		{"with spaces", "user name", "Username"},
		{"multiple spaces", "user  name", "Username"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertToTitleCaseNoSpaces(tt.input)
			if got != tt.want {
				t.Errorf("convertToTitleCaseNoSpaces(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestGetGoModule(t *testing.T) {
	// Create a temporary directory with a go.mod file
	tmpDir, err := os.MkdirTemp("", "gomodule-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Save current directory and change to temp dir
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current dir: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change to temp dir: %v", err)
	}

	t.Run("with valid go.mod", func(t *testing.T) {
		goModContent := "module github.com/example/myproject\n\ngo 1.21\n"
		if err := os.WriteFile("go.mod", []byte(goModContent), 0644); err != nil {
			t.Fatalf("failed to write go.mod: %v", err)
		}
		defer os.Remove("go.mod")

		got := GetGoModule()
		want := "github.com/example/myproject"
		if got != want {
			t.Errorf("GetGoModule() = %q, want %q", got, want)
		}
	})

	t.Run("without go.mod", func(t *testing.T) {
		os.Remove("go.mod") // ensure no go.mod exists

		got := GetGoModule()
		want := "unknown-module"
		if got != want {
			t.Errorf("GetGoModule() = %q, want %q", got, want)
		}
	})
}

func TestWriteHistory(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "history-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current dir: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change to temp dir: %v", err)
	}

	t.Run("write new history", func(t *testing.T) {
		files := []string{"file1.go", "file2.go"}
		err := writeHistory("TestService", files)
		if err != nil {
			t.Errorf("writeHistory() error = %v", err)
		}

		// Verify file was created
		if _, err := os.Stat(".gen_history.json"); os.IsNotExist(err) {
			t.Error("history file was not created")
		}
	})

	t.Run("update existing history", func(t *testing.T) {
		files := []string{"file3.go", "file4.go"}
		err := writeHistory("AnotherService", files)
		if err != nil {
			t.Errorf("writeHistory() error = %v", err)
		}
	})
}

func TestCreateProjectStructure(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "project-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current dir: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change to temp dir: %v", err)
	}

	t.Run("create new project", func(t *testing.T) {
		err := CreateProjectStructure("myproject", true)
		if err != nil {
			t.Errorf("CreateProjectStructure() error = %v", err)
		}

		// Verify directories were created
		expectedDirs := []string{
			"myproject/cmd",
			"myproject/internal/handler",
			"myproject/internal/service",
			"myproject/pkg/logger",
		}
		for _, dir := range expectedDirs {
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				t.Errorf("expected directory %s was not created", dir)
			}
		}

		// Verify Dockerfile exists when withDocker=true
		if _, err := os.Stat("myproject/Dockerfile"); os.IsNotExist(err) {
			t.Error("Dockerfile was not created when withDocker=true")
		}
	})

	t.Run("create project without docker", func(t *testing.T) {
		err := CreateProjectStructure("nodockerproject", false)
		if err != nil {
			t.Errorf("CreateProjectStructure() error = %v", err)
		}

		// Verify Dockerfile does NOT exist when withDocker=false
		dockerfilePath := filepath.Join("nodockerproject", "Dockerfile")
		if _, err := os.Stat(dockerfilePath); !os.IsNotExist(err) {
			t.Error("Dockerfile should not exist when withDocker=false")
		}
	})

	t.Run("empty project name", func(t *testing.T) {
		err := CreateProjectStructure("", true)
		if err == nil {
			t.Error("expected error for empty project name, got nil")
		}
	})

	t.Run("project already exists", func(t *testing.T) {
		// Create directory first
		os.Mkdir("existingproject", 0755)

		err := CreateProjectStructure("existingproject", true)
		if err == nil {
			t.Error("expected error for existing project, got nil")
		}
	})
}

func TestServiceOptions(t *testing.T) {
	// Test that ServiceOptions struct has expected fields
	opt := ServiceOptions{
		Name:       "TestService",
		ModulePath: "github.com/test/project",
		Force:      true,
		DryRun:     false,
	}

	if opt.Name != "TestService" {
		t.Errorf("ServiceOptions.Name = %q, want %q", opt.Name, "TestService")
	}
	if opt.ModulePath != "github.com/test/project" {
		t.Errorf("ServiceOptions.ModulePath = %q, want %q", opt.ModulePath, "github.com/test/project")
	}
	if !opt.Force {
		t.Error("ServiceOptions.Force should be true")
	}
	if opt.DryRun {
		t.Error("ServiceOptions.DryRun should be false")
	}
}
