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
	if !s.done {
		title := tui.StyleTitle.Render(fmt.Sprintf("%s Generando arnés Velo para '%s'...", s.spinner.View(), s.state.ProjectName))
		return tui.StyleBox.Render(title)
	}

	if s.err != nil {
		title := tui.StyleTitle.Render("❌ Error generando el proyecto")
		errMsg := fmt.Sprintf("Detalle del error:\n%v\n\nPresiona [ Enter ] para salir.", s.err)
		return tui.StyleBox.Render(fmt.Sprintf("%s\n\n%s", title, errMsg))
	}

	title := tui.StyleSuccess.Render("🚀 ¡Arnés Velo Generado con Éxito!")
	dir := fmt.Sprintf("./%s", s.state.ProjectSlug)

	instructions := fmt.Sprintf(`El proyecto fue aprovisionado en: %s

Siguientes pasos para despegar:
  1. cd %s
  2. pnpm install
  3. pnpm dev              (Abre en http://localhost:4321)
  4. wrangler pages deploy (Despliega directo a Cloudflare Pages $0)

Cockpit Telegram configurado en:
  → %s/src/pages/api/telegram.ts`,
		tui.StyleSelected.Render(dir),
		s.state.ProjectSlug,
		dir,
	)

	footer := tui.StyleSubtitle.Render("\nPresiona [ Enter ] para finalizar...")
	content := fmt.Sprintf("%s\n\n%s\n%s", title, instructions, footer)
	return tui.StyleBox.Render(content)
}
