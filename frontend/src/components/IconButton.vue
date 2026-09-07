<script setup lang="ts">
withDefaults(defineProps<{ label: string; variant?: 'ghost' | 'secondary' | 'danger'; size?: 'sm' | 'md'; disabled?: boolean; type?: 'button' | 'submit' }>(), { variant: 'ghost', size: 'md', disabled: false, type: 'button' })
</script>
<template>
  <button class="btn btn-square at-icon-button" :class="[`at-icon-${variant}`, `at-icon-${size}`]" :type="type" :aria-label="label" :title="label" :disabled="disabled"><slot /></button>
</template>
