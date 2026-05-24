<script setup lang="ts">
import type { ContentNavigationItem } from "@nuxt/content";

const navigation = inject<Ref<ContentNavigationItem[] | null>>("navigation");
</script>

<template>
  <UContainer class="py-12 sm:py-16">
    <UPage>
      <template #left>
        <UPageAside>
          <UContentNavigation highlight :navigation="navigation ?? undefined" />
        </UPageAside>
      </template>

      <slot />
    </UPage>
  </UContainer>
</template>
