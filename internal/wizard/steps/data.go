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

type dataChoice struct {
	Key     string
	TitleEn string
	TitleEs string
	DescEn  string
	DescEs  string
}

var dataChoices = []dataChoice{
	{
		Key:     "surrealdb",
		TitleEn: "SurrealDB (New default)",
		TitleEs: "SurrealDB (Default nuevo)",
		DescEn:  "Multi-model, HTTP engine ideal for stateless Workers.",
		DescEs:  "Multi-modelo, HTTP engine ideal para Workers sin estado.",
	},
	{
		Key:     "insforge",
		TitleEn: "InsForge (Legacy)",
		TitleEs: "InsForge (Legacy)",
		DescEn:  "BaaS with PostgreSQL, Storage, and Auth.",
		DescEs:  "BaaS con PostgreSQL, Storage y Auth.",
	},
	{
		Key:     "none",
		TitleEn: "None / No Database",
		TitleEs: "Ninguno / Sin Base de Datos",
		DescEn:  "Static project or pure edge without persistence.",
		DescEs:  "Proyecto estático o edge puro sin persistencia.",
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
	if state.IsSpanish() {
		ti.Placeholder = "URL pública del endpoint (opcional, ej. http://127.0.0.1:8000)"
	} else {
		ti.Placeholder = "Public endpoint URL (optional, e.g. http://127.0.0.1:8000)"
	}
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
	isEs := s.state.IsSpanish()

	var title, desc, urlPrompt, helpText string
	if isEs {
		title = tui.StyleTitle.Render("3. Capa de Datos & Persistencia")
		desc = tui.StyleSubtitle.Render("Selecciona el proveedor de datos para tu backend:")
		urlPrompt = "URL del endpoint (opcional, configurable luego en .env/.dev.vars):"
		if s.isInputURL {
			helpText = "[ Enter ] Confirmar   [ Esc ] Cambiar proveedor"
		} else {
			helpText = "[ ↑/↓ ] Navegar   [ Enter ] Seleccionar   [ Esc ] Volver"
		}
	} else {
		title = tui.StyleTitle.Render("3. Data Layer & Persistence")
		desc = tui.StyleSubtitle.Render("Select the data provider for your backend:")
		urlPrompt = "Endpoint URL (optional, configurable later in .env/.dev.vars):"
		if s.isInputURL {
			helpText = "[ Enter ] Confirm   [ Esc ] Change provider"
		} else {
			helpText = "[ ↑/↓ ] Navigate   [ Enter ] Select   [ Esc ] Back"
		}
	}

	var options string
	for i, choice := range dataChoices {
		prefix := "  "
		cursorMarker := "○"
		if s.choiceIdx == i {
			prefix = "> "
			cursorMarker = "●"
		}

		choiceTitle := choice.TitleEn
		choiceDesc := choice.DescEn
		if isEs {
			choiceTitle = choice.TitleEs
			choiceDesc = choice.DescEs
		}

		line := fmt.Sprintf("%s%s %s\n    %s\n",
			prefix,
			cursorMarker,
			choiceTitle,
			tui.StyleSubtitle.Render(choiceDesc),
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
			tui.StyleSubtitle.Render(urlPrompt),
			s.urlInput.View(),
		)
	}

	help := tui.StyleSubtitle.Render(helpText)
	content := fmt.Sprintf("%s\n%s\n\n%s%s\n%s", title, desc, options, urlSection, help)
	return tui.StyleBox.Render(content)
}

// Compatibility constructor for BaaSStep
func NewBaaSStep(state *wizard.State) *DataStep {
	return NewDataStep(state)
}
