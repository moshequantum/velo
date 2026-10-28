package wizard

// State holds the accumulated project configuration across the wizard steps.
type State struct {
	ProjectName       string
	ProjectSlug       string
	Frontend          string // "astro" (default) or "react-vite"
	Runtime           string // "cloudflare-pages"
	BaaS              string // "insforge"
	InsForgeProjectID string
	InsForgeAnonKey   string
	Cockpit           string // "telegram"
	TelegramBotHandle string
	TelegramOwnerID   string
	TargetDir         string
}

// NewDefaultState returns state initialized with Velo recommended defaults ($0 edge stack).
func NewDefaultState() *State {
	return &State{
		ProjectName: "mi-app-velo",
		ProjectSlug: "mi-app-velo",
		Frontend:    "astro",
		Runtime:     "cloudflare-pages",
		BaaS:        "insforge",
		Cockpit:     "telegram",
		TargetDir:   ".",
	}
}
