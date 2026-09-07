<script setup lang="ts">
withDefaults(defineProps<{ modelValue?: string | number; disabled?: boolean; id?: string; name?: string }>(), { modelValue: '', disabled: false })
defineEmits<{ 'update:modelValue': [value: string] }>()
</script>
<template><select class="select select-bordered at-control" :id="id" :name="name" :value="modelValue" :disabled="disabled" @change="$emit('update:modelValue', ($event.target as HTMLSelectElement).value)"><slot /></select></template>
