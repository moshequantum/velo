package steps

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/moshequantum/velo/internal/scaffold"
	"github.com/moshequantum/velo/internal/tui"
	"github.com/moshequantum/velo/internal/wizard"
)

type InstallStep struct {
	state   *wizard.State
	spinner spinner.Model
	err     error
	done    bool
}

type installDoneMsg struct {
	err error
}

func NewInstallStep(state *wizard.State) *InstallStep {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = tui.StyleTitle

	return &InstallStep{
		state:   state,
		spinner: s,
	}
}

func (s *InstallStep) Init() tea.Cmd {
	return tea.Batch(
		s.spinner.Tick,
		func() tea.Msg {
			err := scaffold.Generate(s.state)
			return installDoneMsg{err: err}
		},
	)
}

func (s *InstallStep) Update(msg tea.Msg) (tui.Step, tea.Cmd) {
	switch msg := msg.(type) {
	case installDoneMsg:
		s.done = true
		s.err = msg.err
		return s, nil

	case tea.KeyMsg:
		if s.done {
			switch msg.String() {
			case "enter", "q", "ctrl+c":
				return s, tea.Quit
			}
		}
	}

	var cmd tea.Cmd
	s.spinner, cmd = s.spinner.Update(msg)
	return s, cmd
}

func (s *InstallStep) View() string {
	isEs := s.state.IsSpanish()

	if !s.done {
		var genMsg string
		if isEs {
			genMsg = fmt.Sprintf("%s Generando arnés Velo para '%s'...", s.spinner.View(), s.state.ProjectName)
		} else {
			genMsg = fmt.Sprintf("%s Generating Velo harness for '%s'...", s.spinner.View(), s.state.ProjectName)
		}
		title := tui.StyleTitle.Render(genMsg)
		return tui.StyleBox.Render(title)
	}

	if s.err != nil {
		var errTitle, errFooter string
		if isEs {
			errTitle = tui.StyleTitle.Render("❌ Error generando el proyecto")
			errFooter = "Presiona [ Enter ] para salir."
		} else {
			errTitle = tui.StyleTitle.Render("❌ Error generating project")
			errFooter = "Press [ Enter ] to exit."
		}
		errMsg := fmt.Sprintf("Error detail / Detalle:\n%v\n\n%s", s.err, errFooter)
		return tui.StyleBox.Render(fmt.Sprintf("%s\n\n%s", errTitle, errMsg))
	}

	dir := fmt.Sprintf("./%s", s.state.ProjectSlug)

	// Determine cockpit endpoint path
	cockpitFile := fmt.Sprintf("%s/src/pages/api/telegram.ts", dir)
	if s.state.Frontend == "next" {
		cockpitFile = fmt.Sprintf("%s/src/app/api/telegram/route.ts", dir)
	} else if s.state.Frontend == "react-vite" {
		cockpitFile = fmt.Sprintf("%s/functions/api/telegram.ts", dir)
	}

	var title, instructions, footer string
	if isEs {
		title = tui.StyleSuccess.Render("🚀 ¡Arnés Velo Generado con Éxito!")
		instructions = fmt.Sprintf(`El proyecto fue aprovisionado en: %s

Siguientes pasos para despegar:
  1. cd %s
  2. pnpm install
  3. pnpm dev              (Abre tu servidor local)
  4. wrangler pages deploy (Despliega directo a Cloudflare $0)

Cockpit Telegram configurado en:
  → %s`,
			tui.StyleSelected.Render(dir),
			s.state.ProjectSlug,
			cockpitFile,
		)
		footer = tui.StyleSubtitle.Render("\nPresiona [ Enter ] para finalizar...")
	} else {
		title = tui.StyleSuccess.Render("🚀 Velo Harness Generated Successfully!")
		instructions = fmt.Sprintf(`The project was provisioned at: %s

Next steps to launch:
  1. cd %s
  2. pnpm install
  3. pnpm dev              (Opens your local development server)
  4. wrangler pages deploy (Deploy directly to Cloudflare $0)

Telegram Cockpit configured at:
  → %s`,
			tui.StyleSelected.Render(dir),
			s.state.ProjectSlug,
			cockpitFile,
		)
		footer = tui.StyleSubtitle.Render("\nPress [ Enter ] to exit...")
	}

	content := fmt.Sprintf("%s\n\n%s\n%s", title, instructions, footer)
	return tui.StyleBox.Render(content)
}
