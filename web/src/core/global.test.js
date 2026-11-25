import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import test from 'node:test'

const requiredBusinessMenuIcons = ['Warning', 'TrendCharts', 'ChatDotRound']

test('global icon registry includes every business menu icon used by the sidebar', async () => {
  const source = readFileSync(new URL('./global.js', import.meta.url), 'utf8')
  const registry = source.match(/const elIcons = \{([\s\S]*?)\n\s{2}\}/)
  assert.ok(registry, 'global icon registry was not found')

  const registeredIcons = new Set(registry[1].match(/[A-Z][A-Za-z0-9]*/g) || [])
  const iconLibrary = await import('@element-plus/icons-vue')

  for (const iconName of requiredBusinessMenuIcons) {
    assert.ok(iconLibrary[iconName], `${iconName} is not exported by Element Plus`)
    assert.ok(registeredIcons.has(iconName), `${iconName} is missing from the global sidebar icon registry`)
  }
})
