package steps

import (
	"fmt"
	"regexp"
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
	ti.Placeholder = "mi-proyecto-edge"
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

func (s *NamingStep) Update(msg tea.Msg) (tui.Step, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			val := strings.TrimSpace(s.textInput.Value())
			if val == "" {
				val = "mi-proyecto-edge"
			}
			s.state.ProjectName = val
			s.state.ProjectSlug = slugify(val)
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
	title := tui.StyleTitle.Render("1. Nombre del Proyecto")
	desc := tui.StyleSubtitle.Render("¿Cómo se llamará el proyecto o cliente?")

	slugPreview := slugify(s.textInput.Value())
	if slugPreview == "" {
		slugPreview = "mi-proyecto-edge"
	}

	preview := fmt.Sprintf("\nSlug canónico: %s\nDirectorio:     ./%s",
		tui.StyleSelected.Render(slugPreview),
		tui.StyleSelected.Render(slugPreview),
	)

	help := tui.StyleSubtitle.Render("\n[ Enter ] Continuar   [ Esc ] Volver")

	content := fmt.Sprintf("%s\n%s\n\n%s\n%s\n%s", title, desc, s.textInput.View(), preview, help)
	return tui.StyleBox.Render(content)
}

func slugify(s string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9\-_]+`)
	slug := strings.ToLower(re.ReplaceAllString(s, "-"))
	return strings.Trim(slug, "-")
}
