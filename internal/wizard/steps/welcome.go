package steps

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/moshequantum/velo/internal/tui"
	"github.com/moshequantum/velo/internal/wizard"
)

type WelcomeStep struct {
	state *wizard.State
}

func NewWelcomeStep(state *wizard.State) *WelcomeStep {
	return &WelcomeStep{state: state}
}

func (s *WelcomeStep) Init() tea.Cmd {
	return nil
}

func (s *WelcomeStep) Update(msg tea.Msg) (tui.Step, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter", " ":
			return s, tui.Next()
		case "ctrl+c", "q":
			return s, tea.Quit
		}
	}
	return s, nil
}

func (s *WelcomeStep) View() string {
	title := tui.StyleTitle.Render("⚡ VELO FRAMEWORK — Edge & Headless Harness")
	badge := tui.StyleBadge.Render("0 to 1 • $0 Infra • IA-First")

	desc := `Velo automatiza las decisiones arquitectónicas para que lances sistemas completos
en el Edge sin pagar un solo dólar en servidores.

  • Frontend & API:  Astro + Cloudflare Pages (SSR & Static)
  • BaaS:            InsForge (PostgreSQL, Realtime, Storage, Auth)
  • Cockpit:         Telegram Bot (HITL / Aprobaciones críticas)
  • Contexto:        Compatible con OpenPassport v1.0`

	footer := tui.StyleSubtitle.Render("\nPresiona [ Enter ] para comenzar o [ q ] para salir...")

	content := fmt.Sprintf("%s  %s\n\n%s\n%s", title, badge, desc, footer)
	return tui.StyleBox.Render(content)
}
