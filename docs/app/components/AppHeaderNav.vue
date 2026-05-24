<script setup lang="ts">
type NavItem = {
  label: string;
  to: string;
  icon?: string;
};

defineProps<{
  items: NavItem[];
  block?: boolean;
}>();
</script>

<template>
  <UButton
    v-for="(item, index) in items"
    :key="index"
    :to="item.to"
    :icon="item.icon"
    :label="item.label"
    color="neutral"
    variant="ghost"
    :class="block ? 'w-full justify-start' : ''"
  />
</template>
