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
      size === 'sm' ? 'btn-sm h-8 min-h-8 w-8' : 'h-10 min-h-10 w-10',
    ]"
    :type="type"
    :aria-label="label"
    :title="label"
    :disabled="disabled"
  >
    <slot />
  </button>
</template>
