package steps

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/moshequantum/velo/internal/tui"
	"github.com/moshequantum/velo/internal/wizard"
)

type CockpitStep struct {
	state     *wizard.State
	botHandle textinput.Model
	ownerID   textinput.Model
	focusIdx  int
}

func NewCockpitStep(state *wizard.State) *CockpitStep {
	b := textinput.New()
	b.Placeholder = "@MiProyecto_bot (opcional)"
	b.Focus()
	b.Width = 35
	if state.TelegramBotHandle != "" {
		b.SetValue(state.TelegramBotHandle)
	}

	o := textinput.New()
	o.Placeholder = "Tu ID numérico de Telegram (chat_id)"
	o.Width = 35
	if state.TelegramOwnerID != "" {
		o.SetValue(state.TelegramOwnerID)
	}

	return &CockpitStep{
		state:     state,
		botHandle: b,
		ownerID:   o,
		focusIdx:  0,
	}
}

func (s *CockpitStep) Init() tea.Cmd {
	return textinput.Blink
}

func (s *CockpitStep) Update(msg tea.Msg) (tui.Step, tea.Cmd) {
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
			s.state.TelegramBotHandle = strings.TrimSpace(s.botHandle.Value())
			s.state.TelegramOwnerID = strings.TrimSpace(s.ownerID.Value())
			return s, tui.Next()
		case "esc":
			return s, tui.Back()
		case "ctrl+c":
			return s, tea.Quit
		}
	}

	var cmd tea.Cmd
	if s.focusIdx == 0 {
		s.botHandle, cmd = s.botHandle.Update(msg)
	} else {
		s.ownerID, cmd = s.ownerID.Update(msg)
	}
	return s, cmd
}

func (s *CockpitStep) updateFocus() {
	if s.focusIdx == 0 {
		s.botHandle.Focus()
		s.ownerID.Blur()
	} else {
		s.botHandle.Blur()
		s.ownerID.Focus()
	}
}

func (s *CockpitStep) View() string {
	title := tui.StyleTitle.Render("4. Cockpit de Operaciones (Telegram HITL)")
	desc := tui.StyleSubtitle.Render("Configura el bot para notificaciones y aprobaciones humanas antes de mutar la DB:")

	input1 := fmt.Sprintf("Handle del Bot:\n%s\n", s.botHandle.View())
	input2 := fmt.Sprintf("Telegram Owner ID (chat_id):\n%s\n", s.ownerID.View())

	note := tui.StyleSubtitle.Render("Nota: El endpoint /api/telegram se generará automáticamente en Astro.")
	help := tui.StyleSubtitle.Render("\n[ Tab ] Cambiar campo   [ Enter ] Siguiente   [ Esc ] Volver")

	content := fmt.Sprintf("%s\n%s\n\n%s\n%s\n%s\n%s", title, desc, input1, input2, note, help)
	return tui.StyleBox.Render(content)
}
