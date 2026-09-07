<script setup lang="ts">
defineProps<{ title?: string; subtitle?: string }>()
</script>
<template><article class="card overflow-hidden border border-base-300 bg-base-100 shadow-none"><header v-if="title || subtitle || $slots.action" class="flex min-h-14 shrink-0 items-center justify-between gap-3 border-b border-base-300 px-4 py-3"><div class="min-w-0"><h2 v-if="title" class="text-sm font-semibold">{{ title }}</h2><p v-if="subtitle" class="text-xs text-base-content/60">{{ subtitle }}</p></div><div v-if="$slots.action" class="shrink-0"><slot name="action" /></div></header><div class="card-body min-w-0 overflow-x-auto p-0"><slot /></div></article></template>
