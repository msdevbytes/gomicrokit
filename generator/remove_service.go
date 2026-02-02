package generator

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

func RemoveService(name string, force bool) error {
	if name == "" {
		return fmt.Errorf("please provide a service name")
	}

	name = strings.ToLower(name)
	service := strings.ToUpper(string(name[0])) + name[1:]

	files, createdAt, err := readHistory(name)
	if err != nil {
		return fmt.Errorf("cannot read history: %w", err)
	}

	// Enforce freshness (1 minute max)
	if time.Since(createdAt) > 1*time.Minute && !force {
		return fmt.Errorf("'%s' was created more than 1 minute ago (%s). Use --force to override", name, createdAt.Format(time.RFC822))
	}

	deleteFiles(files)

	cleanFromFile("internal/service/container.go", []string{
		fmt.Sprintf("%s *%sService", service, service),
		fmt.Sprintf("%s: New%sService(repository.New%sRepository(db))", service, service, service),
		fmt.Sprintf("New%sService", service),
	})

	fmt.Println("🔥 Removed:", fmt.Sprintf("%s/internal/repository", GetGoModule()))
	cleanUnusedImport("internal/service/container.go", fmt.Sprintf("%s/internal/repository", GetGoModule()))

	cleanFromFile("internal/api/router.go", []string{
		fmt.Sprintf("New%sHandler(svc.%s).Register(api.Group(\"/%ss\"))", service, service, name),
	})

	if err := removeFromHistory(name); err != nil {
		fmt.Printf("⚠️ Warning: failed to update history: %v\n", err)
	}
	fmt.Printf("✅ Service '%s' removed successfully.\n", name)
	return nil
}


func deleteFiles(paths []string) {
	for _, p := range paths {
		if err := os.Remove(p); err == nil {
			fmt.Println("🗑️ Removed:", p)
		} else if os.IsNotExist(err) {
			fmt.Println("⚠️ Not found:", p)
		} else {
			fmt.Println("❌ Error removing", p, ":", err)
		}
	}
}

func cleanFromFile(path string, patterns []string) {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("❌ Cannot open:", path)
		return
	}

	lines := strings.Split(string(data), "\n")
	filtered := []string{}

	for _, line := range lines {
		shouldSkip := false
		for _, pattern := range patterns {
			// Match ignoring indentation
			if strings.Contains(removeWhitespace(line), removeWhitespace(pattern)) {
				shouldSkip = true
				break
			}
		}
		if !shouldSkip {
			filtered = append(filtered, line)
		}
	}

	if err := os.WriteFile(path, []byte(strings.Join(filtered, "\n")), 0644); err != nil {
		fmt.Println("❌ Error writing cleaned file:", err)
	} else {
		fmt.Println("✂️ Cleaned up:", path)
	}
}

func cleanUnusedImport(path string, importPath string) {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("❌ Cannot open:", path)
		return
	}

	lines := strings.Split(string(data), "\n")
	importUsed := false

	// Check if the import path is used anywhere in code (excluding import line)
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "repository.") && !strings.HasPrefix(trimmed, "import") && !strings.Contains(trimmed, importPath) {
			importUsed = true
			break
		}
	}

	if importUsed {
		return
	}

	// Remove import line from the import block
	var cleaned []string
	inImportBlock := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Detect start of import block
		if strings.HasPrefix(trimmed, "import (") {
			inImportBlock = true
			cleaned = append(cleaned, line)
			continue
		}

		// Detect end of import block
		if inImportBlock && trimmed == ")" {
			inImportBlock = false
			cleaned = append(cleaned, line)
			continue
		}

		// Skip repository import inside block
		if inImportBlock && trimmed == fmt.Sprintf(`"%s"`, importPath) {
			continue
		}

		// Handle single-line import (non-block)
		if strings.HasPrefix(trimmed, "import ") && strings.Contains(trimmed, importPath) {
			continue
		}

		cleaned = append(cleaned, line)
	}

	if err := os.WriteFile(path, []byte(strings.Join(cleaned, "\n")), 0644); err != nil {
		fmt.Println("❌ Failed to update import in:", path)
	} else {
		fmt.Println("🧽 Removed unused import from:", path)
	}
}

func readHistory(name string) ([]string, time.Time, error) {
	const historyFile = ".gen_history.json"

	history := map[string]struct {
		CreatedAt string   `json:"created_at"`
		Files     []string `json:"files"`
	}{}

	data, err := os.ReadFile(historyFile)
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("history not found")
	}
	if err := json.Unmarshal(data, &history); err != nil {
		return nil, time.Time{}, fmt.Errorf("invalid history")
	}

	entry, ok := history[strings.ToLower(name)]
	if !ok {
		return nil, time.Time{}, fmt.Errorf("no record for %s", name)
	}

	t, err := time.Parse(time.RFC3339, entry.CreatedAt)
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("invalid timestamp")
	}

	return entry.Files, t, nil
}

func removeFromHistory(name string) error {
	const historyFile = ".gen_history.json"
	history := map[string]interface{}{}

	data, err := os.ReadFile(historyFile)
	if err != nil {
		return fmt.Errorf("cannot read history file: %w", err)
	}
	if err := json.Unmarshal(data, &history); err != nil {
		return fmt.Errorf("cannot parse history file: %w", err)
	}

	delete(history, strings.ToLower(name))

	newData, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot marshal history: %w", err)
	}
	if err := os.WriteFile(historyFile, newData, 0644); err != nil {
		return fmt.Errorf("cannot write history file: %w", err)
	}
	return nil
}

func removeWhitespace(s string) string {
	return strings.Join(strings.Fields(s), "")
}
