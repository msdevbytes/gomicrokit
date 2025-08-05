package generator

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

func RemoveService(name string, force bool) error {
	flagName := name
	flagForce := force
	if flagName == "" {
		return fmt.Errorf("❌ Usage: go run tools/remove_service.go -name <name>")
	}
	if flagName == "" {
		return fmt.Errorf("❌ Please provide a service name with --name")
	}

	name = strings.ToLower(flagName)
	service := strings.ToUpper(string(name[0])) + name[1:]

	files, createdAt, err := readHistory(name)
	if err != nil {
		fmt.Println("⛔", err)
		os.Exit(1)
	}

	// Enforce freshness (1 minutes max)
	if time.Since(createdAt) > 1*time.Minute && !flagForce {
		fmt.Println("⛔Force", !flagForce)
		fmt.Printf("⛔ '%s' was created more than 5 minutes ago (%s). Use --force to override.\n", name, createdAt.Format(time.RFC822))
		os.Exit(1)
	}

	deleteFiles(files)

	cleanFromFile("internal/service/container.go", []string{
		fmt.Sprintf("%s *%sService", service, service),
		fmt.Sprintf("%s: New%sService(repository.New%sRepository(db))", service, service, service),
		fmt.Sprintf("New%sService", service),
	})

	fmt.Println("🔥 Removed:", fmt.Sprintf("%s/internal/repository", getGoModule()))
	cleanUnusedImport("internal/service/container.go", fmt.Sprintf("%s/internal/repository", getGoModule()))

	cleanFromFile("internal/api/router.go", []string{
		fmt.Sprintf("New%sHandler(svc.%s).Register(api.Group(\"/%ss\"))", service, service, name),
	})

	removeFromHistory(name)
	return nil
}

func getGoModule() string {
	data, err := os.ReadFile("go.mod")
	must(err)

	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module "))
		}
	}
	return "your-module-name"
}

func must(err error) {
	if err != nil {
		fmt.Println("❌", err)
		os.Exit(1)
	}
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

func removeFromHistory(name string) {
	const historyFile = ".gen_history.json"
	history := map[string]interface{}{}

	data, _ := os.ReadFile(historyFile)
	json.Unmarshal(data, &history)

	delete(history, strings.ToLower(name))

	newData, _ := json.MarshalIndent(history, "", "  ")
	_ = os.WriteFile(historyFile, newData, 0644)
}

func removeWhitespace(s string) string {
	return strings.Join(strings.Fields(s), "")
}
