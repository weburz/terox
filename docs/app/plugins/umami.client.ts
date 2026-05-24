export default defineNuxtPlugin(() => {
  const { umamiWebsiteId } = useRuntimeConfig().public;

  if (!umamiWebsiteId) return;

  useHead({
    script: [
      {
        defer: true,
        src: "https://umami.weburz.com/script.js",
        "data-website-id": umamiWebsiteId,
      },
    ],
  });
});
