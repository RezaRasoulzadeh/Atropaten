<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(defineProps<{ name: string; size?: number; strokeWidth?: number }>(), {
  size: 18,
  strokeWidth: 1.8,
})

const iconPaths: Record<string, string[]> = {
  dashboard: ['M4 4h6v6H4z', 'M14 4h6v6h-6z', 'M4 14h6v6H4z', 'M14 14h6v6h-6z'],
  orders: ['M6 3h12v18H6z', 'M9 7h6', 'M9 11h6', 'M9 15h4'],
  production: ['M4 7h16v13H4z', 'M8 7V4h8v3', 'M4 12h16', 'M10 12v3h4v-3'],
  customers: ['M16 20v-1.5a3.5 3.5 0 0 0-3.5-3.5h-5A3.5 3.5 0 0 0 4 18.5V20', 'M10 11a3.5 3.5 0 1 0 0-7 3.5 3.5 0 0 0 0 7Z', 'M17 11a3 3 0 0 0 0-6'],
  services: ['M5 5h14v14H5z', 'M8 9h8', 'M8 13h5', 'M8 16h3'],
  materials: ['M4 7h16v13H4z', 'M4 7l3-4h10l3 4', 'M9 12h6'],
  purchases: ['M5 4h14v16H5z', 'M8 8h8', 'M8 12h8', 'M8 16h5'],
  suppliers: ['M3 19h18', 'M5 19V9l7-5 7 5v10', 'M9 19v-5h6v5', 'M8 10h.01', 'M12 10h.01', 'M16 10h.01'],
  accounting: ['M4 5h16v14H4z', 'M8 9h8', 'M8 13h3', 'M14 13h2', 'M8 16h8'],
  checks: ['M4 6h16v12H4z', 'M4 10h16', 'M8 15h4'],
  loans: ['M12 3v18', 'M17 7c0-2-2-3-5-3S7 5 7 7s1.5 3 5 4 5 2 5 4-2 3-5 3-5-1-5-3'],
  owners: ['M4 20v-1.5a3.5 3.5 0 0 1 3.5-3.5h9a3.5 3.5 0 0 1 3.5 3.5V20', 'M12 11a3.5 3.5 0 1 0 0-7 3.5 3.5 0 0 0 0 7Z'],
  reports: ['M5 20V10', 'M12 20V4', 'M19 20v-7'],
  settings: ['M12 15.5a3.5 3.5 0 1 0 0-7 3.5 3.5 0 0 0 0 7Z', 'M19.4 15a1.7 1.7 0 0 0 .34 1.88l.06.06-1.7 1.7-.06-.06a1.7 1.7 0 0 0-1.88-.34 1.7 1.7 0 0 0-1.03 1.56V20h-2.4v-.2a1.7 1.7 0 0 0-1.03-1.56 1.7 1.7 0 0 0-1.88.34l-.06.06-1.7-1.7.06-.06A1.7 1.7 0 0 0 8.5 15a1.7 1.7 0 0 0-1.56-1.03H6.7v-2.4h.24A1.7 1.7 0 0 0 8.5 10a1.7 1.7 0 0 0-.34-1.88L8.1 8.06l1.7-1.7.06.06a1.7 1.7 0 0 0 1.88.34 1.7 1.7 0 0 0 1.03-1.56V5h2.4v.2a1.7 1.7 0 0 0 1.03 1.56 1.7 1.7 0 0 0 1.88-.34l.06-.06 1.7 1.7-.06.06A1.7 1.7 0 0 0 19.4 10a1.7 1.7 0 0 0 1.56 1.03h.24v2.4h-.24A1.7 1.7 0 0 0 19.4 15Z'],
  search: ['m21 21-4.35-4.35', 'M10.5 18a7.5 7.5 0 1 1 0-15 7.5 7.5 0 0 1 0 15Z'],
  menu: ['M4 7h16', 'M4 12h16', 'M4 17h16'],
  bell: ['M18 8a6 6 0 0 0-12 0c0 7-3 7-3 9h18c0-2-3-2-3-9', 'M10 21h4'],
  sales: ['M4 19V5', 'M4 19h16', 'M7 15l3-4 3 2 5-6'],
  profit: ['M4 18h16', 'M6 15l3-3 3 2 5-6', 'M14 8h3v3'],
  receivables: ['M4 5h16v14H4z', 'M8 10h8', 'M8 14h4'],
  payables: ['M4 5h16v14H4z', 'M8 10h8', 'M8 14h2'],
  receipt: ['M6 3h12v18l-3-2-3 2-3-2-3 2z', 'M9 8h6', 'M9 12h6'],
  'arrow-down': ['M12 4v14', 'm6 12 6 6 6-6'],
}

const paths = computed(() => iconPaths[props.name] ?? iconPaths.dashboard)
</script>

<template>
  <svg :width="size" :height="size" viewBox="0 0 24 24" fill="none" stroke="currentColor" :stroke-width="strokeWidth" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
    <path v-for="path in paths" :key="path" :d="path" />
  </svg>
</template>
