// velo.config.ts — Contrato de Infraestructura Velo
export default {
  project: {
    name: "mi-app-vite",
    slug: "react-vite-surrealdb-telegram",
    version: "1.0.0"
  },
  frontend: {
    framework: "react-vite",
    runtime: "cloudflare-pages"
  },
  baas: {
    provider: "surrealdb",
    url: "http://127.0.0.1:8000"
  },
  cockpit: {
    provider: "telegram",
    botHandle: "@ViteBot"
  }
};
