package ui

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/msdevbytes/gomicrokit/generator"
	"golang.org/x/term"
)

type featureOption struct {
	ID          string
	Title       string
	Description string
}

var featureCatalog = []featureOption{
	{ID: "auth", Title: "Auth", Description: "API key middleware/interceptor guards protected routes"},
	{ID: "config", Title: "Config", Description: "Typed app configuration loader with validation"},
	{ID: "discovery", Title: "Service Discovery", Description: "Consul registration and deregistration hooks"},
	{ID: "messaging", Title: "Messaging", Description: "NATS client bootstrap for event-driven workflows"},
	{ID: "metrics", Title: "Metrics", Description: "Basic HTTP/gRPC metrics handlers/interceptors"},
	{ID: "otel", Title: "OpenTelemetry", Description: "OTLP/stdout tracing provider initialization"},
	{ID: "resilience", Title: "Resilience", Description: "Circuit breaker + timeout policy scaffold"},
	{ID: "tracing", Title: "Tracing", Description: "Lightweight request trace context helpers"},
}

var stdinReader = bufio.NewReader(os.Stdin)

func RunInteractiveMenu(seed map[string]string) map[string]string {
	for {
		values := defaultValues(seed)
		reader := stdinReader

		values["project"] = promptProjectName(reader, values["project"], values["output_dir"])
		values["module"] = promptInput(reader, "Module path", valueOr(values["module"], values["project"]), true)
		values["port"] = promptInput(reader, "Port", valueOr(values["port"], "8000"), true)

		values["protocol"] = chooseOne("Select microservice type", []menuItem{
			{id: "rest", label: "REST"},
			{id: "grpc", label: "gRPC"},
		}, values["protocol"])

		cacheEnabled := hasCSVFeature(values["features"], "cache") || strings.EqualFold(values["cache"], "y")
		if values["protocol"] == "grpc" {
			values["framework"] = "grpc-go"
			values["db"] = chooseOne("Select database for gRPC", []menuItem{
				{id: "none", label: "None"},
				{id: "mysql", label: "MySQL"},
				{id: "postgres", label: "Postgres"},
				{id: "sqlite", label: "SQLite"},
			}, valueOr(values["db"], "none"))
			values["gorm"] = "n"
			values["grpc_version"] = promptInput(reader, "gRPC dependency version", valueOr(values["grpc_version"], "latest"), true)
			values["rest_version"] = "latest"
		} else {
			values["framework"] = chooseOne("Select REST framework", []menuItem{
				{id: "fiber", label: "Fiber"},
				{id: "chi", label: "Chi"},
			}, values["framework"])
			values["db"] = chooseOne("Select database", []menuItem{
				{id: "mysql", label: "MySQL"},
				{id: "postgres", label: "Postgres"},
				{id: "sqlite", label: "SQLite"},
			}, values["db"])
			values["rest_version"] = promptInput(reader, "REST dependency version", valueOr(values["rest_version"], "latest"), true)
			values["grpc_version"] = "latest"
			values["gorm"] = chooseOne("Use GORM?", []menuItem{
				{id: "y", label: "Yes"},
				{id: "n", label: "No"},
			}, values["gorm"])
		}

		values["cache"] = chooseOne("Include cache layer?", []menuItem{
			{id: "y", label: "Yes"},
			{id: "n", label: "No"},
		}, boolToYN(cacheEnabled))
		if values["cache"] == "y" {
			values["cache_store"] = chooseOne("Select cache store", []menuItem{
				{id: "redis", label: "Redis"},
				{id: "memory", label: "In-memory"},
			}, valueOr(values["cache_store"], "redis"))
		} else {
			values["cache_store"] = "none"
		}

		values["docker"] = chooseOne("Include Docker (Dockerfile + Compose)?", []menuItem{
			{id: "y", label: "Yes"},
			{id: "n", label: "No"},
		}, values["docker"])
		values["air"] = chooseOne("Include Air auto-reload config (.air.toml)?", []menuItem{
			{id: "y", label: "Yes"},
			{id: "n", label: "No"},
		}, values["air"])

		values["features"] = chooseFeatures(values["features"])
		values["features"] = setCSVFeature(values["features"], "cache", values["cache"] == "y")
		values["output_dir"] = promptInput(reader, "Output directory", valueOr(values["output_dir"], "."), true)
		if !projectNameAvailable(values["project"], values["output_dir"]) {
			values["project"] = promptProjectName(reader, "", values["output_dir"])
		}
		values["save_config"] = chooseOne("Save .gmkrc.json?", []menuItem{
			{id: "y", label: "Yes"},
			{id: "n", label: "No"},
		}, values["save_config"])

		fmt.Println("\nConfiguration summary:")
		fmt.Printf("  project      : %s\n", values["project"])
		fmt.Printf("  module       : %s\n", values["module"])
		fmt.Printf("  protocol     : %s\n", values["protocol"])
		fmt.Printf("  port         : %s\n", values["port"])
		fmt.Printf("  framework    : %s\n", values["framework"])
		fmt.Printf("  db           : %s\n", values["db"])
		fmt.Printf("  cache_layer  : %s\n", values["cache"])
		fmt.Printf("  cache_store  : %s\n", values["cache_store"])
		fmt.Printf("  rest_version : %s\n", values["rest_version"])
		fmt.Printf("  grpc_version : %s\n", values["grpc_version"])
		fmt.Printf("  docker       : %s\n", values["docker"])
		fmt.Printf("  air          : %s\n", values["air"])
		fmt.Printf("  features     : %s\n", featureSummary(values["features"]))
		fmt.Printf("  output_dir   : %s\n", values["output_dir"])
		fmt.Printf("  save_config  : %s\n\n", values["save_config"])

		next := chooseOne("Next action", []menuItem{
			{id: "generate", label: "Generate project"},
			{id: "restart", label: "Edit selections (restart wizard)"},
			{id: "cancel", label: "Cancel"},
		}, "generate")
		if next == "generate" {
			if !projectNameAvailable(values["project"], values["output_dir"]) {
				continue
			}
			return values
		}
		if next == "cancel" {
			fmt.Println("Cancelled.")
			os.Exit(0)
		}
	}
}

type menuItem struct {
	id    string
	label string
}

func chooseOne(question string, options []menuItem, defaultID string) string {
	if shouldUseBasicMenu() {
		return chooseOneBasic(question, options, defaultID)
	}
	labels := make([]string, 0, len(options))
	defaultLabel := ""
	idByLabel := make(map[string]string, len(options))
	for _, opt := range options {
		labels = append(labels, opt.label)
		idByLabel[opt.label] = opt.id
		if opt.id == defaultID {
			defaultLabel = opt.label
		}
	}
	selected := defaultLabel
	prompt := &survey.Select{
		Message: question + ":",
		Options: labels,
		Default: defaultLabel,
	}
	if err := survey.AskOne(prompt, &selected); err != nil {
		return chooseOneBasic(question, options, defaultID)
	}
	return idByLabel[selected]
}

func chooseFeatures(existingCSV string) string {
	selected := map[string]bool{}
	for _, raw := range strings.Split(existingCSV, ",") {
		v := strings.TrimSpace(raw)
		if v != "" {
			selected[v] = true
		}
	}

	if shouldUseBasicMenu() {
		return chooseFeaturesBasic(selected)
	}

	fmt.Println("\nOptional features (use arrow keys + space to toggle, enter to confirm):")
	options := make([]string, 0, len(featureCatalog))
	optionToID := make(map[string]string, len(featureCatalog))
	defaults := make([]string, 0, len(featureCatalog))
	for _, feature := range featureCatalog {
		label := formatFeatureLabel(feature)
		options = append(options, label)
		optionToID[label] = feature.ID
		if selected[feature.ID] {
			defaults = append(defaults, label)
		}
	}

	var picked []string
	prompt := &survey.MultiSelect{
		Message:  "Select optional features:",
		Options:  options,
		Default:  defaults,
		PageSize: 12,
	}
	if err := survey.AskOne(prompt, &picked); err != nil {
		return chooseFeaturesBasic(selected)
	}
	ids := make([]string, 0, len(picked))
	for _, label := range picked {
		if id, ok := optionToID[label]; ok {
			ids = append(ids, id)
		}
	}
	return strings.Join(ids, ",")
}

func shouldUseBasicMenu() bool {
	if strings.EqualFold(os.Getenv("GMK_FORCE_BASIC_MENU"), "1") {
		return true
	}
	return !term.IsTerminal(int(os.Stdin.Fd()))
}

func chooseOneBasic(question string, options []menuItem, defaultID string) string {
	reader := stdinReader
	defaultIndex := 1
	for i, opt := range options {
		if opt.id == defaultID {
			defaultIndex = i + 1
			break
		}
	}
	for {
		fmt.Printf("\n%s:\n", question)
		for i, opt := range options {
			label := opt.label
			if i+1 == defaultIndex {
				label = label + " (default)"
			}
			fmt.Printf("  %d) %s\n", i+1, label)
		}
		fmt.Printf("Select option [%d]: ", defaultIndex)
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading input: %v\n", err)
			os.Exit(1)
		}
		line = strings.TrimSpace(line)
		if line == "" {
			return options[defaultIndex-1].id
		}
		n, err := strconv.Atoi(line)
		if err != nil || n < 1 || n > len(options) {
			fmt.Println("Invalid selection, try again.")
			continue
		}
		return options[n-1].id
	}
}

func chooseFeaturesBasic(selected map[string]bool) string {
	reader := stdinReader
	for {
		fmt.Println("\nToggle features:")
		for i, feature := range featureCatalog {
			state := "[ ]"
			if selected[feature.ID] {
				state = "[x]"
			}
			fmt.Printf("  %d) %s %-16s %s\n", i+1, state, feature.Title, feature.Description)
		}
		fmt.Println("  d) Done")
		fmt.Println("  c) Clear all")
		fmt.Print("Select option: ")
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading input: %v\n", err)
			os.Exit(1)
		}
		choice := strings.ToLower(strings.TrimSpace(line))
		switch choice {
		case "d":
			out := make([]string, 0, len(featureCatalog))
			for _, feature := range featureCatalog {
				if selected[feature.ID] {
					out = append(out, feature.ID)
				}
			}
			return strings.Join(out, ",")
		case "c":
			for _, feature := range featureCatalog {
				delete(selected, feature.ID)
			}
		default:
			n, err := strconv.Atoi(choice)
			if err != nil || n < 1 || n > len(featureCatalog) {
				fmt.Println("Invalid selection, try again.")
				continue
			}
			key := featureCatalog[n-1].ID
			selected[key] = !selected[key]
		}
	}
}

func formatFeatureLabel(feature featureOption) string {
	return fmt.Sprintf("%-16s %s", feature.Title, feature.Description)
}

func featureSummary(csv string) string {
	if strings.TrimSpace(csv) == "" {
		return "none"
	}
	parts := strings.Split(csv, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		id := strings.TrimSpace(p)
		if id == "" {
			continue
		}
		label := id
		for _, feature := range featureCatalog {
			if feature.ID == id {
				label = feature.Title
				break
			}
		}
		out = append(out, label)
	}
	if len(out) == 0 {
		return "none"
	}
	return strings.Join(out, ", ")
}

func promptInput(reader *bufio.Reader, label, defaultValue string, required bool) string {
	for {
		if defaultValue != "" {
			fmt.Printf("%s [%s]: ", label, defaultValue)
		} else {
			fmt.Printf("%s: ", label)
		}
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading input: %v\n", err)
			os.Exit(1)
		}
		value := strings.TrimSpace(line)
		if value == "" {
			value = defaultValue
		}
		if required && strings.TrimSpace(value) == "" {
			fmt.Println("Value is required.")
			continue
		}
		return value
	}
}

func promptProjectName(reader *bufio.Reader, current, outputDir string) string {
	for {
		name := promptInput(reader, "Project name", current, true)
		if err := generator.ValidateProjectName(name); err != nil {
			fmt.Printf("❌ %v\n", err)
			current = ""
			continue
		}
		if !projectNameAvailable(name, outputDir) {
			current = ""
			continue
		}
		return name
	}
}

func projectNameAvailable(name, outputDir string) bool {
	exists, path, err := generator.ProjectTargetExists(outputDir, name)
	if err != nil {
		fmt.Printf("❌ %v\n", err)
		return false
	}
	if exists {
		fmt.Printf("❌ directory '%s' already exists. Choose a different name or delete the existing directory\n", path)
		return false
	}
	return true
}

func defaultValues(seed map[string]string) map[string]string {
	values := map[string]string{
		"project":      "",
		"module":       "",
		"protocol":     "rest",
		"port":         "8000",
		"framework":    "fiber",
		"db":           "mysql",
		"gorm":         "y",
		"docker":       "y",
		"air":          "n",
		"rest_version": "latest",
		"grpc_version": "latest",
		"features":     "",
		"cache_store":  "redis",
		"output_dir":   ".",
		"save_config":  "y",
	}
	for k, v := range seed {
		if strings.TrimSpace(v) != "" {
			values[k] = strings.TrimSpace(v)
		}
	}
	return values
}

func valueOr(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func boolToYN(v bool) string {
	if v {
		return "y"
	}
	return "n"
}

func hasCSVFeature(csv string, target string) bool {
	for _, raw := range strings.Split(csv, ",") {
		if strings.EqualFold(strings.TrimSpace(raw), target) {
			return true
		}
	}
	return false
}

func setCSVFeature(csv, feature string, enabled bool) string {
	seen := map[string]bool{}
	out := make([]string, 0)
	for _, raw := range strings.Split(csv, ",") {
		part := strings.ToLower(strings.TrimSpace(raw))
		if part == "" || part == strings.ToLower(feature) || seen[part] {
			continue
		}
		seen[part] = true
		out = append(out, part)
	}
	if enabled {
		out = append(out, strings.ToLower(feature))
	}
	return strings.Join(out, ",")
}
