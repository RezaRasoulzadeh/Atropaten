<script setup lang="ts">
defineOptions({ inheritAttrs: false });
defineProps<{ modelValue?: string | number | null; modelModifiers?: Record<string, boolean> }>();
const emit = defineEmits<{ 'update:modelValue': [value: string] }>();
</script>
<template>
  <input
    v-bind="$attrs"
    class="input h-10 w-full min-w-0 text-start text-sm leading-5 text-base-content"
    :value="modelValue"
    @input="emit('update:modelValue', ($event.target as HTMLInputElement).value)"
  />
</template>
