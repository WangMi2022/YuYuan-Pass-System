<template>
  <div class="settings-module">
    <header class="settings-module__header">
      <p class="settings-module__eyebrow">WORKSPACE</p>
      <h3>布局设置</h3>
      <p>选择导航结构，并精确控制侧边栏与页面切换方式。</p>
    </header>

    <section class="settings-section" aria-labelledby="layout-mode-title">
      <div class="settings-section__header">
        <h4 id="layout-mode-title">布局模式</h4>
        <p>切换后立即应用，菜单权限和业务页面不会发生变化。</p>
      </div>
      <LayoutModeCard v-model="config.side_mode" @update:modelValue="appStore.toggleSideMode" />
    </section>

    <section class="settings-section" aria-labelledby="interface-title">
      <div class="settings-section__header">
        <h4 id="interface-title">界面行为</h4>
        <p>控制标签导航和页面内容切换效果。</p>
      </div>
      <div class="settings-panel">
        <SettingItem label="显示标签页" description="在顶部保留已打开页面的快捷入口">
          <el-switch v-model="config.showTabs" aria-label="显示标签页" @change="appStore.toggleTabs" />
        </SettingItem>
        <SettingItem label="页面切换动画" description="选择页面之间的过渡方式">
          <el-select v-model="config.transition_type" aria-label="页面切换动画" @change="appStore.toggleTransition">
            <el-option value="fade" label="淡入淡出" />
            <el-option value="slide" label="滑动" />
            <el-option value="zoom" label="缩放" />
            <el-option value="none" label="无动画" />
          </el-select>
        </SettingItem>
      </div>
    </section>

    <section class="settings-section" aria-labelledby="navigation-size-title">
      <div class="settings-section__header">
        <h4 id="navigation-size-title">导航尺寸</h4>
        <p>数值单位为像素，修改后实时更新工作区布局。</p>
      </div>
      <div class="settings-panel">
        <SettingItem label="侧边栏展开宽度" description="完整显示菜单名称时的宽度">
          <el-input-number
            v-model="config.layout_side_width"
            :min="150"
            :max="400"
            :step="10"
            controls-position="right"
            aria-label="侧边栏展开宽度"
            @change="appStore.toggleConfigSideWidth"
          />
        </SettingItem>
        <SettingItem label="侧边栏收起宽度" description="仅显示菜单图标时的宽度">
          <el-input-number
            v-model="config.layout_side_collapsed_width"
            :min="60"
            :max="100"
            controls-position="right"
            aria-label="侧边栏收起宽度"
            @change="appStore.toggleConfigSideCollapsedWidth"
          />
        </SettingItem>
        <SettingItem label="菜单项高度" description="控制侧边栏菜单的垂直密度">
          <el-input-number
            v-model="config.layout_side_item_height"
            :min="30"
            :max="50"
            controls-position="right"
            aria-label="菜单项高度"
            @change="appStore.toggleConfigSideItemHeight"
          />
        </SettingItem>
      </div>
    </section>
  </div>
</template>

<script setup>
import { storeToRefs } from 'pinia'
import { useAppStore } from '@/pinia'
import LayoutModeCard from '../../components/layoutModeCard.vue'
import SettingItem from '../../components/settingItem.vue'

defineOptions({
  name: 'LayoutSettings'
})

const appStore = useAppStore()
const { config } = storeToRefs(appStore)
</script>
