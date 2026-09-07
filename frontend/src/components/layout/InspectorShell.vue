<script setup lang="ts">
import InspectorHeader from './InspectorHeader.vue';
defineProps<{ title: string; subtitle?: string }>();
</script>
<template>
  <aside class="min-w-0 overflow-hidden rounded-box border border-base-300 bg-base-100" :aria-label="title">
    <InspectorHeader :title="title" :subtitle="subtitle"
      ><slot name="header" /><slot name="action"
    /></InspectorHeader>
    <div class="min-w-0 space-y-4 p-4" :class="$slots.footer ? 'pb-20' : ''"><slot /></div>
    <footer
      v-if="$slots.footer"
      class="sticky bottom-0 z-10 flex shrink-0 flex-wrap items-center justify-end gap-2 rounded-b-box border-t border-base-300 bg-base-100 px-4 py-3"
    >
      <slot name="footer" />
    </footer>
  </aside>
</template>
