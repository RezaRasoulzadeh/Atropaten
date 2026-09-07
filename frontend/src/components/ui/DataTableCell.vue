<script setup lang="ts">
defineProps<{ numeric?: boolean }>();
</script>
<template>
  <td :class="numeric ? 'text-end whitespace-nowrap tabular-nums' : ''"><slot /></td>
</template>
