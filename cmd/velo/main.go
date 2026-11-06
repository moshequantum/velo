package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/moshequantum/velo/internal/scaffold"
	"github.com/moshequantum/velo/internal/tui"
	"github.com/moshequantum/velo/internal/wizard"
	"github.com/moshequantum/velo/internal/wizard/steps"
	"github.com/moshequantum/velo/pkg/manifest"
)

const Version = "0.2.0"

var (
	validLanguages = map[string]bool{"en": true, "es": true}
	validFrontends = map[string]bool{"astro": true, "next": true, "react-vite": true}
	validRuntimes  = map[string]bool{"cloudflare-pages": true, "cloudflare-workers": true}
	validData      = map[string]bool{"surrealdb": true, "insforge": true, "none": true}
	validCockpits  = map[string]bool{"telegram": true, "none": true}
)

func main() {
	if len(os.Args) < 2 {
		if !isTTY() {
			fmt.Fprintln(os.Stderr, "Error: terminal no interactiva detectada (sin TTY). Debe especificar --name en modo headless.")
			os.Exit(2)
		}
		runWizard()
		return
	}

	cmd := os.Args[1]

	switch cmd {
	case "init", "create":
		code := handleInit(os.Args[2:])
		os.Exit(code)

	case "doctor":
		code := handleDoctor()
		os.Exit(code)

	case "version", "--version", "-v":
		fmt.Printf("velo CLI v%s — Edge & Headless Framework\n", Version)
		os.Exit(0)

	case "verify":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Uso: velo verify <ruta/al/directorio-o-manifest.json>")
			os.Exit(2)
		}
		code := handleVerify(os.Args[2])
		os.Exit(code)

	case "help", "--help", "-h":
		printHelp()
		os.Exit(0)

	default:
		fmt.Fprintf(os.Stderr, "Comando desconocido: %s\n\n", cmd)
		printHelp()
		os.Exit(2)
	}
}

func isTTY() bool {
	fd := os.Stdin.Fd()
	return isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd)
}

func handleInit(args []string) int {
	if len(args) == 0 {
		if !isTTY() {
			fmt.Fprintln(os.Stderr, "Error: terminal no interactiva detectada (sin TTY). Debe especificar --name en modo headless.")
			return 2
		}
		runWizard()
		return 0
	}

	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	lang := fs.String("lang", "en", "Idioma / Language (en, es)")
	name := fs.String("name", "", "Nombre del proyecto")
	slug := fs.String("slug", "", "Slug identificador (carpeta)")
	frontend := fs.String("frontend", "astro", "Frontend framework (astro, next, react-vite)")
	runtime := fs.String("runtime", "cloudflare-pages", "Runtime target (cloudflare-pages, cloudflare-workers)")
	data := fs.String("data", "surrealdb", "Provider de datos (surrealdb, insforge, none)")
	dataURL := fs.String("data-url", "", "URL pública del endpoint de datos (NUNCA credenciales)")
	cockpit := fs.String("cockpit", "telegram", "Cockpit HITL (telegram, none)")
	bot := fs.String("bot", "", "Handle del bot de Telegram")
	owner := fs.String("owner", "", "ID numérico de Telegram del dueño")
	target := fs.String("target", ".", "Directorio destino")
	dryRun := fs.Bool("dry-run", false, "Lista los archivos sin escribir en disco")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	if *name == "" {
		if !isTTY() {
			fmt.Fprintln(os.Stderr, "Error: terminal no interactiva detectada. Debe proporcionar el flag --name.")
			return 2
		}
		runWizard()
		return 0
	}

	// Validate enums
	if !validLanguages[*lang] {
		fmt.Fprintf(os.Stderr, "Error: valor inválido para --lang: %q (válidos: en, es)\n", *lang)
		return 2
	}
	if !validFrontends[*frontend] {
		fmt.Fprintf(os.Stderr, "Error: valor inválido para --frontend: %q (válidos: astro, next, react-vite)\n", *frontend)
		return 2
	}
	if !validRuntimes[*runtime] {
		fmt.Fprintf(os.Stderr, "Error: valor inválido para --runtime: %q (válidos: cloudflare-pages, cloudflare-workers)\n", *runtime)
		return 2
	}
	if !validData[*data] {
		fmt.Fprintf(os.Stderr, "Error: valor inválido para --data: %q (válidos: surrealdb, insforge, none)\n", *data)
		return 2
	}
	if !validCockpits[*cockpit] {
		fmt.Fprintf(os.Stderr, "Error: valor inválido para --cockpit: %q (válidos: telegram, none)\n", *cockpit)
		return 2
	}

	finalSlug := *slug
	if finalSlug == "" {
		finalSlug = *name
	}
	finalSlug = wizard.Slugify(finalSlug)

	state := &wizard.State{
		Language:          *lang,
		ProjectName:       *name,
		ProjectSlug:       finalSlug,
		Frontend:          *frontend,
		Runtime:           *runtime,
		DataProvider:      *data,
		DataURL:           *dataURL,
		Cockpit:           *cockpit,
		TelegramBotHandle: *bot,
		TelegramOwnerID:   *owner,
		TargetDir:         *target,
		DryRun:            *dryRun,
	}

	if !*dryRun {
		if state.IsSpanish() {
			fmt.Printf("⚡ Generando arnés Velo para '%s' en %s/%s...\n", state.ProjectName, state.TargetDir, state.ProjectSlug)
		} else {
			fmt.Printf("⚡ Generating Velo harness for '%s' in %s/%s...\n", state.ProjectName, state.TargetDir, state.ProjectSlug)
		}
	}

	if err := scaffold.Generate(state); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error al generar: %v\n", err)
		return 1
	}

	if *dryRun {
		return 0
	}

	if state.IsSpanish() {
		fmt.Println("🚀 ¡Arnés Velo generado exitosamente!")
		fmt.Printf("  • Directorio: %s/%s\n", state.TargetDir, state.ProjectSlug)
		fmt.Printf("  • Stack:      %s + %s + %s\n", state.Frontend, state.Runtime, state.DataProvider)
		if state.Cockpit != "none" {
			fmt.Printf("  • Cockpit:    Telegram HITL (/api/telegram)\n")
		}
		fmt.Println("\nPara arrancar:")
		fmt.Printf("  cd %s/%s && pnpm install && pnpm dev\n", state.TargetDir, state.ProjectSlug)
	} else {
		fmt.Println("🚀 Velo harness generated successfully!")
		fmt.Printf("  • Directory:  %s/%s\n", state.TargetDir, state.ProjectSlug)
		fmt.Printf("  • Stack:      %s + %s + %s\n", state.Frontend, state.Runtime, state.DataProvider)
		if state.Cockpit != "none" {
			fmt.Printf("  • Cockpit:    Telegram HITL (/api/telegram)\n")
		}
		fmt.Println("\nTo launch:")
		fmt.Printf("  cd %s/%s && pnpm install && pnpm dev\n", state.TargetDir, state.ProjectSlug)
	}

	return 0
}

func handleDoctor() int {
	fmt.Println("🩺 Diagnóstico de Entorno Velo (doctor)")
	fmt.Println("--------------------------------------")

	tools := []struct {
		name    string
		cmd     string
		args    []string
		crucial bool
	}{
		{"Node.js", "node", []string{"-v"}, true},
		{"pnpm", "pnpm", []string{"-v"}, true},
		{"Wrangler", "wrangler", []string{"--version"}, true},
		{"SurrealDB", "surreal", []string{"version"}, false},
	}

	allOk := true
	for _, t := range tools {
		path, err := exec.LookPath(t.cmd)
		if err != nil {
			fmt.Printf("  ❌ %-10s : No encontrado en el PATH\n", t.name)
			if t.crucial {
				allOk = false
			}
			continue
		}

		out, err := exec.Command(t.cmd, t.args...).CombinedOutput()
		if err != nil {
			fmt.Printf("  ⚠️  %-10s : Presente en %s pero falló al consultar versión\n", t.name, path)
			if t.crucial {
				allOk = false
			}
			continue
		}

		ver := strings.TrimSpace(string(out))
		// Take first line
		if idx := strings.Index(ver, "\n"); idx != -1 {
			ver = strings.TrimSpace(ver[:idx])
		}
		fmt.Printf("  ✅ %-10s : %s (%s)\n", t.name, ver, path)
	}

	if !allOk {
		fmt.Fprintln(os.Stderr, "\n❌ Falta una o más herramientas requeridas.")
		return 1
	}

	fmt.Println("\n✅ Todas las herramientas requeridas están operativas.")
	return 0
}

func handleVerify(manifestPath string) int {
	fmt.Printf("🔍 Verificando integridad del pasaporte: %s\n", manifestPath)
	m, err := manifest.LoadAndVerify(manifestPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Falló la verificación de integridad: %v\n", err)
		return 1
	}

	fmt.Println("✅ Sello criptográfico SHA-256 VÁLIDO.")
	fmt.Printf("  • Proyecto:             %s\n", m.Name)
	fmt.Printf("  • Versión del paquete:  %s\n", m.PackageVersion)
	fmt.Printf("  • Archivos verificados: %d\n", len(m.VerifiedFiles))
	return 0
}

func printHelp() {
	fmt.Println(`⚡ VELO — Framework & Arnés Edge Moderno

Uso:
  velo [comando] [flags]

Comandos:
  init                     Inicia el wizard interactivo o genera en modo headless
  doctor                   Verifica dependencias del entorno local (node, pnpm, wrangler, surreal)
  verify <ruta>            Verifica el sello criptográfico SHA-256 de un Pasaporte OpenPassport v1
  version                  Muestra la versión de Velo
  help                     Muestra esta ayuda

Flags para 'velo init':
  --lang <str>             Idioma / Language: en (default), es
  --name <str>             Nombre del proyecto (requerido en headless)
  --slug <str>             Slug identificador de la carpeta (auto-sanitizado)
  --frontend <str>         Frontend: astro (default), next, react-vite
  --runtime <str>          Runtime: cloudflare-pages (default), cloudflare-workers
  --data <str>             Provider de datos: surrealdb (default), insforge, none
  --data-url <str>         URL pública del endpoint de datos
  --cockpit <str>          Cockpit HITL: telegram (default), none
  --bot <str>              Handle del bot de Telegram
  --owner <str>            ID numérico del dueño de Telegram
  --target <dir>           Directorio de destino (default: .)
  --dry-run                Muestra la lista de archivos sin escribir en disco`)
}

func runWizard() {
	state := wizard.NewDefaultState()

	wizardSteps := []tui.Step{
		steps.NewWelcomeStep(state),
		steps.NewNamingStep(state),
		steps.NewStackStep(state),
		steps.NewDataStep(state),
		steps.NewCockpitStep(state),
		steps.NewReviewStep(state),
		steps.NewInstallStep(state),
	}

	if err := wizard.Run(wizardSteps); err != nil {
		fmt.Fprintf(os.Stderr, "Error ejecutando el wizard: %v\n", err)
		os.Exit(1)
	}
}
