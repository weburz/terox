<script setup lang="ts">
definePageMeta({
  layout: "default",
});

useSeoMeta({
  title: "Terox · Scaffold projects with a single Go binary",
  description:
    "A small, cross-platform CLI for scaffolding projects from local or GitHub templates.",
});
</script>

<template>
  <UContainer class="py-16 sm:py-24">
    <UPageHero
      title="Scaffold projects with a single Go binary"
      description="Terox renders templates from local folders or GitHub repositories into ready-to-edit projects. No language runtime required on the target machine."
      orientation="vertical"
      class="pb-12"
    >
      <template #links>
        <UButton
          to="/getting-started/introduction"
          size="xl"
          color="primary"
          icon="i-lucide-arrow-right"
          trailing
        >
          Get started
        </UButton>
        <UButton
          to="https://github.com/weburz/terox"
          target="_blank"
          size="xl"
          color="neutral"
          variant="outline"
          icon="i-simple-icons-github"
        >
          View on GitHub
        </UButton>
      </template>
    </UPageHero>

    <UPageGrid class="gap-6">
      <UPageFeature
        title="Single binary"
        description="Distribute a static Go binary for Linux, macOS and Windows. No Python, no Node, no runtime to install on the target machine."
        icon="i-lucide-package"
      />
      <UPageFeature
        title="Local or remote templates"
        description="Scaffold from a folder on disk or a public GitHub repository. Cached locally so subsequent runs are instant."
        icon="i-lucide-globe"
      />
      <UPageFeature
        title="Interactive prompts"
        description="A terox.json manifest declares variables, choices and defaults. Terox prompts the user and renders the result."
        icon="i-lucide-message-square"
      />
      <UPageFeature
        title="CI friendly"
        description="Provide answers with repeatable --set flags and run with --non-interactive for scripts and pipelines."
        icon="i-lucide-bot"
      />
      <UPageFeature
        title="Safe for binary files"
        description="File contents are sniffed and binary assets are copied byte-for-byte, never run through the templating engine."
        icon="i-lucide-image"
      />
      <UPageFeature
        title="Inspired by good neighbours"
        description="Stands on the shoulders of Cookiecutter, copier, and hay-kot/scaffold. Picks a different shape, not a fight."
        icon="i-lucide-heart-handshake"
      />
    </UPageGrid>
  </UContainer>
</template>
