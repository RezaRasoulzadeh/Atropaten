<script setup lang="ts">
import AppPanel from './AppPanel.vue'
defineProps<{ title: string; subtitle?: string }>()
</script>

<template><AppPanel :title="title" :subtitle="subtitle"><template #action><slot name="action" /></template><slot /></AppPanel></template>
