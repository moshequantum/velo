package steps

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/moshequantum/velo/internal/tui"
	"github.com/moshequantum/velo/internal/wizard"
)

type StackStep struct {
	state  *wizard.State
	cursor int
}

var stackChoices = []struct {
	Key         string
	Title       string
	Description string
}{
	{
		Key:         "astro",
		Title:       "Astro + Cloudflare Pages (Recomendado)",
		Description: "Velocidad extrema, SSR/Static híbrido, Islas React y API routes unificadas.",
	},
	{
		Key:         "react-vite",
		Title:       "React + Vite SPA",
		Description: "SPA clásico en Cloudflare Pages sin SSR.",
	},
}

func NewStackStep(state *wizard.State) *StackStep {
	cursor := 0
	if state.Frontend == "react-vite" {
		cursor = 1
	}
	return &StackStep{state: state, cursor: cursor}
}

func (s *StackStep) Init() tea.Cmd {
	return nil
}

func (s *StackStep) Update(msg tea.Msg) (tui.Step, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if s.cursor > 0 {
				s.cursor--
			}
		case "down", "j":
			if s.cursor < len(stackChoices)-1 {
				s.cursor++
			}
		case "enter", " ":
			s.state.Frontend = stackChoices[s.cursor].Key
			return s, tui.Next()
		case "esc":
			return s, tui.Back()
		case "ctrl+c":
			return s, tea.Quit
		}
	}
	return s, nil
}

func (s *StackStep) View() string {
	title := tui.StyleTitle.Render("2. Arquitectura de Frontend & Edge")
	desc := tui.StyleSubtitle.Render("Elige la base del sistema visual y sus rutas en Cloudflare:")

	var options string
	for i, choice := range stackChoices {
		prefix := "  "
		cursorMarker := " "
		if s.cursor == i {
			prefix = "> "
			cursorMarker = "●"
		} else {
			cursorMarker = "○"
		}

		line := fmt.Sprintf("%s%s %s\n    %s\n",
			prefix,
			cursorMarker,
			choice.Title,
			tui.StyleSubtitle.Render(choice.Description),
		)

		if s.cursor == i {
			options += tui.StyleSelected.Render(line) + "\n"
		} else {
			options += tui.StyleNormal.Render(line) + "\n"
		}
	}

	help := tui.StyleSubtitle.Render("[ ↑/↓ ] Navegar   [ Enter ] Seleccionar   [ Esc ] Volver")

	content := fmt.Sprintf("%s\n%s\n\n%s%s", title, desc, options, help)
	return tui.StyleBox.Render(content)
}
