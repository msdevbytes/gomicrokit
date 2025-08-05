package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/msdevbytes/gomicrokit/generator"
)

type generationDoneMsg struct{ Err error }

type loadingModel struct {
	spinner spinner.Model
	values  map[string]string
	err     error
	ready   bool // ← Add this
}

func runGeneration(values map[string]string) tea.Cmd {
	return func() tea.Msg {
		err := generator.GenerateProjectFrom(values)
		if err != nil {
			fmt.Println("❌ Error during project generation:", err)
		}
		return generationDoneMsg{Err: err}
	}
}

func newLoadingModel(values map[string]string) loadingModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	return loadingModel{
		spinner: s,
		values:  values,
	}
}

func (m loadingModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, runGeneration(m.values))
}

func (m loadingModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case generationDoneMsg:
		m.err = msg.Err
		m.ready = true           // <- Mark generation as done
		return m, m.spinner.Tick // <- Keep ticking spinner once more

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)

		if m.ready {
			return m, tea.Quit // <- Quit only after spinner ticked
		}
		return m, cmd
	}

	return m, nil
}

func (m loadingModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("❌ Generation failed:\n\n%v", m.err)
	}
	return fmt.Sprintf("%s Generating project...", m.spinner.View())
}
