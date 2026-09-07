<script setup lang="ts">
defineProps<{ title: string; description?: string }>();
</script>
<template>
  <div class="flex min-w-0 min-h-28 items-center justify-center gap-3 p-5 text-sm">
    <div class="text-base-content/50"><slot name="icon" /></div>
    <div class="min-w-0">
      <h3 class="font-semibold">{{ title }}</h3>
      <p v-if="description" class="text-xs text-base-content/60">{{ description }}</p>
      <div v-if="$slots.action"><slot name="action" /></div>
    </div>
  </div>
</template>
