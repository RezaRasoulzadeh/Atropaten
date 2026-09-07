<script setup lang="ts">
withDefaults(defineProps<{ modelValue?: string | number; type?: string; placeholder?: string; disabled?: boolean; readonly?: boolean; id?: string; name?: string; autocomplete?: string }>(), { modelValue: '', type: 'text', disabled: false, readonly: false })
defineEmits<{ 'update:modelValue': [value: string] }>()
</script>
<template><input class="input input-bordered at-control" :id="id" :name="name" :type="type" :value="modelValue" :placeholder="placeholder" :disabled="disabled" :readonly="readonly" :autocomplete="autocomplete" @input="$emit('update:modelValue', ($event.target as HTMLInputElement).value)" /></template>
