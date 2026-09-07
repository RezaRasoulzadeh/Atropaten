import { ref } from 'vue'
import { normalizeError, useToast } from '../ui/feedback'

export function reportError(error: unknown) {
  useToast().error(normalizeError(error).message)
}

// A workspace serializes explicit mutations, while reads keep their own loading
// state. Failed actions preserve the caller's form and always release the guard.
export function useWorkspaceActions() {
  const busy = ref(false)
  const pageLoading = ref(false)
  async function runLoad<T>(read: () => Promise<T>): Promise<T> {
    pageLoading.value = true
    try { return await read() }
    finally { pageLoading.value = false }
  }
  async function runAction<T>(action: () => Promise<T>): Promise<T | undefined> {
    if (busy.value) return
    busy.value = true
    try { return await action() }
    catch (error) { reportError(error) }
    finally { busy.value = false }
  }
  return { busy, runAction, pageLoading, runLoad }
}
