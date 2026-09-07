<script setup lang="ts">
defineProps<{ title?: string; subtitle?: string }>()
</script>
<template><article class="card overflow-hidden border border-base-300 bg-base-100 shadow-none"><header v-if="title || subtitle || $slots.action" class="flex min-h-14 items-center justify-between gap-3 border-b border-base-300"><div><h2 v-if="title" class="text-sm font-semibold">{{ title }}</h2><p v-if="subtitle" class="text-xs text-base-content/60">{{ subtitle }}</p></div><div v-if="$slots.action"><slot name="action" /></div></header><div class="card-body p-0"><slot /></div></article></template>
