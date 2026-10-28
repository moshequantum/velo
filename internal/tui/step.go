package tui

import tea "github.com/charmbracelet/bubbletea"

// Step defines the contract for any interactive wizard view.
type Step interface {
	Init() tea.Cmd
	Update(tea.Msg) (Step, tea.Cmd)
	View() string
}

// NextMsg requests navigation to the next wizard step.
type NextMsg struct{}

// BackMsg requests navigation to the previous wizard step.
type BackMsg struct{}

// Next emits a NextMsg command.
func Next() tea.Cmd {
	return func() tea.Msg {
		return NextMsg{}
	}
}

// Back emits a BackMsg command.
func Back() tea.Cmd {
	return func() tea.Msg {
		return BackMsg{}
	}
}
