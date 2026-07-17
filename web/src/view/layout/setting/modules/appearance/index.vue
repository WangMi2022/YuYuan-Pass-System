<template>
  <div class="settings-module">
    <header class="settings-module__header">
      <p class="settings-module__eyebrow">APPEARANCE</p>
      <h3>外观设置</h3>
      <p>调整工作区的主题模式、品牌色彩与显示辅助选项。</p>
    </header>

    <section class="settings-section" aria-labelledby="theme-mode-title">
      <div class="settings-section__header">
        <h4 id="theme-mode-title">主题模式</h4>
        <p>根据当前环境选择浅色、深色或自动模式。</p>
      </div>
      <ThemeModeSelector v-model="config.darkMode" @update:modelValue="appStore.toggleDarkMode" />
    </section>

    <section class="settings-section" aria-labelledby="theme-color-title">
      <div class="settings-section__header">
        <h4 id="theme-color-title">主题颜色</h4>
        <p>主题色会应用到导航、按钮、选中状态和焦点提示。</p>
      </div>
      <ThemeColorPicker v-model="config.primaryColor" @update:modelValue="appStore.togglePrimaryColor" />
    </section>

    <section class="settings-section" aria-labelledby="component-size-title">
      <div class="settings-section__header">
        <h4 id="component-size-title">组件尺寸</h4>
        <p>统一调整表单、按钮和交互控件的显示密度。</p>
      </div>
      <div class="settings-panel">
        <SettingItem label="全局组件尺寸" description="影响系统内 Element Plus 组件尺寸">
          <el-select v-model="config.global_size" aria-label="全局组件尺寸" @change="appStore.toggleGlobalSize">
            <el-option label="标准" value="default" />
            <el-option label="宽松" value="large" />
            <el-option label="紧凑" value="small" />
          </el-select>
        </SettingItem>
      </div>
    </section>

    <section class="settings-section" aria-labelledby="accessibility-title">
      <div class="settings-section__header">
        <h4 id="accessibility-title">视觉辅助</h4>
        <p>根据阅读习惯调整页面呈现，不改变业务数据。</p>
      </div>
      <div class="settings-panel">
        <SettingItem label="灰色模式" description="降低界面色彩饱和度">
          <el-switch v-model="config.grey" aria-label="灰色模式" @change="appStore.toggleGrey" />
        </SettingItem>
        <SettingItem label="色弱模式" description="增强关键内容的色彩辨识度">
          <el-switch v-model="config.weakness" aria-label="色弱模式" @change="appStore.toggleWeakness" />
        </SettingItem>
        <SettingItem label="页面水印" description="在工作区显示当前用户水印">
          <el-switch v-model="config.show_watermark" aria-label="页面水印" @change="appStore.toggleConfigWatermark" />
        </SettingItem>
      </div>
    </section>
  </div>
</template>

<script setup>
import { storeToRefs } from 'pinia'
import { useAppStore } from '@/pinia'
import ThemeModeSelector from '../../components/themeModeSelector.vue'
import ThemeColorPicker from '../../components/themeColorPicker.vue'
import SettingItem from '../../components/settingItem.vue'

defineOptions({
  name: 'AppearanceSettings'
})

const appStore = useAppStore()
const { config } = storeToRefs(appStore)
</script>
