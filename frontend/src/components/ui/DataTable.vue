<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { translateUi } from '../../i18n';

const props = withDefaults(defineProps<{ label?: string; paginate?: boolean }>(), { paginate: true });
const root = ref<HTMLElement | null>(null);
const page = ref(1);
const pageSize = ref(10);
const rowCount = ref(0);
let resizeObserver: ResizeObserver | undefined;
let mutationObserver: MutationObserver | undefined;
let frame = 0;

const pageCount = computed(() => Math.max(1, Math.ceil(rowCount.value / pageSize.value)));
const paginationVisible = computed(() => props.paginate && rowCount.value > pageSize.value);
const startRow = computed(() => rowCount.value ? ((page.value - 1) * pageSize.value) + 1 : 0);
const endRow = computed(() => Math.min(page.value * pageSize.value, rowCount.value));

function rows() {
  return root.value ? Array.from(root.value.querySelectorAll<HTMLTableRowElement>('tbody > tr')) : [];
}

function availableHeight() {
  if (!root.value) return 0;
  const rect = root.value.getBoundingClientRect();
  const viewportBottom = typeof window === 'undefined' ? 0 : window.innerHeight - rect.top - 16;
  const parentRect = root.value.parentElement?.getBoundingClientRect();
  const parentBottom = parentRect ? parentRect.bottom - rect.top - 8 : viewportBottom;
  const naturalHeight = root.value.scrollHeight || 0;
  return Math.max(120, Math.min(naturalHeight || viewportBottom, viewportBottom, parentBottom > 0 ? parentBottom : viewportBottom));
}

function syncRows() {
  const tableRows = rows();
  rowCount.value = tableRows.length;
  if (!props.paginate || !tableRows.length) {
    tableRows.forEach((row) => { row.hidden = false; });
    return;
  }

  const firstRowHeight = tableRows[0]?.getBoundingClientRect().height || 42;
  const nextPageSize = Math.max(1, Math.floor((availableHeight() - 48) / firstRowHeight));
  pageSize.value = Math.max(1, nextPageSize);
  if (page.value > pageCount.value) page.value = pageCount.value;
  tableRows.forEach((row, index) => {
    row.hidden = index < (page.value - 1) * pageSize.value || index >= page.value * pageSize.value;
  });
}

function scheduleSync() {
  cancelAnimationFrame(frame);
  frame = requestAnimationFrame(syncRows);
}

function goToPage(value: number) {
  page.value = Math.min(Math.max(value, 1), pageCount.value);
  scheduleSync();
}

watch(() => props.paginate, scheduleSync);
watch(pageCount, () => {
  if (page.value > pageCount.value) page.value = pageCount.value;
  scheduleSync();
});

onMounted(async () => {
  await nextTick();
  resizeObserver = new ResizeObserver(scheduleSync);
  if (root.value) resizeObserver.observe(root.value);
  if (root.value?.parentElement) resizeObserver.observe(root.value.parentElement);
  mutationObserver = new MutationObserver(scheduleSync);
  if (root.value) mutationObserver.observe(root.value, { childList: true, subtree: true });
  scheduleSync();
});

onBeforeUnmount(() => {
  resizeObserver?.disconnect();
  mutationObserver?.disconnect();
  cancelAnimationFrame(frame);
});
</script>
<template>
  <div
    ref="root"
    class="data-table min-w-0 overflow-x-auto"
    tabindex="0"
    role="region"
    :aria-label="translateUi(props.label || 'Data table')"
  >
    <table class="table table-sm w-full text-sm">
      <slot />
    </table>
    <footer v-if="paginationVisible" class="data-table-pagination flex flex-wrap items-center justify-between gap-3 border-t border-base-300 bg-base-100 px-3 py-2 text-xs text-base-content/60">
      <span>{{ translateUi(`Showing ${startRow}–${endRow} of ${rowCount} rows`) }}</span>
      <div class="flex items-center gap-1">
        <button class="btn btn-ghost btn-xs" type="button" :disabled="page === 1" :aria-label="translateUi('Previous page')" @click="goToPage(page - 1)">{{ translateUi('Previous page') }}</button>
        <span class="min-w-16 text-center tabular-nums">{{ page }} / {{ pageCount }}</span>
        <button class="btn btn-ghost btn-xs" type="button" :disabled="page === pageCount" :aria-label="translateUi('Next page')" @click="goToPage(page + 1)">{{ translateUi('Next page') }}</button>
      </div>
    </footer>
  </div>
</template>
<style scoped>
.data-table :deep(th) {
  color: var(--color-base-content);
  font-size: 0.75rem;
  font-weight: 600;
  line-height: 1rem;
  white-space: nowrap;
  background: var(--color-base-200);
}
.data-table :deep(table) {
  width: max-content;
  min-width: 100%;
}
.data-table :deep(th),
.data-table :deep(td) {
  font-size: 0.875rem;
  line-height: 1.25rem;
  padding: 0.625rem 0.75rem;
  border-bottom: 1px solid var(--color-base-300);
  vertical-align: middle;
  white-space: nowrap;
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
.data-table :deep(td) { overflow-wrap: normal; }
.data-table :deep(.badge) { flex-shrink: 0; }
.data-table-pagination { min-height: 2.75rem; }
</style>
