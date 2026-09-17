package ui

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type item struct {
	Name string
	Desc string
	URL  string
}

func (i item) Title() string       { return i.Name }
func (i item) Description() string { return i.Desc }
func (i item) FilterValue() string { return i.Name }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 2 }
func (d itemDelegate) Spacing() int                            { return 1 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	cursor := "  "
	checked := "[]"
	if index == m.Index() {
		cursor = "➤ "
		checked = "[*]"
	}

	fmt.Fprintf(w, "%s%s %s\n   %s\n", cursor, checked, i.Name, i.Desc)
}

var restFrameworkItems = []list.Item{
	item{"fiber", "use gofiber from: https://github.com/gofiber/fiber", "https://github.com/gofiber/fiber"},
	item{"chi", "use go-chi from: https://github.com/go-chi/chi", "https://github.com/go-chi/chi"},
}

var protocolItems = []list.Item{
	item{"rest", "HTTP/JSON microservice", ""},
	item{"grpc", "gRPC/protobuf microservice", ""},
}

var dbItems = []list.Item{
	item{"mysql", "MySQL compatible DB", ""},
	item{"postgres", "PostgreSQL database", ""},
	item{"sqlite", "SQLite embedded DB", ""},
	// item{"mongo", "MongoDB document store", ""},
	// item{"none", "Skip database setup", ""},
}

var featureItems = []item{
	{Name: "auth", Desc: "API key auth middleware/interceptor scaffold"},
	{Name: "cache", Desc: "Redis cache client scaffold"},
	{Name: "metrics", Desc: "Runtime/grpc metrics scaffold"},
	{Name: "tracing", Desc: "Request tracing scaffolding"},
}

type model struct {
	step          int
	projectInput  textinput.Model
	moduleInput   textinput.Model
	versionInput  textinput.Model
	outputInput   textinput.Model
	protocols     list.Model
	frameworks    list.Model
	dbs           list.Model
	yesNoInput    textinput.Model
	values        map[string]string
	done          bool
	confirm       textinput.Model
	featureCursor int
	featureSelect map[string]bool
}

func initialModel(seed map[string]string) model {
	ti := textinput.New()
	ti.Placeholder = "Enter project name"
	ti.Focus()
	ti.CharLimit = 32
	ti.Width = 40

	listFrameworks := list.New(restFrameworkItems, itemDelegate{}, 60, len(restFrameworkItems)*2)
	listFrameworks.Title = "Select REST framework"
	listFrameworks.SetShowFilter(false)
	listFrameworks.SetShowPagination(false)

	listProtocols := list.New(protocolItems, itemDelegate{}, 60, len(protocolItems)*2)
	listProtocols.Title = "Select project type"
	listProtocols.SetShowFilter(false)
	listProtocols.SetShowPagination(false)

	listDBs := list.New(dbItems, itemDelegate{}, 60, len(dbItems)*2)
	listDBs.Title = "Select a database"
	listDBs.SetShowFilter(false)
	listDBs.SetShowPagination(false)

	yesNo := textinput.New()
	yesNo.Placeholder = "y/n"
	yesNo.CharLimit = 1
	yesNo.Width = 4

	mi := textinput.New()
	mi.Placeholder = "Enter module name (e.g. github.com/user/project)"
	mi.Focus()
	mi.CharLimit = 64
	mi.Width = 191

	vi := textinput.New()
	vi.Placeholder = "latest"
	vi.CharLimit = 32
	vi.Width = 24

	oi := textinput.New()
	oi.Placeholder = "."
	oi.CharLimit = 128
	oi.Width = 40

	en := textinput.New()
	en.Placeholder = "Press y to confirm, n to cancel"
	en.CharLimit = 1
	en.Width = 4
	en.Focus()

	m := model{
		projectInput: ti,
		versionInput: vi,
		outputInput:  oi,
		protocols:    listProtocols,
		frameworks:   listFrameworks,
		moduleInput:  mi,
		dbs:          listDBs,
		yesNoInput:   yesNo,
		values:       make(map[string]string),
		confirm:      en,
		featureSelect: map[string]bool{
			"auth":    false,
			"cache":   false,
			"metrics": false,
			"tracing": false,
		},
	}

	m.values["protocol"] = "rest"
	m.values["framework"] = "fiber"
	m.values["db"] = "mysql"
	m.values["docker"] = "y"
	m.values["gorm"] = "y"
	m.values["rest_version"] = "latest"
	m.values["grpc_version"] = "latest"
	m.values["output_dir"] = "."
	m.values["save_config"] = "y"

	if seed != nil {
		if v := strings.TrimSpace(seed["project"]); v != "" {
			m.projectInput.SetValue(v)
			m.values["project"] = v
		}
		if v := strings.TrimSpace(seed["module"]); v != "" {
			m.moduleInput.SetValue(v)
			m.values["module"] = v
		}
		if v := strings.TrimSpace(seed["protocol"]); v == "grpc" || v == "rest" {
			m.values["protocol"] = v
			m.protocols.Select(indexOfListItem(protocolItems, v))
		}
		if v := strings.TrimSpace(seed["framework"]); v != "" {
			m.values["framework"] = v
			m.frameworks.Select(indexOfListItem(restFrameworkItems, v))
		}
		if v := strings.TrimSpace(seed["db"]); v != "" {
			m.values["db"] = v
			m.dbs.Select(indexOfListItem(dbItems, v))
		}
		if v := strings.TrimSpace(seed["docker"]); v == "y" || v == "n" {
			m.values["docker"] = v
		}
		if v := strings.TrimSpace(seed["gorm"]); v == "y" || v == "n" {
			m.values["gorm"] = v
		}
		if v := strings.TrimSpace(seed["rest_version"]); v != "" {
			m.values["rest_version"] = v
			m.versionInput.SetValue(v)
		}
		if v := strings.TrimSpace(seed["grpc_version"]); v != "" {
			m.values["grpc_version"] = v
			m.versionInput.SetValue(v)
		}
		if v := strings.TrimSpace(seed["output_dir"]); v != "" {
			m.values["output_dir"] = v
			m.outputInput.SetValue(v)
		}
		if v := strings.TrimSpace(seed["save_config"]); v == "y" || v == "n" {
			m.values["save_config"] = v
		}
		if v := strings.TrimSpace(seed["features"]); v != "" {
			m.values["features"] = v
			for _, f := range strings.Split(v, ",") {
				key := strings.TrimSpace(f)
				if _, ok := m.featureSelect[key]; ok {
					m.featureSelect[key] = true
				}
			}
		}
	}

	return m
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.done {
		return m, tea.Quit
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			fmt.Println("\n👋 Exiting. Goodbye!")
			return m, tea.Quit
		}
		// continue your logic...
	}

	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch m.step {

		case 0: // Project name input
			var cmd tea.Cmd
			m.projectInput, cmd = m.projectInput.Update(msg)
			if msg.Type == tea.KeyEnter && m.projectInput.Value() != "" {
				m.values["project"] = strings.TrimSpace(m.projectInput.Value())
				m.step++
			}
			return m, cmd
		case 1: // Module name input
			var cmd tea.Cmd
			m.moduleInput, cmd = m.moduleInput.Update(msg)
			if msg.Type == tea.KeyEnter && m.moduleInput.Value() != "" {
				m.values["module"] = strings.TrimSpace(m.moduleInput.Value())
				m.step++
			}
			return m, cmd

		case 2: // Protocol list
			var cmd tea.Cmd
			m.protocols, cmd = m.protocols.Update(msg)
			if msg.Type == tea.KeyEnter {
				selected := m.protocols.SelectedItem().(item)
				m.values["protocol"] = selected.Name
				m.step++
			}
			return m, cmd

		case 3: // Framework for REST or version for gRPC
			if m.values["protocol"] == "grpc" {
				var cmd tea.Cmd
				m.versionInput.Focus()
				m.versionInput, cmd = m.versionInput.Update(msg)
				if msg.Type == tea.KeyEnter {
					v := strings.TrimSpace(m.versionInput.Value())
					if v == "" {
						v = "latest"
					}
					m.values["framework"] = "grpc-go"
					m.values["grpc_version"] = v
					m.values["rest_version"] = "latest"
					m.values["db"] = "none"
					m.values["gorm"] = "n"
					m.step = 7
				}
				return m, cmd
			}

			var cmd tea.Cmd
			m.frameworks, cmd = m.frameworks.Update(msg)
			if msg.Type == tea.KeyEnter {
				selected := m.frameworks.SelectedItem().(item)
				m.values["framework"] = selected.Name
				m.step++
			}
			return m, cmd

		case 4: // DB list
			var cmd tea.Cmd
			m.dbs, cmd = m.dbs.Update(msg)
			if msg.Type == tea.KeyEnter {
				selected := m.dbs.SelectedItem().(item)
				m.values["db"] = selected.Name
				m.step++
			}
			return m, cmd

		case 5: // Version input for REST
			var cmd tea.Cmd
			m.versionInput.Focus()
			m.versionInput, cmd = m.versionInput.Update(msg)
			if msg.Type == tea.KeyEnter {
				v := strings.TrimSpace(m.versionInput.Value())
				if v == "" {
					v = "latest"
				}
				m.values["rest_version"] = v
				m.step++
			}
			return m, cmd

		case 6: // GORM? y/n
			var cmd tea.Cmd
			m.yesNoInput.Focus()
			m.yesNoInput, cmd = m.yesNoInput.Update(msg)
			if msg.Type == tea.KeyEnter {
				v := strings.ToLower(strings.TrimSpace(m.yesNoInput.Value()))
				if v == "y" || v == "n" {
					m.values["gorm"] = v

					// prepare for Docker input
					m.yesNoInput.Blur()
					m.yesNoInput.Reset()
					m.yesNoInput.Placeholder = "y/n"
					m.yesNoInput.Focus()
					m.step++
				}
			}
			return m, cmd

		case 7: // Docker? y/n
			var cmd tea.Cmd
			m.yesNoInput.Focus()
			m.yesNoInput, cmd = m.yesNoInput.Update(msg)
			if msg.Type == tea.KeyEnter {
				v := strings.ToLower(strings.TrimSpace(m.yesNoInput.Value()))
				if v == "y" || v == "n" {
					m.values["docker"] = v
					m.yesNoInput.Blur()
					m.yesNoInput.Reset()
					m.yesNoInput.Placeholder = "y/n"
					m.yesNoInput.Focus()
					m.step++
				}
			}
			return m, cmd
		case 8: // Feature multi-select
			switch msg.String() {
			case "up", "k":
				if m.featureCursor > 0 {
					m.featureCursor--
				}
			case "down", "j":
				if m.featureCursor < len(featureItems)-1 {
					m.featureCursor++
				}
			case " ":
				key := featureItems[m.featureCursor].Name
				m.featureSelect[key] = !m.featureSelect[key]
			}
			if msg.Type == tea.KeyEnter {
				selected := []string{}
				for _, f := range featureItems {
					if m.featureSelect[f.Name] {
						selected = append(selected, f.Name)
					}
				}
				m.values["features"] = strings.Join(selected, ",")
				m.step++
			}
			return m, nil
		case 9: // Output directory input
			var cmd tea.Cmd
			m.outputInput.Focus()
			m.outputInput, cmd = m.outputInput.Update(msg)
			if msg.Type == tea.KeyEnter {
				out := strings.TrimSpace(m.outputInput.Value())
				if out == "" {
					out = "."
				}
				m.values["output_dir"] = out
				m.step++
			}
			return m, cmd
		case 10: // Save config? y/n
			var cmd tea.Cmd
			m.yesNoInput.Focus()
			m.yesNoInput, cmd = m.yesNoInput.Update(msg)
			if msg.Type == tea.KeyEnter {
				v := strings.ToLower(strings.TrimSpace(m.yesNoInput.Value()))
				if v == "y" || v == "n" {
					m.values["save_config"] = v
					m.done = true
				}
			}
			return m, cmd
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.done && m.values != nil {
		return fmt.Sprintf(`
				🚀 Scaffolding project...

				📁 Project:   %s
				📦 Module:    %s
				🛰️  Type:      %s
				🔧 Framework: %s
				🗄️  Database:  %s
				📦 GORM:      %s
				🧩 REST deps: %s
				🧩 gRPC deps: %s
				🧩 Features:  %s
				📂 Output dir: %s
				📝 Save config: %s
				🐳 Docker:    %s
				⤵️ Press Enter to confirm generation...
			`,
			m.values["project"],
			m.values["module"],
			m.values["protocol"],
			m.values["framework"],
			m.values["db"],
			m.values["gorm"],
			valueOrDefault(m.values["rest_version"], "latest"),
			valueOrDefault(m.values["grpc_version"], "latest"),
			valueOrDefault(m.values["features"], "none"),
			valueOrDefault(m.values["output_dir"], "."),
			valueOrDefault(m.values["save_config"], "y"),
			m.values["docker"],
		)
	}

	switch m.step {
	case 0:
		return "📁 Enter Project Name:\n\n" + m.projectInput.View()
	case 1:
		return "📁 Enter Module Name:\n\n" + m.moduleInput.View()
	case 2:
		return m.protocols.View()
	case 3:
		if m.values["protocol"] == "grpc" {
			return "🧩 Enter gRPC dependency version (default: latest):\n\n" + m.versionInput.View()
		}
		return m.frameworks.View()
	case 4:
		if m.values["protocol"] == "grpc" {
			return ""
		}
		return m.dbs.View()
	case 5:
		return "🧩 Enter REST dependency version (default: latest):\n\n" + m.versionInput.View()
	case 6:
		return "📦 Use GORM? (y/n):\n\n" + m.yesNoInput.View()
	case 7:
		return "🐳 Include Docker (Dockerfile + Compose)? (y/n):\n\n" + m.yesNoInput.View()
	case 8:
		return featureSelectionView(m.featureCursor, m.featureSelect)
	case 9:
		return "📂 Enter output directory (default: .):\n\n" + m.outputInput.View()
	case 10:
		return "📝 Save .gmkrc.json? (y/n):\n\n" + m.yesNoInput.View()
	default:
		return "↩ Press Enter to confirm generation..."
	}
}

func valueOrDefault(v, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return v
}

func indexOfListItem(items []list.Item, name string) int {
	for i, it := range items {
		candidate, ok := it.(item)
		if !ok {
			continue
		}
		if candidate.Name == name {
			return i
		}
	}
	return 0
}

func featureSelectionView(cursor int, selected map[string]bool) string {
	var b strings.Builder
	b.WriteString("🧩 Select optional features (space to toggle, enter to continue):\n\n")
	for i, f := range featureItems {
		prefix := "  "
		if i == cursor {
			prefix = "➤ "
		}
		state := "[ ]"
		if selected[f.Name] {
			state = "[x]"
		}
		b.WriteString(fmt.Sprintf("%s%s %s\n    %s\n", prefix, state, f.Name, f.Desc))
	}
	return b.String()
}

func RunInteractive(seed map[string]string) map[string]string {
	p := tea.NewProgram(initialModel(seed))

	if finalModel, err := p.Run(); err == nil {
		if m, ok := finalModel.(model); ok {
			if m.done {
				if _, err := tea.NewProgram(newLoadingModel(m.values)).Run(); err != nil {
					fmt.Println("❌ Spinner failed:", err)
				}
			}
			return m.values
		}
	}

	fmt.Println("Error: failed to run interactive UI.")
	os.Exit(1)
	return nil
}
