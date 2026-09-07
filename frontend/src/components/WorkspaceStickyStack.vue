<script setup lang="ts">
withDefaults(defineProps<{ flush?: boolean }>(), { flush: false })
</script>

<template>
  <div class="workspace-sticky-stack sticky top-0 z-10 min-w-0 bg-base-200"
    :class="{ 'workspace-sticky-stack--flush': flush }">
    <slot />
  </div>
</template>

<style scoped>
.workspace-sticky-stack {
  align-self: flex-start;
  padding-block: 0.5rem;
}

.workspace-sticky-stack--flush {
  padding-bottom: 0;
}

.workspace-sticky-stack > :deep(header) {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 1rem;
  padding-block: 0.5rem;
}

.workspace-sticky-stack > :deep(header > div:last-child) {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: flex-end;
  gap: 0.5rem;
}

.workspace-sticky-stack > :deep(header h1) {
  margin: 0;
  color: var(--color-base-content);
  font-size: 1.625rem;
  font-weight: 700;
  letter-spacing: -0.02em;
  line-height: 1.25;
}

.workspace-sticky-stack > :deep(header p) {
  margin: 0;
}

.workspace-sticky-stack > :deep(header > div:not(:last-child)) {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 0.125rem;
}

.workspace-sticky-stack > :deep(header > div:not(:last-child) > p:first-child) {
  color: var(--color-primary);
  font-size: 0.75rem;
  font-weight: 600;
  letter-spacing: 0.01em;
  line-height: 1.35;
}

.workspace-sticky-stack > :deep(header > div:not(:last-child) > p:last-child) {
  max-width: 62ch;
  color: color-mix(in oklab, var(--color-base-content) 65%, transparent);
  font-size: 0.8125rem;
  line-height: 1.5;
}

@media (max-width: 48rem) {
  .workspace-sticky-stack > :deep(header) {
    flex-direction: column;
    align-items: flex-start;
  }

  .workspace-sticky-stack > :deep(header > div:last-child) {
    justify-content: flex-start;
  }
}
</style>
