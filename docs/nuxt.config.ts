export default defineNuxtConfig({
  modules: ["@nuxt/eslint", "@nuxt/ui", "@nuxt/content"],

  devtools: {
    enabled: true,
  },

  css: ["~/assets/css/main.css"],

  site: {
    url: "https://terox.weburz.com",
    name: "Terox",
  },

  ui: {
    theme: {
      colors: [
        "primary",
        "secondary",
        "neutral",
        "success",
        "info",
        "warning",
        "error",
      ],
    },
  },

  content: {
    build: {
      markdown: {
        toc: {
          searchDepth: 1,
        },
      },
    },
    experimental: {
      sqliteConnector: "native",
    },
  },

  runtimeConfig: {
    public: {
      umamiWebsiteId: "",
    },
  },

  compatibilityDate: "2026-05-23",

  nitro: {
    prerender: {
      crawlLinks: true,
      routes: ["/"],
      autoSubfolderIndex: false,
    },
  },

  icon: {
    provider: "iconify",
  },

  app: {
    head: {
      link: [{ rel: "icon", type: "image/svg+xml", href: "/favicon.svg" }],
    },
  },
});
