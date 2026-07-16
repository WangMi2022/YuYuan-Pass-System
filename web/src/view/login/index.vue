<template>
  <div
    id="userLayout"
    class="na-auth-page"
  >
    <section class="na-auth-visual" :style="loginPageStyle" aria-label="系统登录背景">
      <a class="na-auth-brand" href="#/login" aria-label="资产管理系统登录页">
        <img
          v-if="loginLogoUrl"
          class="na-auth-brand-image"
          :src="loginLogoUrl"
          :alt="`${$GIN_VUE_ADMIN.appName}图标`"
          @error="handleLogoError"
        />
        <Logo v-else :size="2" />
        <span>{{ $GIN_VUE_ADMIN.appName }}</span>
      </a>
    </section>

    <main class="na-auth-main">
      <section class="na-auth-panel" aria-labelledby="login-title">
          <div>
            <div class="na-auth-heading">
              <span class="na-auth-logo">
                <img
                  v-if="loginLogoUrl"
                  :src="loginLogoUrl"
                  :alt="`${$GIN_VUE_ADMIN.appName}登录图标`"
                  @error="handleLogoError"
                />
                <Logo v-else :size="2.5" />
              </span>
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
  import { getCurrentLoginBackground, getCurrentLoginLogo } from '@/api/systemSettings'
  import defaultBackground from '@/assets/login_background.jpg'

  defineOptions({
    name: 'Login'
  })

  const router = useRouter()
  const loginBackgroundUrl = ref('')
  const loginLogoUrl = ref('')
  const backgroundUrl = computed(() => loginBackgroundUrl.value || defaultBackground)
  const loginPageStyle = computed(() => ({
    '--na-login-background-image': `url(${JSON.stringify(backgroundUrl.value)})`
  }))

  const loadLoginBackground = async () => {
    try {
      const res = await getCurrentLoginBackground()
      if (res.code === 0) loginBackgroundUrl.value = res.data?.url || ''
    } catch {
      loginBackgroundUrl.value = ''
    }
  }

  const loadLoginLogo = async () => {
    try {
      const res = await getCurrentLoginLogo()
      if (res.code === 0) loginLogoUrl.value = res.data?.url || ''
    } catch {
      loginLogoUrl.value = ''
    }
  }

  const handleLogoError = () => {
    loginLogoUrl.value = ''
  }

  onMounted(() => {
    loadLoginBackground()
    loadLoginLogo()
  })
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
    display: grid;
    grid-template-columns: minmax(0, 1fr) clamp(440px, 34vw, 520px);
    width: 100%;
    height: 100%;
    overflow: hidden;
    background: var(--na-background);
    color: var(--na-foreground);
  }
  .na-auth-visual {
    position: relative;
    min-width: 0;
    min-height: 0;
    background-color: #0f172a;
    background-image: linear-gradient(rgb(15 23 42 / 10%), rgb(15 23 42 / 24%)), var(--na-login-background-image);
    background-position: center;
    background-size: cover;
  }
  .na-auth-brand {
    position: absolute;
    top: 28px;
    left: 32px;
    display: flex;
    align-items: center;
    gap: 10px;
    color: #fff;
    font-size: 16px;
    font-weight: 650;
    letter-spacing: -.01em;
    text-shadow: 0 1px 8px rgb(0 0 0 / 32%);
  }
  .na-auth-brand-image { width: 32px; height: 32px; border-radius: 8px; object-fit: contain; }
  .na-auth-main {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 100%;
    height: 100%;
    min-width: 0;
    padding: 56px 48px;
    border-left: 1px solid var(--na-border);
    background: var(--na-card);
  }
  .na-auth-panel {
    width: min(100%, 400px);
  }
  .na-auth-heading { margin-bottom: 32px; text-align: center; }
  .na-auth-logo {
    display: grid;
    place-items: center;
    width: 52px;
    height: 52px;
    margin: 0 auto 18px;
    border: 1px solid var(--na-border);
    border-radius: 14px;
    background: var(--na-card);
    box-shadow: var(--na-shadow-sm);
  }
  .na-auth-logo > img { width: 36px; height: 36px; object-fit: contain; }
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
  .na-auth-security { margin: 5px 0 0; color: var(--na-muted-foreground); font-size: 11px; line-height: 1.6; text-align: left; }
  @media (max-width: 800px) {
    .na-auth-page {
      grid-template-columns: minmax(0, 1fr);
      grid-template-rows: clamp(150px, 26vh, 190px) minmax(0, 1fr);
      height: auto;
      min-height: 100%;
      overflow-x: hidden;
      overflow-y: auto;
    }
    .na-auth-brand { top: 20px; left: 20px; }
    .na-auth-main {
      min-height: calc(100vh - 190px);
      height: auto;
      padding: 40px 24px;
      border-top: 1px solid var(--na-border);
      border-left: 0;
    }
  }
  @media (max-width: 480px) {
    .na-auth-main { padding: 32px 20px; }
    .na-captcha-row { grid-template-columns: 1fr 112px; }
  }
</style>
