package generator

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/iancoleman/strcase"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

//go:embed templates/internal/service/*.tmpl
var serviceTemplates embed.FS

type ServiceOptions struct {
	Name       string
	ModulePath string
	Force      bool
	DryRun     bool
}

func GenerateService(opt ServiceOptions) error {
	if !isValidGoIdent(opt.Name) {
		return fmt.Errorf("invalid service name: %s", opt.Name)
	}

	if opt.ModulePath == "" {
		opt.ModulePath = GetGoModule()
	}

	// Derive naming
	service := convertToTitleCaseNoSpaces(opt.Name)
	receiver := convertToTitleCaseNoSpaces(opt.Name)
	repo := strings.ToLower(service[:1]) + service[1:]
	snake := strcase.ToSnake(opt.Name)

	data := map[string]string{
		"Service":    service,
		"Receiver":   receiver,
		"Repository": repo,
		"Module":     cases.Lower(language.Und).String(opt.ModulePath),
	}

	files := map[string]string{
		fmt.Sprintf("internal/model/%s_model.go", snake):           "model.tmpl",
		fmt.Sprintf("internal/repository/%s_repository.go", snake): "repository.tmpl",
		fmt.Sprintf("internal/service/%s_service.go", snake):       "service.tmpl",
		fmt.Sprintf("internal/handler/%s_handler.go", snake):       "handler.tmpl",
		fmt.Sprintf("internal/dto/%s_dto.go", snake):               "dto.tmpl",
		fmt.Sprintf("test/unit/dto/%s_input_test.go", snake):       "dto_test.tmpl",
	}

	generated := []string{}

	for path, tmplName := range files {
		content, err := renderTemplate(tmplName, data)
		if err != nil {
			return err
		}

		if opt.DryRun {
			fmt.Println("🔍 Preview:", path)
			fmt.Println(content)
			fmt.Println(strings.Repeat("-", 60))
			continue
		}

		if _, err := os.Stat(path); err == nil && !opt.Force {
			continue
		}

		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}

		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return err
		}

		generated = append(generated, path)
	}

	if !opt.DryRun {
		fileName := strcase.ToSnake(opt.Name)
		writeHistory(opt.Name, generated)
		updateContainer(service)
		updateRoutes(service)

		generatedFiles := []string{
			fmt.Sprintf("internal/model/%s_model.go", fileName),
			fmt.Sprintf("internal/repository/%s_repository.go", fileName),
			fmt.Sprintf("internal/service/%s_service.go", fileName),
			fmt.Sprintf("internal/handler/%s_handler.go", fileName),
			fmt.Sprintf("internal/dto/%s_dto.go", fileName),
			fmt.Sprintf("test/unit/dto/%s_input_test.go", fileName),
		}
		writeHistory(opt.Name, generatedFiles)
		fmt.Println("📝 History updated in .gen_history.json")
		fmt.Printf("✅ Service '%s' generated and registered in container.\n", fileName)
	}

	return nil
}

func renderTemplate(name string, data map[string]string) (string, error) {
	t, err := template.ParseFS(serviceTemplates, "templates/internal/service/"+name)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	err = t.Execute(&b, data)
	return b.String(), err
}
