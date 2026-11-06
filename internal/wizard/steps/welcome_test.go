package steps

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/moshequantum/velo/internal/wizard"
)

func TestWelcomeStep_LanguageToggle(t *testing.T) {
	state := wizard.NewDefaultState()
	if state.Language != "en" {
		t.Fatalf("expected initial language to be 'en', got %q", state.Language)
	}

	step := NewWelcomeStep(state)

	// Check English view
	viewEn := step.View()
	if !strings.Contains(viewEn, "Edge & Headless Harness") {
		t.Errorf("expected English title in view, got:\n%s", viewEn)
	}
	if !strings.Contains(viewEn, "[•] English") {
		t.Errorf("expected English radio checked, got:\n%s", viewEn)
	}

	// Press Tab to switch to Spanish
	step.Update(tea.KeyMsg{Type: tea.KeyTab})
	if state.Language != "es" {
		t.Errorf("expected state.Language to be 'es' after Tab, got %q", state.Language)
	}
	if state.ProjectName != "mi-app-velo" {
		t.Errorf("expected default project name to adapt to 'mi-app-velo', got %q", state.ProjectName)
	}

	// Check Spanish view
	viewEs := step.View()
	if !strings.Contains(viewEs, "Arnés Edge & Headless") {
		t.Errorf("expected Spanish title in view, got:\n%s", viewEs)
	}
	if !strings.Contains(viewEs, "[•] Español") {
		t.Errorf("expected Spanish radio checked, got:\n%s", viewEs)
	}

	// Press '1' to switch back to English
	step.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	if state.Language != "en" {
		t.Errorf("expected state.Language to be 'en' after pressing '1', got %q", state.Language)
	}
	if state.ProjectName != "my-velo-app" {
		t.Errorf("expected default project name to adapt to 'my-velo-app', got %q", state.ProjectName)
	}
}
