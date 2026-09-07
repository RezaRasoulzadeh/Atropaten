<script setup lang="ts">
withDefaults(defineProps<{
  label: string
  tone?: 'blue' | 'green' | 'amber' | 'red' | 'slate'
}>(), {
  tone: 'slate',
})
</script>

<template>
  <span class="badge badge-sm" :class="tone === 'green' ? 'badge-success' : tone === 'red' ? 'badge-error' : tone === 'amber' ? 'badge-warning' : tone === 'blue' ? 'badge-info' : 'badge-neutral'">
    {{ label }}
  </span>
</template>
