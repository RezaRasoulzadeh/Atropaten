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
  <article class="kpi-card" :class="`kpi-${accent}`">
    <div class="kpi-card-header">
      <span class="kpi-label">{{ title }}</span>
      <span class="kpi-icon"><component :is="icon" :size="16" :stroke-width="1.8" aria-hidden="true" /></span>
    </div>
    <p class="kpi-value">{{ value }}</p>
    <div class="kpi-footer">
      <span>{{ detail }}</span>
      <span class="kpi-trend">{{ trend }}</span>
    </div>
  </article>
</template>

<style scoped>
.kpi-blue {
  --kpi-accent: var(--accent);
  --kpi-soft: var(--accent-soft);
}

.kpi-green {
  --kpi-accent: var(--success);
  --kpi-soft: var(--success-soft);
}

.kpi-amber {
  --kpi-accent: var(--warning);
  --kpi-soft: var(--warning-soft);
}

.kpi-red {
  --kpi-accent: var(--danger);
  --kpi-soft: var(--danger-soft);
}
</style>
