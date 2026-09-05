<script setup lang="ts">
import AppIcon from './AppIcon.vue'

withDefaults(defineProps<{
  title: string
  value: string
  detail: string
  trend: string
  icon: string
  accent?: 'blue' | 'green' | 'amber' | 'red'
}>(), {
  accent: 'blue',
})
</script>

<template>
  <article class="kpi-card" :class="`kpi-${accent}`">
    <div class="kpi-card-header">
      <span class="kpi-label">{{ title }}</span>
      <span class="kpi-icon"><AppIcon :name="icon" :size="16" /></span>
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
  --kpi-accent: #2f6fca;
  --kpi-soft: #eaf2fd;
}

.kpi-green {
  --kpi-accent: #16845b;
  --kpi-soft: #e8f6ef;
}

.kpi-amber {
  --kpi-accent: #a46b14;
  --kpi-soft: #fff5df;
}

.kpi-red {
  --kpi-accent: #c34747;
  --kpi-soft: #fdeeee;
}
</style>
