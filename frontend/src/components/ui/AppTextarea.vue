<script setup lang="ts">
defineOptions({ inheritAttrs: false });
defineProps<{ modelValue?: string }>();
const emit = defineEmits<{ 'update:modelValue': [value: string] }>();
</script>
<template>
  <textarea
    v-bind="$attrs"
    class="textarea w-full min-w-0 text-start text-sm leading-5 text-base-content"
    :value="modelValue"
    @input="emit('update:modelValue', ($event.target as HTMLTextAreaElement).value)"
  />
</template>
