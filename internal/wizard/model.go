package wizard

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/moshequantum/velo/internal/tui"
)

type Model struct {
	state   *State
	steps   []tui.Step
	current int
}

func NewModel() Model {
	state := NewDefaultState()
	return NewModelWithState(state)
}

func NewModelWithState(state *State) Model {
	return Model{
		state:   state,
		current: 0,
	}
}

func (m *Model) SetSteps(steps []tui.Step) {
	m.steps = steps
}

func (m Model) Init() tea.Cmd {
	if len(m.steps) > 0 {
		return m.steps[0].Init()
	}
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case tui.NextMsg:
		if m.current < len(m.steps)-1 {
			m.current++
			return m, m.steps[m.current].Init()
		}
		return m, tea.Quit

	case tui.BackMsg:
		if m.current > 0 {
			m.current--
			return m, m.steps[m.current].Init()
		}
		return m, nil
	}

	if len(m.steps) == 0 {
		return m, tea.Quit
	}

	updatedStep, cmd := m.steps[m.current].Update(msg)
	m.steps[m.current] = updatedStep
	return m, cmd
}

func (m Model) View() string {
	if len(m.steps) == 0 {
		return ""
	}
	return m.steps[m.current].View()
}

// Run executes the full wizard program.
func Run(steps []tui.Step) error {
	m := NewModel()
	m.SetSteps(steps)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
