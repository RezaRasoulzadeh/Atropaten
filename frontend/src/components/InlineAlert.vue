<script setup lang="ts">
withDefaults(defineProps<{ message: string; tone?: 'error' | 'warning' | 'info' | 'success' }>(), { tone: 'error' })
</script>
<template><div class="alert at-inline-alert" :class="`at-alert-${tone}`" role="alert"><span><slot name="icon" /></span><span>{{ message }}</span><button v-if="$slots.action" class="at-alert-action"><slot name="action" /></button></div></template>
