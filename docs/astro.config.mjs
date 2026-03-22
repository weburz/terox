import { defineConfig } from "astro/config";
import starlight from "@astrojs/starlight";

export default defineConfig({
  site: "https://terox.weburz.com/",
  integrations: [
    starlight({
      title: "Terox",
      description: "Scaffold your projects through the power of automation!",
      editLink: {
        baseUrl: "https://github.com/Weburz/terox/edit/main/docs",
      },
      social: [
        {
          icon: "github",
          label: "GitHub",
          href: "https://github.com/Weburz/terox",
        },
        {
          icon: "discord",
          label: "Discord",
          href: "https://discord.gg/QeYqwyxBhR",
        },
        {
          icon: "email",
          label: "Email",
          href: "mailto:contact@weburz.com",
        },
        {
          icon: "facebook",
          label: "Facebook",
          href: "https://www.facebook.com/Weburz",
        },
        {
          icon: "instagram",
          label: "Instagram",
          href: "https://www.instagram.com/weburzit",
        },
        {
          icon: "linkedin",
          label: "LinkedIn",
          href: "https://www.linkedin.com/company/weburz",
        },
        {
          icon: "youtube",
          label: "YouTube",
          href: "https://www.youtube.com/@Weburz",
        },
        {
          icon: "x.com",
          label: "Twitter",
          href: "https://x.com/weburz",
        },
      ],
      lastUpdated: true,
      head: [
        {
          tag: "script",
          attrs: {
            async: true,
            src: "https://umami.weburz.com/script.js",
            "data-website-id": "166f0328-f723-471b-86ec-48e85f284ed8",
          },
        },
      ],
      sidebar: [
        {
          label: "Introduction",
          slug: "introduction",
        },
        {
          label: "Usage Guide",
          items: [
            { label: "Installation Guide", slug: "usage-guide/installation" },
            {
              label: "Command-Line Interface (CLI) Reference",
              slug: "usage-guide/cli-reference",
            },
          ],
        },
        {
          label: "Contribution & Development Guide",
          items: [
            {
              label: "Development Guidelines",
              slug: "development-guide/development",
            },
            {
              label: "Software Requirements Specifications (SRS)",
              slug: "development-guide/spec-sheet",
            },
          ],
        },
      ],
    }),
  ],
});
