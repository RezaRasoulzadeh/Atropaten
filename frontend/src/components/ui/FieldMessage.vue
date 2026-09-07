<script setup lang="ts">
withDefaults(defineProps<{ message?: string; tone?: 'error' | 'help' }>(), { tone: 'help' });
</script>
<template>
  <p
    v-if="message"
    class="text-xs leading-4"
    :class="tone === 'error' ? 'text-error' : 'text-base-content/60'"
  >
    {{ message }}
  </p>
</template>
