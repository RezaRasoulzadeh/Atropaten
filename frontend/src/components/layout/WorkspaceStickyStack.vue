<script setup lang="ts">
defineProps<{ flush?: boolean }>();
</script>
<template>
  <div
    class="workspace-sticky-stack sticky top-0 z-10 min-w-0 space-y-2 bg-base-200 pt-2"
    :class="flush ? '' : 'pb-3'"
  >
    <slot />
  </div>
</template>
