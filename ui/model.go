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

var frameworkItems = []list.Item{
	// item{"chi", "use go-chi from: https://github.com/go-chi/chi", "https://github.com/go-chi/chi"},
	// item{"gin", "use gin-gonic from: https://github.com/gin-gonic/gin", "https://github.com/gin-gonic/gin"},
	item{"fiber", "use gofiber from: https://github.com/gofiber/fiber", "https://github.com/gofiber/fiber"},
	// item{"gorilla/mux", "use gorilla/mux from: https://github.com/gorilla/mux", "https://github.com/gorilla/mux"},
	// item{"httprouter", "use julienschmidt/httprouter from: https://github.com/julienschmidt/httprouter", "https://github.com/julienschmidt/httprouter"},
	// item{"echo", "use echo from: https://github.com/labstack/echo", "https://github.com/labstack/echo"},
}

var dbItems = []list.Item{
	item{"mysql", "MySQL compatible DB", ""},
	// item{"postgres", "PostgreSQL database", ""},
	// item{"sqlite", "SQLite embedded DB", ""},
	// item{"mongo", "MongoDB document store", ""},
	// item{"none", "Skip database setup", ""},
}

type model struct {
	step         int
	projectInput textinput.Model
	moduleInput  textinput.Model
	frameworks   list.Model
	dbs          list.Model
	yesNoInput   textinput.Model
	values       map[string]string
	done         bool
	confirm      textinput.Model
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Enter project name"
	ti.Focus()
	ti.CharLimit = 32
	ti.Width = 40

	listFrameworks := list.New(frameworkItems, itemDelegate{}, 60, len(frameworkItems)*2)
	listFrameworks.Title = "Select a framework"
	listFrameworks.SetShowFilter(false)
	listFrameworks.SetShowPagination(false)

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

	en := textinput.New()
	en.Placeholder = "Press y to confirm, n to cancel"
	en.CharLimit = 1
	en.Width = 4
	en.Focus()

	return model{
		projectInput: ti,
		frameworks:   listFrameworks,
		moduleInput:  mi,
		dbs:          listDBs,
		yesNoInput:   yesNo,
		values:       make(map[string]string),
		confirm:      en,
	}
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

		case 2: // Framework list
			var cmd tea.Cmd
			m.frameworks, cmd = m.frameworks.Update(msg)
			if msg.Type == tea.KeyEnter {
				selected := m.frameworks.SelectedItem().(item)
				m.values["framework"] = selected.Name
				m.step++
			}
			return m, cmd

		case 3: // DB list
			var cmd tea.Cmd
			m.dbs, cmd = m.dbs.Update(msg)
			if msg.Type == tea.KeyEnter {
				selected := m.dbs.SelectedItem().(item)
				m.values["db"] = selected.Name
				m.step++
			}
			return m, cmd

		case 4: // GORM? y/n
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

		case 5: // Docker? y/n
			var cmd tea.Cmd
			m.yesNoInput.Focus()
			m.yesNoInput, cmd = m.yesNoInput.Update(msg)
			if msg.Type == tea.KeyEnter {
				v := strings.ToLower(strings.TrimSpace(m.yesNoInput.Value()))
				if v == "y" || v == "n" {
					m.values["docker"] = v
					m.done = true
				}
			}
			return m, cmd
		case 6: // Final confirmation
			if msg.Type == tea.KeyEnter {
				fmt.Println("🚀 Generating project with the following settings:")
				return m, runGeneration(m.values) // trigger spinner + generation
			}

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
				🔧 Framework: %s
				🗄️  Database:  %s
				📦 GORM:      %s
				🐳 Docker:    %s
				⤵️ Press Enter to confirm generation...
			`,
			m.values["project"],
			m.values["module"],
			m.values["framework"],
			m.values["db"],
			m.values["gorm"],
			m.values["docker"],
		)
	}

	switch m.step {
	case 0:
		return "📁 Enter Project Name:\n\n" + m.projectInput.View()
	case 1:
		return "📁 Enter Module Name:\n\n" + m.moduleInput.View()
	case 2:
		return m.frameworks.View()
	case 3:
		return m.dbs.View()
	case 4:
		return "📦 Use GORM? (y/n):\n\n" + m.yesNoInput.View()
	case 5:
		return "🐳 Include Dockerfile? (y/n):\n\n" + m.yesNoInput.View()
	default:
		return "↩ Press Enter to confirm generation..."
	}
}

func RunInteractive() map[string]string {
	p := tea.NewProgram(initialModel())

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
