<script setup lang="ts">
defineProps<{ title: string; subtitle?: string }>();
</script>
<template>
  <header class="flex min-w-0 items-start justify-between gap-3 border-b border-base-300 px-4 py-3">
    <div class="min-w-0">
      <h2 class="text-sm font-semibold leading-5 wrap-anywhere">{{ title }}</h2>
      <p v-if="subtitle" class="mt-1 text-xs leading-4 text-base-content/60 wrap-anywhere">
        {{ subtitle }}
      </p>
    </div>
    <div v-if="$slots.default" class="flex shrink-0 items-center gap-2"><slot /></div>
  </header>
</template>
