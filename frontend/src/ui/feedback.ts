import { computed, reactive } from 'vue'

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
  return value.replace(/^Error:\s*/i, '').trim() || 'The operation could not be completed.'
}

export function normalizeError(error: unknown, fallback = 'The operation could not be completed.'): NormalizedError {
  if (error && typeof error === 'object') {
    const source = error as Record<string, unknown>
    const message = typeof source.message === 'string' ? source.message : typeof source.error === 'string' ? source.error : ''
    if (message) return { message: cleanMessage(message), code: typeof source.code === 'string' ? source.code : undefined, detail: typeof source.detail === 'string' ? source.detail : undefined }
  }
  if (error instanceof Error) return { message: cleanMessage(error.message), detail: error.stack }
  if (typeof error === 'string') return { message: cleanMessage(error) }
  return { message: fallback }
}

function push(kind: ToastKind, message: unknown, title?: string) {
  const normalized = typeof message === 'string' ? cleanMessage(message) : normalizeError(message).message
  const key = `${kind}:${normalized}`
  const now = Date.now()
  if ((recent.get(key) ?? 0) > now - 800) return
  recent.set(key, now)
  const item: ToastItem = { id: nextToastId++, kind, message: normalized, title, dismissible: true }
  state.items.push(item)
  if (state.items.length > 5) state.items.shift()
  const timeout = kind === 'error' ? 9000 : kind === 'warning' ? 7000 : kind === 'info' ? 5000 : 3600
  window.setTimeout(() => dismiss(item.id), timeout)
}

export function dismiss(id: number) {
  const index = state.items.findIndex((item) => item.id === id)
  if (index >= 0) state.items.splice(index, 1)
}

export function useToast() {
  return {
    items: computed(() => state.items),
    success: (message: unknown, title?: string) => push('success', message, title),
    error: (message: unknown, title?: string) => push('error', message, title),
    warning: (message: unknown, title?: string) => push('warning', message, title),
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

export const confirmState = reactive<ConfirmState>({ open: false, title: '', message: '', confirmLabel: 'Confirm', cancelLabel: 'Cancel', danger: false })

export function confirmAction(options: { title: string; message: string; confirmLabel?: string; cancelLabel?: string; danger?: boolean }): Promise<boolean> {
  return new Promise((resolve) => {
    Object.assign(confirmState, { ...options, confirmLabel: options.confirmLabel ?? 'Confirm', cancelLabel: options.cancelLabel ?? 'Cancel', danger: options.danger ?? false, open: true, resolve })
  })
}

export function resolveConfirm(value: boolean) {
  const resolve = confirmState.resolve
  confirmState.open = false
  confirmState.resolve = undefined
  resolve?.(value)
}
