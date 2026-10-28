package scaffold

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/moshequantum/velo/internal/wizard"
)

// Generate scaffolds the complete Velo harness based on the wizard State.
func Generate(state *wizard.State) error {
	baseDir := filepath.Join(state.TargetDir, state.ProjectSlug)
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return fmt.Errorf("creating project directory: %w", err)
	}

	dirs := []string{
		filepath.Join(baseDir, "src", "pages", "api"),
		filepath.Join(baseDir, "src", "components"),
		filepath.Join(baseDir, "src", "lib", "insforge"),
		filepath.Join(baseDir, "public"),
	}

	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return fmt.Errorf("creating dir %s: %w", d, err)
		}
	}

	// 1. package.json
	pkgJSON := fmt.Sprintf(`{
  "name": "%s",
  "type": "module",
  "version": "1.0.0",
  "scripts": {
    "dev": "astro dev",
    "start": "astro dev",
    "build": "astro build",
    "preview": "astro preview"
  },
  "dependencies": {
    "@astrojs/cloudflare": "^12.0.0",
    "@astrojs/react": "^4.0.0",
    "@astrojs/tailwind": "^5.0.0",
    "astro": "^5.0.0",
    "react": "^19.0.0",
    "react-dom": "^19.0.0",
    "tailwindcss": "^3.4.0"
  }
}
`, state.ProjectSlug)
	if err := os.WriteFile(filepath.Join(baseDir, "package.json"), []byte(pkgJSON), 0644); err != nil {
		return err
	}

	// 2. astro.config.mjs
	astroConfig := `import { defineConfig } from 'astro/config';
import cloudflare from '@astrojs/cloudflare';
import tailwind from '@astrojs/tailwind';
import react from '@astrojs/react';

export default defineConfig({
  output: 'server',
  adapter: cloudflare({
    imageService: 'cloudflare'
  }),
  integrations: [tailwind(), react()]
});
`
	if err := os.WriteFile(filepath.Join(baseDir, "astro.config.mjs"), []byte(astroConfig), 0644); err != nil {
		return err
	}

	// 3. wrangler.toml
	wranglerToml := fmt.Sprintf(`name = "%s"
compatibility_date = "2024-11-01"
pages_build_output_dir = "dist"

[vars]
ENVIRONMENT = "production"
`, state.ProjectSlug)
	if err := os.WriteFile(filepath.Join(baseDir, "wrangler.toml"), []byte(wranglerToml), 0644); err != nil {
		return err
	}

	// 4. velo.config.ts
	veloConfig := fmt.Sprintf(`// velo.config.ts — Contrato de Infraestructura Velo
export default {
  project: {
    name: "%s",
    slug: "%s",
    version: "1.0.0"
  },
  frontend: {
    framework: "%s",
    runtime: "%s"
  },
  baas: {
    provider: "%s",
    url: "%s"
  },
  cockpit: {
    provider: "%s",
    botHandle: "%s"
  }
};
`, state.ProjectName, state.ProjectSlug, state.Frontend, state.Runtime, state.BaaS, state.InsForgeProjectID, state.Cockpit, state.TelegramBotHandle)
	if err := os.WriteFile(filepath.Join(baseDir, "velo.config.ts"), []byte(veloConfig), 0644); err != nil {
		return err
	}

	// 5. Telegram Webhook Endpoint (src/pages/api/telegram.ts)
	telegramAPI := `import type { APIRoute } from 'astro';

export const POST: APIRoute = async ({ request }) => {
  const secret = request.headers.get('X-Telegram-Bot-Api-Secret-Token');
  const expectedSecret = import.meta.env.TELEGRAM_WEBHOOK_SECRET;

  if (expectedSecret && secret !== expectedSecret) {
    return new Response(JSON.stringify({ error: 'Unauthorized' }), { status: 401 });
  }

  const payload = await request.json();
  
  // Procesar comando o HITL callback_query
  if (payload.callback_query) {
    const action = payload.callback_query.data;
    console.log('[VELO HITL] Acción recibida:', action);
    // Autorización humana procesada en el Edge
  }

  return new Response(JSON.stringify({ ok: true }), {
    headers: { 'Content-Type': 'application/json' }
  });
};
`
	if err := os.WriteFile(filepath.Join(baseDir, "src", "pages", "api", "telegram.ts"), []byte(telegramAPI), 0644); err != nil {
		return err
	}

	// 6. Home page (src/pages/index.astro)
	indexAstro := fmt.Sprintf(`---
// Velo Index Page
const projectName = "%s";
---

<html lang="es">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width" />
    <title>{projectName} • Velo Edge</title>
  </head>
  <body class="bg-slate-950 text-white min-h-screen flex items-center justify-center p-6">
    <main class="max-w-xl text-center space-y-4">
      <div class="inline-block bg-cyan-500/10 border border-cyan-500/20 text-cyan-400 text-xs px-3 py-1 rounded-full uppercase tracking-wider font-semibold">
        Velo Framework Active
      </div>
      <h1 class="text-4xl font-extrabold tracking-tight">{projectName}</h1>
      <p class="text-slate-400">Arnés de infraestructura Edge $0 desplegado en Cloudflare con BaaS InsForge y Cockpit Telegram.</p>
    </main>
  </body>
</html>
`, state.ProjectName)
	if err := os.WriteFile(filepath.Join(baseDir, "src", "pages", "index.astro"), []byte(indexAstro), 0644); err != nil {
		return err
	}

	// 7. .env.example
	envExample := fmt.Sprintf(`INSFORGE_URL=%s
INSFORGE_ANON_KEY=%s
TELEGRAM_BOT_TOKEN=
TELEGRAM_WEBHOOK_SECRET=
TELEGRAM_OWNER_ID=%s
`, state.InsForgeProjectID, state.InsForgeAnonKey, state.TelegramOwnerID)
	if err := os.WriteFile(filepath.Join(baseDir, ".env.example"), []byte(envExample), 0644); err != nil {
		return err
	}

	// 8. .gitignore
	gitignore := `node_modules/
dist/
.astro/
.env
.DS_Store
`
	if err := os.WriteFile(filepath.Join(baseDir, ".gitignore"), []byte(gitignore), 0644); err != nil {
		return err
	}

	return nil
}
