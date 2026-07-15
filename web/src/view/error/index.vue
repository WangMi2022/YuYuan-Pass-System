<template>
  <main class="na-error-page">
      <section class="na-error-card">
        <img src="../../assets/404.png" alt="页面不存在" />
        <h1>页面不存在或暂无权限</h1>
        <p>
          常见问题为当前此角色无当前路由，如果确定要使用本路由，请到角色管理进行分配
        </p>
        <el-button type="primary" @click="toDashboard">返回工作台</el-button>
      </section>
  </main>
</template>

<script setup>
  import { useUserStore } from '@/pinia/modules/user'
  import { useRouter } from 'vue-router'
  import { emitter } from '@/utils/bus'

  defineOptions({
    name: 'Error'
  })

  const userStore = useUserStore()
  const router = useRouter()
  const toDashboard = () => {
    try {
      router.push({ name: userStore.userInfo.authority.defaultRouter })
    } catch (error) {
        emitter.emit('show-error', {
        code: '401',
        message: "检测到其他用户修改了路由权限，请重新登录",
        fn: () => {
          userStore.ClearStorage()
          router.push({ name: 'Login', replace: true })
        }
      })
    }
  }
</script>

<style scoped lang="scss">
.na-error-page { display:grid; place-items:center; width:100%; min-height:100%; padding:24px; background:var(--na-background); color:var(--na-foreground); }
.na-error-card { display:flex; align-items:center; flex-direction:column; width:min(100%,620px); padding:36px; border:1px solid var(--na-border); border-radius:var(--na-radius-lg); background:var(--na-card); box-shadow:var(--na-shadow-sm); text-align:center; }
.na-error-card img { width:min(100%,280px); margin-bottom:20px; }
.na-error-card h1 { margin:0; font-size:24px; font-weight:650; }
.na-error-card p { max-width:520px; margin:10px 0 24px; color:var(--na-muted-foreground); font-size:13px; line-height:1.7; }
</style>
