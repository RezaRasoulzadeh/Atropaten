<script setup lang="ts">
withDefaults(defineProps<{ message?: string; tone?: 'error' | 'help' }>(), { tone: 'help' })
</script>
<template><p v-if="message" class="at-field-message" :class="tone === 'error' ? 'at-field-error' : 'at-field-help'">{{ message }}</p></template>
