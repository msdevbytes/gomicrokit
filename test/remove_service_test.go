package gmk_test

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	. "github.com/msdevbytes/gomicrokit/generator"
)

func TestRemoveService(t *testing.T) {
	t.Run("empty name returns error", func(t *testing.T) {
		err := RemoveService("", false)
		if err == nil {
			t.Error("expected error for empty name, got nil")
		}
	})
}

func TestReadHistory(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "readhistory-test")
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

	t.Run("history file not found", func(t *testing.T) {
		_, _, err := ReadHistory("nonexistent")
		if err == nil {
			t.Error("expected error when history file doesn't exist")
		}
	})

	t.Run("service not in history", func(t *testing.T) {
		history := map[string]struct {
			CreatedAt string   `json:"created_at"`
			Files     []string `json:"files"`
		}{
			"otherservice": {
				CreatedAt: time.Now().Format(time.RFC3339),
				Files:     []string{"file1.go"},
			},
		}
		data, _ := json.MarshalIndent(history, "", "  ")
		os.WriteFile(".gen_history.json", data, 0644)
		defer os.Remove(".gen_history.json")

		_, _, err := ReadHistory("nonexistent")
		if err == nil {
			t.Error("expected error when service not in history")
		}
	})

	t.Run("valid history entry", func(t *testing.T) {
		now := time.Now()
		history := map[string]struct {
			CreatedAt string   `json:"created_at"`
			Files     []string `json:"files"`
		}{
			"testservice": {
				CreatedAt: now.Format(time.RFC3339),
				Files:     []string{"file1.go", "file2.go"},
			},
		}
		data, _ := json.MarshalIndent(history, "", "  ")
		os.WriteFile(".gen_history.json", data, 0644)
		defer os.Remove(".gen_history.json")

		files, createdAt, err := ReadHistory("testservice")
		if err != nil {
			t.Errorf("ReadHistory() error = %v", err)
		}
		if len(files) != 2 {
			t.Errorf("expected 2 files, got %d", len(files))
		}
		if createdAt.IsZero() {
			t.Error("createdAt should not be zero")
		}
	})
}

func TestRemoveFromHistory(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "removehistory-test")
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

	t.Run("remove existing service", func(t *testing.T) {
		// Create history file with a service
		history := map[string]interface{}{
			"testservice": map[string]interface{}{
				"created_at": time.Now().Format(time.RFC3339),
				"files":      []string{"file1.go"},
			},
			"otherservice": map[string]interface{}{
				"created_at": time.Now().Format(time.RFC3339),
				"files":      []string{"file2.go"},
			},
		}
		data, _ := json.MarshalIndent(history, "", "  ")
		os.WriteFile(".gen_history.json", data, 0644)

		err := RemoveFromHistory("testservice")
		if err != nil {
			t.Errorf("RemoveFromHistory() error = %v", err)
		}

		// Verify the service was removed
		newData, _ := os.ReadFile(".gen_history.json")
		var newHistory map[string]interface{}
		json.Unmarshal(newData, &newHistory)

		if _, exists := newHistory["testservice"]; exists {
			t.Error("testservice should have been removed from history")
		}
		if _, exists := newHistory["otherservice"]; !exists {
			t.Error("otherservice should still exist in history")
		}

		os.Remove(".gen_history.json")
	})

	t.Run("no history file", func(t *testing.T) {
		err := RemoveFromHistory("anyservice")
		if err == nil {
			t.Error("expected error when history file doesn't exist")
		}
	})
}

func TestDeleteFiles(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "deletefiles-test")
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

	// Create test files
	os.WriteFile("file1.go", []byte("test"), 0644)
	os.WriteFile("file2.go", []byte("test"), 0644)

	// Delete files (including a non-existent one)
	DeleteFiles([]string{"file1.go", "file2.go", "nonexistent.go"})

	// Verify files were deleted
	if _, err := os.Stat("file1.go"); !os.IsNotExist(err) {
		t.Error("file1.go should have been deleted")
	}
	if _, err := os.Stat("file2.go"); !os.IsNotExist(err) {
		t.Error("file2.go should have been deleted")
	}
}

func TestRemoveWhitespace(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"no whitespace", "hello", "hello"},
		{"leading space", "  hello", "hello"},
		{"trailing space", "hello  ", "hello"},
		{"multiple spaces", "hello  world", "helloworld"},
		{"tabs", "hello\tworld", "helloworld"},
		{"mixed whitespace", "  hello \t world  ", "helloworld"},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RemoveWhitespace(tt.input)
			if got != tt.want {
				t.Errorf("RemoveWhitespace(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
