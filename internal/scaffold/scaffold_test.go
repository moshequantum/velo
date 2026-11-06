package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/moshequantum/velo/internal/wizard"
)

func TestGenerate_SecurityGuarantees(t *testing.T) {
	tmpDir := t.TempDir()

	testSecretState := &wizard.State{
		ProjectName:       "Security Test App",
		ProjectSlug:       "sec-app",
		Frontend:          "astro",
		Runtime:           "cloudflare-pages",
		DataProvider:      "insforge",
		DataURL:           "https://my-insforge-cluster.internal",
		Cockpit:           "telegram",
		TelegramBotHandle: "TestSecBot",
		TelegramOwnerID:   "9988776655",
		TargetDir:         tmpDir,
	}

	if err := Generate(testSecretState); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	projectDir := filepath.Join(tmpDir, "sec-app")

	// 1. Verify .env.example contains NO state values (especially secrets)
	envPath := filepath.Join(projectDir, ".env.example")
	envBytes, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatalf("reading .env.example: %v", err)
	}
	envContent := strings.ReplaceAll(string(envBytes), "\r\n", "\n")

	if strings.Contains(envContent, testSecretState.TelegramOwnerID) {
		t.Errorf("SECURITY FLAW: .env.example contains TelegramOwnerID!")
	}
	if strings.Contains(envContent, testSecretState.DataURL) {
		t.Errorf("SECURITY FLAW: .env.example contains DataURL!")
	}

	// All keys must be empty in .env.example
	expectedEmptyKeys := []string{
		"INSFORGE_URL=\n",
		"INSFORGE_ANON_KEY=\n",
		"TELEGRAM_BOT_TOKEN=\n",
		"TELEGRAM_WEBHOOK_SECRET=\n",
		"TELEGRAM_OWNER_ID=\n",
	}
	for _, k := range expectedEmptyKeys {
		if !strings.Contains(envContent, k) {
			t.Errorf("expected .env.example to contain empty key assignment %q, got:\n%s", k, envContent)
		}
	}

	// 2. Verify telegram.ts security invariants
	telegramPath := filepath.Join(projectDir, "src", "pages", "api", "telegram.ts")
	tgBytes, err := os.ReadFile(telegramPath)
	if err != nil {
		t.Fatalf("reading telegram.ts: %v", err)
	}
	tgContent := string(tgBytes)

	// Must NOT contain !== on secret comparison
	if strings.Contains(tgContent, "secret !== expectedSecret") || strings.Contains(tgContent, "secret != expectedSecret") {
		t.Errorf("CRITICAL: telegram.ts uses non-constant-time comparison !== or != on secret")
	}

	// Must contain timingSafeEqual
	if !strings.Contains(tgContent, "timingSafeEqual") {
		t.Errorf("telegram.ts must implement timingSafeEqual")
	}

	// Must fail-closed (status 503 if secret unconfigured)
	if !strings.Contains(tgContent, "status: 503") {
		t.Errorf("telegram.ts must fail closed with status 503 when expectedSecret is unconfigured")
	}

	// Must validate owner ID with status 403
	if !strings.Contains(tgContent, "status: 403") {
		t.Errorf("telegram.ts must reject unauthorized owner with status 403")
	}

	// Must validate approve: / reject: action format
	if !strings.Contains(tgContent, "approve|reject") {
		t.Errorf("telegram.ts must enforce approve/reject action format")
	}

	// 3. Verify wrangler.toml fixed compatibility date
	wranglerPath := filepath.Join(projectDir, "wrangler.toml")
	wranglerBytes, err := os.ReadFile(wranglerPath)
	if err != nil {
		t.Fatalf("reading wrangler.toml: %v", err)
	}
	if !strings.Contains(string(wranglerBytes), `compatibility_date = "2025-01-01"`) {
		t.Errorf("wrangler.toml should have fixed recent compatibility_date 2025-01-01")
	}
}
