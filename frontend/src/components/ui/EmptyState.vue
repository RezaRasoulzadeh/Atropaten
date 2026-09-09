<script setup lang="ts">
import { Inbox } from 'lucide-vue-next'

withDefaults(
  defineProps<{
    title: string
    description?: string
    compact?: boolean
  }>(),
  {
    description: undefined,
    compact: false,
  },
)
</script>

<template>
  <section
    class="empty-state flex min-w-0 flex-col items-center justify-center gap-3 rounded-box border border-dashed border-base-300 bg-base-100 px-6 text-center"
    :class="compact ? 'min-h-28 py-5' : 'min-h-52 py-8'"
    role="status"
    aria-live="polite"
  >
    <div class="grid size-11 shrink-0 place-items-center rounded-full bg-primary/15 text-primary">
      <slot name="icon"><Inbox :size="22" :stroke-width="1.7" aria-hidden="true" /></slot>
    </div>
    <div class="min-w-0 max-w-xl">
      <h3 class="text-sm font-semibold leading-5">{{ title }}</h3>
      <p v-if="description" class="mt-1 text-xs leading-5 text-base-content/60">{{ description }}</p>
      <div v-if="$slots.action" class="empty-state-action mt-3"><slot name="action" /></div>
    </div>
  </section>
</template>
