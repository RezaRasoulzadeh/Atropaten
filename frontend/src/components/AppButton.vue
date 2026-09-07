<script setup lang="ts">
import { computed } from 'vue'

type ButtonVariant = 'primary' | 'secondary' | 'outline' | 'ghost' | 'danger'
type ButtonSize = 'sm' | 'md'

const props = withDefaults(defineProps<{
  variant?: ButtonVariant
  size?: ButtonSize
  loading?: boolean
  disabled?: boolean
  type?: 'button' | 'submit' | 'reset'
}>(), {
  variant: 'secondary',
  size: 'md',
  loading: false,
  disabled: false,
  type: 'button',
})

const variantClass = computed(() => ({
  primary: 'btn-primary',
  secondary: 'btn-outline btn-primary',
  outline: 'btn-outline btn-primary',
  ghost: 'btn-ghost',
  danger: 'btn-error',
}[props.variant]))

const sizeClass = computed(() => props.size === 'sm' ? 'btn-sm' : '')
</script>

<template>
  <button
    class="btn"
    :class="[variantClass, sizeClass]"
    :type="type"
    :disabled="disabled || loading"
    :aria-busy="loading || undefined"
  >
    <span class="inline-flex items-center justify-center gap-2">
      <span v-if="loading" class="loading loading-spinner loading-xs" aria-hidden="true"></span>
      <span v-if="$slots.icon" class="inline-flex shrink-0 items-center leading-none">
        <slot name="icon" />
      </span>
      <span class="relative top-0.5 inline-flex items-center leading-5">
        <slot />
      </span>
    </span>
  </button>
</template>
