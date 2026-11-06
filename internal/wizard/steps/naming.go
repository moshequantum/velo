package steps

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/moshequantum/velo/internal/tui"
	"github.com/moshequantum/velo/internal/wizard"
)

type NamingStep struct {
	state     *wizard.State
	textInput textinput.Model
}

func NewNamingStep(state *wizard.State) *NamingStep {
	ti := textinput.New()
	if state.IsSpanish() {
		ti.Placeholder = "mi-proyecto-edge"
	} else {
		ti.Placeholder = "my-edge-project"
	}
	ti.Focus()
	ti.CharLimit = 50
	ti.Width = 35
	if state.ProjectName != "" {
		ti.SetValue(state.ProjectName)
	}

	return &NamingStep{
		state:     state,
		textInput: ti,
	}
}

func (s *NamingStep) Init() tea.Cmd {
	return textinput.Blink
}

func (s *NamingStep) defaultName() string {
	if s.state.IsSpanish() {
		return "mi-proyecto-edge"
	}
	return "my-edge-project"
}

func (s *NamingStep) Update(msg tea.Msg) (tui.Step, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			val := strings.TrimSpace(s.textInput.Value())
			if val == "" {
				val = s.defaultName()
			}
			s.state.ProjectName = val
			s.state.ProjectSlug = wizard.Slugify(val)
			return s, tui.Next()
		case "esc":
			return s, tui.Back()
		case "ctrl+c":
			return s, tea.Quit
		}
	}

	s.textInput, cmd = s.textInput.Update(msg)
	return s, cmd
}

func (s *NamingStep) View() string {
	isEs := s.state.IsSpanish()

	var title, desc, slugLabel, dirLabel, helpText string
	if isEs {
		title = tui.StyleTitle.Render("1. Identidad del Proyecto & Contratos")
		desc = tui.StyleSubtitle.Render("Nombre del proyecto y espacio de nombres del contrato OpenPassport:")
		slugLabel = "Slug canónico"
		dirLabel = "Directorio"
		helpText = "[ Enter ] Continuar   [ Esc ] Volver"
	} else {
		title = tui.StyleTitle.Render("1. Project Identity & Contracts")
		desc = tui.StyleSubtitle.Render("Project name and OpenPassport contract namespace:")
		slugLabel = "Canonical slug"
		dirLabel = "Directory"
		helpText = "[ Enter ] Continue   [ Esc ] Back"
	}

	slugPreview := wizard.Slugify(s.textInput.Value())
	if slugPreview == "" {
		slugPreview = wizard.Slugify(s.defaultName())
	}

	preview := fmt.Sprintf("\n%s: %s\n%s:     ./%s",
		slugLabel,
		tui.StyleSelected.Render(slugPreview),
		dirLabel,
		tui.StyleSelected.Render(slugPreview),
	)

	help := tui.StyleSubtitle.Render("\n" + helpText)

	content := fmt.Sprintf("%s\n%s\n\n%s\n%s\n%s", title, desc, s.textInput.View(), preview, help)
	return tui.StyleBox.Render(content)
}
