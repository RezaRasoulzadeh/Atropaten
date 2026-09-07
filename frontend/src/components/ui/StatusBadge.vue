<script setup lang="ts">
import { computed } from 'vue';
import {
  CheckCircle2,
  CircleDot,
  Clock3,
  FilePenLine,
  TriangleAlert,
  XCircle,
} from 'lucide-vue-next';

const props = withDefaults(
  defineProps<{
    label: string;
    tone?: 'blue' | 'green' | 'amber' | 'red' | 'slate';
  }>(),
  {
    tone: 'slate',
  },
);

const icon = computed(() => {
  const value = props.label.toLowerCase();
  if (/(cancel|reject|void|failed|error)/.test(value)) return XCircle;
  if (/(overdue|urgent|attention|warning)/.test(value)) return TriangleAlert;
  if (/(pending|waiting|due|partially|paused)/.test(value)) return Clock3;
  if (/(draft|unconfirmed)/.test(value)) return FilePenLine;
  if (/(paid|posted|closed|ready|delivered|active|approved|completed|confirmed)/.test(value))
    return CheckCircle2;
  return CircleDot;
});
</script>

<template>
  <span
    class="badge badge-sm gap-1 text-xs leading-4"
    :class="
      tone === 'green'
        ? 'badge-success'
        : tone === 'red'
          ? 'badge-error'
          : tone === 'amber'
            ? 'badge-warning'
            : tone === 'blue'
              ? 'badge-info'
              : 'badge-neutral'
    "
  >
    <component :is="icon" :size="11" :stroke-width="2" aria-hidden="true" />
    {{ label }}
  </span>
</template>
