import { defineCollection, defineContentConfig, z } from "@nuxt/content";

export default defineContentConfig({
  collections: {
    docs: defineCollection({
      type: "page",
      source: {
        include: "**/*.md",
        exclude: ["index.md"],
      },
      schema: z.object({
        title: z.string(),
        description: z.string().optional(),
      }),
    }),
  },
});
