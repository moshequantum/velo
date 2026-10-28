package steps

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/moshequantum/velo/internal/tui"
	"github.com/moshequantum/velo/internal/wizard"
)

type BaaSStep struct {
	state     *wizard.State
	projectID textinput.Model
	anonKey   textinput.Model
	focusIdx  int
}

func NewBaaSStep(state *wizard.State) *BaaSStep {
	p := textinput.New()
	p.Placeholder = "https://tu-proyecto.insforge.app (o Enter para omitir)"
	p.Focus()
	p.Width = 45
	if state.InsForgeProjectID != "" {
		p.SetValue(state.InsForgeProjectID)
	}

	k := textinput.New()
	k.Placeholder = "anon-key / public-token (opcional en init)"
	k.Width = 45
	if state.InsForgeAnonKey != "" {
		k.SetValue(state.InsForgeAnonKey)
	}

	return &BaaSStep{
		state:     state,
		projectID: p,
		anonKey:   k,
		focusIdx:  0,
	}
}

func (s *BaaSStep) Init() tea.Cmd {
	return textinput.Blink
}

func (s *BaaSStep) Update(msg tea.Msg) (tui.Step, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "down":
			s.focusIdx = (s.focusIdx + 1) % 2
			s.updateFocus()
			return s, nil
		case "shift+tab", "up":
			s.focusIdx = (s.focusIdx + 1) % 2
			s.updateFocus()
			return s, nil
		case "enter":
			s.state.InsForgeProjectID = strings.TrimSpace(s.projectID.Value())
			s.state.InsForgeAnonKey = strings.TrimSpace(s.anonKey.Value())
			return s, tui.Next()
		case "esc":
			return s, tui.Back()
		case "ctrl+c":
			return s, tea.Quit
		}
	}

	var cmd tea.Cmd
	if s.focusIdx == 0 {
		s.projectID, cmd = s.projectID.Update(msg)
	} else {
		s.anonKey, cmd = s.anonKey.Update(msg)
	}
	return s, cmd
}

func (s *BaaSStep) updateFocus() {
	if s.focusIdx == 0 {
		s.projectID.Focus()
		s.anonKey.Blur()
	} else {
		s.projectID.Blur()
		s.anonKey.Focus()
	}
}

func (s *BaaSStep) View() string {
	title := tui.StyleTitle.Render("3. Base de Datos & BaaS (InsForge)")
	desc := tui.StyleSubtitle.Render("Configura el acceso a PostgreSQL, Storage y Auth (se guardará en .env):")

	input1 := fmt.Sprintf("URL de InsForge:\n%s\n", s.projectID.View())
	input2 := fmt.Sprintf("Anon / Public Key:\n%s\n", s.anonKey.View())

	note := tui.StyleSubtitle.Render("Tip: Puedes dejarlo vacío y configurarlo después en el archivo .env.")
	help := tui.StyleSubtitle.Render("\n[ Tab ] Cambiar campo   [ Enter ] Siguiente   [ Esc ] Volver")

	content := fmt.Sprintf("%s\n%s\n\n%s\n%s\n%s\n%s", title, desc, input1, input2, note, help)
	return tui.StyleBox.Render(content)
}
