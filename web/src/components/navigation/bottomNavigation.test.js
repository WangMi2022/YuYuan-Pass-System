import assert from 'node:assert/strict'
import test from 'node:test'
import {
  containsActiveRoute,
  flattenLeafMenuItems,
  splitNavigationItems
} from './bottomNavigation.js'

const menus = [
  { name: 'dashboard', meta: { title: '首页' } },
  {
    name: 'assets',
    meta: { title: '资产' },
    children: [
      { name: 'assetList', meta: { title: '资产列表' } },
      { name: 'hiddenAsset', hidden: true, meta: { title: '隐藏资产' } }
    ]
  },
  { name: 'audit', meta: { title: '审计' } },
  { name: 'settings', meta: { title: '设置' } },
  { name: 'profile', meta: { title: '我的' } },
  { name: 'help', meta: { title: '帮助' } }
]

test('containsActiveRoute matches visible descendants', () => {
  assert.equal(containsActiveRoute(menus[1], 'assetList'), true)
  assert.equal(containsActiveRoute(menus[1], 'hiddenAsset'), false)
  assert.equal(containsActiveRoute(menus[1], 'dashboard'), false)
})

test('flattenLeafMenuItems preserves the visible parent trail', () => {
  assert.deepEqual(flattenLeafMenuItems([menus[1]]), [
    {
      item: menus[1].children[0],
      trail: ['资产']
    }
  ])
})

test('splitNavigationItems reserves the last action for overflow', () => {
  const result = splitNavigationItems(menus, 5)
  assert.deepEqual(
    result.primaryItems.map((item) => item.name),
    ['dashboard', 'assets', 'audit', 'settings']
  )
  assert.deepEqual(
    result.overflowItems.map((item) => item.name),
    ['profile', 'help']
  )
})
