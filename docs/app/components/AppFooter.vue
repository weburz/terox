<script setup lang="ts">
const { footer } = useAppConfig();
</script>

<template>
  <UFooter>
    <template #left>
      <div class="flex items-center gap-3 text-sm text-muted">
        <span>{{ footer.credits }}</span>
      </div>
    </template>

    <template #right>
      <UColorModeButton v-if="footer?.colorMode" />

      <UButton
        v-for="(link, index) of footer?.links"
        :key="index"
        v-bind="{ color: 'neutral', variant: 'ghost', ...link }"
      />
    </template>
  </UFooter>
</template>
