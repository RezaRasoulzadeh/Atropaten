<script setup lang="ts">
defineProps<{
  tabs: string[]
  activeTab: string
}>()

defineEmits<{ change: [tab: string] }>()
</script>

<template>
  <nav class="workspace-tabs" aria-label="Order workspace tabs">
    <button
      v-for="tab in tabs"
      :key="tab"
      class="workspace-tab"
      :class="{ 'is-active': tab === activeTab }"
      type="button"
      :aria-selected="tab === activeTab"
      role="tab"
      @click="$emit('change', tab)"
    >
      {{ tab }}
    </button>
  </nav>
</template>
