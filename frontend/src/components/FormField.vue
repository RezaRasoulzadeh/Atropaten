<script setup lang="ts">
import FieldMessage from './FieldMessage.vue'
withDefaults(defineProps<{ label?: string; error?: string; help?: string; required?: boolean; forId?: string }>(), { required: false })
</script>
<template><label class="form-control w-full gap-1" :for="forId"><span v-if="label" class="label-text text-xs text-base-content/70">{{ label }}<b v-if="required" aria-hidden="true"> *</b></span><slot /><FieldMessage v-if="error" :message="error" tone="error" /><FieldMessage v-else :message="help" /></label></template>
