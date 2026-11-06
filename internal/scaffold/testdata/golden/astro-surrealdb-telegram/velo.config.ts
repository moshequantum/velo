// velo.config.ts — Contrato de Infraestructura Velo
export default {
  project: {
    name: "mi-app-surreal",
    slug: "astro-surrealdb-telegram",
    version: "1.0.0"
  },
  frontend: {
    framework: "astro",
    runtime: "cloudflare-pages"
  },
  baas: {
    provider: "surrealdb",
    url: "http://127.0.0.1:8000"
  },
  cockpit: {
    provider: "telegram",
    botHandle: "@MyBot"
  }
};
