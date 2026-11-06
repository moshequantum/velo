package scaffold

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/moshequantum/velo/internal/wizard"
)

//go:embed all:templates
var templateFS embed.FS

// templateContext supplies template values for text/template execution.
type templateContext struct {
	ProjectName       string
	ProjectSlug       string
	Frontend          string
	Runtime           string
	BaaSProvider      string
	BaaSURL           string
	Cockpit           string
	TelegramBotHandle string
	TelegramOwnerID   string
}

// Generate scaffolds the complete Velo harness based on the wizard State using composable templates (C3).
func Generate(state *wizard.State) error {
	baseDir := filepath.Join(state.TargetDir, state.ProjectSlug)
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return fmt.Errorf("creating project directory: %w", err)
	}

	baasProvider := state.BaaS
	if baasProvider == "" {
		baasProvider = "insforge"
	}
	baasURL := state.InsForgeProjectID

	ctx := templateContext{
		ProjectName:       state.ProjectName,
		ProjectSlug:       state.ProjectSlug,
		Frontend:          state.Frontend,
		Runtime:           state.Runtime,
		BaaSProvider:      baasProvider,
		BaaSURL:           baasURL,
		Cockpit:           state.Cockpit,
		TelegramBotHandle: state.TelegramBotHandle,
		TelegramOwnerID:   state.TelegramOwnerID,
	}

	// Composition order: base -> frontend -> data -> cockpit (last wins)
	layers := []string{
		"templates/base",
		fmt.Sprintf("templates/frontend/%s", state.Frontend),
		fmt.Sprintf("templates/data/%s", baasProvider),
	}
	if state.Cockpit != "" && state.Cockpit != "none" {
		layers = append(layers, fmt.Sprintf("templates/cockpit/%s", state.Cockpit))
	}

	renderedFiles := make(map[string][]byte)

	for _, layer := range layers {
		_, err := fs.ReadDir(templateFS, layer)
		if err != nil {
			// Layer not present, continue
			continue
		}

		err = fs.WalkDir(templateFS, layer, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}

			rel, err := filepath.Rel(layer, path)
			if err != nil {
				return err
			}
			rel = filepath.ToSlash(rel)

			content, err := templateFS.ReadFile(path)
			if err != nil {
				return fmt.Errorf("reading embedded file %s: %w", path, err)
			}

			var destRel string
			var destBytes []byte

			if strings.HasSuffix(rel, ".tmpl") {
				destRel = strings.TrimSuffix(rel, ".tmpl")
				tmpl, err := template.New(destRel).Parse(string(content))
				if err != nil {
					return fmt.Errorf("parsing template %s: %w", path, err)
				}
				var buf bytes.Buffer
				if err := tmpl.Execute(&buf, ctx); err != nil {
					return fmt.Errorf("executing template %s: %w", path, err)
				}
				destBytes = buf.Bytes()
			} else {
				destRel = rel
				destBytes = content
			}

			renderedFiles[destRel] = destBytes
			return nil
		})
		if err != nil {
			return fmt.Errorf("processing layer %s: %w", layer, err)
		}
	}

	for relPath, data := range renderedFiles {
		destFile := filepath.Join(baseDir, filepath.FromSlash(relPath))
		if strings.HasSuffix(relPath, ".gitkeep") {
			if err := os.MkdirAll(filepath.Dir(destFile), 0755); err != nil {
				return fmt.Errorf("creating dir for %s: %w", relPath, err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(destFile), 0755); err != nil {
			return fmt.Errorf("creating dir for %s: %w", relPath, err)
		}
		if err := os.WriteFile(destFile, data, 0644); err != nil {
			return fmt.Errorf("writing file %s: %w", destFile, err)
		}
	}

	return nil
}
