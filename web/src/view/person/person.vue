<template>
  <div class="na-page profile-container">
    <section class="na-panel profile-hero" aria-labelledby="profile-title">
      <div class="profile-banner">
        <div class="profile-banner__copy">
          <span class="profile-kicker">账户档案</span>
          <p>维护个人身份、联系方式与登录安全设置</p>
        </div>
        <el-tag :type="accountStatus.type" effect="light" class="profile-status-tag">
          {{ accountStatus.label }}
        </el-tag>
      </div>

      <div class="profile-identity">
        <div class="profile-avatar-wrapper">
          <SelectImage
            v-model="userStore.userInfo.headerImg"
            file-type="image"
            :preview-url="userStore.userInfo.headerImgPreviewUrl"
            :loading="savingAvatar"
            rounded
          />
          <span class="profile-avatar-note">点击头像更换</span>
        </div>

        <div class="profile-identity__main">
          <div class="profile-name-row">
            <template v-if="!editFlag">
              <h1 id="profile-title">{{ displayName }}</h1>
              <el-button text circle :icon="Edit" aria-label="编辑昵称" @click="openEdit" />
            </template>
            <template v-else>
              <el-input v-model="nickName" class="profile-name-input" aria-label="昵称" />
              <el-button type="primary" plain :loading="savingNickname" @click="enterEdit">确认</el-button>
              <el-button plain @click="closeEdit">取消</el-button>
            </template>
          </div>
          <p class="profile-account-line">
            <span>@{{ userStore.userInfo.userName || '未设置账号' }}</span>
            <span class="profile-dot">·</span>
            <span>{{ primaryRole }}</span>
          </p>
          <div class="profile-identity__tags">
            <el-tag type="info" effect="plain">账号 ID {{ userId }}</el-tag>
            <el-tag v-if="roleNames.length" type="primary" effect="plain">{{ roleNames.length }} 个权限角色</el-tag>
          </div>
        </div>

        <dl class="profile-hero-meta">
          <div>
            <dt>加入系统</dt>
            <dd>{{ createdAtText }}</dd>
          </div>
          <div>
            <dt>资料完整度</dt>
            <dd>{{ profileCompletion }}%</dd>
          </div>
        </dl>
      </div>
    </section>

    <div class="profile-content">
      <main class="profile-main">
        <section class="na-panel profile-card" aria-labelledby="account-info-title">
          <header class="profile-card__header">
            <div>
              <h2 id="account-info-title">账号资料</h2>
              <p>用于识别当前账号和权限范围的信息</p>
            </div>
            <el-icon class="profile-card__icon"><User /></el-icon>
          </header>
          <div class="profile-detail-grid">
            <div class="profile-detail-item">
              <span class="profile-detail-label">登录账号</span>
              <strong>{{ userStore.userInfo.userName || '未设置' }}</strong>
              <small>账号名称不可在此页面修改</small>
            </div>
            <div class="profile-detail-item">
              <span class="profile-detail-label">显示昵称</span>
              <strong>{{ displayName }}</strong>
              <el-button text type="primary" :icon="Edit" @click="openEdit">编辑</el-button>
            </div>
            <div class="profile-detail-item">
              <span class="profile-detail-label">主角色</span>
              <strong>{{ primaryRole }}</strong>
              <small>由管理员分配</small>
            </div>
            <div class="profile-detail-item">
              <span class="profile-detail-label">账号状态</span>
              <strong class="profile-detail-status"><span :class="['status-dot', `status-dot--${accountStatus.type}`]" />{{ accountStatus.label }}</strong>
              <small>{{ accountStatus.description }}</small>
            </div>
          </div>
          <div v-if="roleNames.length > 1" class="profile-roles">
            <span class="profile-detail-label">权限角色</span>
            <div class="profile-role-list">
              <el-tag v-for="role in roleNames" :key="role" effect="plain" type="info">{{ role }}</el-tag>
            </div>
          </div>
        </section>

        <section class="na-panel profile-card" aria-labelledby="contact-info-title">
          <header class="profile-card__header">
            <div>
              <h2 id="contact-info-title">联系方式与安全</h2>
              <p>联系方式用于通知与账号安全验证</p>
            </div>
            <el-icon class="profile-card__icon"><Key /></el-icon>
          </header>
          <div class="profile-contact-list">
            <div class="profile-contact-row">
              <span class="profile-contact-icon profile-contact-icon--phone"><el-icon><Phone /></el-icon></span>
              <div><span class="profile-detail-label">手机号码</span><strong>{{ userStore.userInfo.phone || '未设置' }}</strong></div>
              <el-button link type="primary" class="profile-row-action" @click="changePhoneFlag = true">{{ userStore.userInfo.phone ? '修改' : '添加' }}</el-button>
            </div>
            <div class="profile-contact-row">
              <span class="profile-contact-icon profile-contact-icon--email"><el-icon><Message /></el-icon></span>
              <div><span class="profile-detail-label">邮箱地址</span><strong>{{ userStore.userInfo.email || '未设置' }}</strong></div>
              <el-button link type="primary" class="profile-row-action" @click="changeEmailFlag = true">{{ userStore.userInfo.email ? '修改' : '添加' }}</el-button>
            </div>
            <div class="profile-contact-row">
              <span class="profile-contact-icon profile-contact-icon--password"><el-icon><Lock /></el-icon></span>
              <div><span class="profile-detail-label">账号密码</span><strong>已设置</strong><small>建议定期更新并避免与其他系统复用</small></div>
              <el-button link type="primary" class="profile-row-action" @click="showPassword = true">修改</el-button>
            </div>
          </div>
        </section>
      </main>

      <aside class="profile-aside">
        <section class="na-panel profile-card profile-completion-card" aria-labelledby="completion-title">
          <header class="profile-card__header">
            <div><h2 id="completion-title">资料完整度</h2><p>补齐资料，便于同事准确联系你</p></div>
          </header>
          <div class="profile-progress-row"><el-progress type="circle" :percentage="profileCompletion" :width="92" :stroke-width="8" /><div><strong>{{ completionTitle }}</strong><p>{{ completionHint }}</p></div></div>
          <ul class="profile-check-list">
            <li v-for="item in profileChecks" :key="item.key" :class="{ 'is-complete': item.complete }"><el-icon><CircleCheck v-if="item.complete" /><Warning v-else /></el-icon><span>{{ item.label }}</span><small>{{ item.complete ? '已完善' : '待补充' }}</small></li>
          </ul>
        </section>

        <section class="na-panel profile-card profile-security-card" aria-labelledby="security-title">
          <header class="profile-card__header"><div><h2 id="security-title">账户概览</h2><p>当前登录身份摘要</p></div><el-icon class="profile-card__icon"><Monitor /></el-icon></header>
          <dl class="profile-overview-list">
            <div><dt>用户编号</dt><dd>{{ userId }}</dd></div>
            <div><dt>主权限</dt><dd>{{ primaryRole }}</dd></div>
            <div><dt>创建时间</dt><dd>{{ createdAtText }}</dd></div>
          </dl>
        </section>
      </aside>
    </div>

    <!-- 弹窗 -->
    <el-dialog
      v-model="showPassword"
      title="修改密码"
      width="400px"
      class="custom-dialog"
      @close="clearPassword"
    >
      <el-form
        ref="modifyPwdForm"
        :model="pwdModify"
        :rules="rules"
        label-width="90px"
        class="py-4"
      >
        <el-form-item :minlength="6" label="原密码" prop="password">
          <el-input v-model="pwdModify.password" show-password />
        </el-form-item>
        <el-form-item :minlength="6" label="新密码" prop="newPassword">
          <el-input v-model="pwdModify.newPassword" show-password />
        </el-form-item>
        <el-form-item :minlength="6" label="确认密码" prop="confirmPassword">
          <el-input v-model="pwdModify.confirmPassword" show-password />
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="showPassword = false">取 消</el-button>
          <el-button type="primary" :loading="savingPassword" @click="savePassword">确 定</el-button>
        </div>
      </template>
    </el-dialog>

    <el-dialog
      v-model="changePhoneFlag"
      title="修改手机号"
      width="400px"
      class="custom-dialog"
    >
      <el-form :model="phoneForm" label-width="80px" class="py-4">
        <el-form-item label="手机号">
          <el-input v-model="phoneForm.phone" placeholder="请输入新的手机号码">
            <template #prefix>
              <el-icon><phone /></el-icon>
            </template>
          </el-input>
        </el-form-item>
        <el-form-item label="验证码">
          <div class="flex gap-4">
            <el-input
              v-model="phoneForm.code"
              placeholder="请输入验证码[模拟]"
              class="flex-1"
            >
              <template #prefix>
                <el-icon><key /></el-icon>
              </template>
            </el-input>
            <el-button
              type="primary"
              :disabled="time > 0"
              class="w-32"
              @click="getCode"
            >
              {{ time > 0 ? `${time}s` : '获取验证码' }}
            </el-button>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="closeChangePhone">取 消</el-button>
          <el-button type="primary" :loading="savingPhone" @click="changePhone">确 定</el-button>
        </div>
      </template>
    </el-dialog>

    <el-dialog
      v-model="changeEmailFlag"
      title="修改邮箱"
      width="400px"
      class="custom-dialog"
    >
      <el-form :model="emailForm" label-width="80px" class="py-4">
        <el-form-item label="邮箱">
          <el-input v-model="emailForm.email" placeholder="请输入新的邮箱地址">
            <template #prefix>
              <el-icon><message /></el-icon>
            </template>
          </el-input>
        </el-form-item>
        <el-form-item label="验证码">
          <div class="flex gap-4">
            <el-input
              v-model="emailForm.code"
              placeholder="请输入验证码[模拟]"
              class="flex-1"
            >
              <template #prefix>
                <el-icon><key /></el-icon>
              </template>
            </el-input>
            <el-button
              type="primary"
              :disabled="emailTime > 0"
              class="w-32"
              @click="getEmailCode"
            >
              {{ emailTime > 0 ? `${emailTime}s` : '获取验证码' }}
            </el-button>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="closeChangeEmail">取 消</el-button>
          <el-button type="primary" :loading="savingEmail" @click="changeEmail">确 定</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
  import { setSelfInfo, changePassword } from '@/api/user.js'
  import { computed, reactive, ref, watch } from 'vue'
  import { ElMessage } from 'element-plus'
  import { CircleCheck, Edit, Key, Lock, Message, Monitor, Phone, User, Warning } from '@element-plus/icons-vue'
  import { useUserStore } from '@/pinia/modules/user'
  import SelectImage from '@/components/selectImage/selectImage.vue'
  defineOptions({
    name: 'Person'
  })

  const userStore = useUserStore()
  const modifyPwdForm = ref(null)
  const showPassword = ref(false)
  const pwdModify = ref({})
  const nickName = ref('')
  const editFlag = ref(false)
  const savingNickname = ref(false)
  const savingPassword = ref(false)
  const savingPhone = ref(false)
  const savingEmail = ref(false)
  const savingAvatar = ref(false)

  const displayName = computed(() => userStore.userInfo.nickName || userStore.userInfo.userName || '未命名用户')
  const userId = computed(() => userStore.userInfo.ID || userStore.userInfo.id || '—')
  const roleNames = computed(() => {
    const names = (userStore.userInfo.authorities || [])
      .map((item) => item?.authorityName)
      .filter(Boolean)
    const primary = userStore.userInfo.authority?.authorityName
    if (primary && !names.includes(primary)) names.unshift(primary)
    return names
  })
  const primaryRole = computed(() => roleNames.value[0] || '未分配角色')
  const accountStatus = computed(() => Number(userStore.userInfo.enable) === 2
    ? { label: '已停用', type: 'danger', description: '当前账号无法继续操作' }
    : { label: '正常使用', type: 'success', description: '当前账号状态正常' })
  const createdAtText = computed(() => {
    const value = userStore.userInfo.CreatedAt || userStore.userInfo.createdAt
    if (!value) return '—'
    const date = new Date(value)
    return Number.isNaN(date.getTime()) ? '—' : date.toLocaleDateString('zh-CN')
  })
  const profileChecks = computed(() => [
    { key: 'avatar', label: '头像', complete: Boolean(userStore.userInfo.headerImg || userStore.userInfo.headerImgPreviewUrl) },
    { key: 'nickname', label: '显示昵称', complete: Boolean(userStore.userInfo.nickName) },
    { key: 'phone', label: '手机号码', complete: Boolean(userStore.userInfo.phone) },
    { key: 'email', label: '邮箱地址', complete: Boolean(userStore.userInfo.email) }
  ])
  const profileCompletion = computed(() => Math.round(profileChecks.value.filter((item) => item.complete).length / profileChecks.value.length * 100))
  const completionTitle = computed(() => profileCompletion.value === 100 ? '资料已完善' : '再补充一点')
  const completionHint = computed(() => profileCompletion.value === 100 ? '联系方式和头像均已设置' : '完善联系方式，方便接收重要通知')

  const selfInfoPayload = (value = {}) => ({
    nickName: userStore.userInfo.nickName || '',
    phone: userStore.userInfo.phone || '',
    email: userStore.userInfo.email || '',
    headerImg: userStore.userInfo.headerImg || '',
    ...value
  })

  const rules = reactive({
    password: [
      { required: true, message: '请输入密码', trigger: 'blur' },
      { min: 6, message: '最少6个字符', trigger: 'blur' }
    ],
    newPassword: [
      { required: true, message: '请输入新密码', trigger: 'blur' },
      { min: 6, message: '最少6个字符', trigger: 'blur' }
    ],
    confirmPassword: [
      { required: true, message: '请输入确认密码', trigger: 'blur' },
      { min: 6, message: '最少6个字符', trigger: 'blur' },
      {
        validator: (rule, value, callback) => {
          if (value !== pwdModify.value.newPassword) {
            callback(new Error('两次密码不一致'))
          } else {
            callback()
          }
        },
        trigger: 'blur'
      }
    ]
  })

  const savePassword = async () => {
    const valid = await modifyPwdForm.value.validate().catch(() => false)
    if (!valid) return

    savingPassword.value = true
    try {
      const res = await changePassword({
        password: pwdModify.value.password,
        newPassword: pwdModify.value.newPassword
      })
      if (res.code === 0) {
        ElMessage.success('修改密码成功！')
        showPassword.value = false
      }
    } finally {
      savingPassword.value = false
    }
  }

  const clearPassword = () => {
    pwdModify.value = {
      password: '',
      newPassword: '',
      confirmPassword: ''
    }
    modifyPwdForm.value?.clearValidate()
  }

  const openEdit = () => {
    nickName.value = userStore.userInfo.nickName
    editFlag.value = true
  }

  const closeEdit = () => {
    nickName.value = ''
    editFlag.value = false
  }

  const enterEdit = async () => {
    savingNickname.value = true
    try {
      const res = await setSelfInfo(selfInfoPayload({
        nickName: nickName.value
      }))
      if (res.code === 0) {
        userStore.ResetUserInfo({ nickName: nickName.value })
        ElMessage.success('修改成功')
      }
    } finally {
      savingNickname.value = false
      nickName.value = ''
      editFlag.value = false
    }
  }

  const changePhoneFlag = ref(false)
  const time = ref(0)
  const phoneForm = reactive({
    phone: '',
    code: ''
  })

  const getCode = async () => {
    time.value = 60
    let timer = setInterval(() => {
      time.value--
      if (time.value <= 0) {
        clearInterval(timer)
        timer = null
      }
    }, 1000)
  }

  const closeChangePhone = () => {
    changePhoneFlag.value = false
    phoneForm.phone = ''
    phoneForm.code = ''
  }

  const changePhone = async () => {
    savingPhone.value = true
    try {
      const res = await setSelfInfo(selfInfoPayload({ phone: phoneForm.phone }))
      if (res.code === 0) {
        ElMessage.success('修改成功')
        userStore.ResetUserInfo({ phone: phoneForm.phone })
        closeChangePhone()
      }
    } finally {
      savingPhone.value = false
    }
  }

  const changeEmailFlag = ref(false)
  const emailTime = ref(0)
  const emailForm = reactive({
    email: '',
    code: ''
  })

  const getEmailCode = async () => {
    emailTime.value = 60
    let timer = setInterval(() => {
      emailTime.value--
      if (emailTime.value <= 0) {
        clearInterval(timer)
        timer = null
      }
    }, 1000)
  }

  const closeChangeEmail = () => {
    changeEmailFlag.value = false
    emailForm.email = ''
    emailForm.code = ''
  }

  const changeEmail = async () => {
    savingEmail.value = true
    try {
      const res = await setSelfInfo(selfInfoPayload({ email: emailForm.email }))
      if (res.code === 0) {
        ElMessage.success('修改成功')
        userStore.ResetUserInfo({ email: emailForm.email })
        closeChangeEmail()
      }
    } finally {
      savingEmail.value = false
    }
  }

  watch(() => userStore.userInfo.headerImg, async (val) => {
    const nextAvatar = val || ''
    userStore.ResetUserInfo({ headerImg: nextAvatar, headerImgPreviewUrl: '' })
    savingAvatar.value = true
    try {
      const res = await setSelfInfo(selfInfoPayload({ headerImg: nextAvatar }))
      if (res.code === 0) {
        await userStore.GetUserInfo()
        ElMessage.success(nextAvatar ? '头像设置成功' : '头像已移除')
      } else {
        await userStore.GetUserInfo()
      }
    } catch {
      await userStore.GetUserInfo()
    } finally {
      savingAvatar.value = false
    }
  })

</script>

<style scoped lang="scss">
  .profile-container {
    --profile-gap: var(--na-space-lg);
    max-width: 1280px;
    margin: 0 auto;

    .profile-hero {
      overflow: hidden;
      margin-bottom: var(--profile-gap);
      padding: 0;
    }

    .profile-banner {
      min-height: 122px;
      display: flex;
      align-items: flex-start;
      justify-content: space-between;
      gap: var(--na-space-lg);
      padding: 24px 32px;
      color: var(--na-foreground);
      background: var(--na-primary-soft);
      border-bottom: 1px solid var(--na-border);
    }

    .profile-banner__copy p,
    .profile-card__header p,
    .profile-progress-row p,
    .profile-detail-item small,
    .profile-contact-row small {
      color: var(--na-muted-foreground);
    }

    .profile-banner__copy p {
      margin: 8px 0 0;
      font-size: 13px;
    }

    .profile-kicker {
      display: block;
      color: var(--na-primary);
      font-size: 12px;
      font-weight: 700;
      letter-spacing: 0;
    }

    .profile-status-tag { flex: 0 0 auto; }

    .profile-identity {
      display: grid;
      grid-template-columns: 132px minmax(0, 1fr) auto;
      align-items: center;
      gap: 24px;
      padding: 0 32px 28px;
      margin-top: -38px;
    }

    .profile-avatar-wrapper {
      position: relative;
      z-index: 1;
      display: flex;
      flex-direction: column;
      align-items: center;
      gap: 8px;
    }

    .profile-avatar-wrapper :deep(.select-image-root),
    .profile-avatar-wrapper :deep(.w-40) {
      width: 116px;
      height: 116px;
      border: 4px solid var(--na-card);
      border-radius: 50%;
      background: var(--na-muted);
      box-shadow: 0 4px 12px rgb(25 23 44 / 12%);
    }

    .profile-avatar-note {
      color: var(--na-muted-foreground);
      font-size: 12px;
      white-space: nowrap;
    }

    .profile-identity__main { min-width: 0; padding-top: 32px; }

    .profile-name-row {
      display: flex;
      align-items: center;
      flex-wrap: wrap;
      gap: 8px;
    }

    .profile-name-row h1 {
      margin: 0;
      color: var(--na-foreground);
      font-size: 28px;
      line-height: 1.2;
      font-weight: 700;
      text-wrap: balance;
    }

    .profile-name-input { width: 220px; }

    .profile-account-line {
      display: flex;
      flex-wrap: wrap;
      gap: 8px;
      margin: 8px 0 12px;
      color: var(--na-muted-foreground);
      font-size: 14px;
    }

    .profile-dot { color: var(--na-border-strong); }

    .profile-identity__tags { display: flex; flex-wrap: wrap; gap: 8px; }

    .profile-hero-meta {
      display: grid;
      grid-template-columns: repeat(2, minmax(90px, 1fr));
      gap: 20px;
      min-width: 230px;
      margin: 32px 0 0;
      padding-left: 24px;
      border-left: 1px solid var(--na-border);
    }

    .profile-hero-meta dt,
    .profile-overview-list dt {
      color: var(--na-muted-foreground);
      font-size: 12px;
    }

    .profile-hero-meta dd,
    .profile-overview-list dd {
      margin: 6px 0 0;
      color: var(--na-foreground);
      font-size: 14px;
      font-weight: 600;
    }

    .profile-content {
      display: grid;
      grid-template-columns: minmax(0, 1.55fr) minmax(300px, 0.85fr);
      gap: var(--profile-gap);
      align-items: start;
    }

    .profile-main,
    .profile-aside { min-width: 0; }

    .profile-card {
      margin-bottom: var(--profile-gap);
      padding: 24px;
    }

    .profile-card__header {
      display: flex;
      align-items: flex-start;
      justify-content: space-between;
      gap: 16px;
      margin-bottom: 20px;
    }

    .profile-card__header h2 {
      margin: 0;
      color: var(--na-foreground);
      font-size: 17px;
      line-height: 1.35;
    }

    .profile-card__header p {
      margin: 5px 0 0;
      font-size: 13px;
    }

    .profile-card__icon {
      color: var(--na-primary);
      font-size: 20px;
    }

    .profile-detail-grid {
      display: grid;
      grid-template-columns: repeat(2, minmax(0, 1fr));
      gap: 1px;
      overflow: hidden;
      border: 1px solid var(--na-border);
      border-radius: var(--na-radius-sm);
      background: var(--na-border);
    }

    .profile-detail-item {
      position: relative;
      display: flex;
      flex-direction: column;
      align-items: flex-start;
      min-height: 100px;
      padding: 16px;
      background: var(--na-card);
    }

    .profile-detail-label {
      display: block;
      color: var(--na-muted-foreground);
      font-size: 12px;
      line-height: 1.4;
    }

    .profile-detail-item strong {
      max-width: 100%;
      margin-top: 8px;
      overflow: hidden;
      color: var(--na-foreground);
      font-size: 15px;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    .profile-detail-item small {
      margin-top: 6px;
      font-size: 12px;
    }

    .profile-detail-item .el-button {
      position: absolute;
      right: 12px;
      bottom: 10px;
      padding: 0;
    }

    .profile-detail-status { display: inline-flex; align-items: center; gap: 7px; }

    .status-dot {
      width: 8px;
      height: 8px;
      border-radius: 50%;
      background: var(--na-muted-foreground);
    }

    .status-dot--success { background: var(--na-success); }
    .status-dot--danger { background: var(--na-danger); }

    .profile-roles {
      display: flex;
      flex-wrap: wrap;
      align-items: center;
      gap: 12px;
      margin-top: 18px;
    }

    .profile-role-list { display: flex; flex-wrap: wrap; gap: 8px; }

    .profile-contact-list { display: grid; gap: 4px; }

    .profile-contact-row {
      display: grid;
      grid-template-columns: 34px minmax(0, 1fr) auto;
      align-items: center;
      gap: 12px;
      min-height: 68px;
      padding: 10px 0;
      border-bottom: 1px solid var(--na-border);
    }

    .profile-contact-row:last-child { border-bottom: 0; }

    .profile-contact-row strong {
      display: block;
      margin-top: 4px;
      overflow: hidden;
      color: var(--na-foreground);
      font-size: 14px;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    .profile-contact-row small { display: block; margin-top: 4px; font-size: 12px; }

    .profile-contact-icon {
      display: inline-flex;
      align-items: center;
      justify-content: center;
      width: 32px;
      height: 32px;
      border-radius: 8px;
      background: var(--na-muted);
    }

    .profile-contact-icon--phone { color: var(--na-info); }
    .profile-contact-icon--email { color: var(--na-success); }
    .profile-contact-icon--password { color: var(--na-primary); }
    .profile-row-action { justify-self: end; }

    .profile-progress-row {
      display: flex;
      align-items: center;
      gap: 16px;
      padding: 4px 0 20px;
    }

    .profile-progress-row strong { color: var(--na-foreground); font-size: 15px; }
    .profile-progress-row p { margin: 6px 0 0; font-size: 12px; line-height: 1.5; }

    .profile-check-list {
      display: grid;
      gap: 10px;
      margin: 0;
      padding: 16px 0 0;
      border-top: 1px solid var(--na-border);
      list-style: none;
    }

    .profile-check-list li {
      display: grid;
      grid-template-columns: 18px 1fr auto;
      align-items: center;
      gap: 8px;
      color: var(--na-muted-foreground);
      font-size: 13px;
    }

    .profile-check-list li.is-complete { color: var(--na-foreground); }
    .profile-check-list li.is-complete .el-icon { color: var(--na-success); }
    .profile-check-list li small { font-size: 12px; }

    .profile-overview-list { display: grid; gap: 14px; margin: 0; }
    .profile-overview-list > div { display: flex; align-items: center; justify-content: space-between; gap: 16px; }
    .profile-overview-list dd { margin: 0; text-align: right; }

    .custom-dialog {
      :deep(.el-dialog__header) {
        @apply mb-0 pb-4 border-b border-gray-100 dark:border-gray-700;
      }
      :deep(.el-dialog__footer) {
        @apply mt-0 pt-4 border-t border-gray-100 dark:border-gray-700;
      }
      :deep(.el-input__wrapper) {
        @apply shadow-none;
      }
      :deep(.el-input__prefix) {
        @apply mr-2;
      }
    }

    @media (max-width: 900px) {
      .profile-content { grid-template-columns: 1fr; }
      .profile-aside { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: var(--profile-gap); }
      .profile-aside .profile-card { margin-bottom: 0; }
    }

    @media (max-width: 680px) {
      .profile-banner { min-height: 108px; padding: 20px; }
      .profile-identity { grid-template-columns: 96px minmax(0, 1fr); gap: 16px; padding: 0 20px 22px; margin-top: -28px; }
      .profile-avatar-wrapper :deep(.select-image-root),
      .profile-avatar-wrapper :deep(.w-40) { width: 88px; height: 88px; }
      .profile-identity__main { padding-top: 24px; }
      .profile-name-row h1 { font-size: 22px; }
      .profile-hero-meta { grid-column: 1 / -1; grid-template-columns: repeat(2, 1fr); min-width: 0; margin: 0; padding: 16px 0 0; border-top: 1px solid var(--na-border); border-left: 0; }
      .profile-card { padding: 18px; }
      .profile-detail-grid { grid-template-columns: 1fr; }
      .profile-aside { display: block; }
      .profile-aside .profile-card { margin-bottom: var(--profile-gap); }
    }

    @media (prefers-reduced-motion: reduce) {
      * { transition-duration: 0.01ms !important; animation-duration: 0.01ms !important; }
    }
  }
</style>
