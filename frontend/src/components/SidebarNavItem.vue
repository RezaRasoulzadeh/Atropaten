<script setup lang="ts">
import type { Component } from 'vue'

defineProps<{
  label: string
  icon: Component
  active?: boolean
  collapsed?: boolean
}>()

defineEmits<{ select: [] }>()
</script>

<template>
  <button
    class="btn btn-ghost w-full justify-start gap-3 text-start"
    :class="{ 'btn-primary': active, 'btn-square justify-center': collapsed }"
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
