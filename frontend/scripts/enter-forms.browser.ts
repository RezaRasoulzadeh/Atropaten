import { createApp, h, nextTick, ref } from 'vue'
import AppInput from '../src/components/ui/AppInput.vue'
import SearchField from '../src/components/ui/SearchField.vue'
import { installFormEnter } from '../src/ui/enterSubmit'

const assert = (value: unknown, message: string) => { if (!value) throw new Error(message) }
const fixture = document.createElement('main')
document.body.append(fixture)
const removeListener = installFormEnter(document)
const settle = async () => { await nextTick(); await nextTick(); await nextTick() }
const enter = async (input: Element, overrides: KeyboardEventInit = {}) => {
  const event = new KeyboardEvent('keydown', { key: 'Enter', bubbles: true, cancelable: true, ...overrides })
  input.dispatchEvent(event)
  await settle()
  return event
}

async function run() {
  fixture.innerHTML = '<form id="native"><input required><input><button type="submit">Save</button></form>'
  const form = fixture.querySelector('form')!
  const input = form.querySelector('input')!
  const submitButton = form.querySelector('button')!
  let submitted = 0
  form.addEventListener('submit', (event) => { event.preventDefault(); submitted++ })
  await enter(input)
  assert(submitted === 0, 'Required field must block Enter submission')
  input.value = 'Valid'
  await enter(input)
  assert(submitted === 1, 'Enter must submit a multi-input native form once')
  submitButton.disabled = true
  await enter(input)
  assert(submitted === 1, 'Disabled Save must block Enter')
  submitButton.disabled = false
  form.setAttribute('aria-busy', 'true')
  await enter(input)
  assert(submitted === 1, 'Busy form must block Enter')
  form.removeAttribute('aria-busy')
  for (const flags of [{ isComposing: true }, { ctrlKey: true }, { metaKey: true }, { shiftKey: true }, { altKey: true }]) {
    assert(!(await enter(input, flags)).defaultPrevented, 'Composition and modified Enter must stay untouched')
  }
  await enter(input, { repeat: true })
  assert(submitted === 1, 'Holding Enter must not submit repeatedly')
  input.addEventListener('keydown', (event) => event.preventDefault(), { once: true })
  await enter(input)
  assert(submitted === 1, 'A control-handled event must not submit')

  for (const markup of ['<textarea></textarea>', '<select><option>Cash</option></select>', '<input type="range">', '<input type="checkbox">', '<input type="file">', '<button type="button">Action</button>', '<div contenteditable>Text</div>']) {
    const holder = document.createElement('div')
    holder.innerHTML = markup
    form.append(holder)
    assert(!(await enter(holder.firstElementChild!)).defaultPrevented, `Native keyboard behavior changed for ${markup}`)
    holder.remove()
  }
  assert(submitted === 1, 'Non-text controls must not submit through the handler')

  // Form-associated header/footer buttons are respected even outside the form.
  submitButton.remove()
  const external = document.createElement('button')
  external.type = 'submit'
  external.setAttribute('form', 'native')
  external.disabled = true
  fixture.append(external)
  await enter(input)
  assert(submitted === 1, 'External disabled Save must block Enter')
  external.disabled = false
  await enter(input)
  assert(submitted === 2, 'External form-associated Save must work')

  const nested = document.createElement('section')
  nested.dataset.enterScope = ''
  nested.innerHTML = '<input required><button type="button">Cancel</button><button type="button" data-enter-submit>Record</button><button type="button">Reverse</button>'
  form.append(nested)
  const amount = nested.querySelector('input')!
  const record = nested.querySelector<HTMLButtonElement>('[data-enter-submit]')!
  let recorded = 0, wrongAction = 0
  record.addEventListener('click', () => recorded++)
  nested.querySelectorAll('button:not([data-enter-submit])').forEach((button) => button.addEventListener('click', () => wrongAction++))
  await enter(amount)
  assert(recorded === 0, 'Subform required fields must be validated')
  amount.value = '100'
  input.value = '' // Invalid parent fields must not block the nested action.
  await enter(amount)
  assert(recorded === 1 && submitted === 2 && wrongAction === 0, 'Nested action must not submit the parent or another action')
  record.disabled = true
  await enter(amount)
  assert(recorded === 1 && submitted === 2, 'Disabled nested action must not fall back to parent')
  record.disabled = false
  amount.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter', bubbles: true, cancelable: true }))
  amount.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter', bubbles: true, cancelable: true }))
  await settle()
  assert(recorded === 2, 'Back-to-back Enter events must not double submit')

  const emptyScope = document.createElement('section')
  emptyScope.dataset.enterScope = ''
  emptyScope.innerHTML = '<input>'
  form.append(emptyScope)
  await enter(emptyScope.querySelector('input')!)
  assert(submitted === 2, 'Empty subform must not submit parent')
  const dialog = document.createElement('div')
  dialog.setAttribute('role', 'dialog')
  dialog.innerHTML = '<input>'
  form.append(dialog)
  await enter(dialog.querySelector('input')!)
  assert(submitted === 2, 'Modal inputs must not submit a background form')

  fixture.innerHTML = '<input>'
  assert(!(await enter(fixture.querySelector('input')!)).defaultPrevented, 'Unscoped input must stay untouched')

  // Real Vue inputs: verify live formatting and latest busy bindings before action.
  fixture.innerHTML = ''
  const value = ref(''), busy = ref(false)
  let captured = '', clicks = 0
  const app = createApp({ setup: () => () => h('section', { 'data-enter-scope': '' }, [
    h(AppInput, { modelValue: value.value, money: 'Toman', 'onUpdate:modelValue': (text: string) => { value.value = text } }),
    h(SearchField, { modelValue: '', placeholder: 'Search' }),
    h('button', { type: 'button', 'data-enter-submit': '', disabled: busy.value, onClick: () => { captured = value.value; clicks++ } }, 'Save'),
  ]) })
  app.mount(fixture)
  const money = fixture.querySelector('input')!
  money.value = '1234567'
  money.dispatchEvent(new Event('input', { bubbles: true }))
  await enter(money)
  assert(captured === '1,234,567' && clicks === 1, 'Enter must use the latest formatted v-model')
  busy.value = true
  await enter(money)
  assert(clicks === 1, 'Pending disabled binding must be flushed before submitting')
  busy.value = false
  await enter(fixture.querySelector('input[type="search"]')!)
  assert(clicks === 1, 'Search must not activate the editor action')
  app.unmount()
  removeListener()
  document.body.dataset.testResult = 'passed'
}

void run().catch((error) => {
  document.body.dataset.testResult = 'failed'
  document.body.append(document.createTextNode(String(error.stack || error)))
})
