package steps

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/moshequantum/velo/internal/tui"
	"github.com/moshequantum/velo/internal/wizard"
)

type DataStep struct {
	state      *wizard.State
	choiceIdx  int
	urlInput   textinput.Model
	isInputURL bool
}

var dataChoices = []struct {
	Key         string
	Title       string
	Description string
}{
	{
		Key:         "surrealdb",
		Title:       "SurrealDB (Default nuevo)",
		Description: "Multi-modelo, HTTP engine ideal para Workers sin estado.",
	},
	{
		Key:         "insforge",
		Title:       "InsForge (Legacy)",
		Description: "BaaS con PostgreSQL, Storage y Auth.",
	},
	{
		Key:         "none",
		Title:       "Ninguno / Sin Base de Datos",
		Description: "Proyecto estático o edge puro sin persistencia.",
	},
}

func NewDataStep(state *wizard.State) *DataStep {
	choiceIdx := 0
	for i, c := range dataChoices {
		if c.Key == state.DataProvider {
			choiceIdx = i
			break
		}
	}

	ti := textinput.New()
	ti.Placeholder = "URL pública del endpoint (opcional, ej. http://127.0.0.1:8000)"
	ti.Width = 50
	if state.DataURL != "" {
		ti.SetValue(state.DataURL)
	}

	return &DataStep{
		state:     state,
		choiceIdx: choiceIdx,
		urlInput:  ti,
	}
}

func (s *DataStep) Init() tea.Cmd {
	return nil
}

func (s *DataStep) Update(msg tea.Msg) (tui.Step, tea.Cmd) {
	if s.isInputURL {
		var cmd tea.Cmd
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				s.state.DataURL = strings.TrimSpace(s.urlInput.Value())
				return s, tui.Next()
			case "esc":
				s.isInputURL = false
				s.urlInput.Blur()
				return s, nil
			case "ctrl+c":
				return s, tea.Quit
			}
		}
		s.urlInput, cmd = s.urlInput.Update(msg)
		return s, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if s.choiceIdx > 0 {
				s.choiceIdx--
			}
		case "down", "j":
			if s.choiceIdx < len(dataChoices)-1 {
				s.choiceIdx++
			}
		case "enter", " ":
			chosen := dataChoices[s.choiceIdx].Key
			s.state.DataProvider = chosen
			if chosen == "none" {
				s.state.DataURL = ""
				return s, tui.Next()
			}
			s.isInputURL = true
			s.urlInput.Focus()
			return s, textinput.Blink
		case "esc":
			return s, tui.Back()
		case "ctrl+c":
			return s, tea.Quit
		}
	}
	return s, nil
}

func (s *DataStep) View() string {
	title := tui.StyleTitle.Render("3. Capa de Datos & Persistencia")
	desc := tui.StyleSubtitle.Render("Selecciona el proveedor de datos para tu backend:")

	var options string
	for i, choice := range dataChoices {
		prefix := "  "
		cursorMarker := "○"
		if s.choiceIdx == i {
			prefix = "> "
			cursorMarker = "●"
		}

		line := fmt.Sprintf("%s%s %s\n    %s\n",
			prefix,
			cursorMarker,
			choice.Title,
			tui.StyleSubtitle.Render(choice.Description),
		)

		if s.choiceIdx == i {
			options += tui.StyleSelected.Render(line) + "\n"
		} else {
			options += tui.StyleNormal.Render(line) + "\n"
		}
	}

	var urlSection string
	if s.isInputURL {
		urlSection = fmt.Sprintf("\n%s\n%s\n",
			tui.StyleSubtitle.Render("URL del endpoint (opcional, configurable luego en .env/.dev.vars):"),
			s.urlInput.View(),
		)
	}

	help := tui.StyleSubtitle.Render("[ ↑/↓ ] Navegar   [ Enter ] Seleccionar   [ Esc ] Volver")
	if s.isInputURL {
		help = tui.StyleSubtitle.Render("[ Enter ] Confirmar   [ Esc ] Cambiar proveedor")
	}

	content := fmt.Sprintf("%s\n%s\n\n%s%s\n%s", title, desc, options, urlSection, help)
	return tui.StyleBox.Render(content)
}

// Compatibility constructor for BaaSStep
func NewBaaSStep(state *wizard.State) *DataStep {
	return NewDataStep(state)
}
