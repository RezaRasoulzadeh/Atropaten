<script setup lang="ts">
import AppPanel from './AppPanel.vue'
defineProps<{ title: string; subtitle?: string; flush?: boolean }>()
</script>

<template><AppPanel :title="title" :subtitle="subtitle" :flush="flush"><template #action><slot name="action" /></template><slot /></AppPanel></template>
