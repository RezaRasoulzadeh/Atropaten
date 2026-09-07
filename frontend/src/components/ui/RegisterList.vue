<script setup lang="ts">
withDefaults(
  defineProps<{
    title: string;
    subtitle?: string;
    count?: number;
  }>(),
  {
    subtitle: undefined,
    count: undefined,
  },
);
</script>

<template>
  <section class="min-w-0 overflow-hidden rounded-box border border-base-300 bg-base-100">
    <header
      class="flex min-w-0 flex-wrap items-center justify-between gap-3 border-b border-base-300 px-4 py-3"
    >
      <div class="min-w-0">
        <h2 class="truncate text-sm font-semibold">{{ title }}</h2>
        <p v-if="subtitle" class="mt-0.5 text-xs text-base-content/60">{{ subtitle }}</p>
      </div>
      <div class="flex shrink-0 items-center gap-2">
        <span v-if="count !== undefined" class="badge badge-ghost tabular-nums">{{ count }}</span>
        <slot name="action" />
      </div>
    </header>
    <div class="divide-y divide-base-300">
      <slot />
    </div>
  </section>
</template>
