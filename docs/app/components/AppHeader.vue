<script setup lang="ts">
const { header } = useAppConfig();
</script>

<template>
  <UHeader
    :ui="{ center: 'hidden lg:flex lg:flex-1 lg:justify-center gap-1' }"
    :to="header?.to || '/'"
  >
    <template #title>
      <span class="flex items-center gap-2">
        <AppLogo class="w-auto h-6 shrink-0" />
        <span v-if="header?.title" class="font-semibold text-default">
          {{ header.title }}
        </span>
      </span>
    </template>

    <template v-if="header?.nav?.length" #default>
      <AppHeaderNav :items="header.nav" />
    </template>

    <template #right>
      <UContentSearchButton v-if="header?.search" />
      <UColorModeButton v-if="header?.colorMode" />
      <template v-if="header?.links">
        <UButton
          v-for="(link, index) of header.links"
          :key="index"
          v-bind="{ color: 'neutral', variant: 'ghost', ...link }"
        />
      </template>
    </template>

    <template #body>
      <nav class="flex flex-col gap-1">
        <AppHeaderNav :items="header?.nav || []" block />
      </nav>
    </template>
  </UHeader>
</template>
