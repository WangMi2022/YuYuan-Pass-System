import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import test from 'node:test'
import { compile } from '@vue/compiler-dom'
import { babelParse, compileScript, parse } from '@vue/compiler-sfc'
import * as Vue from 'vue'
import * as VueRouter from 'vue-router'

const dashboardSource = parse(readFileSync(new URL('./index.vue', import.meta.url), 'utf8')).descriptor
const layoutTemplate = parse(readFileSync(new URL('../layout/index.vue', import.meta.url), 'utf8')).descriptor.template.content
const routeTemplate = layoutTemplate.slice(layoutTemplate.indexOf('<router-view'), layoutTemplate.indexOf('</router-view>') + '</router-view>'.length)

// Run the real dashboard and layout route template with Vue's renderer.
// Only business API responses and leaf UI are fixtures.
function renderTemplate(template, runtime = Vue) {
  return new Function('Vue', compile(template, { mode: 'function', prefixIdentifiers: true }).code)(runtime)
}

function compileDashboard(resolveImport) {
  let source = compileScript(dashboardSource, { id: 'dashboard-test', inlineTemplate: true }).content
  const imports = babelParse(source, { sourceType: 'module' }).program.body.filter((statement) => statement.type === 'ImportDeclaration')
  const names = []
  const values = []
  for (const declaration of imports) {
    for (const specifier of declaration.specifiers) {
      names.push(specifier.local.name)
      values.push(resolveImport(declaration.source.value, specifier.imported?.name || 'default'))
    }
  }
  for (const declaration of imports.reverse()) source = source.slice(0, declaration.start) + source.slice(declaration.end)
  return new Function(...names, source.replace('export default', 'return'))(...values)
}

function createHost() {
  const node = (type, text = '') => ({ type, text, children: [], props: {}, parent: null })
  const body = node('body')
  const root = node('app')
  const remove = (child) => {
    if (child.parent) child.parent.children.splice(child.parent.children.indexOf(child), 1)
    child.parent = null
  }
  const renderer = Vue.createRenderer({
    createElement: node,
    createText: (text) => node('#text', text),
    createComment: (text) => node('#comment', text),
    setText: (child, text) => { child.text = text },
    setElementText: (child, text) => { child.text = text; child.children = [] },
    patchProp: (child, key, previous, value) => { child.props[key] = value },
    parentNode: (child) => child.parent,
    nextSibling: (child) => child.parent?.children[child.parent.children.indexOf(child) + 1] || null,
    querySelector: (selector) => selector === 'body' ? body : null,
    remove,
    insert: (child, parent, anchor = null) => {
      remove(child)
      const index = anchor ? parent.children.indexOf(anchor) : -1
      parent.children.splice(index < 0 ? parent.children.length : index, 0, child)
      child.parent = parent
    }
  })
  const find = (parent, predicate) => predicate(parent) ? parent : parent.children.map((child) => find(child, predicate)).find(Boolean)
  return { ...renderer, root, body, find }
}

for (const cached of [false, true]) {
  for (const direct of [false, true]) {
    test(`wallboard exit restores dashboard (cached=${cached}, direct=${direct})`, async (context) => {
      const host = createHost()
      const errors = []
      const previousWindow = globalThis.window
      globalThis.window = { setInterval: (...args) => setInterval(...args).unref(), clearInterval }
      context.after(() => {
        if (previousWindow === undefined) delete globalThis.window
        else globalThis.window = previousWindow
      })
      let wallboardOpen = false
      const Wallboard = {
        props: ['snapshot', 'loading'],
        emits: ['exit', 'refresh'],
        setup(props, { emit }) {
          Vue.onMounted(() => { wallboardOpen = true })
          Vue.onActivated(() => { wallboardOpen = true })
          Vue.onBeforeUnmount(() => { wallboardOpen = false })
          Vue.onDeactivated(() => { wallboardOpen = false })
          return () => Vue.h(Vue.Teleport, { to: 'body' }, [Vue.h('button', { onClick: () => emit('exit') }, 'Exit')])
        }
      }
      const Button = { setup: (props, { slots }) => () => Vue.h('button', slots.default?.()) }
      const Dashboard = compileDashboard((source, name) => {
        if (source === 'vue') return Vue[name]
        if (source === 'vue-router') return VueRouter[name]
        if (source === '@element-plus/icons-vue') return Button
        if (source === '@/view/dashboard/LeadershipWallboard.vue') return Wallboard
        if (source === '@/view/dashboard/wallboard/AnimatedValue.vue') return { props: ['value', 'format', 'animated'], render: () => Vue.h('span') }
        if (source === '@/components/three/HeroCanvas.vue') return { render: () => Vue.h('div', { class: 'workbench-hero-canvas' }) }
        if (source === '@/view/dashboard/PendingTasks.vue') return { render: () => Vue.h('main', { class: 'pending-tasks' }) }
        if (source === '@/components/page/AppPageHeader.vue') return { setup: (props, { slots }) => () => Vue.h('header', slots.actions?.()) }
        if (source === '@/pinia/modules/user') return () => ({ userInfo: { nickName: 'Test' } })
        if (source === '@/utils/format' || source === '@/plugin/invoice/utils/invoice') return String
        if (source === '@/utils/workCalendar') return () => ''
        if (source.startsWith('@/api/') || source.includes('/api/')) return async () => ({ code: 0, data: {} })
        throw new Error(`Unexpected dashboard import: ${source}`)
      })
      const router = VueRouter.createRouter({
        history: VueRouter.createMemoryHistory(),
        routes: [
          { name: 'dashboard', path: '/layout/dashboard', component: Dashboard },
          { name: 'assetDashboard', path: '/assets', component: { render: () => Vue.h('main', { class: 'other-route' }) } }
        ]
      })
      const Transition = (props, { slots }) => Vue.h(Vue.Transition, {
        ...props,
        css: false,
        onLeave: (element, done) => queueMicrotask(() => {
          try { done() } catch (error) { errors.push(error.message) }
        })
      }, slots)
      const app = host.createApp({
        components: { RouterView: VueRouter.RouterView },
        setup: () => ({ reloadFlag: true, config: { transition_type: 'fade' }, routerStore: { keepAliveRouters: cached ? ['Dashboard'] : [] } }),
        render: renderTemplate(routeTemplate, { ...Vue, Transition })
      })
      app.config.errorHandler = (error) => errors.push(error.message)
      app.config.warnHandler = (warning) => errors.push(warning)
      app.component('el-button', Button).component('el-icon', Button)
      app.use(router)
      await router.push({ path: '/layout/dashboard', query: { scope: 'test', ...(direct ? { view: 'wallboard' } : {}) } })
      await router.isReady()
      app.mount(host.root)
      const settle = async () => { await Vue.nextTick(); await new Promise((resolve) => setImmediate(resolve)); await Vue.nextTick() }
      const hasDashboard = () => Boolean(host.find(host.root, (child) => child.props.class === 'dashboard-page'))
      try {
        for (let cycle = 0; cycle < 3; cycle++) {
          if (!direct || cycle > 0) {
            assert.ok(hasDashboard(), 'dashboard is mounted before entering the wallboard')
            await host.find(host.root, (child) => child.type === 'button' && child.props.onClick)?.props.onClick()
            await settle()
          }
          assert.equal(wallboardOpen, true, JSON.stringify(errors))
          const exit = host.find(host.body, (child) => child.type === 'button')
          assert.ok(exit, 'wallboard is teleported to the body')
          assert.equal(host.body.children.filter((child) => child.type === 'button').length, 1, 'cached dashboards must not leave duplicate wallboards')
          exit.props.onClick()
          await settle()
          assert.equal(router.currentRoute.value.query.view, undefined)
          assert.equal(router.currentRoute.value.query.scope, 'test')
          assert.equal(wallboardOpen, false)
          assert.ok(hasDashboard(), `dashboard must be mounted after exiting the wallboard: ${JSON.stringify(errors)}`)
          assert.equal(host.find(host.body, (child) => child.type === 'button'), undefined)
          assert.deepEqual(errors, [])
        }
        await router.push({ path: '/layout/dashboard', query: { view: 'pending' } })
        await settle()
        assert.ok(host.find(host.root, (child) => child.props.class === 'pending-tasks'))
        await router.replace('/layout/dashboard')
        await settle()
        assert.ok(hasDashboard(), 'pending view still returns to the dashboard')
        await router.push('/layout/dashboard?view=wallboard')
        await settle()
        await router.push('/assets?view=wallboard')
        await settle()
        assert.ok(host.find(host.root, (child) => child.props.class === 'other-route'))
        assert.equal(wallboardOpen, false, 'deactivated dashboards must release the overlay')
        assert.equal(host.find(host.body, (child) => child.type === 'button'), undefined)
        await router.push('/layout/dashboard?view=wallboard')
        await settle()
        assert.equal(wallboardOpen, true, 'reactivating the dashboard restores its wallboard')
        assert.equal(host.body.children.filter((child) => child.type === 'button').length, 1)
        host.find(host.body, (child) => child.type === 'button').props.onClick()
        await settle()
        assert.ok(hasDashboard())
        assert.deepEqual(errors, [])
      } finally {
        try { app.unmount() } catch (error) { errors.push(error.message) }
      }
      assert.deepEqual(errors, [])
    })
  }
}
