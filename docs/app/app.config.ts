export default defineAppConfig({
  ui: {
    colors: {
      primary: "burzyellow",
      secondary: "burzblue",
      neutral: "slate",
      success: "green",
      info: "blue",
      warning: "amber",
      error: "red",
    },
    footer: {
      slots: {
        root: "border-t border-default",
        left: "text-sm text-muted",
      },
    },
  },
  seo: {
    siteName: "Terox",
  },
  header: {
    title: "Terox",
    to: "/",
    search: true,
    colorMode: true,
    nav: [
      { label: "Docs", to: "/getting-started/introduction" },
      { label: "CLI Reference", to: "/reference/cli" },
    ],
    links: [
      {
        icon: "i-simple-icons-github",
        to: "https://github.com/weburz/terox",
        target: "_blank",
        "aria-label": "Terox on GitHub",
      },
    ],
  },
  footer: {
    credits: `© ${new Date().getFullYear()} Weburz`,
    colorMode: false,
    links: [
      {
        icon: "i-simple-icons-github",
        to: "https://github.com/weburz/terox",
        target: "_blank",
        "aria-label": "Terox on GitHub",
      },
      {
        icon: "i-lucide-globe",
        to: "https://weburz.com",
        target: "_blank",
        "aria-label": "weburz.com",
      },
    ],
  },
  toc: {
    title: "On this page",
    bottom: {
      title: "Page",
      edit: "https://github.com/weburz/terox/edit/main/docs/content",
      links: [],
    },
  },
});
