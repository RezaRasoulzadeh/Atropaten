<script setup lang="ts">
withDefaults(defineProps<{ message?: string; tone?: 'error' | 'warning' | 'info' | 'success' }>(), {
  tone: 'error',
});
</script>
<template>
  <div
    class="alert min-w-0 border border-base-300 bg-base-100 text-sm"
    :role="tone === 'error' ? 'alert' : 'status'"
  >
    <span
      :class="tone === 'error' ? 'text-error' : tone === 'warning' ? 'text-warning' : 'text-info'"
      ><slot name="icon"
    /></span>
    <div class="min-w-0 wrap-anywhere">
      <slot>{{ message }}</slot>
    </div>
    <slot name="action" />
  </div>
</template>
