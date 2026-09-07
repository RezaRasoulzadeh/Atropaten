<script setup lang="ts">
withDefaults(
  defineProps<{
    label: string;
    variant?: 'ghost' | 'secondary' | 'danger';
    size?: 'sm' | 'md';
    disabled?: boolean;
    type?: 'button' | 'submit';
  }>(),
  { variant: 'ghost', size: 'md', disabled: false, type: 'button' },
);
</script>
<template>
  <button
    class="btn btn-square"
    :class="[
      variant === 'danger' ? 'btn-error' : variant === 'secondary' ? 'btn-outline' : 'btn-ghost',
      size === 'sm' ? 'btn-sm' : '',
    ]"
    :type="type"
    :aria-label="label"
    :title="label"
    :disabled="disabled"
  >
    <slot />
  </button>
</template>
