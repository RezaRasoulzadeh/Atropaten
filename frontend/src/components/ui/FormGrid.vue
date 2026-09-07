<script setup lang="ts">
withDefaults(defineProps<{ columns?: 1 | 2 | 3 }>(), { columns: 2 });
</script>
<template>
  <div
    class="grid min-w-0 items-start gap-3"
    :class="
      columns === 3 ? 'sm:grid-cols-2 2xl:grid-cols-3' : columns === 2 ? 'sm:grid-cols-2' : ''
    "
  >
    <slot />
  </div>
</template>
