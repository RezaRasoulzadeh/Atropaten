<script setup lang="ts">
import { translateUi } from '../../i18n';
defineProps<{
  tabs: string[];
  activeTab: string;
}>();

defineEmits<{ change: [tab: string] }>();
</script>

<template>
  <nav
    class="flex gap-1 overflow-x-auto border-b border-base-300"
    :aria-label='$t("Workspace tabs")'
    role="tablist"
  >
    <button
      v-for="tab in tabs"
      :key="tab"
      class="btn btn-sm shrink-0 rounded-b-none text-sm leading-4"
      :class="tab === activeTab ? 'btn-primary' : 'btn-ghost'"
      type="button"
      :aria-selected="tab === activeTab"
      role="tab"
      @click="$emit('change', tab)"
    >
      {{ translateUi(tab) }}
    </button>
  </nav>
</template>
