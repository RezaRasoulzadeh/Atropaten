<script setup lang="ts">
defineProps<{ title?: string; subtitle?: string }>()
</script>
<template><article class="panel at-panel"><header v-if="title || subtitle || $slots.action" class="panel-heading at-panel-heading"><div><h2 v-if="title" class="panel-title">{{ title }}</h2><p v-if="subtitle" class="panel-subtitle">{{ subtitle }}</p></div><div v-if="$slots.action"><slot name="action" /></div></header><div class="panel-body"><slot /></div></article></template>
