<template>
  <div class="na-page na-page--list profile-page">
    <!-- 1. 顶部宽幅个人档案名片 Studio Banner (占满宽度，高度紧凑精致) -->
    <header class="profile-hero-banner" aria-labelledby="profile-title">
      <div class="hero-left">
        <!-- 头像区 -->
        <div class="avatar-box">
          <div class="avatar-ring">
            <SelectImage
              v-model="userStore.userInfo.headerImg"
              file-type="image"
              :preview-url="userStore.userInfo.headerImgPreviewUrl"
              :loading="savingAvatar"
              rounded
            />
          </div>
          <span class="avatar-badge">
            <el-icon><Camera /></el-icon>
            更换
          </span>
        </div>

        <!-- 身份与昵称区 -->
        <div class="identity-info">
          <div class="name-row">
            <template v-if="!editFlag">
              <h1 id="profile-title" class="user-display-name">{{ displayName }}</h1>
              <el-tooltip content="编辑显示昵称" placement="top">
                <el-button text circle :icon="Edit" class="edit-btn" aria-label="编辑昵称" @click="openEdit" />
              </el-tooltip>
            </template>
            <template v-else>
              <div class="name-edit-wrap">
                <el-input v-model="nickName" class="name-input" placeholder="请输入新昵称" />
                <el-button type="primary" size="small" :loading="savingNickname" @click="enterEdit">保存</el-button>
                <el-button plain size="small" @click="closeEdit">取消</el-button>
              </div>
            </template>
            <span class="status-pill" :class="`is-${accountStatus.type}`">
              <span class="status-dot" />
              {{ accountStatus.label }} · 企业安全认证
            </span>
          </div>

          <div class="meta-row">
            <span class="user-handle font-mono">@{{ userStore.userInfo.userName || '未设置账号' }}</span>
            <span class="sep-dot">·</span>
            <span class="primary-role-chip">
              <el-icon><Stamp /></el-icon>
              {{ primaryRole }}
            </span>
            <span class="sep-dot">·</span>
            <span class="uid-tag font-mono">UID #{{ userId }}</span>
            <span class="sep-dot">·</span>
            <span class="sec-tag">
              <el-icon><CircleCheck /></el-icon>
              安全策略正常
            </span>
          </div>

          <div v-if="roleNames.length > 1" class="multi-roles-row">
            <span class="roles-label">所属权限组：</span>
            <span v-for="role in roleNames" :key="role" class="sub-role-tag">
              <el-icon><Stamp /></el-icon>
              {{ role }}
            </span>
          </div>
        </div>
      </div>

      <!-- 右侧指标卡片与快捷操作 -->
      <div class="hero-right">
        <div class="metric-tiles">
          <div class="metric-tile">
            <div class="tile-icon"><el-icon><Calendar /></el-icon></div>
            <div class="tile-body">
              <span class="tile-label">加入系统</span>
              <strong class="tile-val">{{ createdAtText }}</strong>
            </div>
          </div>
          <div class="metric-tile">
            <div class="tile-icon is-health" :class="{ 'is-perfect': profileCompletion === 100 }">
              <el-icon><Compass /></el-icon>
            </div>
            <div class="tile-body">
              <span class="tile-label">资料完整度</span>
              <strong class="tile-val" :class="{ 'is-perfect-text': profileCompletion === 100 }">
                {{ profileCompletion }}%
              </strong>
            </div>
          </div>
          <div class="metric-tile">
            <div class="tile-icon is-security"><el-icon><Lock /></el-icon></div>
            <div class="tile-body">
              <span class="tile-label">密码保护</span>
              <strong class="tile-val">已激活</strong>
            </div>
          </div>
          <div class="metric-tile">
            <div class="tile-icon is-audit"><el-icon><Monitor /></el-icon></div>
            <div class="tile-body">
              <span class="tile-label">操作审计</span>
              <strong class="tile-val font-mono">实时生效</strong>
            </div>
          </div>
        </div>

        <div class="hero-actions">
          <el-button type="primary" :icon="Lock" @click="showPassword = true">
            修改登录密码
          </el-button>
          <el-button plain :icon="Calendar" @click="navTo('workSchedule')">
            工作日历
          </el-button>
        </div>
      </div>
    </header>

    <!-- 2. 下半部分：全屏等高 3 栏 Bento Studio 网格 (填满页面，宽幅舒展) -->
    <div class="profile-grid">
      <!-- 栏目 1：账号核心档案与权限范围 -->
      <section class="profile-card col-account" aria-labelledby="account-card-title">
        <div class="card-head">
          <div class="head-icon is-user"><el-icon><User /></el-icon></div>
          <div class="head-text">
            <h2 id="account-card-title">账号核心资料</h2>
            <p>系统登录唯一凭据与组织授权范围</p>
          </div>
        </div>

        <div class="account-items-grid">
          <div class="account-item">
            <div class="item-label">
              <span>系统登录账号</span>
              <el-icon class="icon-muted"><User /></el-icon>
            </div>
            <div class="item-val font-mono">{{ userStore.userInfo.userName || '未设置' }}</div>
            <div class="item-desc">全局唯一凭据，不可在此页面变更</div>
          </div>

          <div class="account-item">
            <div class="item-label">
              <span>当前显示昵称</span>
              <el-button text type="primary" size="small" :icon="Edit" @click="openEdit">编辑</el-button>
            </div>
            <div class="item-val">{{ displayName }}</div>
            <div class="item-desc">用于内部协同沟通与操作日志展示</div>
          </div>

          <div class="account-item">
            <div class="item-label">
              <span>核心授权角色</span>
              <el-icon class="icon-muted"><Stamp /></el-icon>
            </div>
            <div class="item-val font-semibold">{{ primaryRole }}</div>
            <div class="item-desc">由超级管理员统一设定与授权</div>
          </div>

          <div class="account-item">
            <div class="item-label">
              <span>账号运行状态</span>
              <el-icon class="icon-muted"><CircleCheck /></el-icon>
            </div>
            <div class="item-val status-val">
              <span class="status-dot" :class="`status-dot--${accountStatus.type}`" />
              {{ accountStatus.label }}
            </div>
            <div class="item-desc">{{ accountStatus.description }}</div>
          </div>
        </div>

        <!-- 关联角色组展示 -->
        <div class="authority-section">
          <div class="section-title">
            <span>全部所属权限角色组 ({{ roleNames.length }})</span>
          </div>
          <div class="role-tags-wrap">
            <span v-for="role in roleNames" :key="role" class="role-badge">
              <el-icon><Stamp /></el-icon>
              {{ role }}
            </span>
          </div>
        </div>

        <!-- 系统元信息清单 -->
        <div class="sys-meta-panel">
          <div class="meta-row-item">
            <span class="meta-k">系统用户编号</span>
            <span class="meta-v font-mono">UID #{{ userId }}</span>
          </div>
          <div class="meta-row-item">
            <span class="meta-k">注册创建时间</span>
            <span class="meta-v">{{ createdAtText }}</span>
          </div>
          <div class="meta-row-item">
            <span class="meta-k">认证防护协议</span>
            <span class="meta-v">JWT Bearer 企业级安全认证</span>
          </div>
        </div>
      </section>

      <!-- 栏目 2：联系方式与安全中枢 -->
      <section class="profile-card col-security" aria-labelledby="security-card-title">
        <div class="card-head">
          <div class="head-icon is-security"><el-icon><Lock /></el-icon></div>
          <div class="head-text">
            <h2 id="security-card-title">联系方式与安全中心</h2>
            <p>密保手机、通知邮箱与核心凭证安全</p>
          </div>
        </div>

        <!-- 企业级安全态势横幅 -->
        <div class="security-banner">
          <el-icon class="shield-icon"><CircleCheck /></el-icon>
          <div class="banner-body">
            <strong>企业级安全防护已激活</strong>
            <p>所有关键凭据均进行硬件加密存储，涉及敏感操作需二次校验</p>
          </div>
        </div>

        <!-- 密保手机 -->
        <div class="contact-card-row">
          <div class="contact-icon is-phone"><el-icon><Phone /></el-icon></div>
          <div class="contact-main">
            <div class="contact-top">
              <span class="contact-title">密保手机</span>
              <span v-if="userStore.userInfo.phone" class="tag-verified">已绑定</span>
              <span v-else class="tag-unverified">待绑定</span>
            </div>
            <strong class="contact-number font-mono">{{ userStore.userInfo.phone || '未设置手机号' }}</strong>
            <span class="contact-tip">{{ contactCapabilities.phone.reason }}</span>
          </div>
          <el-button type="primary" plain size="small" class="contact-action-btn" @click="openPhoneDialog">
            {{ userStore.userInfo.phone ? '修改手机号' : '绑定手机号' }}
          </el-button>
        </div>

        <!-- 密保邮箱 -->
        <div class="contact-card-row">
          <div class="contact-icon is-email"><el-icon><Message /></el-icon></div>
          <div class="contact-main">
            <div class="contact-top">
              <span class="contact-title">密保邮箱</span>
              <span v-if="userStore.userInfo.email" class="tag-verified">已绑定</span>
              <span v-else class="tag-unverified">待绑定</span>
            </div>
            <strong class="contact-number font-mono">{{ userStore.userInfo.email || '未设置邮箱' }}</strong>
            <span class="contact-tip">{{ contactCapabilities.email.reason }}</span>
          </div>
          <el-button type="primary" plain size="small" class="contact-action-btn" @click="openEmailDialog">
            {{ userStore.userInfo.email ? '修改邮箱' : '绑定邮箱' }}
          </el-button>
        </div>

        <!-- 登录密码 -->
        <div class="contact-card-row">
          <div class="contact-icon is-pwd"><el-icon><Key /></el-icon></div>
          <div class="contact-main">
            <div class="contact-top">
              <span class="contact-title">账号登录密码</span>
              <span class="tag-verified is-pwd">安全强度 良好</span>
            </div>
            <strong class="contact-number font-mono">••••••••••••</strong>
            <span class="contact-tip">建议每 90 天定期更新一次密码，保障账户安全</span>
          </div>
          <el-button type="primary" plain size="small" class="contact-action-btn" @click="showPassword = true">
            修改密码
          </el-button>
        </div>
      </section>

      <!-- 栏目 3：健康度仪表盘 & 快捷协同入口 -->
      <div class="col-side-group">
        <!-- 资料健康度卡片 -->
        <section class="profile-card card-health" aria-labelledby="health-card-title">
          <div class="card-head">
            <div class="head-icon is-health"><el-icon><CircleCheck /></el-icon></div>
            <div class="head-text">
              <h2 id="health-card-title">资料健康度</h2>
              <p>完善个人档案以保障系统高效协同</p>
            </div>
          </div>

          <div class="health-progress-wrap">
            <div class="ring-wrapper">
              <el-progress
                type="circle"
                :percentage="profileCompletion"
                :width="82"
                :stroke-width="7"
                :color="progressColor"
              />
            </div>
            <div class="health-progress-info">
              <strong>{{ completionTitle }}</strong>
              <p>{{ completionHint }}</p>
            </div>
          </div>

          <ul class="health-check-list">
            <li v-for="item in profileChecks" :key="item.key" :class="{ 'is-done': item.complete }">
              <div class="check-title">
                <el-icon v-if="item.complete" class="icon-done"><CircleCheck /></el-icon>
                <el-icon v-else class="icon-pending"><Warning /></el-icon>
                <span>{{ item.label }}</span>
              </div>
              <span class="check-chip" :class="item.complete ? 'is-chip-done' : 'is-chip-pending'">
                {{ item.complete ? '已完善' : '待补充' }}
              </span>
            </li>
          </ul>
        </section>

        <!-- 快捷协同功能导航卡片 -->
        <section class="profile-card card-quick" aria-labelledby="quick-card-title">
          <div class="card-head">
            <div class="head-icon is-nav"><el-icon><Compass /></el-icon></div>
            <div class="head-text">
              <h2 id="quick-card-title">快捷协同直达</h2>
              <p>系统常用核心模块与工作台入口</p>
            </div>
          </div>

          <div class="quick-nav-group">
            <button type="button" class="quick-btn" @click="navTo('workSchedule')">
              <div class="btn-left">
                <el-icon class="icon-amber"><Calendar /></el-icon>
                <span>查看我的工作日历</span>
              </div>
              <el-icon class="arrow-icon"><Right /></el-icon>
            </button>
            <button type="button" class="quick-btn" @click="navTo('dashboard')">
              <div class="btn-left">
                <el-icon class="icon-blue"><Monitor /></el-icon>
                <span>返回首屏驾驶舱</span>
              </div>
              <el-icon class="arrow-icon"><Right /></el-icon>
            </button>
            <button type="button" class="quick-btn" @click="navTo('aiOperations')">
              <div class="btn-left">
                <el-icon class="icon-purple"><Cpu /></el-icon>
                <span>智能能力与模型接入</span>
              </div>
              <el-icon class="arrow-icon"><Right /></el-icon>
            </button>
          </div>
        </section>
      </div>
    </div>

    <!-- 弹窗 1：修改密码 -->
    <el-dialog
      v-model="showPassword"
      title="修改密码"
      width="440px"
      class="custom-dialog modern-profile-dialog"
      append-to-body
      destroy-on-close
      @close="clearPassword"
    >
      <template #header>
        <div class="dialog-header-custom">
          <div class="dialog-icon-circle is-lock">
            <el-icon><Lock /></el-icon>
          </div>
          <div>
            <h3>修改账户登录密码</h3>
            <p>更新密码后，请妥善保管新密码并避免与常用外网密码重复</p>
          </div>
        </div>
      </template>

      <el-form
        ref="modifyPwdForm"
        :model="pwdModify"
        :rules="rules"
        label-position="top"
        class="dialog-modern-form"
      >
        <el-form-item label="原密码" prop="password">
          <el-input
            v-model="pwdModify.password"
            show-password
            placeholder="请输入当前正在使用的旧密码"
            size="large"
          >
            <template #prefix><el-icon><Lock /></el-icon></template>
          </el-input>
        </el-form-item>
        <el-form-item label="新密码" prop="newPassword">
          <el-input
            v-model="pwdModify.newPassword"
            show-password
            placeholder="最少 6 位，建议混合字母与数字"
            size="large"
          >
            <template #prefix><el-icon><Key /></el-icon></template>
          </el-input>
        </el-form-item>
        <el-form-item label="确认新密码" prop="confirmPassword">
          <el-input
            v-model="pwdModify.confirmPassword"
            show-password
            placeholder="请再次输入新密码"
            size="large"
          >
            <template #prefix><el-icon><Check /></el-icon></template>
          </el-input>
        </el-form-item>
      </el-form>

      <template #footer>
        <div class="dialog-footer">
          <el-button @click="showPassword = false">取 消</el-button>
          <el-button type="primary" :loading="savingPassword" @click="savePassword">确认修改</el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 弹窗 2：修改手机号 -->
    <el-dialog
      v-model="changePhoneFlag"
      title="修改手机号"
      width="480px"
      class="contact-dialog modern-profile-dialog"
      append-to-body
      :close-on-click-modal="false"
      @closed="resetPhoneForm"
    >
      <template #header>
        <div class="contact-dialog__heading">
          <span class="contact-dialog__icon" aria-hidden="true"><el-icon><Phone /></el-icon></span>
          <div>
            <h3>修改密保手机号</h3>
            <p>完成短信验证码核验后，将实时同步更新系统联系方式</p>
          </div>
        </div>
      </template>

      <div
        class="contact-verification-state"
        :class="{ 'is-ready': contactCapabilities.phone.enabled }"
        aria-live="polite"
      >
        <el-icon><CircleCheck v-if="contactCapabilities.phone.enabled" /><Warning v-else /></el-icon>
        <div>
          <strong>{{ contactCapabilities.phone.enabled ? '短信验证服务正常就绪' : '暂不可修改手机号' }}</strong>
          <span>{{ contactCapabilities.phone.reason }}</span>
        </div>
      </div>

      <el-form
        ref="phoneFormRef"
        :model="phoneForm"
        :rules="phoneRules"
        label-position="top"
        class="contact-form"
        :disabled="!contactCapabilities.phone.enabled"
      >
        <el-form-item label="新手机号码" prop="phone">
          <el-input
            v-model.trim="phoneForm.phone"
            placeholder="请输入 11 位中国大陆手机号码"
            maxlength="11"
            inputmode="tel"
            autocomplete="tel"
            clearable
            size="large"
          >
            <template #prefix><el-icon><Phone /></el-icon></template>
          </el-input>
        </el-form-item>
        <el-form-item label="短信验证码" prop="code">
          <div class="contact-code-row">
            <el-input
              v-model.trim="phoneForm.code"
              placeholder="请输入 6 位数字验证码"
              maxlength="6"
              inputmode="numeric"
              autocomplete="one-time-code"
              size="large"
            >
              <template #prefix><el-icon><Key /></el-icon></template>
            </el-input>
            <el-button
              type="primary"
              plain
              size="large"
              :loading="sendingPhoneCode"
              :disabled="time > 0 || sendingPhoneCode || !phoneForm.phone || !contactCapabilities.phone.enabled"
              @click="getCode"
            >
              {{ time > 0 ? `${time}s` : '获取验证码' }}
            </el-button>
          </div>
          <p class="contact-field-hint">验证码发送成功后 5 分钟内有效，请勿泄露给他人</p>
        </el-form-item>
      </el-form>

      <template #footer>
        <div class="dialog-footer">
          <el-button @click="closeChangePhone">取 消</el-button>
          <el-button
            type="primary"
            :loading="savingPhone"
            :disabled="!contactCapabilities.phone.enabled"
            @click="changePhone"
          >确认绑定</el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 弹窗 3：修改邮箱 -->
    <el-dialog
      v-model="changeEmailFlag"
      title="修改邮箱"
      width="480px"
      class="contact-dialog modern-profile-dialog"
      append-to-body
      :close-on-click-modal="false"
      @closed="resetEmailForm"
    >
      <template #header>
        <div class="contact-dialog__heading">
          <span class="contact-dialog__icon contact-dialog__icon--email" aria-hidden="true"><el-icon><Message /></el-icon></span>
          <div>
            <h3>修改密保邮箱</h3>
            <p>完成邮件验证后，将用于接收系统资产风险与重要待办通知</p>
          </div>
        </div>
      </template>

      <div
        class="contact-verification-state"
        :class="{ 'is-ready': contactCapabilities.email.enabled }"
        aria-live="polite"
      >
        <el-icon><CircleCheck v-if="contactCapabilities.email.enabled" /><Warning v-else /></el-icon>
        <div>
          <strong>{{ contactCapabilities.email.enabled ? '邮件验证服务正常就绪' : '暂不可修改邮箱' }}</strong>
          <span>{{ contactCapabilities.email.reason }}</span>
        </div>
      </div>

      <el-form
        ref="emailFormRef"
        :model="emailForm"
        :rules="emailRules"
        label-position="top"
        class="contact-form"
        :disabled="!contactCapabilities.email.enabled"
      >
        <el-form-item label="新邮箱地址" prop="email">
          <el-input
            v-model.trim="emailForm.email"
            placeholder="请输入规范的邮箱地址（如 name@company.com）"
            maxlength="120"
            inputmode="email"
            autocomplete="email"
            clearable
            size="large"
          >
            <template #prefix><el-icon><Message /></el-icon></template>
          </el-input>
        </el-form-item>
        <el-form-item label="邮件验证码" prop="code">
          <div class="contact-code-row">
            <el-input
              v-model.trim="emailForm.code"
              placeholder="请输入 6 位数字验证码"
              maxlength="6"
              inputmode="numeric"
              autocomplete="one-time-code"
              size="large"
            >
              <template #prefix><el-icon><Key /></el-icon></template>
            </el-input>
            <el-button
              type="primary"
              plain
              size="large"
              :loading="sendingEmailCode"
              :disabled="emailTime > 0 || sendingEmailCode || !emailForm.email || !contactCapabilities.email.enabled"
              @click="getEmailCode"
            >
              {{ emailTime > 0 ? `${emailTime}s` : '获取验证码' }}
            </el-button>
          </div>
          <p class="contact-field-hint">验证码发送成功后 5 分钟内有效，请注意检查垃圾邮件箱</p>
        </el-form-item>
      </el-form>

      <template #footer>
        <div class="dialog-footer">
          <el-button @click="closeChangeEmail">取 消</el-button>
          <el-button
            type="primary"
            :loading="savingEmail"
            :disabled="!contactCapabilities.email.enabled"
            @click="changeEmail"
          >确认绑定</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import {
  changePassword,
  getContactVerificationCapabilities,
  sendContactVerificationCode,
  setSelfInfo,
  updateSelfContact
} from '@/api/user.js'
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  Calendar,
  Camera,
  Check,
  CircleCheck,
  Compass,
  Cpu,
  Edit,
  Key,
  Lock,
  Message,
  Monitor,
  Phone,
  Right,
  Stamp,
  User,
  Warning
} from '@element-plus/icons-vue'
import { useUserStore } from '@/pinia/modules/user'
import SelectImage from '@/components/selectImage/selectImage.vue'

defineOptions({
  name: 'Person'
})

const router = useRouter()
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
const sendingPhoneCode = ref(false)
const sendingEmailCode = ref(false)
const savingAvatar = ref(false)
const phoneFormRef = ref(null)
const emailFormRef = ref(null)
const capabilitiesLoading = ref(false)
const contactCapabilities = reactive({
  phone: { configured: false, enabled: false, reason: '正在检查短信验证配置' },
  email: { configured: false, enabled: false, reason: '正在检查邮件验证配置' }
})

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
  ? { label: '已停用', type: 'danger', description: '当前账号无法继续操作业务' }
  : { label: '正常使用', type: 'success', description: '当前账号权限与状态正常' })

const createdAtText = computed(() => {
  const value = userStore.userInfo.CreatedAt || userStore.userInfo.createdAt
  if (!value) return '—'
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? '—' : date.toLocaleDateString('zh-CN')
})

const profileChecks = computed(() => [
  { key: 'avatar', label: '个性头像', complete: Boolean(userStore.userInfo.headerImg || userStore.userInfo.headerImgPreviewUrl) },
  { key: 'nickname', label: '显示昵称', complete: Boolean(userStore.userInfo.nickName) },
  { key: 'phone', label: '密保手机', complete: Boolean(userStore.userInfo.phone) },
  { key: 'email', label: '密保邮箱', complete: Boolean(userStore.userInfo.email) }
])

const profileCompletion = computed(() => Math.round(profileChecks.value.filter((item) => item.complete).length / profileChecks.value.length * 100))
const completionTitle = computed(() => profileCompletion.value === 100 ? '个人资料极佳' : '资料待补充')
const completionHint = computed(() => profileCompletion.value === 100 ? '密保手机、邮箱与头像均已配置就绪' : '完善联系方式，以便接收重要系统与资产告警')
const progressColor = computed(() => profileCompletion.value === 100 ? '#059669' : '#6d5dfb')

const selfInfoPayload = (value = {}) => ({
  nickName: userStore.userInfo.nickName || '',
  headerImg: userStore.userInfo.headerImg || '',
  ...value
})

const rules = reactive({
  password: [
    { required: true, message: '请输入原密码', trigger: 'blur' },
    { min: 6, message: '密码最少 6 个字符', trigger: 'blur' }
  ],
  newPassword: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, message: '密码最少 6 个字符', trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, message: '请再次输入新密码', trigger: 'blur' },
    { min: 6, message: '密码最少 6 个字符', trigger: 'blur' },
    {
      validator: (rule, value, callback) => {
        if (value !== pwdModify.value.newPassword) {
          callback(new Error('两次输入的新密码不一致'))
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
      ElMessage.success('密码修改成功，请使用新密码登录')
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
  nickName.value = userStore.userInfo.nickName || ''
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
      ElMessage.success('昵称修改成功')
    }
  } finally {
    savingNickname.value = false
    nickName.value = ''
    editFlag.value = false
  }
}

const changePhoneFlag = ref(false)
const time = ref(0)
const phoneTimer = ref(null)
const phoneForm = reactive({
  phone: '',
  code: ''
})

const phoneRules = {
  phone: [
    { required: true, message: '请输入手机号码', trigger: 'blur' },
    { pattern: /^1[3-9]\d{9}$/, message: '请输入有效的 11 位中国大陆手机号码', trigger: 'blur' }
  ],
  code: [
    { required: true, message: '请输入短信验证码', trigger: 'blur' },
    { pattern: /^\d{6}$/, message: '请输入 6 位数字验证码', trigger: 'blur' }
  ]
}

const emailRules = {
  email: [
    { required: true, message: '请输入邮箱地址', trigger: 'blur' },
    { type: 'email', message: '请输入格式有效的电子邮箱地址', trigger: 'blur' }
  ],
  code: [
    { required: true, message: '请输入邮件验证码', trigger: 'blur' },
    { pattern: /^\d{6}$/, message: '请输入 6 位数字验证码', trigger: 'blur' }
  ]
}

const loadContactCapabilities = async () => {
  if (capabilitiesLoading.value) return
  capabilitiesLoading.value = true
  try {
    const res = await getContactVerificationCapabilities()
    if (res.code === 0) {
      Object.assign(contactCapabilities.phone, res.data.phone)
      Object.assign(contactCapabilities.email, res.data.email)
    }
  } catch {
    contactCapabilities.phone.reason = '短信验证状态读取失败，请稍后重试'
    contactCapabilities.email.reason = '邮件验证状态读取失败，请稍后重试'
  } finally {
    capabilitiesLoading.value = false
  }
}

const openPhoneDialog = () => {
  changePhoneFlag.value = true
  loadContactCapabilities()
}

const startCountdown = (counter, timerRef) => {
  if (timerRef.value) clearInterval(timerRef.value)
  counter.value = 60
  timerRef.value = window.setInterval(() => {
    counter.value--
    if (counter.value <= 0) {
      clearInterval(timerRef.value)
      timerRef.value = null
    }
  }, 1000)
}

const getCode = async () => {
  if (sendingPhoneCode.value || time.value > 0) return
  if (!contactCapabilities.phone.enabled) {
    ElMessage.warning(contactCapabilities.phone.reason)
    return
  }
  sendingPhoneCode.value = true
  try {
    const valid = await phoneFormRef.value?.validateField('phone').then(() => true).catch(() => false)
    if (!valid) return
    const res = await sendContactVerificationCode({
      channel: 'phone',
      target: phoneForm.phone
    })
    if (res.code === 0) {
      startCountdown(time, phoneTimer)
      ElMessage.success('验证码已成功发送到新手机号')
    }
  } finally {
    sendingPhoneCode.value = false
  }
}

const closeChangePhone = () => {
  changePhoneFlag.value = false
}

const resetPhoneForm = () => {
  phoneForm.phone = ''
  phoneForm.code = ''
  phoneFormRef.value?.clearValidate()
}

const changePhone = async () => {
  if (savingPhone.value) return
  if (!contactCapabilities.phone.enabled) {
    ElMessage.warning(contactCapabilities.phone.reason)
    return
  }
  savingPhone.value = true
  try {
    const valid = await phoneFormRef.value?.validate().catch(() => false)
    if (!valid) return
    const res = await updateSelfContact({
      channel: 'phone',
      target: phoneForm.phone,
      code: phoneForm.code
    })
    if (res.code === 0) {
      ElMessage.success('手机号绑定成功')
      userStore.ResetUserInfo({ phone: phoneForm.phone })
      closeChangePhone()
    }
  } finally {
    savingPhone.value = false
  }
}

const changeEmailFlag = ref(false)
const emailTime = ref(0)
const emailTimer = ref(null)
const emailForm = reactive({
  email: '',
  code: ''
})

const openEmailDialog = () => {
  changeEmailFlag.value = true
  loadContactCapabilities()
}

const getEmailCode = async () => {
  if (sendingEmailCode.value || emailTime.value > 0) return
  if (!contactCapabilities.email.enabled) {
    ElMessage.warning(contactCapabilities.email.reason)
    return
  }
  sendingEmailCode.value = true
  try {
    const valid = await emailFormRef.value?.validateField('email').then(() => true).catch(() => false)
    if (!valid) return
    const res = await sendContactVerificationCode({
      channel: 'email',
      target: emailForm.email
    })
    if (res.code === 0) {
      startCountdown(emailTime, emailTimer)
      ElMessage.success('验证码已成功发送到新邮箱')
    }
  } finally {
    sendingEmailCode.value = false
  }
}

const closeChangeEmail = () => {
  changeEmailFlag.value = false
}

const resetEmailForm = () => {
  emailForm.email = ''
  emailForm.code = ''
  emailFormRef.value?.clearValidate()
}

const changeEmail = async () => {
  if (savingEmail.value) return
  if (!contactCapabilities.email.enabled) {
    ElMessage.warning(contactCapabilities.email.reason)
    return
  }
  savingEmail.value = true
  try {
    const valid = await emailFormRef.value?.validate().catch(() => false)
    if (!valid) return
    const target = emailForm.email.trim().toLowerCase()
    const res = await updateSelfContact({
      channel: 'email',
      target,
      code: emailForm.code
    })
    if (res.code === 0) {
      ElMessage.success('邮箱绑定成功')
      userStore.ResetUserInfo({ email: target })
      closeChangeEmail()
    }
  } finally {
    savingEmail.value = false
  }
}

const navTo = (routeName) => {
  if (router.hasRoute(routeName)) {
    router.push({ name: routeName })
  }
}

onMounted(loadContactCapabilities)

onBeforeUnmount(() => {
  if (phoneTimer.value) clearInterval(phoneTimer.value)
  if (emailTimer.value) clearInterval(emailTimer.value)
})

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
/* 页面满宽根容器 */
.profile-page {
  width: 100% !important;
  max-width: none !important;
  margin: 0 !important;
  padding: 16px 20px 28px !important;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

/* 1. 顶部宽幅个人档案 Studio Hero */
.profile-hero-banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 20px;
  padding: 20px 24px;
  border-radius: 16px;
  border: 1px solid var(--na-border);
  background: linear-gradient(135deg, color-mix(in srgb, var(--na-primary) 6%, var(--na-card)) 0%, var(--na-card) 60%);
  box-shadow: 0 4px 20px -4px rgb(0 0 0 / 4%);
}

.hero-left {
  display: flex;
  align-items: center;
  gap: 20px;
  min-width: 0;
  flex: 1 1 auto;
}

.avatar-box {
  position: relative;
  flex-shrink: 0;

  .avatar-ring {
    padding: 3px;
    border-radius: 50%;
    background: linear-gradient(135deg, var(--na-primary) 0%, #38bdf8 100%);
    box-shadow: 0 6px 16px color-mix(in srgb, var(--na-primary) 24%, transparent);
    transition: transform 200ms ease;

    &:hover {
      transform: scale(1.03);
    }
  }

  :deep(.select-image-root),
  :deep(.w-40) {
    width: 84px !important;
    height: 84px !important;
    border: 3px solid var(--na-card);
    border-radius: 50%;
    background: var(--na-muted);
    overflow: hidden;
  }
}

.avatar-badge {
  position: absolute;
  bottom: 0;
  right: -2px;
  display: inline-flex;
  align-items: center;
  gap: 3px;
  padding: 2px 7px;
  border-radius: 20px;
  background: var(--na-card);
  border: 1px solid var(--na-border);
  color: var(--na-foreground);
  font-size: 11px;
  font-weight: 600;
  box-shadow: 0 2px 6px rgb(0 0 0 / 8%);
  pointer-events: none;
}

.identity-info {
  display: flex;
  flex-direction: column;
  gap: 6px;
  min-width: 0;
}

.name-row {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;

  .user-display-name {
    margin: 0;
    color: var(--na-foreground);
    font-size: 22px;
    font-weight: 750;
    line-height: 1.2;
    letter-spacing: -0.01em;
  }
}

.edit-btn {
  color: var(--na-muted-foreground);
  transition: all 150ms ease;

  &:hover {
    color: var(--na-primary);
    background: var(--na-primary-soft);
  }
}

.name-edit-wrap {
  display: inline-flex;
  align-items: center;
  gap: 6px;

  .name-input {
    width: 180px;
  }
}

.status-pill {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 3px 10px;
  border-radius: 100px;
  font-size: 12px;
  font-weight: 600;
  background: var(--na-card);
  border: 1px solid var(--na-border);

  &.is-success {
    color: var(--na-success);
    border-color: color-mix(in srgb, var(--na-success) 30%, transparent);

    .status-dot {
      background: var(--na-success);
      box-shadow: 0 0 0 2px color-mix(in srgb, var(--na-success) 24%, transparent);
    }
  }

  &.is-danger {
    color: var(--na-danger);
    border-color: color-mix(in srgb, var(--na-danger) 30%, transparent);

    .status-dot {
      background: var(--na-danger);
      box-shadow: 0 0 0 2px color-mix(in srgb, var(--na-danger) 24%, transparent);
    }
  }
}

.status-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  animation: beacon-pulse 2s infinite ease-out;
}

@keyframes beacon-pulse {
  0% { transform: scale(0.95); opacity: 0.8; }
  50% { transform: scale(1.15); opacity: 1; }
  100% { transform: scale(0.95); opacity: 0.8; }
}

.meta-row {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
  font-size: 13px;
  color: var(--na-muted-foreground);

  .user-handle {
    font-weight: 600;
    color: var(--na-foreground);
  }

  .sep-dot {
    color: var(--na-border-strong);
  }

  .primary-role-chip {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    padding: 1px 8px;
    border-radius: 6px;
    background: var(--na-primary-soft);
    color: var(--na-primary);
    font-size: 12px;
    font-weight: 650;
  }

  .uid-tag {
    font-weight: 700;
    color: var(--na-muted-foreground);
  }

  .sec-tag {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    color: var(--na-success);
    font-size: 12px;
    font-weight: 600;
  }
}

.multi-roles-row {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 2px;
  font-size: 12px;

  .roles-label {
    color: var(--na-muted-foreground);
  }

  .sub-role-tag {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    padding: 1px 7px;
    border-radius: 5px;
    background: color-mix(in srgb, var(--na-muted) 60%, var(--na-card));
    border: 1px solid var(--na-border);
    color: var(--na-foreground);
    font-size: 11.5px;
  }
}

/* 右侧指标与操作 */
.hero-right {
  display: flex;
  align-items: center;
  gap: 20px;
  flex-wrap: wrap;
}

.metric-tiles {
  display: grid;
  grid-template-columns: repeat(4, auto);
  gap: 12px;
}

.metric-tile {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 14px;
  border-radius: 12px;
  background: var(--na-muted);
  border: 1px solid var(--na-border);

  .tile-icon {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 32px;
    height: 32px;
    border-radius: 8px;
    background: var(--na-card);
    color: var(--na-primary);
    font-size: 15px;

    &.is-health {
      color: #0284c7;

      &.is-perfect {
        color: var(--na-success);
      }
    }

    &.is-security {
      color: #f59e0b;
    }

    &.is-audit {
      color: var(--na-success);
    }
  }

  .tile-body {
    display: flex;
    flex-direction: column;
  }

  .tile-label {
    font-size: 11px;
    color: var(--na-muted-foreground);
  }

  .tile-val {
    font-size: 13.5px;
    font-weight: 700;
    color: var(--na-foreground);

    &.is-perfect-text {
      color: var(--na-success);
    }
  }
}

.hero-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

/* 2. 下半部分 3 列全屏 Bento 布局 */
.profile-grid {
  display: grid;
  grid-template-columns: minmax(0, 1.15fr) minmax(0, 1.25fr) minmax(0, 1fr);
  gap: 16px;
  align-items: stretch;
}

/* 通用卡片外壳 */
.profile-card {
  padding: 20px 22px;
  border-radius: 16px;
  border: 1px solid var(--na-border);
  background: var(--na-card);
  box-shadow: 0 4px 16px -2px rgb(0 0 0 / 4%);
  display: flex;
  flex-direction: column;
  transition: all 180ms ease;
}

.card-head {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 18px;

  .head-icon {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 36px;
    height: 36px;
    border-radius: 10px;
    font-size: 18px;
    flex-shrink: 0;

    &.is-user {
      background: var(--na-primary-soft);
      color: var(--na-primary);
    }

    &.is-security {
      background: rgba(14, 165, 233, 0.12);
      color: #0284c7;
    }

    &.is-health {
      background: var(--na-success-soft);
      color: var(--na-success);
    }

    &.is-nav {
      background: var(--na-primary-soft);
      color: var(--na-primary);
    }
  }

  .head-text {
    h2 {
      margin: 0;
      font-size: 16px;
      font-weight: 700;
      color: var(--na-foreground);
      line-height: 1.3;
    }

    p {
      margin: 2px 0 0;
      font-size: 12px;
      color: var(--na-muted-foreground);
    }
  }
}

/* 栏目 1：账号核心资料 */
.account-items-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
  margin-bottom: 16px;
}

.account-item {
  padding: 12px 14px;
  border-radius: 12px;
  border: 1px solid var(--na-border);
  background: color-mix(in srgb, var(--na-muted) 50%, var(--na-card));
  transition: all 160ms ease;

  &:hover {
    border-color: color-mix(in srgb, var(--na-primary) 35%, var(--na-border));
    background: var(--na-card);
  }

  .item-label {
    display: flex;
    align-items: center;
    justify-content: space-between;
    font-size: 11.5px;
    color: var(--na-muted-foreground);

    .icon-muted {
      font-size: 13px;
    }
  }

  .item-val {
    margin: 6px 0 2px;
    font-size: 14px;
    font-weight: 700;
    color: var(--na-foreground);

    &.status-val {
      display: inline-flex;
      align-items: center;
      gap: 6px;
      color: var(--na-success);
    }
  }

  .item-desc {
    font-size: 11px;
    color: var(--na-muted-foreground);
  }
}

.authority-section {
  margin-bottom: 16px;
  padding: 12px 14px;
  border-radius: 12px;
  background: color-mix(in srgb, var(--na-muted) 35%, var(--na-card));
  border: 1px solid var(--na-border);

  .section-title {
    font-size: 12px;
    font-weight: 600;
    color: var(--na-muted-foreground);
    margin-bottom: 8px;
  }

  .role-tags-wrap {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 6px;
  }

  .role-badge {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    padding: 3px 9px;
    border-radius: 6px;
    background: var(--na-primary-soft);
    color: var(--na-primary);
    font-size: 12px;
    font-weight: 600;
    border: 1px solid color-mix(in srgb, var(--na-primary) 20%, transparent);
  }
}

.sys-meta-panel {
  margin-top: auto;
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 12px 14px;
  border-radius: 12px;
  background: color-mix(in srgb, var(--na-muted) 45%, var(--na-card));
  border: 1px solid var(--na-border);

  .meta-row-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    font-size: 12px;

    .meta-k {
      color: var(--na-muted-foreground);
    }

    .meta-v {
      font-weight: 600;
      color: var(--na-foreground);
    }
  }
}

/* 栏目 2：联系方式与安全中心 */
.security-banner {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 10px 14px;
  border-radius: 12px;
  background: var(--na-success-soft);
  border: 1px solid color-mix(in srgb, var(--na-success) 25%, transparent);
  margin-bottom: 14px;

  .shield-icon {
    color: var(--na-success);
    font-size: 16px;
    margin-top: 2px;
    flex-shrink: 0;
  }

  .banner-body {
    strong {
      display: block;
      color: var(--na-success);
      font-size: 12.5px;
      font-weight: 700;
    }

    p {
      margin: 2px 0 0;
      color: var(--na-muted-foreground);
      font-size: 11.5px;
      line-height: 1.4;
    }
  }
}

.contact-card-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 16px;
  border-radius: 12px;
  border: 1px solid var(--na-border);
  background: color-mix(in srgb, var(--na-muted) 35%, var(--na-card));
  margin-bottom: 12px;
  transition: all 160ms ease;

  &:hover {
    border-color: color-mix(in srgb, var(--na-primary) 35%, var(--na-border));
    background: var(--na-card);
  }

  .contact-icon {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 36px;
    height: 36px;
    border-radius: 10px;
    font-size: 16px;
    flex-shrink: 0;

    &.is-phone {
      background: rgba(14, 165, 233, 0.12);
      color: #0284c7;
    }

    &.is-email {
      background: var(--na-primary-soft);
      color: var(--na-primary);
    }

    &.is-pwd {
      background: rgba(245, 158, 11, 0.12);
      color: #d97706;
    }
  }

  .contact-main {
    flex: 1 1 auto;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .contact-top {
    display: flex;
    align-items: center;
    gap: 8px;

    .contact-title {
      font-size: 12px;
      color: var(--na-muted-foreground);
    }

    .tag-verified {
      display: inline-flex;
      align-items: center;
      padding: 0 6px;
      border-radius: 4px;
      font-size: 10.5px;
      font-weight: 650;
      background: var(--na-success-soft);
      color: var(--na-success);

      &.is-pwd {
        background: var(--na-primary-soft);
        color: var(--na-primary);
      }
    }

    .tag-unverified {
      display: inline-flex;
      align-items: center;
      padding: 0 6px;
      border-radius: 4px;
      font-size: 10.5px;
      font-weight: 650;
      background: color-mix(in srgb, var(--na-muted) 80%, var(--na-card));
      color: var(--na-muted-foreground);
    }
  }

  .contact-number {
    font-size: 14px;
    font-weight: 700;
    color: var(--na-foreground);
  }

  .contact-tip {
    font-size: 11px;
    color: var(--na-muted-foreground);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .contact-action-btn {
    border-radius: 8px;
    flex-shrink: 0;
  }
}

/* 栏目 3：健康度与快捷直达 */
.col-side-group {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.card-health {
  .health-progress-wrap {
    display: flex;
    align-items: center;
    gap: 16px;
    margin-bottom: 16px;

    .ring-wrapper {
      flex-shrink: 0;
    }

    .health-progress-info {
      strong {
        display: block;
        font-size: 14.5px;
        font-weight: 700;
        color: var(--na-foreground);
      }

      p {
        margin: 4px 0 0;
        font-size: 12px;
        color: var(--na-muted-foreground);
        line-height: 1.4;
      }
    }
  }

  .health-check-list {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 6px;

    li {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 8px 12px;
      border-radius: 8px;
      background: color-mix(in srgb, var(--na-muted) 45%, var(--na-card));
      border: 1px solid transparent;
      font-size: 12.5px;
      transition: all 140ms ease;

      &.is-done {
        border-color: color-mix(in srgb, var(--na-success) 20%, transparent);

        .icon-done {
          color: var(--na-success);
        }
      }

      &:not(.is-done) {
        border-color: color-mix(in srgb, var(--na-border) 60%, transparent);

        .icon-pending {
          color: var(--na-muted-foreground);
        }
      }
    }

    .check-title {
      display: flex;
      align-items: center;
      gap: 8px;
      color: var(--na-foreground);
      font-weight: 550;
    }

    .check-chip {
      padding: 1px 7px;
      border-radius: 4px;
      font-size: 11px;
      font-weight: 600;

      &.is-chip-done {
        background: var(--na-success-soft);
        color: var(--na-success);
      }

      &.is-chip-pending {
        background: var(--na-muted);
        color: var(--na-muted-foreground);
      }
    }
  }
}

.card-quick {
  .quick-nav-group {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .quick-btn {
    display: flex;
    align-items: center;
    justify-content: space-between;
    width: 100%;
    padding: 10px 14px;
    border-radius: 10px;
    border: 1px solid var(--na-border);
    background: color-mix(in srgb, var(--na-muted) 35%, var(--na-card));
    cursor: pointer;
    transition: all 160ms ease;

    &:hover {
      border-color: var(--na-primary);
      background: var(--na-primary-soft);
      transform: translateX(2px);

      .arrow-icon {
        color: var(--na-primary);
        transform: translateX(2px);
      }
    }

    .btn-left {
      display: flex;
      align-items: center;
      gap: 10px;
      font-size: 13px;
      font-weight: 600;
      color: var(--na-foreground);

      .icon-amber { color: #f59e0b; font-size: 16px; }
      .icon-blue { color: #0ea5e9; font-size: 16px; }
      .icon-purple { color: #8b5cf6; font-size: 16px; }
    }

    .arrow-icon {
      color: var(--na-muted-foreground);
      font-size: 14px;
      transition: all 160ms ease;
    }
  }
}

/* 弹窗通用样式 */
.dialog-header-custom,
.contact-dialog__heading {
  display: flex;
  align-items: center;
  gap: 12px;

  h3 {
    margin: 0;
    font-size: 16px;
    font-weight: 700;
    color: var(--na-foreground);
  }

  p {
    margin: 2px 0 0;
    font-size: 12px;
    color: var(--na-muted-foreground);
  }
}

.dialog-icon-circle,
.contact-dialog__icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border-radius: 12px;
  font-size: 18px;
  flex-shrink: 0;

  &.is-lock {
    background: var(--na-primary-soft);
    color: var(--na-primary);
  }

  &--email {
    background: rgba(14, 165, 233, 0.12);
    color: #0284c7;
  }
}

.contact-dialog__icon:not(.contact-dialog__icon--email) {
  background: rgba(14, 165, 233, 0.12);
  color: #0284c7;
}

.dialog-modern-form,
.contact-form {
  padding-top: 10px;

  :deep(.el-form-item__label) {
    font-weight: 600;
    font-size: 13px;
    color: var(--na-foreground);
    padding-bottom: 6px;
  }
}

.contact-verification-state {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 10px 14px;
  border-radius: 10px;
  background: var(--na-muted);
  border: 1px solid var(--na-border);
  font-size: 12px;
  margin-bottom: 14px;

  &.is-ready {
    background: var(--na-success-soft);
    border-color: color-mix(in srgb, var(--na-success) 30%, transparent);
    color: var(--na-success);
  }

  strong {
    display: block;
    margin-bottom: 2px;
  }

  span {
    color: var(--na-muted-foreground);
  }
}

.contact-code-row {
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 10px;
  width: 100%;
}

.contact-field-hint {
  margin: 6px 0 0;
  font-size: 11.5px;
  color: var(--na-muted-foreground);
}

.dialog-footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 10px;
}

/* 响应式适配 */
@media (max-width: 1320px) {
  .metric-tiles {
    grid-template-columns: repeat(2, auto);
  }

  .profile-grid {
    grid-template-columns: 1fr 1fr;
  }

  .col-side-group {
    grid-column: 1 / -1;
    display: grid;
    grid-template-columns: 1fr 1fr;
  }
}

@media (max-width: 960px) {
  .profile-hero-banner {
    flex-direction: column;
    align-items: stretch;
  }

  .hero-right {
    justify-content: space-between;
  }

  .profile-grid {
    grid-template-columns: 1fr;
  }

  .col-side-group {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 640px) {
  .hero-left {
    flex-direction: column;
    align-items: flex-start;
  }

  .metric-tiles {
    grid-template-columns: 1fr;
    width: 100%;
  }

  .account-items-grid {
    grid-template-columns: 1fr;
  }
}
</style>
