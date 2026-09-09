<script setup lang="ts">
import { formatMoneyInputWhileTyping, type CurrencyUnit } from '../../utils/currency';
defineOptions({ inheritAttrs: false });
const props = defineProps<{
  modelValue?: string | number | null;
  modelModifiers?: Record<string, boolean>;
  money?: CurrencyUnit;
}>();
const emit = defineEmits<{ 'update:modelValue': [value: string] }>();
function displayValue() {
  return props.modelValue === null || props.modelValue === undefined ? '' : String(props.modelValue);
}
function update(event: Event) {
  const value = (event.target as HTMLInputElement).value;
  emit('update:modelValue', props.money ? formatMoneyInputWhileTyping(value, props.money) : value);
}
</script>
<template>
  <input
    v-bind="$attrs"
    class="input h-10 w-full min-w-0 text-start text-sm leading-5 text-base-content"
    :value="displayValue()"
    @input="update"
  />
</template>
