<script setup lang="ts">
import { translateUi } from '../../i18n';
withDefaults(defineProps<{ message?: string; tone?: 'error' | 'help' }>(), { tone: 'help' });
</script>
<template>
  <p
    v-if="message"
    class="text-xs leading-4"
    :class="tone === 'error' ? 'text-error' : 'text-base-content/60'"
  >
    {{ translateUi(message) }}
  </p>
</template>
