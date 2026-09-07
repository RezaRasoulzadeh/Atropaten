<script setup lang="ts">
defineProps<{
  title: string;
  description?: string;
}>();
</script>

<template>
  <section class="min-w-0 border-b border-base-300 pb-4 last:border-0 last:pb-0">
    <header class="mb-3">
      <h3 class="text-sm font-semibold">{{ title }}</h3>
      <p v-if="description" class="mt-0.5 text-xs text-base-content/60">{{ description }}</p>
    </header>
    <div class="grid gap-3 sm:grid-cols-2">
      <slot />
    </div>
  </section>
</template>
