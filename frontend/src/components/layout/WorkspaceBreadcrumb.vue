<script setup lang="ts">
import { ChevronRight } from 'lucide-vue-next'

export type WorkspaceBreadcrumbItem = {
  label: string
  current?: boolean
  href?: string
}

defineProps<{
  items: readonly WorkspaceBreadcrumbItem[]
}>()
const emit = defineEmits<{ navigate: [index: number] }>()
</script>

<template>
  <nav class="workspace-breadcrumb flex min-w-0 items-center gap-1.5 text-sm" aria-label="Breadcrumb">
    <template v-for="(item, index) in items" :key="`${item.label}-${index}`">
      <ChevronRight v-if="index > 0" :size="15" class="shrink-0 text-base-content/45" aria-hidden="true" />
      <span v-if="item.current || index === items.length - 1" class="min-w-0 truncate text-base-content">{{ item.label }}</span>
      <a v-else class="workspace-breadcrumb-link min-w-0 truncate" :href="item.href || '#'" @click.prevent="emit('navigate', index)">{{ item.label }}</a>
    </template>
  </nav>
</template>
