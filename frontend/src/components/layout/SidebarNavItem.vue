<script setup lang="ts">
import type { Component } from 'vue';

defineProps<{
  label: string;
  icon: Component;
  active?: boolean;
  collapsed?: boolean;
}>();

defineEmits<{ select: [] }>();
</script>

<template>
  <button
    class="btn btn-ghost relative min-h-10 gap-3 text-start text-sm"
    :class="[
      active ? 'btn-primary' : '',
      collapsed ? 'btn-square justify-center px-0' : 'w-full justify-start px-3',
    ]"
    type="button"
    :title="collapsed ? label : undefined"
    :aria-label="collapsed ? label : undefined"
    :aria-current="active ? 'page' : undefined"
    @click="$emit('select')"
  >
    <component :is="icon" :size="17" :stroke-width="1.8" aria-hidden="true" />
    <span v-if="!collapsed">{{ label }}</span>
  </button>
</template>
