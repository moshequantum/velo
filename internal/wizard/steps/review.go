package steps

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/moshequantum/velo/internal/tui"
	"github.com/moshequantum/velo/internal/wizard"
)

type ReviewStep struct {
	state *wizard.State
}

func NewReviewStep(state *wizard.State) *ReviewStep {
	return &ReviewStep{state: state}
}

func (s *ReviewStep) Init() tea.Cmd {
	return nil
}

func (s *ReviewStep) Update(msg tea.Msg) (tui.Step, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			return s, tui.Next()
		case "esc":
			return s, tui.Back()
		case "ctrl+c", "q":
			return s, tea.Quit
		}
	}
	return s, nil
}

func (s *ReviewStep) View() string {
	isEs := s.state.IsSpanish()

	bot := s.state.TelegramBotHandle
	if bot == "" {
		if isEs {
			bot = "(Sin bot)"
		} else {
			bot = "(No bot)"
		}
	}

	dataDesc := s.state.DataProvider
	if s.state.DataURL != "" {
		dataDesc += fmt.Sprintf(" (%s)", s.state.DataURL)
	}

	var title, desc, helpText, summary string

	if isEs {
		title = tui.StyleTitle.Render("5. Revisión del Blueprint Velo")
		desc = tui.StyleSubtitle.Render("Verifica la arquitectura completa antes de generar el arnés:")
		summary = fmt.Sprintf(`  • 1. Contratos:    %s (%s) · OpenPassport v1.0
  • 2. Base Datos:   %s
  • 3. Endpoints:    Telegram HITL %s (/api/telegram)
  • 4. Frontend:     %s en %s
  • Idioma:          %s
  • Modo:            Free-tier first ($0)`,
			tui.StyleSelected.Render(s.state.ProjectName),
			s.state.ProjectSlug,
			tui.StyleSelected.Render(dataDesc),
			tui.StyleSelected.Render(bot),
			tui.StyleSelected.Render(s.state.Frontend),
			tui.StyleSelected.Render(s.state.Runtime),
			tui.StyleSelected.Render("Español"),
		)
		helpText = "\n[ Enter ] Generar Proyecto   [ Esc ] Modificar   [ q ] Cancelar"
	} else {
		title = tui.StyleTitle.Render("5. Velo Blueprint Review")
		desc = tui.StyleSubtitle.Render("Verify complete architecture before generating the harness:")
		summary = fmt.Sprintf(`  • 1. Contracts:    %s (%s) · OpenPassport v1.0
  • 2. Database:     %s
  • 3. Endpoints:    Telegram HITL %s (/api/telegram)
  • 4. Frontend:     %s on %s
  • Language:        %s
  • Mode:            Free-tier first ($0)`,
			tui.StyleSelected.Render(s.state.ProjectName),
			s.state.ProjectSlug,
			tui.StyleSelected.Render(dataDesc),
			tui.StyleSelected.Render(bot),
			tui.StyleSelected.Render(s.state.Frontend),
			tui.StyleSelected.Render(s.state.Runtime),
			tui.StyleSelected.Render("English"),
		)
		helpText = "\n[ Enter ] Generate Project   [ Esc ] Modify   [ q ] Cancel"
	}

	help := tui.StyleSubtitle.Render(helpText)
	content := fmt.Sprintf("%s\n%s\n\n%s\n\n%s", title, desc, summary, help)
	return tui.StyleBox.Render(content)
}
