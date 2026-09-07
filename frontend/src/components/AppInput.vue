<script setup lang="ts">
withDefaults(defineProps<{ modelValue?: string | number; type?: string; placeholder?: string; disabled?: boolean; readonly?: boolean; id?: string; name?: string; autocomplete?: string }>(), { modelValue: '', type: 'text', disabled: false, readonly: false })
defineEmits<{ 'update:modelValue': [value: string] }>()
</script>
<template><input class="input input-bordered h-10 w-full min-w-0 border border-base-300 bg-base-100 px-3 py-2 text-sm shadow-none outline-none transition-colors placeholder:text-base-content/50 focus:border-primary focus:outline-none focus:ring-0 disabled:cursor-not-allowed disabled:opacity-60" :id="id" :name="name" :type="type" :value="modelValue" :placeholder="placeholder" :disabled="disabled" :readonly="readonly" :autocomplete="autocomplete" @input="$emit('update:modelValue', ($event.target as HTMLInputElement).value)" /></template>
