<script setup lang="ts">
import { computed } from 'vue';
import type { Component } from 'vue';

const props = withDefaults(
  defineProps<{
    title: string;
    value: string;
    detail: string;
    trend: string;
    icon: Component;
    accent?: 'blue' | 'green' | 'amber' | 'red';
    loading?: boolean;
  }>(),
  {
    accent: 'blue',
    loading: false,
  },
);

const accentClass: Record<NonNullable<typeof props.accent>, string> = {
  blue: 'text-info',
  green: 'text-success',
  amber: 'text-warning',
  red: 'text-error',
};

const auraTone = computed(() => accentClass[props.accent]);
</script>

<template>
  <div class="block w-full min-w-0">
    <article class="card h-full rounded-box border border-base-300 bg-base-100 shadow-none">
      <div class="card-body gap-2 p-4">
        <div class="flex items-center justify-between gap-3">
          <span class="text-sm font-semibold text-base-content/80">{{ title }}</span>
          <span :class="auraTone"
            ><component :is="icon" :size="16" :stroke-width="1.8" aria-hidden="true"
          /></span>
        </div>
        <p v-if="!loading" class="m-0 text-2xl font-bold">{{ value }}</p>
        <div v-else class="skeleton h-8 w-32 bg-base-300/70" aria-label="Loading value"></div>
        <div class="flex items-center justify-between gap-2 text-xs text-base-content/60">
          <span v-if="!loading">{{ detail }}</span>
          <span v-else class="skeleton h-4 w-36 bg-base-300/70" aria-hidden="true"></span>
          <span v-if="!loading" :class="auraTone">{{ trend }}</span>
          <span v-else class="skeleton h-4 w-16 bg-base-300/70" aria-hidden="true"></span>
        </div>
      </div>
    </article>
  </div>
</template>
