<script setup lang="ts">
import { translateUi } from '../../i18n';
withDefaults(defineProps<{ label?: string }>(), { label: 'Loading…' });
</script>
<template>
  <div
    class="flex min-h-24 items-center justify-center gap-2 text-sm text-base-content/60"
    role="status"
  >
    <span class="loading"></span><span>{{ translateUi(label) }}</span>
  </div>
</template>
