package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/msdevbytes/gomicrokit/generator"
)

type removeModel struct {
	step    int
	input   textinput.Model
	confirm textinput.Model
	spinner spinner.Model
	loading bool
	err     error
	done    bool
	service string
}

func NewRemoveModel() removeModel {
	ti := textinput.New()
	ti.Placeholder = "Enter service name"
	ti.Focus()
	ti.CharLimit = 32
	ti.Width = 40

	confirm := textinput.New()
	confirm.Placeholder = "y/n"
	confirm.CharLimit = 1
	confirm.Width = 4

	sp := spinner.New()
	sp.Spinner = spinner.Dot

	return removeModel{
		input:   ti,
		confirm: confirm,
		spinner: sp,
	}
}

func (m removeModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m removeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.done {
		return m, tea.Quit
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.step {
		case 0: // Service name input
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			if msg.Type == tea.KeyEnter && m.input.Value() != "" {
				m.service = strings.TrimSpace(m.input.Value())
				m.step++
				m.confirm.Focus() // ✅ Focus confirm input
			}
			return m, cmd

		case 1: // Confirm deletion
			var cmd tea.Cmd
			m.confirm, cmd = m.confirm.Update(msg)

			if msg.Type == tea.KeyEnter {
				v := strings.ToLower(strings.TrimSpace(m.confirm.Value()))
				switch v {
				case "y":
					m.loading = true
					m.confirm.Reset()
					return m, tea.Batch(m.spinner.Tick, runRemoval(m.service))
				case "n":
					m.done = true
					return m, tea.Quit
				default:
					m.confirm.SetValue("")
					return m, cmd
				}
			}

			return m, cmd
		}

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case removalDoneMsg:
		m.loading = false
		m.err = msg.Err
		m.done = true
		return m, tea.Quit
	}

	return m, nil
}

func (m removeModel) View() string {
	if m.done {
		if m.err != nil {
			return fmt.Sprintf("❌ Failed to remove service: %v\n", m.err)
		}
		return fmt.Sprintf("🔥 Successfully removed service: %s\n", m.service)
	}

	switch m.step {
	case 0:
		return "🧼 Enter the service name to remove:\n\n" + m.input.View()
	case 1:
		return fmt.Sprintf("⚠️ Are you sure you want to delete '%s'? (y/n):\n\n%s", m.service, m.confirm.View())
	}

	if m.loading {
		return fmt.Sprintf("%s Removing service '%s'...", m.spinner.View(), m.service)
	}

	return ""
}

type removalDoneMsg struct{ Err error }

func runRemoval(name string) tea.Cmd {
	return func() tea.Msg {
		err := generator.RemoveService(name, true)
		return removalDoneMsg{Err: err}
	}
}

func RunRemoveFlow() {
	p := tea.NewProgram(NewRemoveModel())
	if _, err := p.Run(); err != nil {
		fmt.Println("❌ Remove UI failed:", err)
		os.Exit(1)
	}
}
