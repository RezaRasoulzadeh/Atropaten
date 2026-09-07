<script setup lang="ts">
withDefaults(defineProps<{ flush?: boolean }>(), { flush: false })
</script>

<template>
  <div class="sticky top-0 z-30 min-w-0 border-b border-base-300 bg-base-200" :class="{ 'pb-0': flush }">
    <slot />
  </div>
</template>
