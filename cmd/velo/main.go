package main

import (
	"fmt"
	"os"

	"github.com/moshequantum/velo/internal/tui"
	"github.com/moshequantum/velo/internal/wizard"
	"github.com/moshequantum/velo/internal/wizard/steps"
	"github.com/moshequantum/velo/pkg/manifest"
)

const Version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		runWizard()
		return
	}

	cmd := os.Args[1]

	switch cmd {
	case "init", "create":
		runWizard()

	case "version", "--version", "-v":
		fmt.Printf("velo CLI v%s — Edge & Headless Framework\n", Version)

	case "verify":
		if len(os.Args) < 3 {
			fmt.Println("Uso: velo verify <ruta/al/manifest.json>")
			os.Exit(1)
		}
		handleVerify(os.Args[2])

	case "help", "--help", "-h":
		printHelp()

	default:
		fmt.Printf("Comando desconocido: %s\n\n", cmd)
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println(`⚡ VELO — Framework & Arnés Edge $0

Uso:
  velo [comando]

Comandos:
  init                     Inicia el wizard interactivo de creación de arnés
  verify <manifest.json>   Verifica el sello criptográfico SHA-256 de un Pasaporte de Contexto
  version                  Muestra la versión de Velo
  help                     Muestra esta ayuda

Stack por Defecto ($0/mes):
  • Frontend & API:  Astro + Cloudflare Pages
  • BaaS:            InsForge (PostgreSQL, Storage, Auth)
  • Cockpit:         Telegram Bot (Human-in-the-Loop)`)
}

func runWizard() {
	state := wizard.NewDefaultState()

	wizardSteps := []tui.Step{
		steps.NewWelcomeStep(state),
		steps.NewNamingStep(state),
		steps.NewStackStep(state),
		steps.NewBaaSStep(state),
		steps.NewCockpitStep(state),
		steps.NewReviewStep(state),
		steps.NewInstallStep(state),
	}

	if err := wizard.Run(wizardSteps); err != nil {
		fmt.Printf("Error ejecutando el wizard: %v\n", err)
		os.Exit(1)
	}
}

func handleVerify(manifestPath string) {
	fmt.Printf("🔍 Verificando integridad del pasaporte: %s\n", manifestPath)
	m, err := manifest.LoadAndVerify(manifestPath)
	if err != nil {
		fmt.Printf("❌ Falló la verificación de integridad: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Sello criptográfico SHA-256 VÁLIDO.")
	fmt.Printf("  • Proyecto: %s (v%s)\n", m.Project.Name, m.Version)
	fmt.Printf("  • Capas registradas: %d\n", len(m.Layers))
	fmt.Printf("  • Frontend objetivo: %s\n", m.Stack.Frontend)
	fmt.Printf("  • BaaS objetivo:     %s\n", m.Stack.BaaS)
}
