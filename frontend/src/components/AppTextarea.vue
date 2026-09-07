<script setup lang="ts">
withDefaults(defineProps<{ modelValue?: string; placeholder?: string; disabled?: boolean; rows?: number; id?: string; name?: string }>(), { modelValue: '', disabled: false, rows: 3 })
defineEmits<{ 'update:modelValue': [value: string] }>()
</script>
<template><textarea class="textarea textarea-bordered w-full min-w-0" :id="id" :name="name" :rows="rows" :value="modelValue" :placeholder="placeholder" :disabled="disabled" @input="$emit('update:modelValue', ($event.target as HTMLTextAreaElement).value)"></textarea></template>
