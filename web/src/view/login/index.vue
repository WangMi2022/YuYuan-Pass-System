<template>
  <div
    id="userLayout"
    class="na-auth-page"
    :class="{ 'has-custom-background': loginBackgroundUrl }"
    :style="loginPageStyle"
  >
    <a class="na-auth-brand" href="#/login" aria-label="资产管理系统登录页">
      <Logo :size="2" />
      <span>{{ $GIN_VUE_ADMIN.appName }}</span>
    </a>

    <div class="na-auth-glow na-auth-glow--one" aria-hidden="true" />
    <div class="na-auth-glow na-auth-glow--two" aria-hidden="true" />

    <main class="na-auth-main">
      <section class="na-auth-panel" aria-labelledby="login-title">
          <div>
            <div class="na-auth-heading">
              <span class="na-auth-logo"><Logo :size="2.5" /></span>
              <h1 id="login-title">登录资产管理系统</h1>
              <p>使用管理员或有效账号继续访问工作台</p>
            </div>
            <el-form
              ref="loginForm"
              :model="loginFormData"
              :rules="rules"
              :validate-on-rule-change="false"
              label-position="top"
              class="na-auth-form"
              @keyup.enter="submitForm"
            >
              <el-form-item label="用户名" prop="username">
                <el-input
                  v-model="loginFormData.username"
                  size="large"
                  placeholder="请输入用户名"
                  autocomplete="username"
                />
              </el-form-item>
              <el-form-item label="密码" prop="password">
                <el-input
                  v-model="loginFormData.password"
                  show-password
                  size="large"
                  type="password"
                  placeholder="请输入密码"
                  autocomplete="current-password"
                />
              </el-form-item>
              <el-form-item
                v-if="loginFormData.openCaptcha"
                label="验证码"
                prop="captcha"
              >
                <div class="na-captcha-row">
                  <el-input
                    v-model="loginFormData.captcha"
                    placeholder="请输入验证码"
                    size="large"
                    inputmode="numeric"
                    autocomplete="off"
                  />
                  <button type="button" class="na-captcha" aria-label="刷新验证码" @click="loginVerify()">
                    <img
                      v-if="picPath"
                      :src="picPath"
                      alt="登录验证码"
                    />
                  </button>
                </div>
              </el-form-item>
              <el-form-item class="na-submit-item">
                <el-button
                  class="na-login-button"
                  type="primary"
                  size="large"
                  :loading="submitting"
                  @click="submitForm"
                  >登录</el-button
                >
              </el-form-item>
              <el-form-item v-if="isDev">
                <el-button
                  class="na-login-button"
                  size="large"
                  @click="checkInit"
                  >前往初始化</el-button
                >
              </el-form-item>
            </el-form>
            <p class="na-auth-security">
              登录会话采用 JWT 鉴权，账号权限由系统角色统一控制。
            </p>
          </div>
      </section>
    </main>
  </div>
</template>

<script setup>
  import { captcha } from '@/api/user'
  import { checkDB } from '@/api/initdb'
  import { computed, onMounted, reactive, ref } from 'vue'
  import { ElMessage } from 'element-plus'
  import { useRouter } from 'vue-router'
  import { useUserStore } from '@/pinia/modules/user'
  import Logo from '@/components/logo/index.vue'
  import { isDev } from '@/utils/env.js'
  import { getCurrentLoginBackground } from '@/api/systemSettings'

  defineOptions({
    name: 'Login'
  })

  const router = useRouter()
  const loginBackgroundUrl = ref('')
  const loginPageStyle = computed(() => loginBackgroundUrl.value
    ? { '--na-login-background-image': `url(${JSON.stringify(loginBackgroundUrl.value)})` }
    : {})

  const loadLoginBackground = async () => {
    try {
      const res = await getCurrentLoginBackground()
      if (res.code === 0) loginBackgroundUrl.value = res.data?.url || ''
    } catch {
      loginBackgroundUrl.value = ''
    }
  }

  onMounted(loadLoginBackground)
  const captchaRequiredLength = ref(6)
  // 验证函数
  const checkUsername = (rule, value, callback) => {
    if (value.length < 5) {
      return callback(new Error('请输入正确的用户名'))
    } else {
      callback()
    }
  }
  const checkPassword = (rule, value, callback) => {
    if (value.length < 6) {
      return callback(new Error('请输入正确的密码'))
    } else {
      callback()
    }
  }
  const checkCaptcha = (rule, value, callback) => {
    if (!loginFormData.openCaptcha) {
      return callback()
    }
    const sanitizedValue = (value || '').replace(/\s+/g, '')
    if (!sanitizedValue) {
      return callback(new Error('请输入验证码'))
    }
    if (!/^\d+$/.test(sanitizedValue)) {
      return callback(new Error('验证码须为数字'))
    }
    if (sanitizedValue.length < captchaRequiredLength.value) {
      return callback(
        new Error(`请输入至少${captchaRequiredLength.value}位数字验证码`)
      )
    }
    if (sanitizedValue !== value) {
      loginFormData.captcha = sanitizedValue
    }
    callback()
  }

  // 获取验证码
  const loginVerify = async () => {
    const ele = await captcha()
    captchaRequiredLength.value = Number(ele.data?.captchaLength) || 0
    picPath.value = ele.data?.picPath
    loginFormData.captchaId = ele.data?.captchaId
    loginFormData.openCaptcha = ele.data?.openCaptcha
  }
  loginVerify()

  // 登录相关操作
  const loginForm = ref(null)
  const submitting = ref(false)
  const picPath = ref('')
  const loginFormData = reactive({
    username: 'admin',
    password: '',
    captcha: '',
    captchaId: '',
    openCaptcha: false
  })
  const rules = reactive({
    username: [{ validator: checkUsername, trigger: 'blur' }],
    password: [{ validator: checkPassword, trigger: 'blur' }],
    captcha: [{ validator: checkCaptcha, trigger: 'blur' }]
  })

  const userStore = useUserStore()
  const login = async () => {
    return await userStore.LoginIn(loginFormData)
  }
  const submitForm = () => {
    loginForm.value.validate(async (v) => {
      if (!v) {
        // 未通过前端静态验证
        ElMessage({
          type: 'error',
          message: '请正确填写登录信息',
          showClose: true
        })
        return false
      }

      // 通过验证，请求登陆
      submitting.value = true
      const flag = await login()
      submitting.value = false

      // 登陆失败，刷新验证码
      if (!flag) {
        await loginVerify()
        return false
      }

      // 登陆成功
      return true
    })
  }

  // 跳转初始化
  const checkInit = async () => {
    const res = await checkDB()
    if (res.code === 0) {
      if (res.data?.needInit) {
        userStore.NeedInit()
        await router.push({ name: 'Init' })
      } else {
        ElMessage({
          type: 'info',
          message: '已配置数据库信息，无法初始化'
        })
      }
    }
  }
</script>

<style scoped lang="scss">
  .na-auth-page {
    position: relative;
    width: 100%;
    height: 100%;
    overflow: hidden;
    background:
      linear-gradient(var(--na-border) 1px, transparent 1px),
      linear-gradient(90deg, var(--na-border) 1px, transparent 1px),
      var(--na-background);
    background-size: 48px 48px;
    color: var(--na-foreground);
  }
  .na-auth-page::after {
    position: absolute;
    inset: 0;
    content: '';
    background: linear-gradient(to bottom, color-mix(in srgb, var(--na-background) 35%, transparent), var(--na-background) 78%);
    pointer-events: none;
  }
  .na-auth-page.has-custom-background {
    background-color: #0f172a;
    background-image: linear-gradient(rgb(15 23 42 / 24%), rgb(15 23 42 / 42%)), var(--na-login-background-image);
    background-position: center;
    background-size: cover;
  }
  .na-auth-page.has-custom-background::after {
    background: linear-gradient(to bottom, rgb(15 23 42 / 6%), rgb(15 23 42 / 26%));
  }
  .na-auth-page.has-custom-background .na-auth-brand {
    color: #fff;
    text-shadow: 0 1px 8px rgb(0 0 0 / 32%);
  }
  .na-auth-brand {
    position: absolute;
    z-index: 10;
    top: 28px;
    left: 32px;
    display: flex;
    align-items: center;
    gap: 10px;
    color: var(--na-foreground);
    font-size: 16px;
    font-weight: 650;
    letter-spacing: -.01em;
    transition: opacity 180ms ease;
  }
  .na-auth-brand:hover { opacity: .75; }
  .na-auth-main {
    position: relative;
    z-index: 3;
    display: grid;
    place-items: center;
    width: 100%;
    height: 100%;
    padding: 84px 20px 72px;
  }
  .na-auth-panel {
    width: min(100%, 448px);
    padding: 34px 36px 30px;
    border: 1px solid var(--na-border);
    border-radius: var(--na-radius-lg);
    background: color-mix(in srgb, var(--na-card) 94%, transparent);
    box-shadow: 0 18px 70px rgb(0 0 0 / 8%);
    backdrop-filter: blur(14px);
  }
  .na-auth-heading { margin-bottom: 28px; text-align: center; }
  .na-auth-logo {
    display: inline-grid;
    place-items: center;
    width: 52px;
    height: 52px;
    margin-bottom: 18px;
    border: 1px solid var(--na-border);
    border-radius: 14px;
    background: var(--na-card);
    box-shadow: var(--na-shadow-sm);
  }
  .na-auth-heading h1 { margin: 0; color: var(--na-foreground); font-size: 24px; font-weight: 680; letter-spacing: -.025em; }
  .na-auth-heading p { margin: 8px 0 0; color: var(--na-muted-foreground); font-size: 13px; }
  .na-auth-form :deep(.el-form-item) { margin-bottom: 20px; }
  .na-auth-form :deep(.el-form-item__label) { padding-bottom: 7px; }
  .na-auth-form :deep(.el-input__wrapper) { min-height: 44px; }
  .na-captcha-row { display: grid; grid-template-columns: 1fr 128px; gap: 10px; width: 100%; }
  .na-captcha {
    overflow: hidden;
    height: 44px;
    padding: 0;
    border: 1px solid var(--na-input);
    border-radius: var(--na-radius-sm);
    background: var(--na-muted);
  }
  .na-captcha img { width: 100%; height: 100%; object-fit: cover; }
  .na-submit-item { margin-top: 4px; margin-bottom: 14px !important; }
  .na-login-button { width: 100%; min-height: 44px; }
  .na-auth-security { margin: 5px 0 0; color: var(--na-muted-foreground); font-size: 11px; line-height: 1.6; text-align: center; }
  .na-auth-glow { position: absolute; z-index: 1; border-radius: 50%; filter: blur(2px); pointer-events: none; }
  .na-auth-glow--one { top: -150px; left: 12%; width: 460px; height: 460px; background: radial-gradient(circle, rgb(63 158 232 / 13%), transparent 68%); }
  .na-auth-glow--two { right: 8%; bottom: -190px; width: 520px; height: 520px; background: radial-gradient(circle, rgb(120 95 220 / 9%), transparent 70%); }
  :global(html.dark) .na-auth-panel { box-shadow: 0 24px 80px rgb(0 0 0 / 28%); }
  @media (max-width: 640px) {
    .na-auth-brand { top: 20px; left: 20px; }
    .na-auth-main { padding-inline: 14px; }
    .na-auth-panel { padding: 28px 20px 24px; }
    .na-captcha-row { grid-template-columns: 1fr 112px; }
  }
</style>
