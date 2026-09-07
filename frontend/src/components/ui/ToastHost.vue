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
    class="toast toast-end toast-bottom bottom-24 z-50 w-[min(24rem,calc(100vw-2rem))]"
    aria-live="polite"
    aria-atomic="false"
  >
    <TransitionGroup name="toast"
      ><div
        v-for="item in toastItems"
        :key="item.id"
        class="alert"
        :class="`alert-${item.kind}`"
        :role="item.kind === 'error' ? 'alert' : 'status'"
      >
        <component :is="iconFor(item.kind)" :size="17" aria-hidden="true" />
        <div class="flex min-w-0 flex-1 flex-col gap-0.5">
          <strong v-if="item.title" class="text-xs">{{ item.title }}</strong
          ><span class="text-xs">{{ item.message }}</span>
        </div>
        <button
          class="btn btn-ghost btn-xs btn-square"
          type="button"
          aria-label="Dismiss notification"
          @click="dismiss(item.id)"
        >
          <X :size="15" />
        </button></div
    ></TransitionGroup>
  </div>
</template>
