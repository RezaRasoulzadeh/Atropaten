import { computed, reactive } from 'vue'
import { formatQuantityUnits } from '../utils/quantity'

export type ToastKind = 'success' | 'error' | 'warning' | 'info'

export interface NormalizedError {
  message: string
  code?: string
  detail?: string
}

export interface ToastItem {
  id: number
  kind: ToastKind
  message: string
  title?: string
  dismissible: boolean
}

const state = reactive<{ items: ToastItem[] }>({ items: [] })
let nextToastId = 1
const recent = new Map<string, number>()

function cleanMessage(value: string): string {
  return (
    value.replace(/^Error:\s*/i, '').trim() ||
    'The operation could not be completed.'
  )
}

export function normalizeError(
  error: unknown,
  fallback = 'The operation could not be completed.',
): NormalizedError {
  if (error && typeof error === 'object') {
    const source = error as Record<string, unknown>
    const message =
      typeof source.message === 'string'
        ? source.message
        : typeof source.error === 'string'
          ? source.error
          : ''
    if (message)
      return {
        message: cleanMessage(message),
        code: typeof source.code === 'string' ? source.code : undefined,
        detail: typeof source.detail === 'string' ? source.detail : undefined,
      }
  }
  if (error instanceof Error)
    return { message: cleanMessage(error.message), detail: error.stack }
  if (typeof error === 'string') return { message: cleanMessage(error) }
  return { message: fallback }
}

function push(kind: ToastKind, message: unknown, title?: string) {
  const normalized =
    typeof message === 'string'
      ? cleanMessage(message)
      : normalizeError(message).message
  const key = `${kind}:${normalized}`
  const now = Date.now()
  if ((recent.get(key) ?? 0) > now - 800) return
  recent.set(key, now)
  const duration =
    kind === 'error'
      ? 9000
      : kind === 'warning'
        ? 7000
        : kind === 'info'
          ? 5000
          : 3600
  const item: ToastItem = {
    id: nextToastId++,
    kind,
    message: normalized,
    title,
    dismissible: true,
  }
  state.items.push(item)
  if (state.items.length > 5) state.items.shift()
  window.setTimeout(() => dismiss(item.id), duration)
}

export function dismiss(id: number) {
  const index = state.items.findIndex((item) => item.id === id)
  if (index >= 0) state.items.splice(index, 1)
}

export function useToast() {
  return {
    items: computed(() => state.items),
    success: (message: unknown, title?: string) =>
      push('success', message, title),
    error: (message: unknown, title?: string) => push('error', message, title),
    warning: (message: unknown, title?: string) =>
      push('warning', message, title),
    info: (message: unknown, title?: string) => push('info', message, title),
    dismiss,
  }
}

interface ConfirmState {
  open: boolean
  title: string
  message: string
  confirmLabel: string
  cancelLabel: string
  danger: boolean
  resolve?: (value: boolean) => void
}

export const confirmState = reactive<ConfirmState>({
  open: false,
  title: '',
  message: '',
  confirmLabel: 'Confirm',
  cancelLabel: 'Cancel',
  danger: false,
})

export function confirmAction(options: {
  title: string
  message: string
  confirmLabel?: string
  cancelLabel?: string
  danger?: boolean
}): Promise<boolean> {
  return new Promise((resolve) => {
    Object.assign(confirmState, {
      ...options,
      confirmLabel: options.confirmLabel ?? 'Confirm',
      cancelLabel: options.cancelLabel ?? 'Cancel',
      danger: options.danger ?? false,
      open: true,
      resolve,
    })
  })
}

export function resolveConfirm(value: boolean) {
  const resolve = confirmState.resolve
  confirmState.open = false
  confirmState.resolve = undefined
  resolve?.(value)
}

// Inbox read state is presentation-only and lasts for this application session.
// Reading an alert never resolves its underlying financial or inventory condition.
interface NotificationItem {
  id: string
  title: string
  detail: string
  destination: string
  date?: string
  amountRial?: number
}
const notifications = reactive<{
  items: NotificationItem[]
  read: Set<string>
}>({ items: [], read: new Set() })
export function useNotifications() {
  return {
    items: computed(() => notifications.items),
    unreadCount: computed(
      () =>
        notifications.items.filter((item) => !notifications.read.has(item.id))
          .length,
    ),
    isRead: (id: string) => notifications.read.has(id),
    markRead: (id: string) => notifications.read.add(id),
    markAllRead: () =>
      notifications.items.forEach((item) => notifications.read.add(item.id)),
    update: (data: import('../api/reports').DashboardRecord) => {
      notifications.items = [
        ...(data.attention ?? []).map((item) => ({
          id: `attention:${item.kind}:${item.detail}:${item.date}:${item.amountRial}`,
          title: item.label,
          detail: item.detail,
          date: item.date,
          amountRial: item.amountRial,
          destination:
            item.kind === 'invoice'
              ? 'Invoices'
              : item.kind === 'loan'
                ? 'Loans'
                : 'Checks',
        })),
        ...(data.lowStock ?? []).map((item) => ({
          id: `stock:${item.id}:${item.availableUnits}:${item.reorderLevelUnits}`,
          title: 'Low stock',
          detail: `${item.name} · ${formatQuantityUnits(String(item.availableUnits))} ${item.unit} available · Reorder at ${formatQuantityUnits(String(item.reorderLevelUnits))}`,
          destination: 'Materials',
        })),
        ...(data.recentActivity ?? []).map((item) => ({
          id: `payment:${item.id}`,
          title: item.label,
          detail: item.detail,
          date: item.date,
          amountRial: item.amountRial,
          destination: 'Accounting',
        })),
      ]
    },
  }
}
