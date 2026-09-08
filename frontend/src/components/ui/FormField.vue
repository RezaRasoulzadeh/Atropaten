<script setup lang="ts">
import FieldMessage from './FieldMessage.vue';
defineProps<{
  label?: string;
  help?: string;
  required?: boolean;
  forId?: string;
}>();
</script>
<template>
  <label
    class="flex w-full min-w-0 flex-col gap-1 text-xs leading-4 text-base-content/65"
    :for="forId"
  >
    <span v-if="label">{{ label }}<span v-if="required" aria-hidden="true"> *</span></span>
    <slot /><FieldMessage
      v-if="help"
      :message="help"
    />
  </label>
</template>
