import { nextTick } from 'vue'

const PRINT_PREPARING_CLASS = 'print-preparing'

function nextFrame(): Promise<void> {
  return new Promise((resolve) => requestAnimationFrame(() => resolve()))
}

async function prepareFonts(): Promise<void> {
  if (!document.fonts) return

  try {
    await Promise.race([
      Promise.all([400, 600, 700].map((weight) =>
        document.fonts.load(`${weight} 12px "Estedad"`, 'Invoice فاکتور ۱۲۳۴۵۶۷۸۹۰'),
      )),
      new Promise<void>((resolve) => window.setTimeout(resolve, 2500)),
    ])
  } catch {
    // A system fallback is safer than preventing printing when WebKit reports
    // an embedded font as unavailable.
  }
}

// WebKit on macOS may snapshot print-only content before it has participated in
// layout. Prepare it off-screen, wait for layout and fonts, then open printing.
export async function printDocument(): Promise<void> {
  await nextTick()
  document.body.classList.add(PRINT_PREPARING_CLASS)

  let fallbackCleanup = 0
  const cleanup = () => {
    document.body.classList.remove(PRINT_PREPARING_CLASS)
    window.removeEventListener('afterprint', cleanup)
    if (fallbackCleanup) window.clearTimeout(fallbackCleanup)
  }

  try {
    await prepareFonts()
    await nextFrame()
    await nextFrame()
    window.addEventListener('afterprint', cleanup, { once: true })
    window.print()
    fallbackCleanup = window.setTimeout(cleanup, 30_000)
  } catch (error) {
    cleanup()
    throw error
  }
}
