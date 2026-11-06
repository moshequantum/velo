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
	if state.Language == "" {
		state.Language = "en"
	}
	return &WelcomeStep{state: state}
}

func (s *WelcomeStep) Init() tea.Cmd {
	return nil
}

func (s *WelcomeStep) Update(msg tea.Msg) (tui.Step, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "shift+tab", "left", "right", "up", "down":
			if s.state.IsSpanish() {
				s.setLanguage("en")
			} else {
				s.setLanguage("es")
			}
			return s, nil

		case "1", "e", "E":
			s.setLanguage("en")
			return s, nil

		case "2", "s", "S":
			s.setLanguage("es")
			return s, nil

		case "enter", " ":
			return s, tui.Next()

		case "ctrl+c", "q":
			return s, tea.Quit
		}
	}
	return s, nil
}

func (s *WelcomeStep) setLanguage(lang string) {
	prev := s.state.Language
	s.state.Language = lang

	// Adapt default names if user hasn't changed them yet
	if prev != lang {
		if lang == "es" && s.state.ProjectName == "my-velo-app" {
			s.state.ProjectName = "mi-app-velo"
			s.state.ProjectSlug = "mi-app-velo"
		} else if lang == "en" && s.state.ProjectName == "mi-app-velo" {
			s.state.ProjectName = "my-velo-app"
			s.state.ProjectSlug = "my-velo-app"
		}
	}
}

func (s *WelcomeStep) View() string {
	isEs := s.state.IsSpanish()

	var title, badge, desc, selectorLabel, hint, footer string

	if isEs {
		title = tui.StyleTitle.Render("⚡ VELO FRAMEWORK — Arnés Edge & Headless")
		badge = tui.StyleBadge.Render("0 a 1 • $0 Infra • IA-First")
		desc = `Velo automatiza las decisiones arquitectónicas para que lances sistemas completos
en el Edge sin pagar un solo dólar en servidores.

  • Frontend & API:  Astro / Next.js / React+Vite en Cloudflare
  • BaaS / DB:       SurrealDB (Multi-modelo) / InsForge
  • Cockpit:         Telegram Bot (HITL / Aprobaciones críticas)
  • Contexto:        Compatible con OpenPassport v1.0`

		selectorLabel = "Idioma / Language:"
		hint = "(Usa [ Tab ], [ ←/→ ] o [ 1/2 ] para alternar idioma)"
		footer = tui.StyleSubtitle.Render("\nPresiona [ Enter ] para comenzar o [ q ] para salir...")
	} else {
		title = tui.StyleTitle.Render("⚡ VELO FRAMEWORK — Edge & Headless Harness")
		badge = tui.StyleBadge.Render("0 to 1 • $0 Infra • AI-First")
		desc = `Velo automates architectural decisions to help you launch production systems
on the Edge without paying a single dollar in servers.

  • Frontend & API:  Astro / Next.js / React+Vite on Cloudflare
  • BaaS / DB:       SurrealDB (Multi-model) / InsForge
  • Cockpit:         Telegram Bot (HITL / Critical human approvals)
  • Context:         OpenPassport v1.0 Compliant`

		selectorLabel = "Language / Idioma:"
		hint = "(Use [ Tab ], [ ←/→ ] or [ 1/2 ] to switch language)"
		footer = tui.StyleSubtitle.Render("\nPress [ Enter ] to continue or [ q ] to exit...")
	}

	enRadio := "[ ] English"
	esRadio := "[ ] Español"
	if isEs {
		esRadio = tui.StyleSelected.Render("[•] Español")
		enRadio = tui.StyleSubtitle.Render("[ ] English")
	} else {
		enRadio = tui.StyleSelected.Render("[•] English")
		esRadio = tui.StyleSubtitle.Render("[ ] Español")
	}

	langSelector := fmt.Sprintf("%s\n  %s    %s\n  %s",
		tui.StyleNormal.Render(selectorLabel),
		enRadio,
		esRadio,
		tui.StyleSubtitle.Render(hint),
	)

	content := fmt.Sprintf("%s  %s\n\n%s\n\n%s\n%s", title, badge, desc, langSelector, footer)
	return tui.StyleBox.Render(content)
}
