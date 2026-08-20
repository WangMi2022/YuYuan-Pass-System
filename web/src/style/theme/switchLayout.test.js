import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import test from 'node:test'

const themeSource = readFileSync(new URL('./components.scss', import.meta.url), 'utf8')
const reportSource = readFileSync(new URL('../../plugin/smart/view/report.vue', import.meta.url), 'utf8')
const operationsSource = readFileSync(new URL('../../plugin/aioperations/view/operations.vue', import.meta.url), 'utf8')

test('switch labels stay horizontal when a settings row becomes narrow', () => {
  assert.ok(
    /\.el-switch\.el-switch\s*\{[^}]*flex:\s*0\s+0\s+auto[^}]*flex-direction:\s*row[^}]*\}/s.test(themeSource),
    'switches must stay horizontal and must not shrink as flex children'
  )
  assert.ok(
    /\.el-switch\.el-switch\s+\.el-switch__label\s*\{[^}]*flex:\s*0\s+0\s+auto[^}]*white-space:\s*nowrap[^}]*\}/s.test(themeSource),
    'switch labels must stay on one line'
  )
  assert.ok(
    /\.subscription-panel\s+\.panel-header__copy\s*\{[^}]*min-width:\s*0[^}]*flex:\s*1\s+1\s+auto[^}]*\}/s.test(reportSource),
    'the report subscription heading must yield space to its switch'
  )
})

test('AI settings copy styles do not turn switch internals into a column', () => {
  assert.ok(
    /\.setting-toggle__copy\s*\{[^}]*flex-direction:\s*column[^}]*\}/s.test(operationsSource),
    'settings copy styles must target the named copy container'
  )
  assert.ok(
    !/\.setting-toggle\s*(?:>|\s)\s*div\s*\{[^}]*flex-direction:\s*column[^}]*\}/s.test(operationsSource),
    'settings copy styles must not target any switch root element'
  )
})
