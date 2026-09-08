<script setup lang="ts">
defineProps<{
  eyebrow?: string;
  title: string;
  description?: string;
}>();
</script>

<template>
  <header class="workspace-header flex min-w-0 flex-wrap items-end justify-between gap-3 py-2">
    <div class="flex min-w-0 items-end gap-3">
      <div v-if="$slots.leading" class="shrink-0"><slot name="leading" /></div>
      <div class="min-w-0">
      <p class="mb-0.5 text-xs font-semibold leading-4 text-primary">{{ eyebrow }}</p>
      <div class="flex min-w-0 items-center gap-2">
        <h1 class="m-0 min-w-0 truncate text-2xl font-bold leading-8 tracking-tight">{{ title }}</h1>
        <slot name="title-suffix" />
      </div>
      <p v-if="description" class="mt-0.5 max-w-[62ch] text-xs leading-4 text-base-content/65">
        {{ description }}
      </p>
      </div>
    </div>
    <div class="workspace-header-actions flex shrink-0 flex-wrap items-end justify-end gap-2">
      <slot />
    </div>
  </header>
</template>
