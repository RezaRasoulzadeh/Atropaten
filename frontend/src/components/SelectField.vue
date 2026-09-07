<script setup lang="ts">
type SelectOption = {
  label: string
  value: string
}

const props = withDefaults(defineProps<{
  modelValue: string
  options: readonly SelectOption[]
  label?: string
  ariaLabel?: string
}>(), {
  label: '',
  ariaLabel: '',
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()
</script>

<template>
  <div class="form-control min-w-0 gap-1">
    <span v-if="label" class="text-xs text-base-content/60">{{ label }}</span>
    <select
      class="select select-bordered h-10 min-h-10 w-full text-start font-normal"
      :value="modelValue"
      :disabled="!options.length"
      :aria-label="ariaLabel || label || 'Select an option'"
      @change="emit('update:modelValue', ($event.target as HTMLSelectElement).value)"
    >
      <option v-if="!options.length" value="">No options available</option>
      <option v-for="option in options" :key="option.value" :value="option.value">{{ option.label }}</option>
    </select>
  </div>
</template>
