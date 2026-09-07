<script setup lang="ts">
defineProps<{ wide?: boolean }>();
</script>
<template>
  <div
    class="master-detail grid min-w-0 items-start gap-4"
    :class="
      wide
        ? 'xl:grid-cols-[minmax(0,1fr)_30rem]'
        : 'xl:grid-cols-[minmax(0,1fr)_24rem]'
    "
  >
    <slot />
  </div>
</template>
