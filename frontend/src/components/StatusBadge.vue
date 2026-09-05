<script setup lang="ts">
withDefaults(defineProps<{
  label: string
  tone?: 'blue' | 'green' | 'amber' | 'red' | 'slate'
}>(), {
  tone: 'slate',
})
</script>

<template>
  <span class="status-badge" :class="`badge-${tone}`">
    <span class="status-badge-dot" aria-hidden="true"></span>
    {{ label }}
  </span>
</template>
