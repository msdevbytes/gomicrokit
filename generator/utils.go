package generator

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
	"unicode"
)

func GetGoModule() string {
	data, err := os.ReadFile("go.mod")
	if err != nil {
		return "unknown-module"
	}
	for line := range strings.SplitSeq(string(data), "\n") {
		if after, ok := strings.CutPrefix(line, "module "); ok {
			return strings.TrimSpace(after)
		}
	}
	return "unknown-module"
}

// IsValidGoIdent checks if s is a valid Go identifier.
// Go identifiers must start with a letter or underscore and contain only
// letters, digits, and underscores.
func IsValidGoIdent(s string) bool {
	if s == "" {
		return false
	}
	for i, r := range s {
		if i == 0 {
			// First character must be a letter or underscore
			if !unicode.IsLetter(r) && r != '_' {
				return false
			}
		} else {
			// Subsequent characters can be letters, digits, or underscores
			if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
				return false
			}
		}
	}
	return true
}

func WriteHistory(name string, files []string) error {
	const historyFile = ".gen_history.json"

	history := map[string]struct {
		CreatedAt string   `json:"created_at"`
		Files     []string `json:"files"`
	}{}

	// Read existing history (ignore error - file may not exist yet)
	data, _ := os.ReadFile(historyFile)
	_ = json.Unmarshal(data, &history)

	history[strings.ToLower(name)] = struct {
		CreatedAt string   `json:"created_at"`
		Files     []string `json:"files"`
	}{
		CreatedAt: time.Now().Format(time.RFC3339),
		Files:     files,
	}

	historyData, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot marshal history: %w", err)
	}
	if err := os.WriteFile(historyFile, historyData, 0644); err != nil {
		return fmt.Errorf("cannot write history file: %w", err)
	}
	return nil
}

func updateRoutes(serviceName string) error {
	routesFile := "internal/api/router.go"

	handlerLine := fmt.Sprintf("\thandler.New%sHandler(svc.%s).Register(api.Group(\"/%ss\"))", serviceName, serviceName, strings.ToLower(serviceName))

	// Check if already registered
	data, err := os.ReadFile(routesFile)
	if err != nil {
		return fmt.Errorf("cannot read routes file: %w", err)
	}
	if strings.Contains(string(data), handlerLine) {
		fmt.Println("📍 Route already exists in router.go")
		return nil
	}

	lines := []string{}
	file, err := os.Open(routesFile)
	if err != nil {
		return fmt.Errorf("cannot open routes file: %w", err)
	}
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
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error scanning routes file: %w", err)
	}

	if err := os.WriteFile(routesFile, []byte(strings.Join(lines, "\n")), 0644); err != nil {
		return fmt.Errorf("cannot write routes file: %w", err)
	}
	fmt.Println("📍 Updated: internal/api/router.go")
	return nil
}

func updateRoutesChi(serviceName string) error {
	routesFile := "internal/api/router.go"
	openLine := fmt.Sprintf("\t\tapi.Route(\"/%ss\", func(r chi.Router) {", strings.ToLower(serviceName))
	registerLine := fmt.Sprintf("\t\t\thandler.New%sHandler(svc.%s).Register(r)", serviceName, serviceName)
	closeLine := "\t\t})"

	data, err := os.ReadFile(routesFile)
	if err != nil {
		return fmt.Errorf("cannot read routes file: %w", err)
	}
	if strings.Contains(string(data), registerLine) {
		fmt.Println("📍 Route already exists in chi router")
		return nil
	}

	lines := []string{}
	inserted := false
	for _, line := range strings.Split(string(data), "\n") {
		lines = append(lines, line)
		if !inserted && strings.Contains(line, "// gmk:register-services") {
			lines = append(lines, openLine)
			lines = append(lines, registerLine)
			lines = append(lines, closeLine)
			inserted = true
		}
	}
	if !inserted {
		return fmt.Errorf("could not find insertion point in chi router")
	}

	if err := os.WriteFile(routesFile, []byte(strings.Join(lines, "\n")), 0644); err != nil {
		return fmt.Errorf("cannot write chi routes file: %w", err)
	}
	fmt.Println("📍 Updated chi router: internal/api/router.go")
	return nil
}

func ConvertToTitleCaseNoSpaces(s string) string {
	if s == "" {
		return ""
	}
	// Remove all spaces from the string
	s = strings.ReplaceAll(s, " ", "")

	var result strings.Builder
	result.Grow(len(s)) // Pre-allocate memory for efficiency

	// Capitalize the first character of the entire string
	result.WriteRune(unicode.ToUpper(rune(s[0])))

	for i := 1; i < len(s); i++ {
		r := rune(s[i])
		prevR := rune(s[i-1])

		// If the current character is an uppercase letter and the previous is lowercase,
		// it indicates a new "word" in a camelCase/PascalCase string.
		if unicode.IsUpper(r) && unicode.IsLower(prevR) {
			result.WriteRune(r)
		} else {
			// Otherwise, add the character as is (lowercase for subsequent letters of a "word")
			result.WriteRune(unicode.ToLower(r))
		}
	}
	return result.String()
}

func updateContainer(serviceName string) error {
	path := "internal/service/container.go"
	lines := []string{}

	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("container.go not found: %w", err)
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

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error scanning container file: %w", err)
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
	if err := os.WriteFile(path, []byte(output), 0644); err != nil {
		return fmt.Errorf("cannot write container file: %w", err)
	}
	fmt.Println("📦 Updated: internal/service/container.go")
	return nil
}

func updateGRPCContainer(serviceName string) error {
	path := "internal/service/container.go"
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("cannot read grpc container file: %w", err)
	}

	fieldLine := fmt.Sprintf("\t%s *%sService", serviceName, serviceName)
	assignLine := fmt.Sprintf("\t\t%s: New%sService(),", serviceName, serviceName)
	content := string(data)

	if strings.Contains(content, fieldLine) && strings.Contains(content, assignLine) {
		fmt.Println("📦 gRPC container already up to date")
		return nil
	}

	lines := []string{}
	insertedField := false
	insertedAssign := false
	for _, line := range strings.Split(content, "\n") {
		lines = append(lines, line)
		if !insertedField && strings.Contains(line, "type Container struct {") {
			lines = append(lines, fieldLine)
			insertedField = true
		}
		if !insertedAssign && strings.Contains(line, "return &Container{") {
			lines = append(lines, assignLine)
			insertedAssign = true
		}
	}

	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644); err != nil {
		return fmt.Errorf("cannot write grpc container file: %w", err)
	}
	fmt.Println("📦 Updated gRPC container: internal/service/container.go")
	return nil
}

func updateGRPCServer(serviceName string) error {
	path := "internal/server/grpc/server.go"
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("cannot read grpc server file: %w", err)
	}

	registerLine := fmt.Sprintf("\t_ = New%sServer(service.New%sService())", serviceName, serviceName)
	content := string(data)
	if strings.Contains(content, registerLine) {
		fmt.Println("🛰️ gRPC service already registered")
		return nil
	}

	importLine := "\t\"{{MODULE}}/internal/service\""
	module := GetGoModule()
	importLine = strings.ReplaceAll(importLine, "{{MODULE}}", module)

	lines := []string{}
	insertedImport := false
	insertedRegister := false
	inImportBlock := false
	alreadyImported := strings.Contains(content, fmt.Sprintf("\"%s/internal/service\"", module))

	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "import (") {
			inImportBlock = true
		}
		if inImportBlock && trimmed == ")" && !alreadyImported {
			lines = append(lines, importLine)
			insertedImport = true
		}
		if strings.Contains(line, "// gmk:register-services") && !insertedRegister {
			lines = append(lines, registerLine)
			insertedRegister = true
		}
		lines = append(lines, line)
		if inImportBlock && trimmed == ")" {
			inImportBlock = false
		}
	}

	if !alreadyImported && !insertedImport {
		return fmt.Errorf("grpc server import block missing for service import insertion")
	}
	if !insertedRegister {
		return fmt.Errorf("grpc server registration marker not found")
	}

	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644); err != nil {
		return fmt.Errorf("cannot write grpc server file: %w", err)
	}
	fmt.Println("🛰️ Updated gRPC server registrations")
	return nil
}
