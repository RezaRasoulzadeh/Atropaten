import { nextTick } from 'vue'

const scopeSelector = '[data-enter-scope], form'
const textInputTypes = new Set(['text', 'number', 'email', 'password', 'tel', 'url'])
const pending = new WeakSet<Element>()

/** Native forms submit normally. Non-form panels explicitly mark their primary
 * button with data-enter-submit and their boundary with data-enter-scope.
 * The nearest boundary wins, so a subform never submits its parent wizard.
 */
export async function handleFormEnter(event: KeyboardEvent) {
  if (event.key !== 'Enter' || event.defaultPrevented || event.isComposing || event.keyCode === 229
    || event.altKey || event.ctrlKey || event.metaKey || event.shiftKey) return

  const input = event.target
  if (!(input instanceof HTMLInputElement) || !textInputTypes.has(input.type)
    || input.disabled || input.readOnly || input.isContentEditable
    || input.closest('[data-enter-ignore], [role="combobox"], [role="search"]')) return

  const scope = input.closest<HTMLElement>(scopeSelector) ?? input.form
  if (!scope || scope.closest('form[method="dialog"]')) return
  const dialog = input.closest('dialog, [role="dialog"]')
  if (dialog && !dialog.contains(scope)) {
    event.preventDefault()
    return
  }
  event.preventDefault()
  if (event.repeat || pending.has(scope)) return

  pending.add(scope)
  try {
    // Flush v-model and disabled/busy bindings before choosing the action.
    await nextTick()
    if (!input.isConnected || !scope.isConnected || scope.closest('[aria-busy="true"], [inert]')) return

    const button = Array.from(scope.querySelectorAll<HTMLButtonElement>('button[data-enter-submit]'))
      .find((candidate) => candidate.closest(scopeSelector) === scope)
    if (button) {
      if (button.matches(':disabled') || button.getAttribute('aria-disabled') === 'true') return
      const fields = Array.from(scope.querySelectorAll<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>('input, select, textarea'))
        .filter((field) => field.closest(scopeSelector) === scope)
      if (fields.some((field) => !field.reportValidity())) return
      button.click()
    } else if (scope instanceof HTMLFormElement) {
      // Include submit buttons linked from wizard headers/footers via form=id.
      const submitters = Array.from(scope.elements).filter((element) =>
        (element instanceof HTMLButtonElement || element instanceof HTMLInputElement) && element.type === 'submit')
      if (submitters.length && submitters.every((element) => element.matches(':disabled'))) return
      scope.requestSubmit()
    }
    await nextTick()
  } finally {
    pending.delete(scope)
  }
}

export function installFormEnter(target: Document) {
  const listener = (event: KeyboardEvent) => { void handleFormEnter(event) }
  target.addEventListener('keydown', listener)
  return () => target.removeEventListener('keydown', listener)
}
