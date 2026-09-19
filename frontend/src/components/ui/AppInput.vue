<script setup lang="ts">
import { computed, useAttrs } from 'vue';
import { formatMoneyInputWhileTyping, type CurrencyUnit } from '../../utils/currency';
import { localizeDigits, normalizeDigits } from '../../utils/number';
defineOptions({ inheritAttrs: false });
const attrs = useAttrs();
const props = defineProps<{
  modelValue?: string | number | null;
  modelModifiers?: Record<string, boolean>;
  money?: CurrencyUnit;
}>();
const emit = defineEmits<{ 'update:modelValue': [value: string] }>();
const numericInput = computed(() => Boolean(props.money || attrs.type === 'number' || attrs.inputmode === 'numeric' || attrs.inputmode === 'decimal'));
const inputType = computed(() => typeof attrs.type === 'string' ? (attrs.type === 'number' ? 'text' : attrs.type) : undefined);
function displayValue() {
  if (props.modelValue === null || props.modelValue === undefined) return '';
  const value = String(props.modelValue);
  return numericInput.value ? localizeDigits(value) : value;
}
function update(event: Event) {
  const value = (event.target as HTMLInputElement).value;
  const normalized = numericInput.value ? normalizeDigits(value) : value;
  emit('update:modelValue', props.money ? formatMoneyInputWhileTyping(normalized, props.money) : normalized);
}
</script>
<template>
  <input
    v-bind="$attrs"
    :type="inputType"
    class="input h-10 w-full min-w-0 text-start text-sm leading-5 text-base-content"
    :value="displayValue()"
    @input="update"
  />
</template>
