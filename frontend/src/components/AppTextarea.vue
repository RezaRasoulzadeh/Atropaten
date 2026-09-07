<script setup lang="ts">
withDefaults(defineProps<{ modelValue?: string; placeholder?: string; disabled?: boolean; rows?: number; id?: string; name?: string }>(), { modelValue: '', disabled: false, rows: 3 })
defineEmits<{ 'update:modelValue': [value: string] }>()
</script>
<template><textarea class="textarea textarea-bordered at-control at-textarea" :id="id" :name="name" :rows="rows" :value="modelValue" :placeholder="placeholder" :disabled="disabled" @input="$emit('update:modelValue', ($event.target as HTMLTextAreaElement).value)"></textarea></template>
