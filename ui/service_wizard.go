package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/msdevbytes/gomicrokit/generator"
)

type ServiceInput struct {
	Name       string
	ModulePath string
	Force      bool
	DryRun     bool
}

type wizardModel struct {
	step      int
	nameInput textinput.Model
	module    string
	force     bool
	dryRun    bool
	done      bool
}

func initialServiceModel() wizardModel {
	ti := textinput.New()
	ti.Placeholder = "e.g. event"
	ti.Focus()
	ti.CharLimit = 32
	ti.Width = 30

	return wizardModel{
		nameInput: ti,
		module:    generator.GetGoModule(),
	}
}

func (m wizardModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m wizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.done {
		return m, tea.Quit
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.step {
		case 0: // Service name
			var cmd tea.Cmd
			m.nameInput, cmd = m.nameInput.Update(msg)
			if msg.Type == tea.KeyEnter && strings.TrimSpace(m.nameInput.Value()) != "" {
				m.step++
			}
			return m, cmd

		case 1: // Overwrite?
			if msg.String() == "y" || msg.String() == "Y" {
				m.force = true
				m.step++
			} else if msg.String() == "n" || msg.String() == "N" {
				m.force = false
				m.step++
			}

		case 2: // Dry-run?
			if msg.String() == "y" || msg.String() == "Y" {
				m.dryRun = true
				m.step++
			} else if msg.String() == "n" || msg.String() == "N" {
				m.dryRun = false
				m.step++
			}

		case 3: // Done
			m.done = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m wizardModel) View() string {
	if m.done {
		return "✅ Generating service...\n"
	}

	switch m.step {
	case 0:
		return "📛 Enter service name:\n\n" + m.nameInput.View() + "\n\n(Press Enter)"
	case 1:
		return "❓ Overwrite existing files? (y/n):"
	case 2:
		return "👁️  Use dry-run (preview only)? (y/n):"
	default:
		return ""
	}
}

func RunServiceWizard() ServiceInput {
	p := tea.NewProgram(initialServiceModel())
	finalModel, err := p.Run()
	if err != nil {
		fmt.Println("❌ Error running wizard:", err)
		os.Exit(1)
	}

	m := finalModel.(wizardModel)

	return ServiceInput{
		Name:       strings.TrimSpace(m.nameInput.Value()),
		ModulePath: strings.ToLower(generator.GetGoModule()),
		Force:      m.force,
		DryRun:     m.dryRun,
	}
}
