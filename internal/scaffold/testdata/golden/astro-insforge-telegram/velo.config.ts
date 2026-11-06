// velo.config.ts — Contrato de Infraestructura Velo
export default {
  project: {
    name: "mi-app-velo",
    slug: "astro-insforge-telegram",
    version: "1.0.0"
  },
  frontend: {
    framework: "astro",
    runtime: "cloudflare-pages"
  },
  baas: {
    provider: "insforge",
    url: ""
  },
  cockpit: {
    provider: "telegram",
    botHandle: ""
  }
};
