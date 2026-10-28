# ⚡ Velo Framework

**The Zero-Cost Edge & Headless Infrastructure Harness**

Velo is an open-source development ecosystem and CLI designed to cut at least 50% of boilerplate setup time when launching production-ready systems from 0 to 1 on the modern Edge.

## The $0 Stack
* **Frontend & Edge API:** [Astro](https://astro.build) with `@astrojs/cloudflare` on Cloudflare Pages ($0/mo).
* **BaaS:** [InsForge](https://insforge.app) (PostgreSQL, Storage, Realtime, Auth) ($0/mo free tier).
* **Operational Cockpit:** Telegram Bot API for Human-in-the-Loop (HITL) gatekeepers and notifications.
* **Deterministic Context:** Native integration with [OpenPassport](https://github.com/moshequantum/openpassport) v1.0.

## Installation & Usage

```bash
# Interactive TUI Wizard (BubbleTea)
velo init

# Headless / Automated Scaffolding
velo init --name "Mi Proyecto" --slug "mi-proyecto" --frontend astro

# Cryptographic Passport Verification
velo verify ./docs/passport-context/manifest.json
```

## Architecture Contract (`velo.config.ts`)
```typescript
export default {
  project: { name: "mi-app", slug: "mi-app", version: "1.0.0" },
  frontend: { framework: "astro", runtime: "cloudflare-pages" },
  baas: { provider: "insforge", url: "https://mi-app.insforge.app" },
  cockpit: { provider: "telegram", botHandle: "@MiApp_bot" }
};
```
