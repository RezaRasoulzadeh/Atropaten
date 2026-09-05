<script setup lang="ts">
import AppIcon from './AppIcon.vue'

defineProps<{
  label: string
  icon: string
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
    <span class="nav-item-icon"><AppIcon :name="icon" :size="17" /></span>
    <span v-if="!collapsed" class="nav-item-label">{{ label }}</span>
  </button>
</template>
