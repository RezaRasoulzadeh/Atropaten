import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

const source = readFileSync(new URL('../src/features/services/rollSizeInputs.ts', import.meta.url), 'utf8')
const js = ts.transpileModule(source, { compilerOptions: { module: ts.ModuleKind.ESNext, target: ts.ScriptTarget.ES2022 } }).outputText
const { ensureRollSizeInputs, usesMaterialRollWidth, selectedRollMaterial, rollWidthValue } = await import(`data:text/javascript;base64,${Buffer.from(js).toString('base64')}`)
const material = { id: 'roll', name: 'Banner 320', kind: 'roll-media', attributes: [{ key: 'width_mm', valueType: 'decimal', decimalValue: '3200' }] }
for (const category of ['Large format', 'Banners & signage', 'Paper printing']) {
  const service = { id: 'legacy', category, parameters: [{ key: 'paper', type: 'choice', defaultValue: '3200', materialSource: { allowedKinds: ['roll-media'], exposedAttributeKeys: ['width_mm'] } }], components: [] }
  ensureRollSizeInputs(service)
  assert.equal(usesMaterialRollWidth(service), true)
  assert.equal(service.finishedSize.heightParameterKey, 'finished_height_mm')
  assert.equal(service.parameters.find(p => p.key === 'finished_height_mm').defaultValue, '1000')
  assert.equal(selectedRollMaterial(service, [material], { paper: '3200' }), material)
  assert.equal(rollWidthValue(material, '0'), '3200')
  assert.equal(rollWidthValue(material, '10'), '3180')
  const snapshot = JSON.stringify(service)
  ensureRollSizeInputs(service)
  assert.equal(JSON.stringify(service), snapshot)
}
for (const category of ['Design', 'Finishing', 'Paper printing']) {
  const service = { category, parameters: [] }
  ensureRollSizeInputs(service)
  assert.equal(usesMaterialRollWidth(service), false)
  assert.equal(service.finishedSize, undefined)
}
console.log('Roll inputs: legacy services, width selection, margins, category isolation and idempotence passed')
const projected = { id: 'old-print-size', category: 'Large format', parameters: [], finishedSize: { parameterKey: 'paper', allowCustom: false, options: [] } }
ensureRollSizeInputs(projected)
assert.equal(usesMaterialRollWidth(projected), true)
