import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import test from 'node:test'
import { fileURLToPath } from 'node:url'
import { compileScript, parse } from '@vue/compiler-sfc'

test('category page exposes every keyword v-model root from script setup', async () => {
  const filename = fileURLToPath(new URL('./categories.vue', import.meta.url))
  const source = await readFile(filename, 'utf8')
  const { descriptor } = parse(source, { filename })

  assert.ok(descriptor.template, 'categories.vue must have a template')
  assert.ok(descriptor.scriptSetup, 'categories.vue must use script setup')

  const bindings = compileScript(descriptor, { id: 'asset-categories-binding-test' }).bindings
  const modelRoots = [...descriptor.template.content.matchAll(/v-model(?::[\w-]+)?="([A-Za-z_$][\w$]*)\.keyword"/g)]
    .map((match) => match[1])
  const missingBindings = [...new Set(modelRoots)].filter((name) => !bindings[name])

  assert.deepEqual(missingBindings, [], `missing script setup bindings: ${missingBindings.join(', ')}`)
})
