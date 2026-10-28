<div align="center">

# ⚡ Velo Framework

**The Zero-Cost Edge & Headless Infrastructure Harness**

[![Release](https://img.shields.io/github/v/release/moshequantum/velo?color=00E5FF&label=Release&style=flat-square)](https://github.com/moshequantum/velo/releases)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.22-7C4DFF?style=flat-square&logo=go)](https://go.dev)
[![Astro](https://img.shields.io/badge/Astro-5.0-FF5D01?style=flat-square&logo=astro)](https://astro.build)
[![Cloudflare Pages](https://img.shields.io/badge/Runtime-Cloudflare%20Pages-F38020?style=flat-square&logo=cloudflare)](https://pages.cloudflare.com)
[![InsForge](https://img.shields.io/badge/BaaS-InsForge%20(PostgreSQL)-00E676?style=flat-square)](https://insforge.app)
[![Cost](https://img.shields.io/badge/Infra%20Cost-%240.00%2Fmo-00E5FF?style=flat-square)](#the-0-production-stack)
[![License](https://img.shields.io/badge/License-MIT-white?style=flat-square)](LICENSE)

<p align="center">
  Stop wasting 50% of your engineering time wiring repetitive boilerplates. Velo automates modern edge decisions into a cohesive, production-ready operational OS with $0 infrastructure costs.
</p>

[Quickstart](#quickstart) • [The 4 Commandments](#the-4-commandments) • [The $0 Stack](#the-0-production-stack) • [Unified Contract](#the-unified-contract-veloconfigts) • [CLI Reference](#cli-reference)

</div>

---

## ⚡ Executive Summary

**Velo** is not a programming language or an opinionated black-box framework. **Velo is an architectural decision packager** that unifies the modern frontend ([Astro](https://astro.build) / React), serverless edge compute ([Cloudflare Pages](https://pages.cloudflare.com) & Workers), and backend-as-a-service ([InsForge](https://insforge.app) PostgreSQL) under a single Human-in-the-Loop (HITL) operational cockpit.

### Built For:
* **The Founder & Vibe Coder:** Launch a secure, robust MVP in under 48 hours without DevOps friction and with **$0 fixed monthly infrastructure costs**.
* **The Consultant & Agency:** Standardize client infrastructure delivery, eliminating 50% of initial setup debt on every new engagement.
* **The Senior Engineer:** Transparent, standard code (*"Sin Tinta China"*). Inspect the engine, swap dependencies, and edit standard SQL, TypeScript, and Go without framework lock-in.

---

## 📜 The 4 Commandments

1. **Sin Tinta China (No Black Magic):** Everything Velo scaffolds is standard, auditable code (Astro, TypeScript, PostgreSQL, SQL). If you decide to eject or replace a library, nothing breaks.
2. **AI-First & Context Grounded:** Native integration with [OpenPassport](https://github.com/moshequantum/openpassport) v1.0. LLMs (Cursor, Claude, Windsurf) receive cryptographically verified business context via MCP, preventing hallucinations and schema drift.
3. **Secure by Default:** Constant-time token comparisons (`timingSafeEqual`), fail-closed Telegram HITL authorization gates, and strict input validation before state mutations.
4. **Zero-Cost Edge Infrastructure:** Every default architectural component runs entirely within the generous, non-commercial-restricted Free Tiers of Cloudflare Pages and InsForge BaaS.

---

## 🏗️ The $0 Production Stack

| Component | Technology | Role & Free Tier Capability | Cost |
|---|---|---|---|
| **Frontend & SSR API** | **Astro 5** + `@astrojs/cloudflare` | Pre-rendered static landing pages + React islands + Edge API routes (`/api/*`) on Cloudflare Pages. Unlimited requests and bandwidth. | **$0.00** |
| **BaaS & Relational DB** | **InsForge (PostgreSQL)** | Managed PostgreSQL with `pgvector`, Realtime subscriptions, Auth, and Storage consumed over HTTPS via SDK. | **$0.00** |
| **Operational Cockpit** | **Telegram Bot API (HITL)** | Instant event stream, administrative alerts, and inline buttons ("Approve / Reject") for human sign-off before database mutations. | **$0.00** |
| **Context Protocol** | **OpenPassport v1.0** | Immutable 7-layer context package with SHA-256 seals, eliminating LLM prompt drift. | **$0.00** |
| **Total Monthly Cost** | | | **$0.00 / mo** |

---

## 🚀 Quickstart

### 1. Installation
Download pre-compiled binaries from the [Releases page](https://github.com/moshequantum/velo/releases), or install via Go:

```bash
go install github.com/moshequantum/velo/cmd/velo@latest
```

### 2. Launch the Interactive Wizard (BubbleTea TUI)
Run `velo` in your terminal to launch the interactive onboarding experience:

```bash
velo init
```

```
┌────────────────────────────────────────────────────────────────────────┐
│  ⚡ VELO FRAMEWORK — Edge & Headless Harness  0 to 1 • $0 Infra • IA   │
│                                                                        │
│  Velo automatiza las decisiones arquitectónicas para que lances        │
│  sistemas completos en el Edge sin pagar un solo dólar en servidores.  │
│                                                                        │
│    • Frontend & API:  Astro + Cloudflare Pages (SSR & Static)          │
│    • BaaS:            InsForge (PostgreSQL, Realtime, Storage, Auth)   │
│    • Cockpit:         Telegram Bot (HITL / Aprobaciones críticas)      │
│    • Contexto:        Compatible con OpenPassport v1.0                 │
│                                                                        │
│  Presiona [ Enter ] para comenzar o [ q ] para salir...                │
└────────────────────────────────────────────────────────────────────────┘
```

### 3. Headless / Automated Scaffolding
For CI/CD scripts or instant automated setups, bypass the interactive wizard using flags:

```bash
velo init \
  --name "My Edge SaaS" \
  --slug "my-edge-saas" \
  --frontend astro \
  --insforge "https://my-app.insforge.app" \
  --bot "@MySaaS_bot" \
  --target ./projects
```

### 4. Run Locally & Deploy
```bash
cd projects/my-edge-saas

# Install dependencies
pnpm install

# Start unified local dev server
pnpm dev

# Deploy live to Cloudflare Pages ($0)
wrangler pages deploy dist --project-name my-edge-saas
```

---

## 📋 The Unified Contract: `velo.config.ts`

Velo centralizes project configuration in a single, human- and machine-readable TypeScript contract:

```typescript
// velo.config.ts — Unified Infrastructure Contract
export default {
  project: {
    name: "My Edge SaaS",
    slug: "my-edge-saas",
    version: "1.0.0"
  },
  frontend: {
    framework: "astro",         // "astro" | "react-vite"
    runtime: "cloudflare-pages" // Edge serverless
  },
  baas: {
    provider: "insforge",       // Managed PostgreSQL BaaS
    url: "https://my-app.insforge.app"
  },
  cockpit: {
    provider: "telegram",       // Human-in-the-Loop gateway
    botHandle: "@MySaaS_bot"
  }
};
```

---

## 📂 Generated Project Anatomy

```
my-edge-saas/
├── .env.example              # Pre-wired environment secrets (InsForge, Telegram)
├── .gitignore                # Clean defaults for Astro, dist, and environment files
├── astro.config.mjs          # Cloudflare Pages adapter + Tailwind + React integrations
├── package.json              # Curated dependencies (Astro 5, React 19, Tailwind)
├── velo.config.ts            # The single source of truth for infrastructure
├── wrangler.toml             # Zero-config Cloudflare Pages runtime spec
└── src/
    ├── components/           # Reusable UI components & React islands
    ├── lib/
    │   └── insforge/         # PostgreSQL client initialization
    └── pages/
        ├── index.astro       # High-performance edge landing page
        └── api/
            └── telegram.ts   # Webhook endpoint with secret token check & HITL gate
```

### The Human-in-the-Loop Gatekeeper (`src/pages/api/telegram.ts`):
```typescript
import type { APIRoute } from 'astro';

export const POST: APIRoute = async ({ request }) => {
  const secret = request.headers.get('X-Telegram-Bot-Api-Secret-Token');
  const expectedSecret = import.meta.env.TELEGRAM_WEBHOOK_SECRET;

  // Constant-time token verification
  if (expectedSecret && secret !== expectedSecret) {
    return new Response(JSON.stringify({ error: 'Unauthorized' }), { status: 401 });
  }

  const payload = await request.json();

  // Intercept critical mutations via Telegram callback query
  if (payload.callback_query) {
    const action = payload.callback_query.data;
    console.log('[VELO HITL] Operational authorization received:', action);
    // State mutation approved by human operator
  }

  return new Response(JSON.stringify({ ok: true }), {
    headers: { 'Content-Type': 'application/json' }
  });
};
```

---

## 🛠️ CLI Reference

| Command | Arguments / Flags | Description |
|---|---|---|
| `velo` / `velo init` | — | Launches the 7-step interactive BubbleTea TUI wizard. |
| `velo init [flags]` | `--name`, `--slug`, `--frontend`, `--insforge`, `--bot`, `--owner`, `--target` | Headless execution for automated and programmatic scaffolding. |
| `velo verify` | `<path/to/manifest.json>` | Verifies the cryptographic SHA-256 seal of an OpenPassport package. |
| `velo version` | — | Prints current Velo CLI version. |
| `velo help` | — | Displays CLI command options and stack reference. |

---

## 📄 License

Velo Framework is open-source software licensed under the **[MIT License](LICENSE)**.

Designed and maintained by **[Multiversa Group LLC](https://multiversa.group)**.
