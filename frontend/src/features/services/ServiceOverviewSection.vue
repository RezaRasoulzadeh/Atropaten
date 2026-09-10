<script setup lang="ts">
defineProps<{
  title: string
  description?: string
}>()
</script>

<template>
  <section class="service-overview-section min-w-0 rounded-box border border-base-300/80 bg-transparent p-3">
    <header class="mb-3 flex min-w-0 items-start justify-between gap-3 border-b border-base-300/75 pb-2.5">
      <div class="min-w-0">
        <h3 class="text-xs font-semibold uppercase tracking-[0.08em] text-base-content/75">{{ title }}</h3>
        <p v-if="description" class="mt-1 text-[0.6875rem] leading-4 text-base-content/55">{{ description }}</p>
      </div>
      <slot name="meta" />
    </header>
    <div class="min-w-0">
      <slot />
    </div>
  </section>
</template>
