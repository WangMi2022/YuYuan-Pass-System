<template>
  <div class="profile-container">
    <!-- 顶部个人信息卡片 -->
    <div class="bg-white dark:bg-slate-800 rounded-2xl shadow-sm mb-8">
      <!-- 顶部背景图 -->
      <div class="h-48 bg-blue-50 dark:bg-slate-600 relative">
        <div class="absolute inset-0 bg-pattern opacity-7"></div>
      </div>

      <!-- 个人信息区 -->
      <div class="px-8 -mt-20 pb-8">
        <div class="flex flex-col lg:flex-row items-start gap-8">
          <!-- 左侧头像 -->
          <div class="profile-avatar-wrapper flex-shrink-0 mx-auto lg:mx-0">
            <SelectImage
                v-model="userStore.userInfo.headerImg"
                file-type="image"
                rounded
            />
          </div>

          <!-- 右侧信息 -->
          <div class="flex-1 pt-12 lg:pt-20 w-full">
            <div
              class="flex flex-col lg:flex-row items-start lg:items-start justify-between gap-4"
            >
              <div class="lg:mt-4">
                <div class="flex items-center gap-4 mb-4">
                  <div
                    v-if="!editFlag"
                    class="text-2xl font-bold flex items-center gap-3 text-gray-800 dark:text-gray-100"
                  >
                    {{ userStore.userInfo.nickName }}
                    <el-icon
                      class="cursor-pointer text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 transition-colors duration-200"
                      @click="openEdit"
                    >
                      <edit />
                    </el-icon>
                  </div>
                  <div v-else class="flex items-center">
                    <el-input v-model="nickName" class="w-48 mr-4" />
                    <el-button type="primary" plain @click="enterEdit">
                      确认
                    </el-button>
                    <el-button type="danger" plain @click="closeEdit">
                      取消
                    </el-button>
                  </div>
                </div>

              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 主要内容区 -->
    <div class="profile-content">
      <!-- 左侧信息栏 -->
      <div>
        <div
          class="bg-white dark:bg-slate-800 rounded-xl p-6 mb-6 profile-card"
        >
          <h2 class="text-lg font-semibold mb-4 flex items-center gap-2">
            <el-icon class="text-blue-500"><info-filled /></el-icon>
            基本信息
          </h2>
          <div class="space-y-4">
            <div
              class="flex items-center gap-1 lg:gap-3 text-gray-600 dark:text-gray-300"
            >
              <el-icon class="text-blue-500"><phone /></el-icon>
              <span class="font-medium">手机号码：</span>
              <span>{{ userStore.userInfo.phone || '未设置' }}</span>
              <el-button
                link
                type="primary"
                class="ml-auto"
                @click="changePhoneFlag = true"
              >
                修改
              </el-button>
            </div>
            <div
              class="flex items-center gap-1 lg:gap-3 text-gray-600 dark:text-gray-300"
            >
              <el-icon class="text-green-500"><message /></el-icon>
              <span class="font-medium flex-shrink-0">邮箱地址：</span>
              <span>{{ userStore.userInfo.email || '未设置' }}</span>
              <el-button
                link
                type="primary"
                class="ml-auto"
                @click="changeEmailFlag = true"
              >
                修改
              </el-button>
            </div>
            <div
              class="flex items-center gap-1 lg:gap-3 text-gray-600 dark:text-gray-300"
            >
              <el-icon class="text-purple-500"><lock /></el-icon>
              <span class="font-medium">账号密码：</span>
              <span>已设置</span>
              <el-button
                link
                type="primary"
                class="ml-auto"
                @click="showPassword = true"
              >
                修改
              </el-button>
            </div>
          </div>
        </div>

      </div>
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
          <el-button type="primary" @click="savePassword">确 定</el-button>
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
          <el-button type="primary" @click="changePhone">确 定</el-button>
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
          <el-button type="primary" @click="changeEmail">确 定</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
  import { setSelfInfo, changePassword } from '@/api/user.js'
  import { reactive, ref, watch } from 'vue'
  import { ElMessage } from 'element-plus'
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
    modifyPwdForm.value.validate((valid) => {
      if (valid) {
        changePassword({
          password: pwdModify.value.password,
          newPassword: pwdModify.value.newPassword
        }).then((res) => {
          if (res.code === 0) {
            ElMessage.success('修改密码成功！')
          }
          showPassword.value = false
        })
      }
    })
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
    const res = await setSelfInfo({
      nickName: nickName.value
    })
    if (res.code === 0) {
      userStore.ResetUserInfo({ nickName: nickName.value })
      ElMessage.success('修改成功')
    }
    nickName.value = ''
    editFlag.value = false
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
    const res = await setSelfInfo({ phone: phoneForm.phone })
    if (res.code === 0) {
      ElMessage.success('修改成功')
      userStore.ResetUserInfo({ phone: phoneForm.phone })
      closeChangePhone()
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
    const res = await setSelfInfo({ email: emailForm.email })
    if (res.code === 0) {
      ElMessage.success('修改成功')
      userStore.ResetUserInfo({ email: emailForm.email })
      closeChangeEmail()
    }
  }

  watch(() => userStore.userInfo.headerImg, async(val) => {
    const res = await setSelfInfo({ headerImg: val })
    if (res.code === 0) {
      userStore.ResetUserInfo({ headerImg: val })
      ElMessage({
        type: 'success',
        message: '设置成功',
      })
    }
  })

</script>

<style scoped lang="scss">
  .profile-container {
    @apply p-4 lg:p-6 min-h-screen bg-gray-50 dark:bg-slate-900;

    .bg-pattern {
      background-image: url("data:image/svg+xml,%3Csvg width='60' height='60' viewBox='0 0 60 60' xmlns='http://www.w3.org/2000/svg'%3E%3Cg fill='none' fill-rule='evenodd'%3E%3Cg fill='%23000000' fill-opacity='0.1'%3E%3Cpath d='M36 34v-4h-2v4h-4v2h4v4h2v-4h4v-2h-4zm0-30V0h-2v4h-4v2h4v4h2V6h4V4h-4zM6 34v-4H4v4H0v2h4v4h2v-4h4v-2H6zM6 4V0H4v4H0v2h4v4h2V6h4V4H6z'/%3E%3C/g%3E%3C/g%3E%3C/svg%3E");
    }

    .profile-card {
      @apply shadow-sm hover:shadow-md transition-shadow duration-300;
    }

    .profile-content {
      max-width: 860px;
    }

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

    .edit-input {
      :deep(.el-input__wrapper) {
        @apply bg-white/10 border-white/20 shadow-none;
        input {
          @apply text-white;
          &::placeholder {
            @apply text-white/60;
          }
        }
      }
    }
  }
</style>
