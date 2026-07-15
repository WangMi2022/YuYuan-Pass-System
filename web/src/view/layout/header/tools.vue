<template>
  <div class="na-tools">
    <el-dropdown v-if="isDev" @command="toDoc">
      <button type="button" class="na-icon-button" aria-label="视频教程">
        <el-icon>
        <Film />
      </el-icon>
      </button>
      <template #dropdown>
        <el-dropdown-menu>
          <el-dropdown-item
            v-for="item in videoList"
            :key="item.link"
            :command="item.link"
            >{{ item.title }}</el-dropdown-item
          >
        </el-dropdown-menu>
      </template>
    </el-dropdown>

    <button type="button" class="na-search-trigger" aria-label="全局搜索" @click="handleCommand">
      <el-icon><Search /></el-icon>
      <span>搜索菜单与功能</span>
      <kbd>{{ first }} K</kbd>
    </button>

    <notification-center />

    <button type="button" class="na-icon-button" aria-label="系统设置" @click="toggleSetting">
      <el-icon>
        <Setting />
      </el-icon>
    </button>

    <button type="button" class="na-icon-button" aria-label="刷新页面" @click="toggleRefresh">
      <el-icon
          :class="showRefreshAnmite ? 'animate-spin' : ''"
      >
        <Refresh />
      </el-icon>
    </button>

    <button
      type="button"
      class="na-icon-button"
      :aria-label="appStore.isDark ? '切换到亮色主题' : '切换到深色主题'"
      @click="appStore.toggleTheme(!appStore.isDark)"
    >
      <el-icon
          v-if="appStore.isDark"
      >
        <Sunny />
      </el-icon>
      <el-icon
          v-else
      >
        <Moon />
      </el-icon>
    </button>

    <gva-setting v-model:drawer="showSettingDrawer"></gva-setting>
    <command-menu ref="command" />
  </div>
</template>

<script setup>
  import { useAppStore } from '@/pinia'
  import GvaSetting from '@/view/layout/setting/index.vue'
  import { ref } from 'vue'
  import { emitter } from '@/utils/bus.js'
  import CommandMenu from '@/components/commandMenu/index.vue'
  import { toDoc } from '@/utils/doc'
  import { isDev } from '@/utils/env.js'
  import NotificationCenter from '@/components/notification/NotificationCenter.vue'

  const appStore = useAppStore()
  const showSettingDrawer = ref(false)
  const showRefreshAnmite = ref(false)
  const toggleRefresh = () => {
    showRefreshAnmite.value = true
    emitter.emit('reload')
    setTimeout(() => {
      showRefreshAnmite.value = false
    }, 1000)
  }

  const toggleSetting = () => {
    showSettingDrawer.value = true
  }

  const first = ref('')
  const command = ref()

  const handleCommand = () => {
    command.value.open()
  }
  const initPage = () => {
    // 判断当前用户的操作系统
    if (window.localStorage.getItem('osType') === 'WIN') {
      first.value = 'Ctrl'
    } else {
      first.value = '⌘'
    }
    // 当用户同时按下ctrl和k键的时候
    const handleKeyDown = (e) => {
      if (e.ctrlKey && e.key === 'k') {
        // 阻止浏览器默认事件
        e.preventDefault()
        handleCommand()
      }
    }
    window.addEventListener('keydown', handleKeyDown)
  }

  initPage()

  const videoList = [
    {
      title: '1.clone项目和安装依赖',
      link: 'https://www.bilibili.com/video/BV1jx4y1s7xx'
    },
    {
      title: '2.初始化项目',
      link: 'https://www.bilibili.com/video/BV1sr421K7sv'
    },
    {
      title: '3.开启调试工具+创建初始化包',
      link: 'https://www.bilibili.com/video/BV1iH4y1c7Na'
    },
    {
      title: '4.手动使用自动化创建功能',
      link: 'https://www.bilibili.com/video/BV1UZ421T7fV'
    },
    {
      title: '5.使用已有表格创建业务',
      link: 'https://www.bilibili.com/video/BV1NE4m1977s'
    },
    {
      title: '6.使用AI创建业务和创建数据源模式的可选项',
      link: 'https://www.bilibili.com/video/BV17i421a7DE'
    },
    {
      title: '7.创建自己的后端方法',
      link: 'https://www.bilibili.com/video/BV1Yw4m1k7fg'
    },
    {
      title: '8.新增一个前端页面',
      link: 'https://www.bilibili.com/video/BV12y411i7oE'
    },
    {
      title: '9.配置一个前端二级页面',
      link: 'https://www.bilibili.com/video/BV1ZM4m1y7i3'
    },
    {
      title: '10.配置一个前端菜单参数',
      link: 'https://www.bilibili.com/video/BV1WS42197DZ'
    },
    {
      title: '11.菜单参数实战+动态菜单标题+菜单高亮配置',
      link: 'https://www.bilibili.com/video/BV1NE4m1979c'
    },
    {
      title: '12.增加菜单可控按钮',
      link: 'https://www.bilibili.com/video/BV1Sw4m1k746'
    },
    {
      title: '14.新增客户角色和其相关配置教学',
      link: 'https://www.bilibili.com/video/BV1Ki421a7X2'
    },
    {
      title: '15.发布项目上线',
      link: 'https://www.bilibili.com/video/BV1Lx4y1s77D'
    }
  ]
</script>

<style scoped lang="scss"></style>
