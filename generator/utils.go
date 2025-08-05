package generator

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func GetGoModule() string {
	data, err := os.ReadFile("go.mod")
	if err != nil {
		return "unknown-module"
	}
	for line := range strings.SplitSeq(string(data), "\n") {
		if after, ok := strings.CutPrefix(line, "module "); ok {
			return cases.Lower(language.Und).String(strings.TrimSpace(after))
		}
	}
	return "unknown-module"
}

func isValidGoIdent(s string) bool {
	if s == "" || !isLetter(rune(s[0])) {
		return false
	}
	for _, r := range s {
		if !isLetter(r) && !isDigit(r) {
			return false
		}
	}
	return true
}

func isLetter(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func writeHistory(name string, files []string) {
	const historyFile = ".gen_history.json"

	history := map[string]struct {
		CreatedAt string   `json:"created_at"`
		Files     []string `json:"files"`
	}{}

	data, _ := os.ReadFile(historyFile)
	_ = json.Unmarshal(data, &history)

	history[strings.ToLower(name)] = struct {
		CreatedAt string   `json:"created_at"`
		Files     []string `json:"files"`
	}{
		CreatedAt: time.Now().Format(time.RFC3339),
		Files:     files,
	}

	historyData, _ := json.MarshalIndent(history, "", "  ")
	_ = os.WriteFile(historyFile, historyData, 0644)
}

func updateRoutes(serviceName string) {
	// module := getGoModule()
	routesFile := "internal/api/router.go"

	handlerLine := fmt.Sprintf("\thandler.New%sHandler(svc.%s).Register(api.Group(\"/%ss\"))", serviceName, serviceName, strings.ToLower(serviceName))

	// Check if already registered
	data, err := os.ReadFile(routesFile)
	must(err)
	if strings.Contains(string(data), handlerLine) {
		fmt.Println("📍 Route already exists in index.go")
		return
	}

	lines := []string{}
	file, err := os.Open(routesFile)
	must(err)
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		// Insert route just after api := app.Group(...)
		if strings.Contains(scanner.Text(), "api := app.Group") {
			lines = append(lines, "")
			lines = append(lines, handlerLine)
		}
	}
	must(scanner.Err())

	must(os.WriteFile(routesFile, []byte(strings.Join(lines, "\n")), 0644))
	fmt.Println("📍 Updated: internal/api/router.go")
}

func updateContainer(serviceName string) {
	path := "internal/service/container.go"
	lines := []string{}

	file, err := os.Open(path)
	if err != nil {
		fmt.Println("❌ container.go not found")
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	insertedField, insertedAssign, insertedImport := false, false, false
	importBlockStarted := false
	importLines := []string{}
	alreadyImportedRepository := false

	module := GetGoModule()

	for scanner.Scan() {
		line := scanner.Text()

		// Check for existing import
		if strings.HasPrefix(strings.TrimSpace(line), `"`+module+`/internal/repository"`) {
			alreadyImportedRepository = true
		}

		// Handle import block start
		if strings.HasPrefix(strings.TrimSpace(line), "import (") {
			importBlockStarted = true
		}

		// Handle import block end
		if importBlockStarted && strings.HasPrefix(strings.TrimSpace(line), ")") && !alreadyImportedRepository {
			importLines = append(importLines, `  "`+module+`/internal/repository"`)
			insertedImport = true
		}

		// Collect import lines separately
		if importBlockStarted {
			importLines = append(importLines, line)
			if strings.HasPrefix(strings.TrimSpace(line), ")") {
				importBlockStarted = false
				lines = append(lines, importLines...)
				continue
			}
			continue
		}

		// Insert service field into Container struct
		if strings.Contains(line, "type Container struct {") && !insertedField {
			lines = append(lines, line)
			lines = append(lines, fmt.Sprintf("\t%s *%sService", serviceName, serviceName))
			insertedField = true
			continue
		}

		// Insert service initialization inside NewContainer
		if strings.Contains(line, "return &Container{") && !insertedAssign {
			lines = append(lines, line)
			lines = append(lines, fmt.Sprintf("\t\t%s: New%sService(repository.New%sRepository(db)),", serviceName, serviceName, serviceName))
			insertedAssign = true
			continue
		}

		lines = append(lines, line)
	}

	if !alreadyImportedRepository && !insertedImport {

		// Add import outside block if block not found
		for i, l := range lines {
			if strings.HasPrefix(strings.TrimSpace(l), "import") {
				lines = append(lines[:i+1], append([]string{` "` + module + `/internal/repository"`}, lines[i+1:]...)...)
				break
			}
		}
	}

	output := strings.Join(lines, "\n")
	must(os.WriteFile(path, []byte(output), 0644))
	fmt.Println("📦 Updated: internal/service/container.go")
}
