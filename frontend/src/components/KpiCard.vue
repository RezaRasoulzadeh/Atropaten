<script setup lang="ts">
import type { Component } from 'vue'

withDefaults(defineProps<{
  title: string
  value: string
  detail: string
  trend: string
  icon: Component
  accent?: 'blue' | 'green' | 'amber' | 'red'
}>(), {
  accent: 'blue',
})
</script>

<template>
  <article class="card border border-base-300 bg-base-100 shadow-none">
    <div class="card-body gap-2 p-4">
      <div class="flex items-center justify-between gap-3">
        <span class="text-sm font-semibold text-base-content/80">{{ title }}</span>
        <span class="text-primary"><component :is="icon" :size="16" :stroke-width="1.8" aria-hidden="true" /></span>
      </div>
      <p class="m-0 text-2xl font-bold">{{ value }}</p>
      <div class="flex items-center justify-between gap-2 text-xs text-base-content/60">
      <span>{{ detail }}</span>
      <span class="text-primary">{{ trend }}</span>
      </div>
    </div>
  </article>
</template>
