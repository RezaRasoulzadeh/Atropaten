<template>
  <section
    class="search-filter-bar flex min-w-0 flex-wrap items-end gap-2 rounded-box border border-base-300 bg-base-100 p-3"
  >
    <div class="min-w-[220px] flex-1">
      <slot name="search" />
    </div>
    <div class="flex min-w-0 flex-wrap items-end gap-2 *:min-w-28 *:flex-1">
      <slot name="filters" />
    </div>
    <div v-if="$slots.count" class="flex h-10 shrink-0 items-center text-xs leading-4 text-base-content/60">
      <slot name="count" />
    </div>
    <div v-if="$slots.actions" class="flex h-10 shrink-0 items-center">
      <slot name="actions" />
    </div>
  </section>
</template>
