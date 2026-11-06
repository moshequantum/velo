package wizard

import (
	"regexp"
	"strings"
)

// State holds the accumulated project configuration across the wizard steps and CLI flags (Contract C2).
type State struct {
	ProjectName       string
	ProjectSlug       string
	Frontend          string // "astro" | "next" | "react-vite"
	Runtime           string // "cloudflare-pages" | "cloudflare-workers"
	DataProvider      string // "surrealdb" (default) | "insforge" (legacy) | "none"
	DataURL           string // URL pública del endpoint; NUNCA claves
	Cockpit           string // "telegram" | "none"
	TelegramBotHandle string
	TelegramOwnerID   string
	TargetDir         string
	DryRun            bool
}

// NewDefaultState returns state initialized with Velo recommended defaults.
func NewDefaultState() *State {
	return &State{
		ProjectName:  "mi-app-velo",
		ProjectSlug:  "mi-app-velo",
		Frontend:     "astro",
		Runtime:      "cloudflare-pages",
		DataProvider: "surrealdb",
		DataURL:      "",
		Cockpit:      "telegram",
		TargetDir:    ".",
		DryRun:       false,
	}
}

// Slugify converts an arbitrary project name into a safe, canonical directory slug.
func Slugify(s string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9_]+`)
	slug := strings.ToLower(re.ReplaceAllString(s, "-"))
	reDashes := regexp.MustCompile(`-+`)
	slug = reDashes.ReplaceAllString(slug, "-")
	return strings.Trim(slug, "-")
}
