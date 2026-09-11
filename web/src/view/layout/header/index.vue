<template>
  <header class="na-header">
    <div class="na-header-left">
      <button
        type="button"
        class="na-brand"
        aria-label="返回工作台"
        @click="router.push({ path: '/' })"
      >
        <Logo class="na-brand-logo" />
        <span
          v-if="!isMobile"
          class="na-brand-name"
        >
          <strong>{{ brandingStore.systemName }}</strong>
          <small v-if="brandingStore.subtitle">{{ brandingStore.subtitle }}</small>
        </span>
      </button>

      <el-breadcrumb
        v-show="!isMobile"
        v-if="config.side_mode !== 'head' && config.side_mode !== 'combination'"
        aria-label="面包屑导航"
      >
        <el-breadcrumb-item
          v-for="item in matched.slice(1, matched.length)"
          :key="item.path"
        >
          {{ fmtTitle(item.meta.title, route) }}
        </el-breadcrumb-item>
      </el-breadcrumb>
      <gva-aside
        v-if="config.side_mode === 'head' && !isMobile"
        class="flex-1"
      />
      <gva-aside
        v-if="config.side_mode === 'combination' && !isMobile"
        mode="head"
        class="flex-1"
      />
    </div>

    <div class="na-header-right">
      <tools />
      <div class="na-header-divider" aria-hidden="true" />
      <el-dropdown
        trigger="click"
        popper-class="na-user-dropdown-popper"
        placement="bottom-end"
      >
        <button type="button" class="na-profile-trigger" aria-label="打开用户菜单">
          <CustomPic class="na-profile-pic" />
          <span v-show="!isMobile" class="na-profile-name">{{
            userStore.userInfo.nickName || userStore.userInfo.userName
          }}</span>
          <el-icon class="na-profile-arrow">
            <ArrowDown />
          </el-icon>
        </button>
        <template #dropdown>
          <el-dropdown-menu class="na-user-dropdown-menu">
            <!-- 1. 用户基本信息展示卡片 -->
            <div class="dropdown-user-header">
              <div class="user-info-row">
                <CustomPic class="header-avatar" />
                <div class="user-meta">
                  <span class="user-nickname">{{ userStore.userInfo.nickName || userStore.userInfo.userName }}</span>
                  <span class="user-handle font-mono">@{{ userStore.userInfo.userName || 'user' }}</span>
                </div>
              </div>
              <div class="active-role-chip">
                <span class="role-beacon" />
                <span class="role-chip-label">当前角色</span>
                <span class="role-chip-name">{{ userStore.userInfo.authority?.authorityName }}</span>
              </div>
            </div>

            <!-- 2. 角色切换区域 -->
            <template v-if="otherAuthorities.length">
              <div class="dropdown-divider" />
              <div class="dropdown-section-title">
                <span>切换所属角色 ({{ otherAuthorities.length }})</span>
              </div>
              <el-dropdown-item
                v-for="item in otherAuthorities"
                :key="item.authorityId"
                class="role-switch-item"
                @click="changeUserAuth(item.authorityId)"
              >
                <div class="role-switch-row">
                  <div class="role-name-wrap">
                    <el-icon class="role-icon"><Stamp /></el-icon>
                    <span class="role-title">{{ item.authorityName }}</span>
                  </div>
                  <span class="role-switch-badge">切换</span>
                </div>
              </el-dropdown-item>
            </template>

            <div class="dropdown-divider" />

            <!-- 3. 功能操作入口 -->
            <div class="dropdown-section-title">
              <span>快捷导航与管理</span>
            </div>
            <el-dropdown-item class="menu-action-item" @click="toPerson">
              <el-icon class="action-icon is-user"><User /></el-icon>
              <span>个人中心与凭证设置</span>
            </el-dropdown-item>
            <el-dropdown-item class="menu-action-item" @click="toCalendar">
              <el-icon class="action-icon is-calendar"><Calendar /></el-icon>
              <span>我的工作日历日程</span>
            </el-dropdown-item>

            <div class="dropdown-divider" />

            <!-- 4. 安全登出 -->
            <el-dropdown-item class="menu-action-item is-logout" @click="userStore.LoginOut">
              <el-icon class="action-icon is-danger"><SwitchButton /></el-icon>
              <span>安全退出登录</span>
            </el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>
  </header>
</template>

<script setup>
  import tools from './tools.vue'
  import CustomPic from '@/components/customPic/index.vue'
  import { useUserStore } from '@/pinia/modules/user'
  import { useRoute, useRouter } from 'vue-router'
  import { useAppStore, useBrandingStore } from '@/pinia'
  import { storeToRefs } from 'pinia'
  import { computed } from 'vue'
  import { setUserAuthority } from '@/api/user'
  import { fmtTitle } from '@/utils/fmtRouterTitle'
  import gvaAside from '@/view/layout/aside/index.vue'
  import Logo from '@/components/logo/index.vue'
  import {
    ArrowDown,
    Calendar,
    Stamp,
    SwitchButton,
    User
  } from '@element-plus/icons-vue'

  const userStore = useUserStore()
  const router = useRouter()
  const route = useRoute()
  const appStore = useAppStore()
  const brandingStore = useBrandingStore()
  const { device, config } = storeToRefs(appStore)
  const isMobile = computed(() => {
    return device.value === 'mobile'
  })
  const toPerson = () => {
    router.push({ name: 'person' })
  }
  const toCalendar = () => {
    if (router.hasRoute('workSchedule')) {
      router.push({ name: 'workSchedule' })
    } else if (router.hasRoute('workCalendar')) {
      router.push({ name: 'workCalendar' })
    }
  }
  const matched = computed(() => route.meta.matched)

  const otherAuthorities = computed(() => {
    if (!userStore.userInfo.authorities) return []
    return userStore.userInfo.authorities.filter(
      (i) => i.authorityId !== userStore.userInfo.authorityId
    )
  })

  const changeUserAuth = async (id) => {
    const res = await setUserAuthority({
      authorityId: id
    })
    if (res.code === 0) {
      window.sessionStorage.setItem('needCloseAll', 'true')
      window.sessionStorage.setItem('needToHome', 'true')
      window.location.reload()
    }
  }
</script>

<style scoped lang="scss"></style>
