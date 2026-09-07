<script setup lang="ts">
withDefaults(defineProps<{
  label: string
  tone?: 'blue' | 'green' | 'amber' | 'red' | 'slate'
}>(), {
  tone: 'slate',
})
</script>

<template>
  <span class="badge status-badge" :class="[`badge-${tone}`, `at-status-${tone}`]">
    <span class="status-badge-dot" aria-hidden="true"></span>
    {{ label }}
  </span>
</template>
