<script setup lang="ts">
const route = useRoute();

definePageMeta({
  layout: "docs",
});

const { data: page } = await useAsyncData(route.path, () =>
  queryCollection("docs").path(route.path).first(),
);

if (!page.value) {
  throw createError({
    statusCode: 404,
    statusMessage: "Page not found",
    fatal: true,
  });
}

const { toc } = useAppConfig();

useSeoMeta({
  title: page.value.title,
  description: page.value.description,
});
</script>

<template>
  <UPage v-if="page">
    <UPageBody>
      <UPageHeader
        :title="page.title"
        :description="page.description"
        class="mb-8"
      />
      <ContentRenderer :value="page" />
    </UPageBody>

    <template #right>
      <UContentToc
        v-if="page.body?.toc?.links?.length"
        :title="toc.title"
        :links="page.body.toc.links"
      />
    </template>
  </UPage>
</template>
