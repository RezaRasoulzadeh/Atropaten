import { nextTick } from 'vue'

// Print-only content is display:none until the dialog opens, so it cannot be
// relied on to trigger font loading. Load every weight used by documents first.
export async function printDocument(): Promise<void> {
  await nextTick()
  const faces = await Promise.all([400, 600, 700].map((weight) =>
    document.fonts.load(`${weight} 12px "Estedad"`, 'Invoice فاکتور ۱۲۳۴۵۶۷۸۹۰'),
  ))
  if (faces.some((loaded) => !loaded.length || loaded.some((face) => face.status !== 'loaded'))) {
    throw new Error('The invoice font could not be loaded. Please reopen the application and try again.')
  }
  await document.fonts.ready
  window.print()
}
