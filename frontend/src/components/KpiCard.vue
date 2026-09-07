<script setup lang="ts">
import { computed } from 'vue'
import type { Component } from 'vue'

const props = withDefaults(defineProps<{
  title: string
  value: string
  detail: string
  trend: string
  icon: Component
  accent?: 'blue' | 'green' | 'amber' | 'red'
}>(), {
  accent: 'blue',
})

const accentClass: Record<NonNullable<typeof props.accent>, string> = {
  blue: 'text-info',
  green: 'text-success',
  amber: 'text-warning',
  red: 'text-error',
}

const auraTone = computed(() => accentClass[props.accent])
</script>

<template>
  <div class="aura aura-glow aura-sm block w-full" :class="auraTone">
    <article class="card h-full rounded-box bg-base-100 shadow-none">
      <div class="card-body gap-2 p-4">
        <div class="flex items-center justify-between gap-3">
          <span class="text-sm font-semibold text-base-content/80">{{ title }}</span>
          <span :class="auraTone"><component :is="icon" :size="16" :stroke-width="1.8" aria-hidden="true" /></span>
        </div>
        <p class="m-0 text-2xl font-bold">{{ value }}</p>
        <div class="flex items-center justify-between gap-2 text-xs text-base-content/60">
          <span>{{ detail }}</span>
          <span :class="auraTone">{{ trend }}</span>
        </div>
      </div>
    </article>
  </div>
</template>
