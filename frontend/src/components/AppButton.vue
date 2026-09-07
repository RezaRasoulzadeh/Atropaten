<script setup lang="ts">
withDefaults(defineProps<{ variant?: 'primary' | 'secondary' | 'ghost' | 'danger'; size?: 'sm' | 'md'; loading?: boolean; disabled?: boolean; type?: 'button' | 'submit' | 'reset' }>(), { variant: 'secondary', size: 'md', loading: false, disabled: false, type: 'button' })
</script>
<template>
  <button class="btn at-btn" :class="[`at-btn-${variant}`, `at-btn-${size}`]" :type="type" :disabled="disabled || loading">
    <span v-if="loading" class="loading loading-spinner loading-xs" aria-hidden="true"></span>
    <span class="at-button-content"><slot /></span>
  </button>
</template>
