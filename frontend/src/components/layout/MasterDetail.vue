<script setup lang="ts">
defineProps<{ wide?: boolean }>();
</script>
<template>
  <div
    class="grid min-w-0 items-start gap-4"
    :class="
      wide
        ? 'min-[1440px]:grid-cols-[minmax(0,1fr)_30rem]'
        : 'min-[1440px]:grid-cols-[minmax(0,1fr)_24rem]'
    "
  >
    <slot />
  </div>
</template>
