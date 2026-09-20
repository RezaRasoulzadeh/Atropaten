import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch, type ComputedRef, type Ref } from 'vue'

type DynamicPaginationOptions = {
  fallbackPageSize?: number
  minPageSize?: number
  viewportSelector?: string
}

/**
 * Paginates a flexible list using the actual height of its viewport.
 * The first rendered row is measured so compact and mobile register rows are
 * both handled correctly. The page size is recalculated on resize and after
 * the source list changes.
 */
export function useDynamicPagination<T>(
  items: ComputedRef<T[]> | Ref<T[]>,
  options: DynamicPaginationOptions = {},
) {
  const page = ref(1)
  const pageSize = ref(Math.max(options.minPageSize ?? 1, options.fallbackPageSize ?? 10))
  const viewport = ref<HTMLElement | null>(null)
  let resizeObserver: ResizeObserver | undefined
  let frame = 0

  const pageCount = computed(() => Math.max(1, Math.ceil(items.value.length / pageSize.value)))
  const pagedItems = computed(() => items.value.slice((page.value - 1) * pageSize.value, page.value * pageSize.value))
  const pageNumbers = computed(() => Array.from({ length: pageCount.value }, (_, index) => index + 1))

  function clampPage() {
    if (page.value > pageCount.value) page.value = pageCount.value
    if (page.value < 1) page.value = 1
  }

  function resolveViewport() {
    if (viewport.value || !options.viewportSelector || typeof document === 'undefined') return viewport.value
    const row = document.querySelector<HTMLElement>(options.viewportSelector)
    viewport.value = row?.parentElement ?? null
    if (viewport.value && resizeObserver) resizeObserver.observe(viewport.value)
    return viewport.value
  }

  function recalculate() {
    resolveViewport()
    if (!viewport.value) return
    const firstRow = viewport.value.firstElementChild as HTMLElement | null
    const rowHeight = firstRow?.getBoundingClientRect().height || 64
    const availableHeight = viewport.value.clientHeight
    if (!availableHeight) return

    const nextPageSize = Math.max(options.minPageSize ?? 1, Math.floor(availableHeight / rowHeight))
    if (nextPageSize !== pageSize.value) pageSize.value = nextPageSize
    clampPage()
  }

  function scheduleRecalculate() {
    cancelAnimationFrame(frame)
    frame = requestAnimationFrame(recalculate)
  }

  function goToPage(value: number) {
    page.value = Math.min(Math.max(value, 1), pageCount.value)
    scheduleRecalculate()
  }

  watch(items, async () => {
    clampPage()
    await nextTick()
    resolveViewport()
    scheduleRecalculate()
  }, { immediate: true })

  watch(pageCount, clampPage)

  onMounted(() => {
    resizeObserver = new ResizeObserver(scheduleRecalculate)
    resolveViewport()
    scheduleRecalculate()
  })

  onBeforeUnmount(() => {
    resizeObserver?.disconnect()
    cancelAnimationFrame(frame)
  })

  return { page, pageSize, pageCount, pageNumbers, pagedItems, viewport, goToPage }
}
