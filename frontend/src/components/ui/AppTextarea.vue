<script setup lang="ts">
defineProps<{ modelValue?: string }>();
const emit = defineEmits<{ 'update:modelValue': [value: string] }>();
</script>
<template>
  <textarea
    class="textarea w-full min-w-0 text-start text-sm text-base-content"
    :value="modelValue"
    @input="emit('update:modelValue', ($event.target as HTMLTextAreaElement).value)"
  />
</template>
