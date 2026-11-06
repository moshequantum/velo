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
	goldenDir := filepath.Join("testdata", "golden", "astro-insforge-telegram")

	state := &wizard.State{
		ProjectName:       "mi-app-velo",
		ProjectSlug:       "mi-app-velo",
		Frontend:          "astro",
		Runtime:           "cloudflare-pages",
		BaaS:              "insforge",
		InsForgeProjectID: "",
		Cockpit:           "telegram",
		TelegramBotHandle: "",
		TargetDir:         t.TempDir(),
	}

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
	state.ProjectSlug = "astro-insforge-telegram"

	if err := Generate(state); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	outDir := filepath.Join(tmpDir, "astro-insforge-telegram")

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

		// Normalize line endings for comparison
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
