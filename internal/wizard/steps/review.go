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
	title := tui.StyleTitle.Render("5. Revisión del Blueprint Velo")
	desc := tui.StyleSubtitle.Render("Verifica la configuración antes de generar el arnés:")

	bot := s.state.TelegramBotHandle
	if bot == "" {
		bot = "(Sin configurar)"
	}
	insforge := s.state.InsForgeProjectID
	if insforge == "" {
		insforge = "(Configurar en .env)"
	}

	summary := fmt.Sprintf(`  • Proyecto:     %s (%s)
  • Frontend:     %s
  • Runtime:      %s
  • Base de datos: PostgreSQL vía %s
  • Cockpit:      Telegram HITL %s
  • Infra Cost:   $0.00 / mes`,
		tui.StyleSelected.Render(s.state.ProjectName),
		s.state.ProjectSlug,
		tui.StyleSelected.Render(s.state.Frontend),
		tui.StyleSelected.Render(s.state.Runtime),
		tui.StyleSelected.Render(insforge),
		tui.StyleSelected.Render(bot),
	)

	help := tui.StyleSubtitle.Render("\n[ Enter ] Generar Proyecto   [ Esc ] Modificar   [ q ] Cancelar")

	content := fmt.Sprintf("%s\n%s\n\n%s\n\n%s", title, desc, summary, help)
	return tui.StyleBox.Render(content)
}
