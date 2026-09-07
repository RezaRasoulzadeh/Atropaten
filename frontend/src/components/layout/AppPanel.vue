<script setup lang="ts">
withDefaults(defineProps<{ title?: string; subtitle?: string; flush?: boolean }>(), {
  flush: false,
});
</script>
<template>
  <article class="min-w-0 rounded-box border border-base-300 bg-base-100">
    <header
      v-if="title || subtitle || $slots.action"
      class="flex min-h-14 shrink-0 items-center justify-between gap-3 border-b border-base-300 px-4 py-3"
    >
      <div class="min-w-0">
        <h2 v-if="title" class="text-sm font-semibold">{{ title }}</h2>
        <p v-if="subtitle" class="text-xs text-base-content/60">{{ subtitle }}</p>
      </div>
      <div v-if="$slots.action" class="shrink-0"><slot name="action" /></div>
    </header>
    <div class="min-w-0 space-y-3" :class="flush ? 'p-0' : 'p-4'"><slot /></div>
  </article>
</template>
