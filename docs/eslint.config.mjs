// @ts-check
import eslintConfigPrettier from "eslint-config-prettier";
import withNuxt from "./.nuxt/eslint.config.mjs";

export default withNuxt()
  .prepend({
    ignores: [
      "**/.nuxt/**/*",
      "**/.output/**/*",
      "**/.data/**/*",
      "**/node_modules/**/*",
      "**/dist/**/*",
      "content/**/*",
    ],
  })
  .override("nuxt/vue/rules", {
    rules: {
      "vue/multi-word-component-names": "off",
      "vue/html-self-closing": [
        "warn",
        {
          html: {
            void: "always",
          },
        },
      ],
    },
  })
  .append(eslintConfigPrettier);
