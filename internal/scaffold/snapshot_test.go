package scaffold

import (
	"bytes"
	"flag"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/moshequantum/velo/internal/wizard"
)

var update = flag.Bool("update", false, "update golden snapshot files")

func TestSnapshot_AstroInsforgeTelegram(t *testing.T) {
	testComboSnapshot(t, "astro-insforge-telegram", &wizard.State{
		ProjectName:       "mi-app-velo",
		ProjectSlug:       "astro-insforge-telegram",
		Frontend:          "astro",
		Runtime:           "cloudflare-pages",
		DataProvider:      "insforge",
		DataURL:           "",
		Cockpit:           "telegram",
		TelegramBotHandle: "",
	})
}

func TestSnapshot_AstroSurrealdbTelegram(t *testing.T) {
	testComboSnapshot(t, "astro-surrealdb-telegram", &wizard.State{
		ProjectName:       "mi-app-surreal",
		ProjectSlug:       "astro-surrealdb-telegram",
		Frontend:          "astro",
		Runtime:           "cloudflare-pages",
		DataProvider:      "surrealdb",
		DataURL:           "http://127.0.0.1:8000",
		Cockpit:           "telegram",
		TelegramBotHandle: "@MyBot",
	})
}

func TestSnapshot_NextSurrealdbTelegram(t *testing.T) {
	testComboSnapshot(t, "next-surrealdb-telegram", &wizard.State{
		ProjectName:       "mi-app-next",
		ProjectSlug:       "next-surrealdb-telegram",
		Frontend:          "next",
		Runtime:           "cloudflare-workers",
		DataProvider:      "surrealdb",
		DataURL:           "http://127.0.0.1:8000",
		Cockpit:           "telegram",
		TelegramBotHandle: "@NextBot",
	})
}

func TestSnapshot_ReactViteSurrealdbTelegram(t *testing.T) {
	testComboSnapshot(t, "react-vite-surrealdb-telegram", &wizard.State{
		ProjectName:       "mi-app-vite",
		ProjectSlug:       "react-vite-surrealdb-telegram",
		Frontend:          "react-vite",
		Runtime:           "cloudflare-pages",
		DataProvider:      "surrealdb",
		DataURL:           "http://127.0.0.1:8000",
		Cockpit:           "telegram",
		TelegramBotHandle: "@ViteBot",
	})
}

func testComboSnapshot(t *testing.T, comboName string, state *wizard.State) {
	goldenDir := filepath.Join("testdata", "golden", comboName)

	if *update {
		if err := os.RemoveAll(goldenDir); err != nil {
			t.Fatalf("removing old golden dir: %v", err)
		}
		if err := os.MkdirAll(goldenDir, 0755); err != nil {
			t.Fatalf("creating golden dir: %v", err)
		}
		state.TargetDir = filepath.Dir(goldenDir)
		state.ProjectSlug = filepath.Base(goldenDir)
		if err := Generate(state); err != nil {
			t.Fatalf("Generate failed on update: %v", err)
		}
		t.Logf("Golden snapshot updated in %s", goldenDir)
		return
	}

	// Render to temp dir
	tmpDir := t.TempDir()
	state.TargetDir = tmpDir
	state.ProjectSlug = comboName

	if err := Generate(state); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	outDir := filepath.Join(tmpDir, comboName)

	// Compare outDir with goldenDir
	err := filepath.Walk(goldenDir, func(goldenPath string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(goldenDir, goldenPath)
		if err != nil {
			return err
		}

		outPath := filepath.Join(outDir, rel)
		outBytes, err := os.ReadFile(outPath)
		if err != nil {
			t.Errorf("file %s missing in generated output: %v", rel, err)
			return nil
		}

		goldenBytes, err := os.ReadFile(goldenPath)
		if err != nil {
			return err
		}

		normGolden := bytes.ReplaceAll(goldenBytes, []byte("\r\n"), []byte("\n"))
		normOut := bytes.ReplaceAll(outBytes, []byte("\r\n"), []byte("\n"))

		if !bytes.Equal(normGolden, normOut) {
			t.Errorf("snapshot mismatch for %s:\n--- GOLDEN ---\n%s\n--- GENERATED ---\n%s", rel, normGolden, normOut)
		}
		return nil
	})

	if err != nil {
		t.Fatalf("comparing snapshots: %v", err)
	}
}
