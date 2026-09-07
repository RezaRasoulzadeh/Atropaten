<script setup lang="ts">
defineProps<{ label?: string }>();
</script>
<template>
  <div
    class="data-table min-w-0 overflow-x-auto"
    tabindex="0"
    role="region"
    :aria-label="label || 'Data table'"
  >
    <table class="table table-sm w-full text-sm">
      <slot />
    </table>
  </div>
</template>
<style scoped>
.data-table :deep(th) {
  color: var(--color-base-content);
  font-size: 0.75rem;
  font-weight: 600;
  white-space: nowrap;
  background: var(--color-base-200);
}
.data-table :deep(th),
.data-table :deep(td) {
  padding: 0.625rem 0.75rem;
  border-bottom: 1px solid var(--color-base-300);
  vertical-align: middle;
}
.data-table :deep(tbody tr:last-child td) {
  border-bottom: 0;
}
.data-table :deep(tbody tr:hover),
.data-table :deep(tbody tr:focus-visible) {
  background: var(--color-base-200);
}
.data-table :deep(tbody tr:focus-visible) {
  outline: 1px solid var(--color-base-content);
  outline-offset: -1px;
}
.data-table :deep(td) {
  overflow-wrap: anywhere;
}
</style>
