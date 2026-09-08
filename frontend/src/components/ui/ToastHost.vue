<script setup lang="ts">
import { CheckCircle2, CircleX, Info, TriangleAlert, X } from 'lucide-vue-next';
import { dismiss, useToast, type ToastKind } from '../../ui/feedback';
const toastItems = useToast().items;
const icons: Record<ToastKind, typeof CheckCircle2> = {
  success: CheckCircle2,
  error: CircleX,
  warning: TriangleAlert,
  info: Info,
};
const iconFor = (kind: ToastKind) => icons[kind];
</script>
<template>
  <div
    class="toast toast-end toast-bottom right-3 bottom-3 z-50 w-[min(24rem,calc(100vw-1.5rem))]"
    aria-live="polite"
    aria-atomic="false"
  >
    <TransitionGroup name="toast"
      ><div
        v-for="item in toastItems"
        :key="item.id"
        class="toast-card"
        :class="`toast-card-${item.kind}`"
        :role="item.kind === 'error' ? 'alert' : 'status'"
      >
        <div class="flex min-w-0 items-center gap-3 p-3">
          <span class="toast-icon shrink-0" aria-hidden="true">
            <component :is="iconFor(item.kind)" :size="16" />
          </span>
          <div class="flex min-w-0 flex-1 flex-col gap-0.5">
            <strong v-if="item.title" class="text-xs font-semibold leading-4">{{ item.title }}</strong
            ><span class="text-xs leading-5">{{ item.message }}</span>
          </div>
          <button
            class="toast-close btn btn-ghost btn-xs btn-square -me-1 -mt-1 shrink-0"
            type="button"
            aria-label="Dismiss notification"
            @click="dismiss(item.id)"
          >
            <X :size="15" />
          </button>
        </div>
      </div
    ></TransitionGroup>
  </div>
</template>
