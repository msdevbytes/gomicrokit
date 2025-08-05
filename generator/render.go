package generator

import (
	"embed"
	"os"
	"path/filepath"
	"sort"
	"text/template"
)

//go:embed templates/*
var tmplFS embed.FS

type TemplateData struct {
	ProjectName string
	ModulePath  string
	GoVersion   string
}

// RenderTemplates renders all embedded templates to the project folder
func RenderTemplates(projectPath string, data TemplateData) error {
	files := map[string]string{
		"cmd/main.go":                          "main.tmpl",
		"go.mod":                               "go.mod.tmpl",
		".gitignore":                           "gitignore.tmpl",
		"Dockerfile":                           "Dockerfile.tmpl",
		".env":                                 "env.tmpl",
		"internal/bootstrap/app.go":            "internal/bootstrap/app.tmpl",
		"internal/api/router.go":               "internal/api/router.tmpl",
		"internal/config/db_config.go":         "internal/config/db_config.tmpl",
		"internal/config/pagination_config.go": "internal/config/pagination_config.tmpl",
		"internal/handler/default.go":          "internal/handler/default.tmpl",
		"internal/handler/response.go":         "internal/handler/response.tmpl",
		"internal/db/db.go":                    "internal/db/db.tmpl",
		"internal/db/migrations.go":            "internal/db/migrations.tmpl",
		"internal/dto/pagination.go":           "internal/dto/pagination.tmpl",
		"internal/model/base.go":               "internal/model/base.tmpl",
		"internal/service/container.go":        "internal/service/container.tmpl",
		"pkg/logger/logger.go":                 "pkg/logger/logger.tmpl",
	}

	// Sort the output file names for deterministic rendering order
	var outFiles []string
	for out := range files {
		outFiles = append(outFiles, out)
	}
	sort.Strings(outFiles)

	for out, tmpl := range files {
		t, err := template.ParseFS(tmplFS, "templates/"+tmpl)
		if err != nil {
			return err
		}
		f, err := os.Create(filepath.Join(projectPath, out))
		if err != nil {
			return err
		}
		defer f.Close()

		if err := t.Execute(f, data); err != nil {
			return err
		}

	}

	return nil
}
