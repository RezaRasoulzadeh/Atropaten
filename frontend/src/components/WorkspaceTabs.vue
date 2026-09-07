<script setup lang="ts">
defineProps<{
  tabs: string[]
  activeTab: string
}>()

defineEmits<{ change: [tab: string] }>()
</script>

<template>
<nav class="flex gap-1 overflow-x-auto border-b border-base-300" aria-label="Workspace tabs" role="tablist">
    <button
      v-for="tab in tabs"
      :key="tab"
      class="btn btn-sm shrink-0 rounded-b-none"
      :class="tab === activeTab ? 'btn-primary' : 'btn-ghost'"
      type="button"
      :aria-selected="tab === activeTab"
      role="tab"
      @click="$emit('change', tab)"
    >
      {{ tab }}
    </button>
  </nav>
</template>
