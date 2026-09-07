<script setup lang="ts">
defineProps<{ title: string; description?: string }>();
</script>
<template>
  <section class="min-w-0 space-y-3 border-b border-base-300 pb-4 last:border-0 last:pb-0">
    <div>
      <h3 class="text-sm font-semibold">{{ title }}</h3>
      <p v-if="description" class="mt-1 text-xs text-base-content/60">{{ description }}</p>
    </div>
    <slot />
  </section>
</template>
