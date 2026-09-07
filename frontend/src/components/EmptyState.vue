<script setup lang="ts">
defineProps<{ title: string; description?: string }>()
</script>
<template><div class="at-empty-state"><div class="at-empty-icon"><slot name="icon" /></div><div><h3>{{ title }}</h3><p v-if="description">{{ description }}</p><div v-if="$slots.action" class="at-empty-action"><slot name="action" /></div></div></div></template>
