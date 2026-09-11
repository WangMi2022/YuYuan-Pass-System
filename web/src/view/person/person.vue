<template>
  <div class="na-page profile-container">
    <!-- 头部个人档案名片 Hero -->
    <section class="profile-hero" aria-labelledby="profile-title">
      <div class="profile-banner">
        <div class="profile-banner__copy">
          <span class="profile-kicker">✦ 账户档案 · 个人主页</span>
          <p>维护个人核心档案、通讯凭据与系统权限范围</p>
        </div>
        <div class="profile-banner__status">
          <span class="status-chip" :class="`status--${accountStatus.type}`">
            <span class="status-beacon" />
            {{ accountStatus.label }} · 企业安全认证
          </span>
        </div>
      </div>

      <div class="profile-identity">
        <!-- 头像区 -->
        <div class="profile-avatar-wrapper">
          <div class="avatar-ring-outer">
            <SelectImage
              v-model="userStore.userInfo.headerImg"
              file-type="image"
              :preview-url="userStore.userInfo.headerImgPreviewUrl"
              :loading="savingAvatar"
              rounded
            />
          </div>
          <span class="profile-avatar-note">
            <el-icon><Camera /></el-icon>
            点击更换头像
          </span>
        </div>

        <!-- 姓名、账号、角色 -->
        <div class="profile-identity__main">
          <div class="profile-name-row">
            <template v-if="!editFlag">
              <h1 id="profile-title">{{ displayName }}</h1>
              <el-tooltip content="编辑显示昵称" placement="top">
                <el-button text circle :icon="Edit" class="edit-nickname-btn" aria-label="编辑昵称" @click="openEdit" />
              </el-tooltip>
            </template>
            <template v-else>
              <div class="profile-name-edit-box">
                <el-input v-model="nickName" class="profile-name-input" placeholder="请输入新昵称" aria-label="昵称" />
                <el-button type="primary" :loading="savingNickname" @click="enterEdit">保存</el-button>
                <el-button plain @click="closeEdit">取消</el-button>
              </div>
            </template>
          </div>

          <div class="profile-account-line">
            <span class="account-handle">@{{ userStore.userInfo.userName || '未设置账号' }}</span>
            <span class="profile-dot">·</span>
            <span class="primary-role-badge">
              <el-icon><Stamp /></el-icon>
              {{ primaryRole }}
            </span>
          </div>

          <div class="profile-identity__tags">
            <span class="identity-tag is-uid">UID #{{ userId }}</span>
            <span v-if="roleNames.length" class="identity-tag is-role-count">
              <el-icon><User /></el-icon>
              {{ roleNames.length }} 个权限组
            </span>
            <span class="identity-tag is-security">
              <el-icon><CircleCheck /></el-icon>
              安全策略正常
            </span>
          </div>
        </div>

        <!-- 右侧核心指标小卡片 -->
        <div class="profile-hero-stats">
          <div class="stat-card">
            <div class="stat-card__icon"><el-icon><Calendar /></el-icon></div>
            <div class="stat-card__info">
              <span class="stat-card__label">加入系统</span>
              <strong class="stat-card__value">{{ createdAtText }}</strong>
            </div>
          </div>
          <div class="stat-card">
            <div class="stat-card__icon" :class="{ 'is-full': profileCompletion === 100 }">
              <el-icon><Compass /></el-icon>
            </div>
            <div class="stat-card__info">
              <span class="stat-card__label">资料完整度</span>
              <strong class="stat-card__value" :class="{ 'text-success': profileCompletion === 100 }">
                {{ profileCompletion }}%
              </strong>
            </div>
          </div>
          <div class="stat-card">
            <div class="stat-card__icon"><el-icon><Lock /></el-icon></div>
            <div class="stat-card__info">
              <span class="stat-card__label">密码安全</span>
              <strong class="stat-card__value">已启用</strong>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- 下半部分两栏 Bento 布局 -->
    <div class="profile-content">
      <!-- 左侧：资料卡片与安全中心 -->
      <main class="profile-main">
        <!-- 卡片 1：账号核心资料 -->
        <section class="profile-card" aria-labelledby="account-info-title">
          <header class="profile-card__header">
            <div class="card-title-group">
              <div class="card-title-icon is-user">
                <el-icon><User /></el-icon>
              </div>
              <div>
                <h2 id="account-info-title">账号基本资料</h2>
                <p>用于识别当前账号体系与全局权限范围的基本标识</p>
              </div>
            </div>
          </header>

          <div class="profile-detail-grid">
            <div class="profile-detail-item">
              <div class="detail-item-top">
                <span class="profile-detail-label">系统登录账号</span>
                <el-icon class="detail-item-icon"><User /></el-icon>
              </div>
              <strong class="detail-item-val font-mono">{{ userStore.userInfo.userName || '未设置' }}</strong>
              <small class="detail-item-desc">唯一凭证，不可在此页面变更</small>
            </div>

            <div class="profile-detail-item">
              <div class="detail-item-top">
                <span class="profile-detail-label">当前显示昵称</span>
                <el-button text type="primary" size="small" :icon="Edit" @click="openEdit">编辑</el-button>
              </div>
              <strong class="detail-item-val">{{ displayName }}</strong>
              <small class="detail-item-desc">用于内部协同沟通与操作日志展示</small>
            </div>

            <div class="profile-detail-item">
              <div class="detail-item-top">
                <span class="profile-detail-label">核心主角色</span>
                <el-icon class="detail-item-icon"><Stamp /></el-icon>
              </div>
              <strong class="detail-item-val">{{ primaryRole }}</strong>
              <small class="detail-item-desc">由超级管理员统一设定与授权</small>
            </div>

            <div class="profile-detail-item">
              <div class="detail-item-top">
                <span class="profile-detail-label">账号当前状态</span>
                <el-icon class="detail-item-icon"><CircleCheck /></el-icon>
              </div>
              <strong class="detail-item-val status-badge-inline">
                <span :class="['status-dot', `status-dot--${accountStatus.type}`]" />
                {{ accountStatus.label }}
              </strong>
              <small class="detail-item-desc">{{ accountStatus.description }}</small>
            </div>
          </div>

          <!-- 关联角色组展示 -->
          <div v-if="roleNames.length > 1" class="profile-roles-card">
            <div class="roles-header">
              <span class="profile-detail-label">全部所属权限角色组 ({{ roleNames.length }})</span>
            </div>
            <div class="profile-role-list">
              <span v-for="role in roleNames" :key="role" class="role-pill-tag">
                <el-icon><Stamp /></el-icon>
                {{ role }}
              </span>
            </div>
          </div>
        </section>

        <!-- 卡片 2：联系方式与安全中枢 -->
        <section class="profile-card" aria-labelledby="contact-info-title">
          <header class="profile-card__header">
            <div class="card-title-group">
              <div class="card-title-icon is-security">
                <el-icon><Lock /></el-icon>
              </div>
              <div>
                <h2 id="contact-info-title">联系方式与安全中心</h2>
                <p>管理密保手机、通知邮箱与登录密码，保障核心资产与数据安全</p>
              </div>
            </div>
          </header>

          <!-- 安全合规提示条 -->
          <div class="security-posture-banner">
            <el-icon><CircleCheck /></el-icon>
            <div>
              <strong>企业级安全防护已激活</strong>
              <span>所有关键凭据均进行硬件加密存储，涉及敏感操作前需二次校验密保手机或邮箱。</span>
            </div>
          </div>

          <div class="profile-contact-list">
            <!-- 手机绑定 -->
            <div class="profile-contact-row">
              <div class="profile-contact-icon is-phone">
                <el-icon><Phone /></el-icon>
              </div>
              <div class="profile-contact-info">
                <div class="contact-title-line">
                  <span class="profile-detail-label">密保手机</span>
                  <span v-if="userStore.userInfo.phone" class="verified-tag">已绑定</span>
                  <span v-else class="unverified-tag">待绑定</span>
                </div>
                <strong class="contact-val font-mono">{{ userStore.userInfo.phone || '未设置手机号' }}</strong>
                <small class="contact-hint">{{ contactCapabilities.phone.reason }}</small>
              </div>
              <el-button
                type="primary"
                plain
                class="profile-row-action"
                @click="openPhoneDialog"
              >
                {{ userStore.userInfo.phone ? '修改手机号' : '绑定手机号' }}
              </el-button>
            </div>

            <!-- 邮箱绑定 -->
            <div class="profile-contact-row">
              <div class="profile-contact-icon is-email">
                <el-icon><Message /></el-icon>
              </div>
              <div class="profile-contact-info">
                <div class="contact-title-line">
                  <span class="profile-detail-label">密保邮箱</span>
                  <span v-if="userStore.userInfo.email" class="verified-tag">已绑定</span>
                  <span v-else class="unverified-tag">待绑定</span>
                </div>
                <strong class="contact-val font-mono">{{ userStore.userInfo.email || '未设置邮箱' }}</strong>
                <small class="contact-hint">{{ contactCapabilities.email.reason }}</small>
              </div>
              <el-button
                type="primary"
                plain
                class="profile-row-action"
                @click="openEmailDialog"
              >
                {{ userStore.userInfo.email ? '修改邮箱' : '绑定邮箱' }}
              </el-button>
            </div>

            <!-- 登录密码 -->
            <div class="profile-contact-row">
              <div class="profile-contact-icon is-password">
                <el-icon><Key /></el-icon>
              </div>
              <div class="profile-contact-info">
                <div class="contact-title-line">
                  <span class="profile-detail-label">账号登录密码</span>
                  <span class="verified-tag is-pwd">安全强度 良好</span>
                </div>
                <strong class="contact-val font-mono">••••••••••••</strong>
                <small class="contact-hint">建议每 90 天定期更新一次密码，避免与其他第三方系统相同</small>
              </div>
              <el-button
                type="primary"
                plain
                class="profile-row-action"
                @click="showPassword = true"
              >
                修改密码
              </el-button>
            </div>
          </div>
        </section>
      </main>

      <!-- 右侧辅助侧栏 -->
      <aside class="profile-aside">
        <!-- 卡片 3：资料健康度与检查清单 -->
        <section class="profile-card profile-completion-card" aria-labelledby="completion-title">
          <header class="profile-card__header">
            <div class="card-title-group">
              <div class="card-title-icon is-health">
                <el-icon><CircleCheck /></el-icon>
              </div>
              <div>
                <h2 id="completion-title">资料健康度</h2>
                <p>完善个人档案以保障系统高效协同</p>
              </div>
            </div>
          </header>

          <div class="profile-progress-row">
            <div class="progress-ring-box">
              <el-progress
                type="circle"
                :percentage="profileCompletion"
                :width="88"
                :stroke-width="7"
                :color="progressColor"
              />
            </div>
            <div class="progress-copy">
              <strong>{{ completionTitle }}</strong>
              <p>{{ completionHint }}</p>
            </div>
          </div>

          <ul class="profile-check-list">
            <li
              v-for="item in profileChecks"
              :key="item.key"
              :class="{ 'is-complete': item.complete }"
            >
              <div class="check-item-left">
                <span class="check-icon">
                  <el-icon v-if="item.complete"><CircleCheck /></el-icon>
                  <el-icon v-else><Warning /></el-icon>
                </span>
                <span class="check-label">{{ item.label }}</span>
              </div>
              <span class="check-badge" :class="item.complete ? 'is-done' : 'is-pending'">
                {{ item.complete ? '已完善' : '待补充' }}
              </span>
            </li>
          </ul>
        </section>

        <!-- 卡片 4：账户系统概况 -->
        <section class="profile-card profile-security-card" aria-labelledby="security-title">
          <header class="profile-card__header">
            <div class="card-title-group">
              <div class="card-title-icon is-monitor">
                <el-icon><Monitor /></el-icon>
              </div>
              <div>
                <h2 id="security-title">账户系统概况</h2>
                <p>当前登录身份凭证元数据</p>
              </div>
            </div>
          </header>

          <dl class="profile-overview-list">
            <div class="overview-item">
              <dt>系统用户编号</dt>
              <dd><span class="font-mono uid-chip">UID #{{ userId }}</span></dd>
            </div>
            <div class="overview-item">
              <dt>核心权限组</dt>
              <dd>{{ primaryRole }}</dd>
            </div>
            <div class="overview-item">
              <dt>注册创建时间</dt>
              <dd>{{ createdAtText }}</dd>
            </div>
            <div class="overview-item">
              <dt>全链路操作审计</dt>
              <dd><span class="audit-chip">● 实时审计中</span></dd>
            </div>
          </dl>
        </section>

        <!-- 卡片 5：快捷工作导航 -->
        <section class="profile-card profile-quick-nav">
          <header class="profile-card__header">
            <div class="card-title-group">
              <div class="card-title-icon is-nav">
                <el-icon><Compass /></el-icon>
              </div>
              <div>
                <h2>快捷功能直达</h2>
                <p>常用协同中心与系统设置入口</p>
              </div>
            </div>
          </header>

          <div class="quick-nav-links">
            <button type="button" class="quick-nav-btn" @click="navTo('workSchedule')">
              <el-icon class="icon-amber"><Calendar /></el-icon>
              <span>查看我的工作日历</span>
              <el-icon class="nav-arrow"><Right /></el-icon>
            </button>
            <button type="button" class="quick-nav-btn" @click="navTo('dashboard')">
              <el-icon class="icon-blue"><Monitor /></el-icon>
              <span>返回首屏驾驶舱</span>
              <el-icon class="nav-arrow"><Right /></el-icon>
            </button>
            <button type="button" class="quick-nav-btn" @click="navTo('aiOperations')">
              <el-icon class="icon-purple"><Cpu /></el-icon>
              <span>智能能力与模型接入</span>
              <el-icon class="nav-arrow"><Right /></el-icon>
            </button>
          </div>
        </section>
      </aside>
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
.profile-container {
  --profile-gap: 20px;
  max-width: 1320px;
  margin: 0 auto;
  padding-bottom: 36px;
}

/* 头部 Hero 卡片 */
.profile-hero {
  position: relative;
  overflow: hidden;
  margin-bottom: var(--profile-gap);
  border-radius: 18px;
  border: 1px solid var(--na-border, #e2e0ec);
  background: var(--na-card, #ffffff);
  box-shadow: 0 10px 30px -6px rgba(109, 93, 251, 0.08), 0 2px 8px -2px rgba(17, 24, 39, 0.04);
}

.profile-banner {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 24px;
  min-height: 126px;
  padding: 24px 32px 42px;
  background: linear-gradient(135deg, color-mix(in srgb, var(--na-primary, #6d5dfb) 12%, #ffffff) 0%, color-mix(in srgb, var(--na-primary, #6d5dfb) 4%, var(--na-card, #ffffff)) 100%);
  border-bottom: 1px solid color-mix(in srgb, var(--na-primary, #6d5dfb) 12%, var(--na-border, #e2e0ec));

  &__copy {
    p {
      margin: 8px 0 0;
      color: var(--na-muted-foreground, #706b82);
      font-size: 13.5px;
    }
  }

  &__status {
    flex-shrink: 0;
  }
}

.profile-kicker {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 3px 10px;
  border-radius: 100px;
  background: var(--na-card, #ffffff);
  border: 1px solid color-mix(in srgb, var(--na-primary) 20%, transparent);
  color: var(--na-primary, #6d5dfb);
  font-size: 12px;
  font-weight: 700;
  box-shadow: 0 2px 6px rgba(109, 93, 251, 0.08);
}

.status-chip {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 12px;
  border-radius: 100px;
  font-size: 12px;
  font-weight: 650;
  background: var(--na-card, #ffffff);
  border: 1px solid var(--na-border, #e2e0ec);
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.04);

  &.status--success {
    color: var(--na-success, #059669);
    border-color: color-mix(in srgb, var(--na-success) 30%, transparent);

    .status-beacon {
      background: var(--na-success, #059669);
      box-shadow: 0 0 0 2px rgba(5, 150, 105, 0.2);
    }
  }

  &.status--danger {
    color: var(--na-danger, #dc2626);
    border-color: color-mix(in srgb, var(--na-danger) 30%, transparent);

    .status-beacon {
      background: var(--na-danger, #dc2626);
      box-shadow: 0 0 0 2px rgba(220, 38, 38, 0.2);
    }
  }
}

.status-beacon {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  animation: beacon-pulse 2s infinite ease-out;
}

/* 个人信息主区 */
.profile-identity {
  display: grid;
  grid-template-columns: 140px minmax(0, 1fr) auto;
  align-items: center;
  gap: 28px;
  padding: 0 32px 28px;
  margin-top: -46px;
}

.profile-avatar-wrapper {
  position: relative;
  z-index: 2;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;

  .avatar-ring-outer {
    padding: 4px;
    border-radius: 50%;
    background: linear-gradient(135deg, var(--na-primary, #6d5dfb) 0%, #38bdf8 100%);
    box-shadow: 0 8px 22px rgba(109, 93, 251, 0.25);
    transition: transform 0.25s ease;

    &:hover {
      transform: scale(1.03);
    }
  }

  :deep(.select-image-root),
  :deep(.w-40) {
    width: 114px;
    height: 114px;
    border: 3px solid var(--na-card, #ffffff);
    border-radius: 50%;
    background: var(--na-muted, #f3f2f7);
    overflow: hidden;
  }
}

.profile-avatar-note {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  color: var(--na-muted-foreground, #706b82);
  font-size: 11.5px;
  white-space: nowrap;
}

.profile-identity__main {
  min-width: 0;
  padding-top: 36px;
}

.profile-name-row {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;

  h1 {
    margin: 0;
    color: var(--na-foreground, #19172c);
    font-size: 26px;
    font-weight: 750;
    line-height: 1.2;
    letter-spacing: -0.01em;
  }
}

.edit-nickname-btn {
  color: var(--na-muted-foreground, #706b82);
  transition: all 0.15s ease;

  &:hover {
    color: var(--na-primary, #6d5dfb);
    background: var(--na-primary-soft, rgba(109, 93, 251, 0.12));
  }
}

.profile-name-edit-box {
  display: flex;
  align-items: center;
  gap: 8px;

  .profile-name-input {
    width: 220px;
  }
}

.profile-account-line {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
  margin: 6px 0 10px;
  color: var(--na-muted-foreground, #706b82);
  font-size: 13.5px;

  .account-handle {
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
    font-weight: 600;
    color: var(--na-foreground, #19172c);
  }

  .primary-role-badge {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    padding: 1px 8px;
    border-radius: 6px;
    background: var(--na-primary-soft, rgba(109, 93, 251, 0.1));
    color: var(--na-primary, #6d5dfb);
    font-size: 12px;
    font-weight: 650;
  }

  .profile-dot {
    color: var(--na-border-strong, #d5d1e2);
  }
}

.profile-identity__tags {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
}

.identity-tag {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 3px 9px;
  border-radius: 6px;
  font-size: 11.5px;
  font-weight: 600;
  border: 1px solid var(--na-border, #e2e0ec);
  background: var(--na-muted, #f3f2f7);
  color: var(--na-muted-foreground, #706b82);

  &.is-uid {
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
    color: var(--na-foreground, #19172c);
    font-weight: 700;
  }

  &.is-role-count {
    background: var(--na-primary-soft, rgba(109, 93, 251, 0.08));
    color: var(--na-primary, #6d5dfb);
    border-color: color-mix(in srgb, var(--na-primary) 20%, transparent);
  }

  &.is-security {
    background: var(--na-success-soft, rgba(5, 150, 105, 0.08));
    color: var(--na-success, #059669);
    border-color: color-mix(in srgb, var(--na-success) 20%, transparent);
  }
}

/* 顶部右侧指标卡片 */
.profile-hero-stats {
  display: grid;
  grid-template-columns: repeat(3, minmax(100px, 1fr));
  gap: 12px;
  padding-left: 24px;
  margin-top: 32px;
  border-left: 1px solid var(--na-border, #e2e0ec);
}

.stat-card {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 14px;
  border-radius: 12px;
  background: var(--na-muted, #f3f2f7);
  border: 1px solid var(--na-border, #e2e0ec);

  &__icon {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 34px;
    height: 34px;
    border-radius: 8px;
    background: var(--na-card, #ffffff);
    color: var(--na-primary, #6d5dfb);
    font-size: 16px;
    flex-shrink: 0;

    &.is-full {
      color: var(--na-success, #059669);
    }
  }

  &__info {
    display: flex;
    flex-direction: column;
    min-width: 0;
  }

  &__label {
    color: var(--na-muted-foreground, #706b82);
    font-size: 11px;
    white-space: nowrap;
  }

  &__value {
    color: var(--na-foreground, #19172c);
    font-size: 14px;
    font-weight: 700;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;

    &.text-success {
      color: var(--na-success, #059669);
    }
  }
}

/* 下半部分双栏布局 */
.profile-content {
  display: grid;
  grid-template-columns: minmax(0, 1.6fr) minmax(320px, 0.9fr);
  gap: var(--profile-gap);
  align-items: start;
}

.profile-main,
.profile-aside {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: var(--profile-gap);
}

/* 通用模块卡片 */
.profile-card {
  padding: 24px;
  border-radius: 16px;
  border: 1px solid var(--na-border, #e2e0ec);
  background: var(--na-card, #ffffff);
  box-shadow: 0 4px 16px -2px rgba(17, 24, 39, 0.04);
  transition: all 0.2s ease;

  &__header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 16px;
    margin-bottom: 20px;

    .card-title-group {
      display: flex;
      align-items: center;
      gap: 12px;
    }

    .card-title-icon {
      display: flex;
      align-items: center;
      justify-content: center;
      width: 38px;
      height: 38px;
      border-radius: 10px;
      font-size: 18px;
      flex-shrink: 0;

      &.is-user {
        background: var(--na-primary-soft, rgba(109, 93, 251, 0.12));
        color: var(--na-primary, #6d5dfb);
      }

      &.is-security {
        background: rgba(14, 165, 233, 0.12);
        color: #0284c7;
      }

      &.is-health {
        background: var(--na-success-soft, rgba(5, 150, 105, 0.12));
        color: var(--na-success, #059669);
      }

      &.is-monitor {
        background: rgba(245, 158, 11, 0.12);
        color: #d97706;
      }

      &.is-nav {
        background: var(--na-primary-soft, rgba(109, 93, 251, 0.12));
        color: var(--na-primary, #6d5dfb);
      }
    }

    h2 {
      margin: 0;
      color: var(--na-foreground, #19172c);
      font-size: 16.5px;
      font-weight: 700;
      line-height: 1.3;
    }

    p {
      margin: 4px 0 0;
      color: var(--na-muted-foreground, #706b82);
      font-size: 12.5px;
    }
  }
}

/* 账号基本资料网格 */
.profile-detail-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px;
}

.profile-detail-item {
  display: flex;
  flex-direction: column;
  padding: 16px 18px;
  border-radius: 12px;
  border: 1px solid var(--na-border, #e2e0ec);
  background: var(--na-table-header, #faf9fc);
  transition: all 0.2s ease;

  &:hover {
    border-color: color-mix(in srgb, var(--na-primary) 35%, var(--na-border));
    background: var(--na-card, #ffffff);
    box-shadow: 0 4px 12px rgba(109, 93, 251, 0.05);
  }

  .detail-item-top {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .detail-item-icon {
    color: var(--na-muted-foreground, #706b82);
    font-size: 15px;
  }

  .profile-detail-label {
    color: var(--na-muted-foreground, #706b82);
    font-size: 12px;
    font-weight: 600;
  }

  .detail-item-val {
    margin: 8px 0 4px;
    color: var(--na-foreground, #19172c);
    font-size: 15px;
    font-weight: 700;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;

    &.font-mono {
      font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
    }
  }

  .detail-item-desc {
    color: var(--na-muted-foreground, #706b82);
    font-size: 11.5px;
  }
}

.status-badge-inline {
  display: inline-flex;
  align-items: center;
  gap: 7px;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;

  &--success { background: var(--na-success, #059669); }
  &--danger { background: var(--na-danger, #dc2626); }
}

/* 权限角色组展示卡片 */
.profile-roles-card {
  margin-top: 16px;
  padding: 14px 18px;
  border-radius: 12px;
  background: var(--na-muted, #f3f2f7);
  border: 1px solid var(--na-border, #e2e0ec);

  .roles-header {
    margin-bottom: 10px;
  }
}

.profile-role-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.role-pill-tag {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 4px 10px;
  border-radius: 8px;
  background: var(--na-card, #ffffff);
  border: 1px solid var(--na-border, #e2e0ec);
  color: var(--na-foreground, #19172c);
  font-size: 12px;
  font-weight: 600;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);

  .el-icon {
    color: var(--na-primary, #6d5dfb);
  }
}

/* 联系方式与安全中枢卡片 */
.security-posture-banner {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 12px 16px;
  margin-bottom: 18px;
  border-radius: 10px;
  background: var(--na-success-soft, rgba(5, 150, 105, 0.08));
  border: 1px solid color-mix(in srgb, var(--na-success) 22%, transparent);
  color: var(--na-success, #059669);

  .el-icon {
    font-size: 18px;
    margin-top: 2px;
    flex-shrink: 0;
  }

  strong {
    display: block;
    font-size: 13px;
    font-weight: 650;
    color: var(--na-foreground, #19172c);
  }

  span {
    display: block;
    margin-top: 2px;
    color: var(--na-muted-foreground, #706b82);
    font-size: 12px;
    line-height: 1.5;
  }
}

.profile-contact-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.profile-contact-row {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px 18px;
  border-radius: 12px;
  border: 1px solid var(--na-border, #e2e0ec);
  background: var(--na-card, #ffffff);
  transition: all 0.2s ease;

  &:hover {
    border-color: var(--na-border-strong, #d5d1e2);
    background: var(--na-table-hover, #f6f4fc);
    transform: translateX(2px);
  }
}

.profile-contact-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 42px;
  height: 42px;
  border-radius: 10px;
  font-size: 18px;
  flex-shrink: 0;

  &.is-phone {
    background: rgba(2, 132, 199, 0.12);
    color: #0284c7;
  }

  &.is-email {
    background: var(--na-success-soft, rgba(5, 150, 105, 0.12));
    color: var(--na-success, #059669);
  }

  &.is-password {
    background: var(--na-primary-soft, rgba(109, 93, 251, 0.12));
    color: var(--na-primary, #6d5dfb);
  }
}

.profile-contact-info {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-width: 0;

  .contact-title-line {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .contact-val {
    margin: 4px 0 2px;
    color: var(--na-foreground, #19172c);
    font-size: 14.5px;
    font-weight: 650;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .contact-hint {
    color: var(--na-muted-foreground, #706b82);
    font-size: 11.5px;
  }
}

.verified-tag {
  display: inline-flex;
  align-items: center;
  padding: 1px 6px;
  border-radius: 4px;
  background: var(--na-success-soft, rgba(5, 150, 105, 0.1));
  color: var(--na-success, #059669);
  font-size: 10.5px;
  font-weight: 700;

  &.is-pwd {
    background: var(--na-primary-soft, rgba(109, 93, 251, 0.1));
    color: var(--na-primary, #6d5dfb);
  }
}

.unverified-tag {
  display: inline-flex;
  align-items: center;
  padding: 1px 6px;
  border-radius: 4px;
  background: var(--na-warning-soft, rgba(217, 119, 6, 0.1));
  color: var(--na-warning, #d97706);
  font-size: 10.5px;
  font-weight: 700;
}

.profile-row-action {
  flex-shrink: 0;
  border-radius: 8px;
  font-weight: 600;
}

/* 侧边栏卡片：资料完整度 */
.profile-progress-row {
  display: flex;
  align-items: center;
  gap: 18px;
  padding: 6px 0 18px;
  border-bottom: 1px solid var(--na-border, #e2e0ec);

  .progress-ring-box {
    flex-shrink: 0;
  }

  .progress-copy {
    strong {
      display: block;
      color: var(--na-foreground, #19172c);
      font-size: 15px;
      font-weight: 700;
    }

    p {
      margin: 4px 0 0;
      color: var(--na-muted-foreground, #706b82);
      font-size: 12px;
      line-height: 1.5;
    }
  }
}

.profile-check-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin: 16px 0 0;
  padding: 0;
  list-style: none;

  li {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 8px 12px;
    border-radius: 8px;
    background: var(--na-muted, #f3f2f7);
    transition: all 0.15s ease;

    .check-item-left {
      display: flex;
      align-items: center;
      gap: 8px;
    }

    .check-icon {
      display: inline-flex;
      font-size: 15px;
      color: var(--na-muted-foreground, #706b82);
    }

    .check-label {
      color: var(--na-foreground, #19172c);
      font-size: 12.5px;
      font-weight: 600;
    }

    .check-badge {
      display: inline-flex;
      font-size: 11px;
      font-weight: 600;
      padding: 1px 6px;
      border-radius: 4px;

      &.is-done {
        color: var(--na-success, #059669);
        background: var(--na-success-soft, rgba(5, 150, 105, 0.12));
      }

      &.is-pending {
        color: var(--na-warning, #d97706);
        background: var(--na-warning-soft, rgba(217, 119, 6, 0.12));
      }
    }

    &.is-complete {
      .check-icon {
        color: var(--na-success, #059669);
      }
    }
  }
}

/* 侧边栏卡片：账户系统概览 */
.profile-overview-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin: 0;

  .overview-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 8px 0;
    border-bottom: 1px dashed var(--na-border, #e2e0ec);

    &:last-child {
      border-bottom: 0;
    }

    dt {
      color: var(--na-muted-foreground, #706b82);
      font-size: 12.5px;
    }

    dd {
      margin: 0;
      color: var(--na-foreground, #19172c);
      font-size: 13px;
      font-weight: 650;
      text-align: right;
    }
  }
}

.uid-chip {
  padding: 2px 6px;
  border-radius: 4px;
  background: var(--na-muted, #f3f2f7);
  font-size: 12px;
}

.audit-chip {
  color: var(--na-success, #059669);
  font-size: 12px;
}

/* 快捷直达卡片 */
.quick-nav-links {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.quick-nav-btn {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  padding: 11px 14px;
  border: 1px solid var(--na-border, #e2e0ec);
  border-radius: 10px;
  background: var(--na-card, #ffffff);
  color: var(--na-foreground, #19172c);
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;

  &:hover {
    border-color: var(--na-primary, #6d5dfb);
    background: var(--na-primary-soft, rgba(109, 93, 251, 0.06));
    color: var(--na-primary, #6d5dfb);
    transform: translateX(3px);

    .nav-arrow {
      color: var(--na-primary, #6d5dfb);
      transform: translateX(2px);
    }
  }

  .icon-amber { color: #d97706; font-size: 16px; margin-right: 8px; }
  .icon-blue { color: #0284c7; font-size: 16px; margin-right: 8px; }
  .icon-purple { color: #8b5cf6; font-size: 16px; margin-right: 8px; }

  .nav-arrow {
    color: var(--na-muted-foreground, #706b82);
    font-size: 13px;
    transition: all 0.2s ease;
  }
}

/* 弹窗通用现代设计 */
.modern-profile-dialog {
  border-radius: 16px;
  overflow: hidden;

  :deep(.el-dialog__header) {
    margin: 0;
    padding: 20px 24px 16px;
    border-bottom: 1px solid var(--na-border, #e2e0ec);
  }

  :deep(.el-dialog__body) {
    padding: 20px 24px 8px;
  }

  :deep(.el-dialog__footer) {
    padding: 14px 24px 20px;
    border-top: 1px solid var(--na-border, #e2e0ec);
  }
}

.dialog-header-custom {
  display: flex;
  align-items: center;
  gap: 12px;

  .dialog-icon-circle {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 40px;
    height: 40px;
    border-radius: 10px;
    background: var(--na-primary-soft, rgba(109, 93, 251, 0.12));
    color: var(--na-primary, #6d5dfb);
    font-size: 18px;
    flex-shrink: 0;
  }

  h3 {
    margin: 0;
    color: var(--na-foreground, #19172c);
    font-size: 17px;
    font-weight: 700;
  }

  p {
    margin: 3px 0 0;
    color: var(--na-muted-foreground, #706b82);
    font-size: 12px;
  }
}

.dialog-modern-form {
  :deep(.el-form-item__label) {
    font-size: 13px;
    font-weight: 600;
    color: var(--na-foreground, #19172c);
    padding-bottom: 6px;
  }

  :deep(.el-input__wrapper) {
    border-radius: 10px;
    background: var(--na-table-header, #faf9fc);
    box-shadow: 0 0 0 1px var(--na-border, #e2e0ec) inset;
    min-height: 42px;

    &:hover {
      box-shadow: 0 0 0 1px var(--na-primary, #6d5dfb) inset;
    }

    &.is-focus {
      box-shadow: 0 0 0 2px var(--na-primary, #6d5dfb) inset;
      background: var(--na-card, #ffffff);
    }
  }
}

/* 手机与邮箱弹窗特定样式 */
.contact-dialog__heading {
  display: flex;
  align-items: center;
  gap: 12px;
}

.contact-dialog__icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border-radius: 10px;
  background: rgba(2, 132, 199, 0.12);
  color: #0284c7;
  font-size: 18px;
  flex-shrink: 0;

  &--email {
    background: var(--na-success-soft, rgba(5, 150, 105, 0.12));
    color: var(--na-success, #059669);
  }
}

.contact-dialog__heading h3 {
  margin: 0;
  color: var(--na-foreground, #19172c);
  font-size: 17px;
  font-weight: 700;
}

.contact-dialog__heading p {
  margin: 3px 0 0;
  color: var(--na-muted-foreground, #706b82);
  font-size: 12px;
}

.contact-verification-state {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 12px 14px;
  border-radius: 10px;
  background: var(--na-warning-soft, rgba(217, 119, 6, 0.1));
  color: var(--na-warning, #d97706);
  font-size: 12px;

  &.is-ready {
    background: var(--na-success-soft, rgba(5, 150, 105, 0.1));
    color: var(--na-success, #059669);
  }

  strong {
    display: block;
    color: var(--na-foreground, #19172c);
    font-size: 13px;
    font-weight: 650;
  }

  span {
    display: block;
    margin-top: 2px;
    color: var(--na-muted-foreground, #706b82);
  }
}

.contact-form {
  margin-top: 16px;

  :deep(.el-form-item__label) {
    font-size: 13px;
    font-weight: 600;
    color: var(--na-foreground, #19172c);
    padding-bottom: 6px;
  }

  :deep(.el-input__wrapper) {
    border-radius: 10px;
    min-height: 42px;
  }
}

.contact-code-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 130px;
  gap: 10px;
  width: 100%;

  .el-button {
    border-radius: 10px;
  }
}

.contact-field-hint {
  margin: 6px 0 0;
  color: var(--na-muted-foreground, #706b82);
  font-size: 11.5px;
}

.dialog-footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 10px;

  .el-button {
    border-radius: 8px;
    min-width: 80px;
  }
}

/* 响应式适配 */
@media (max-width: 960px) {
  .profile-content {
    grid-template-columns: 1fr;
  }

  .profile-hero-stats {
    grid-column: 1 / -1;
    border-left: 0;
    border-top: 1px solid var(--na-border, #e2e0ec);
    padding-left: 0;
    padding-top: 16px;
    margin-top: 16px;
  }
}

@media (max-width: 680px) {
  .profile-identity {
    grid-template-columns: 100px minmax(0, 1fr);
    padding: 0 20px 20px;
    gap: 16px;
    margin-top: -32px;
  }

  .profile-avatar-wrapper {
    :deep(.select-image-root),
    :deep(.w-40) {
      width: 90px;
      height: 90px;
    }
  }

  .profile-detail-grid {
    grid-template-columns: 1fr;
  }

  .profile-hero-stats {
    grid-template-columns: 1fr;
  }
}
</style>
