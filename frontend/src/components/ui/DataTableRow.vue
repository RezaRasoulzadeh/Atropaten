<script setup lang="ts">
defineProps<{ selected?: boolean; interactive?: boolean }>();
const emit = defineEmits<{ activate: [] }>();
</script>
<template>
  <tr
    :class="[
      { 'bg-base-300': selected, 'cursor-pointer': interactive },
      interactive ? 'focus-visible:outline focus-visible:outline-1 focus-visible:outline-base-content focus-visible:outline-offset-[-1px]' : '',
    ]"
    :tabindex="interactive ? 0 : undefined"
    @click="interactive && emit('activate')"
    @keydown.enter.self.prevent="interactive && emit('activate')"
    @keydown.space.self.prevent="interactive && emit('activate')"
  >
    <slot />
  </tr>
</template>
