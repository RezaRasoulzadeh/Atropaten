<script setup lang="ts">
withDefaults(defineProps<{ label?: string }>(), { label: 'Loading…' })
</script>
<template><div class="at-loading-state" role="status"><span class="loading loading-spinner loading-sm"></span><span>{{ label }}</span></div></template>
