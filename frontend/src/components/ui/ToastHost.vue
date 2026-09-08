<script setup lang="ts">
import { useToast } from '../../ui/feedback';
const toastItems = useToast().items;
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
        <div class="flex min-w-0 items-center justify-center p-3 text-center">
          <span class="text-xs leading-5">{{ item.message }}</span>
        </div>
      </div
    ></TransitionGroup>
  </div>
</template>
