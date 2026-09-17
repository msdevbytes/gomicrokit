package generator

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/iancoleman/strcase"
)

//go:embed templates/internal/service/*.tmpl templates/grpc/service/*.tmpl
var serviceTemplates embed.FS

type ServiceOptions struct {
	Name        string
	ModulePath  string
	Force       bool
	DryRun      bool
	ProjectType ProjectType
	Framework   string
}

func GenerateService(opt ServiceOptions) error {
	if !IsValidGoIdent(opt.Name) {
		return fmt.Errorf("invalid service name: %s", opt.Name)
	}

	if opt.ModulePath == "" {
		opt.ModulePath = GetGoModule()
	}
	if opt.ProjectType == "" {
		opt.ProjectType = DetectProjectType(".")
	}
	if opt.Framework == "" {
		cfg := DetectProjectConfig(".")
		opt.Framework = cfg.Framework
	}
	if opt.Framework == "" && opt.ProjectType == ProjectTypeREST {
		opt.Framework = string(FrameworkFiber)
	}

	// Derive naming
	service := ConvertToTitleCaseNoSpaces(opt.Name)
	receiver := ConvertToTitleCaseNoSpaces(opt.Name)
	repo := strings.ToLower(service[:1]) + service[1:]
	snake := strcase.ToSnake(opt.Name)

	data := map[string]string{
		"Service":      service,
		"Receiver":     receiver,
		"Repository":   repo,
		"Module":       opt.ModulePath,
		"ProtoPackage": snake,
	}

	files := map[string]string{
		fmt.Sprintf("internal/model/%s_model.go", snake):           "model.tmpl",
		fmt.Sprintf("internal/repository/%s_repository.go", snake): "repository.tmpl",
		fmt.Sprintf("internal/service/%s_service.go", snake):       "service.tmpl",
		fmt.Sprintf("internal/handler/%s_handler.go", snake):       "handler.tmpl",
		fmt.Sprintf("internal/dto/%s_dto.go", snake):               "dto.tmpl",
		fmt.Sprintf("test/unit/dto/%s_input_test.go", snake):       "dto_test.tmpl",
	}
	templatePrefix := "templates/internal/service/"
	if opt.ProjectType == ProjectTypeREST && opt.Framework == string(FrameworkChi) {
		files[fmt.Sprintf("internal/handler/%s_handler.go", snake)] = "handler_chi.tmpl"
	}
	if opt.ProjectType == ProjectTypeGRPC {
		files = map[string]string{
			fmt.Sprintf("api/proto/%s.proto", snake):                "proto.tmpl",
			fmt.Sprintf("internal/service/%s_service.go", snake):    "service.tmpl",
			fmt.Sprintf("internal/server/grpc/%s_server.go", snake): "server.tmpl",
		}
		templatePrefix = "templates/grpc/service/"
	}

	generated := []string{}
	skipped := []string{}

	for path, tmplName := range files {
		content, err := renderTemplate(templatePrefix, tmplName, data)
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
			skipped = append(skipped, path)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return fmt.Errorf("cannot create directory for %s: %w", path, err)
		}

		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("cannot write file %s: %w", path, err)
		}

		generated = append(generated, path)
	}

	// Notify user about skipped files
	if len(skipped) > 0 && !opt.DryRun {
		fmt.Printf("⏭️ Skipped %d existing file(s) (use --force to overwrite):\n", len(skipped))
		for _, f := range skipped {
			fmt.Printf("   - %s\n", f)
		}
	}

	if !opt.DryRun {
		fileName := strcase.ToSnake(opt.Name)
		if opt.ProjectType == ProjectTypeGRPC {
			if err := updateGRPCContainer(service); err != nil {
				return fmt.Errorf("failed to update grpc container: %w", err)
			}
			if err := updateGRPCServer(service); err != nil {
				return fmt.Errorf("failed to update grpc server registration: %w", err)
			}
		} else {
			if err := updateContainer(service); err != nil {
				return fmt.Errorf("failed to update container: %w", err)
			}
			var err error
			if opt.Framework == string(FrameworkChi) {
				err = updateRoutesChi(service)
			} else {
				err = updateRoutes(service)
			}
			if err != nil {
				return fmt.Errorf("failed to update routes: %w", err)
			}
		}
		if err := WriteHistory(opt.Name, generated); err != nil {
			fmt.Printf("⚠️ Warning: failed to write history: %v\n", err)
		} else {
			fmt.Println("📝 History updated in .gen_history.json")
		}
		fmt.Printf("✅ Service '%s' generated and registered in container.\n", fileName)
	}

	return nil
}

func renderTemplate(prefix, name string, data map[string]string) (string, error) {
	t, err := template.ParseFS(serviceTemplates, prefix+name)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	err = t.Execute(&b, data)
	return b.String(), err
}
