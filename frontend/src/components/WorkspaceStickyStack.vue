<script setup lang="ts">
withDefaults(defineProps<{ flush?: boolean }>(), { flush: false })
</script>

<template>
  <div class="workspace-sticky-stack" :class="{ 'workspace-sticky-stack--flush': flush }">
    <slot />
  </div>
</template>
