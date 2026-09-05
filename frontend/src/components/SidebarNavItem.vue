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
    class="nav-item"
    :class="{ 'is-active': active, 'is-collapsed': collapsed }"
    type="button"
    :title="collapsed ? label : undefined"
    :aria-label="collapsed ? label : undefined"
    :aria-current="active ? 'page' : undefined"
    @click="$emit('select')"
  >
    <span class="nav-item-icon"><component :is="icon" :size="17" :stroke-width="1.8" aria-hidden="true" /></span>
    <span v-if="!collapsed" class="nav-item-label">{{ label }}</span>
  </button>
</template>
