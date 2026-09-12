import assert from 'node:assert/strict'
import { spawn } from 'node:child_process'
import { mkdtemp, readFile, rm } from 'node:fs/promises'
import { createServer } from 'node:http'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { build } from 'esbuild'
import { compileScript, parse } from '@vue/compiler-sfc'

// Use a real browser for requestSubmit, validation, DOM boundaries and Vue updates.
// No desktop APIs or user database are involved. Set CHROME_BIN when necessary.
const root = fileURLToPath(new URL('../', import.meta.url))
const bundle = await build({
  entryPoints: [join(root, 'scripts/enter-forms.browser.ts')], bundle: true,
  write: false, format: 'iife', platform: 'browser',
  define: { __VUE_OPTIONS_API__: 'true', __VUE_PROD_DEVTOOLS__: 'false', __VUE_PROD_HYDRATION_MISMATCH_DETAILS__: 'false' },
  plugins: [{ name: 'vue', setup(builder) {
    builder.onLoad({ filter: /\.vue$/ }, async ({ path }) => {
      const { descriptor } = parse(await readFile(path, 'utf8'), { filename: path })
      return { contents: compileScript(descriptor, { id: path, inlineTemplate: true }).content, loader: 'ts' }
    })
  } }],
})
const server = createServer((request, response) => {
  response.setHeader('Content-Type', request.url === '/test.js' ? 'text/javascript' : 'text/html')
  response.end(request.url === '/test.js' ? bundle.outputFiles[0].text : '<!doctype html><body><script src="/test.js"></script></body>')
})
await new Promise((resolve, reject) => { server.once('error', reject); server.listen(0, '127.0.0.1', resolve) })
const profile = await mkdtemp(join(tmpdir(), 'atropaten-enter-test-'))
try {
  const output = await new Promise((resolve, reject) => {
    const chrome = spawn(process.env.CHROME_BIN || 'google-chrome', [
      '--headless', '--no-sandbox', '--disable-gpu', '--disable-dev-shm-usage', '--no-first-run',
      `--user-data-dir=${profile}`, '--dump-dom', '--virtual-time-budget=5000',
      `http://127.0.0.1:${server.address().port}`,
    ])
    let stdout = '', stderr = ''
    const timeout = setTimeout(() => { chrome.kill(); reject(new Error('Enter form browser test timed out')) }, 30000)
    chrome.stdout.on('data', (data) => { stdout += data })
    chrome.stderr.on('data', (data) => { stderr += data })
    chrome.once('error', (error) => { clearTimeout(timeout); reject(error) })
    chrome.once('exit', (code) => {
      clearTimeout(timeout)
      if (code !== 0) reject(new Error(`Chrome exited ${code}: ${stderr}`))
      else resolve(stdout)
    })
  })
  assert.match(output, /data-test-result="passed"/, output)
  console.log('Enter form browser tests passed (native forms, nested actions, validation, keyboard guards, Vue bindings).')
} finally {
  server.close()
  await rm(profile, { recursive: true, force: true })
}
