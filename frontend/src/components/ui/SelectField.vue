<script setup lang="ts" generic="T extends string">
import { useId } from 'vue';
defineOptions({ inheritAttrs: false });
defineProps<{
  modelValue: T;
  options: readonly { label: string; value: T; disabled?: boolean }[];
  label?: string;
  ariaLabel?: string;
  disabled?: boolean;
}>();
const emit = defineEmits<{ 'update:modelValue': [value: T] }>();
const id = useId();
</script>
<template>
  <div class="flex min-w-0 flex-col gap-1">
    <label
      v-if="label"
      :for="String($attrs.id || id)"
      class="text-xs leading-4 text-base-content/65"
      >{{ label }}</label
    >
    <select
      v-bind="$attrs"
      :id="String($attrs.id || id)"
      class="select h-10 min-h-10 w-full min-w-0 text-start text-sm font-normal leading-5"
      :value="modelValue"
      :disabled="disabled || !options.length"
      :aria-label="ariaLabel || label || 'Select an option'"
      @change="emit('update:modelValue', ($event.target as HTMLSelectElement).value as T)"
    >
      <option v-if="!options.length" value="">No options available</option>
      <option
        v-for="option in options"
        :key="option.value"
        :value="option.value"
        :disabled="option.disabled"
      >
        {{ option.label }}
      </option>
    </select>
  </div>
</template>
