<template>
  <main class="na-page na-page--list legacy-admin-page legacy-admin-page--user">
    <AppPageHeader
      title-id="user-management-title"
      title="用户管理"
      description="维护系统账户、角色归属和启用状态，支持快速检索与密码重置。"
    >
      <template #actions>
        <el-button type="primary" :icon="Plus" @click="addUser">新增用户</el-button>
      </template>
    </AppPageHeader>

    <warning-bar title="注：右上角头像下拉可切换角色" />

    <div class="gva-search-box">
      <el-form ref="searchForm" :inline="true" :model="searchInfo" class="search-form-inline">
        <el-form-item label="用户名">
          <el-input v-model="searchInfo.username" placeholder="请输入用户名" :prefix-icon="User" clearable class="compact-search-input" />
        </el-form-item>
        <el-form-item label="昵称">
          <el-input v-model="searchInfo.nickname" placeholder="请输入昵称" :prefix-icon="EditPen" clearable class="compact-search-input" />
        </el-form-item>
        <el-form-item label="手机号">
          <el-input v-model="searchInfo.phone" placeholder="请输入手机号" :prefix-icon="Iphone" clearable class="compact-search-input" />
        </el-form-item>
        <el-form-item label="邮箱">
          <el-input v-model="searchInfo.email" placeholder="请输入邮箱" :prefix-icon="Message" clearable class="compact-search-input" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="onSubmit">
            查询
          </el-button>
          <el-button :icon="Refresh" @click="onReset"> 重置 </el-button>
        </el-form-item>
      </el-form>
    </div>

    <div class="gva-table-box">
      <el-table
        :data="tableData"
        row-key="ID"
        :default-sort="{ prop: 'ID', order: 'descending' }"
        @sort-change="sortChange"
      >
        <el-table-column align="center" label="头像" width="76">
          <template #default="scope">
            <div class="table-avatar-cell">
              <CustomPic :pic-src="scope.row.headerImgPreviewUrl || scope.row.headerImg" />
            </div>
          </template>
        </el-table-column>
        <el-table-column align="center" label="ID" width="70" prop="ID" sortable="custom">
          <template #default="scope">
            <span class="font-mono text-xs text-slate-500 font-semibold">#{{ scope.row.ID }}</span>
          </template>
        </el-table-column>
        <el-table-column
          align="left"
          label="用户名"
          min-width="140"
          prop="userName"
        >
          <template #default="scope">
            <span class="font-mono text-xs font-semibold text-slate-700 dark:text-slate-200">@{{ scope.row.userName }}</span>
          </template>
        </el-table-column>
        <el-table-column
          align="left"
          label="昵称"
          min-width="140"
          prop="nickName"
        >
          <template #default="scope">
            <span class="font-medium">{{ scope.row.nickName || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column
          align="left"
          label="手机号"
          min-width="150"
          prop="phone"
        >
          <template #default="scope">
            <span class="font-mono text-xs text-slate-600 dark:text-slate-300">{{ scope.row.phone || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column
          align="left"
          label="邮箱"
          min-width="180"
          prop="email"
        >
          <template #default="scope">
            <span class="font-mono text-xs text-slate-600 dark:text-slate-300">{{ scope.row.email || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column align="left" label="用户角色" min-width="200">
          <template #default="scope">
            <el-cascader
              v-model="scope.row.authorityIds"
              :options="authOptions"
              :show-all-levels="false"
              collapse-tags
              collapse-tags-tooltip
              :max-collapse-tags="2"
              :props="{
                multiple: true,
                checkStrictly: true,
                label: 'authorityName',
                value: 'authorityId',
                disabled: 'disabled',
                emitPath: false
              }"
              :clearable="false"
              @visible-change="
                (flag) => {
                  changeAuthority(scope.row, flag, 0)
                }
              "
              @remove-tag="
                (removeAuth) => {
                  changeAuthority(scope.row, false, removeAuth)
                }
              "
            />
          </template>
        </el-table-column>
        <el-table-column align="center" label="状态" width="100">
          <template #default="scope">
            <el-switch
              v-model="scope.row.enable"
              inline-prompt
              :active-value="1"
              :inactive-value="2"
              active-text="启用"
              inactive-text="禁用"
              @change="
                () => {
                  switchEnable(scope.row)
                }
              "
            />
          </template>
        </el-table-column>

        <el-table-column label="操作" :min-width="appStore.operateMinWith" fixed="right">
          <template #default="scope">
            <div class="table-actions">
              <el-button
                type="primary"
                link
                :icon="Edit"
                @click="openEdit(scope.row)"
              >编辑</el-button>
              <el-button
                type="primary"
                link
                :icon="Key"
                @click="resetPasswordFunc(scope.row)"
              >重置密码</el-button>
              <el-button
                type="danger"
                link
                :icon="Delete"
                :loading="deletingUserId === scope.row.ID"
                @click="deleteUserFunc(scope.row)"
              >删除</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <div class="gva-pagination">
        <el-pagination
          :current-page="page"
          :page-size="pageSize"
          :page-sizes="[10, 30, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next"
          @current-change="handleCurrentChange"
          @size-change="handleSizeChange"
        />
      </div>
    </div>

    <!-- 重置密码对话框 -->
    <el-dialog
      v-model="resetPwdDialog"
      title="重置登录密码"
      width="480px"
      :close-on-click-modal="false"
      :close-on-press-escape="false"
      class="reset-pwd-dialog"
    >
      <div class="reset-pwd-header-card">
        <div class="reset-icon-pill">
          <el-icon><Key /></el-icon>
        </div>
        <div class="reset-info">
          <div class="reset-title">为「{{ resetPwdInfo.nickName || resetPwdInfo.userName }}」重置密码</div>
          <div class="reset-desc">密码重置后立即生效，原密码失效。建议生成高强度密码。</div>
        </div>
      </div>

      <el-form :model="resetPwdInfo" ref="resetPwdForm" label-position="top" class="compact-form">
        <div class="form-row-2">
          <el-form-item label="用户账号">
            <el-input v-model="resetPwdInfo.userName" disabled :prefix-icon="User" />
          </el-form-item>
          <el-form-item label="用户昵称">
            <el-input v-model="resetPwdInfo.nickName" disabled :prefix-icon="EditPen" />
          </el-form-item>
        </div>
        <el-form-item label="新登录密码" required>
          <div class="flex w-full gap-2 items-center">
            <el-input
              class="flex-1"
              v-model="resetPwdInfo.password"
              placeholder="请输入新密码（至少6位）"
              show-password
              :prefix-icon="Lock"
            />
            <el-button type="primary" plain :icon="MagicStick" @click="generateRandomPassword">
              生成密码
            </el-button>
          </div>
        </el-form-item>
      </el-form>

      <template #footer>
        <div class="dialog-footer">
          <el-button @click="closeResetPwdDialog">取 消</el-button>
          <el-button type="primary" :loading="resetPwdSubmitting" @click="confirmResetPassword">确认重置</el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 用户新增 / 编辑抽屉 (轻量紧凑优雅风) -->
    <el-drawer
      v-model="addUserDialog"
      :size="userDrawerSize"
      :show-close="true"
      :close-on-click-modal="false"
      class="user-compact-drawer"
      @close="closeAddUserDialog"
    >
      <template #header>
        <div class="drawer-header-clean">
          <div class="header-main">
            <h3 class="header-title">{{ dialogFlag === 'add' ? '新增系统用户' : '编辑用户资料' }}</h3>
            <span v-if="dialogFlag === 'edit' && userInfo.ID" class="header-uid-badge font-mono">
              UID #{{ userInfo.ID }}
            </span>
          </div>
          <p class="header-sub">
            {{ dialogFlag === 'add' ? '录入账号凭证、通讯方式并分配权限角色' : `维护用户 @${userInfo.userName || ''} 的账户资料与授权` }}
          </p>
        </div>
      </template>

      <div class="drawer-body-clean">
        <!-- 顶部轻量头像设置区 -->
        <div class="avatar-hero-clean">
          <div class="avatar-wrapper">
            <SelectImage
              v-model="userInfo.headerImg"
              file-type="image"
              rounded
            />
          </div>
          <div class="avatar-meta">
            <div class="avatar-label">{{ dialogFlag === 'add' ? '设置用户头像' : (userInfo.nickName || '用户头像') }}</div>
            <div class="avatar-help">点击图片即可上传或从媒体库选择，静态资源已启用本地离线缓存</div>
          </div>
        </div>

        <!-- 紧凑表单容器：彻底解决输入框过长问题 -->
        <div class="user-form-container">
          <el-form
            ref="userForm"
            :rules="rules"
            :model="userInfo"
            label-position="top"
            class="compact-user-form"
          >
            <!-- 分组 1: 账号与凭据 -->
            <div class="form-sub-header">
              <span>账号凭证</span>
            </div>

            <el-form-item label="登录账号" prop="userName">
              <el-input
                v-model.trim="userInfo.userName"
                :disabled="dialogFlag === 'edit'"
                placeholder="请输入用户名（至少5位）"
                :prefix-icon="User"
                clearable
              />
              <span v-if="dialogFlag === 'edit'" class="field-hint">用户名为系统唯一标识，创建后不可修改</span>
            </el-form-item>

            <el-form-item
              v-if="dialogFlag === 'add'"
              label="初始密码"
              prop="password"
            >
              <el-input
                v-model.trim="userInfo.password"
                type="password"
                show-password
                placeholder="设置初始密码（至少6位）"
                :prefix-icon="Lock"
              />
            </el-form-item>

            <div v-else class="edit-pwd-hint-row">
              <el-icon><Lock /></el-icon>
              <span>密码已加密保护。如需更改，可在列表点击「重置密码」</span>
            </div>

            <el-form-item label="用户昵称" prop="nickName">
              <el-input
                v-model.trim="userInfo.nickName"
                placeholder="例如：张三、财务主管"
                :prefix-icon="EditPen"
                clearable
              />
            </el-form-item>

            <!-- 分组 2: 联络信息 -->
            <div class="form-sub-header">
              <span>联络通讯</span>
            </div>

            <el-form-item label="手机号码" prop="phone">
              <el-input
                v-model.trim="userInfo.phone"
                placeholder="11位手机号（选填）"
                :prefix-icon="Iphone"
                clearable
              />
            </el-form-item>

            <el-form-item label="电子邮箱" prop="email">
              <el-input
                v-model.trim="userInfo.email"
                placeholder="user@example.com（选填）"
                :prefix-icon="Message"
                clearable
              />
            </el-form-item>

            <!-- 分组 3: 角色与状态 -->
            <div class="form-sub-header">
              <span>权限与状态</span>
            </div>

            <el-form-item label="分配角色" prop="authorityIds">
              <el-cascader
                v-model="userInfo.authorityIds"
                :options="authOptions"
                :show-all-levels="false"
                collapse-tags
                collapse-tags-tooltip
                :max-collapse-tags="2"
                :props="{
                  multiple: true,
                  checkStrictly: true,
                  label: 'authorityName',
                  value: 'authorityId',
                  disabled: 'disabled',
                  emitPath: false
                }"
                placeholder="请选择分配的角色组（支持多选）"
                clearable
                @change="handleAuthorityChange"
              />
            </el-form-item>

            <el-form-item label="账号启用状态" class="mb-0">
              <div class="status-clean-row">
                <el-switch
                  v-model="userInfo.enable"
                  inline-prompt
                  :active-value="1"
                  :inactive-value="2"
                  active-text="启用"
                  inactive-text="禁用"
                />
                <span class="status-clean-tip" :class="{ 'is-active': userInfo.enable === 1 }">
                  {{ userInfo.enable === 1 ? '账号处于「正常启用」状态，允许登录系统' : '账号处于「已禁用」状态，禁止登录系统' }}
                </span>
              </div>
            </el-form-item>
          </el-form>
        </div>
      </div>

      <template #footer>
        <div class="drawer-footer-clean">
          <el-button @click="closeAddUserDialog">取消</el-button>
          <el-button type="primary" :loading="formSubmitting" @click="enterAddUserDialog">
            {{ dialogFlag === 'add' ? '立即创建' : '保存修改' }}
          </el-button>
        </div>
      </template>
    </el-drawer>
  </main>
</template>

<script setup>
  import {
    getUserList,
    setUserAuthorities,
    register,
    deleteUser,
    setUserInfo,
    resetPassword
  } from '@/api/user'

  import { getAuthorityList } from '@/api/authority'
  import CustomPic from '@/components/customPic/index.vue'
  import WarningBar from '@/components/warningBar/warningBar.vue'
  import { fetchAndCacheAvatar } from '@/utils/avatarCache'

  import { computed, nextTick, ref, watch } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import SelectImage from '@/components/selectImage/selectImage.vue'
  import { useAppStore } from '@/pinia'
  import { toSQLLine } from '@/utils/stringFun'
  import { usePagedList } from '@/hooks/usePagedList'
  import AppPageHeader from '@/components/page/AppPageHeader.vue'

  import {
    Delete,
    Edit,
    EditPen,
    Iphone,
    Key,
    Lock,
    MagicStick,
    Message,
    Plus,
    Refresh,
    Search,
    User
  } from '@element-plus/icons-vue'

  defineOptions({
    name: 'User'
  })

  const appStore = useAppStore()

  // 抽屉宽度适配：紧凑 480px，杜绝输入框被横向拉得过长
  const userDrawerSize = computed(() => {
    if (appStore.drawerSize === '100%') return '100%'
    return 'min(92vw, 480px)'
  })

  // 初始化角色树选项
  const setAuthorityOptions = (AuthorityData, optionsData) => {
    AuthorityData &&
      AuthorityData.forEach((item) => {
        if (item.children && item.children.length) {
          const option = {
            authorityId: item.authorityId,
            authorityName: item.authorityName,
            children: []
          }
          setAuthorityOptions(item.children, option.children)
          optionsData.push(option)
        } else {
          const option = {
            authorityId: item.authorityId,
            authorityName: item.authorityName
          }
          optionsData.push(option)
        }
      })
  }

  const {
    search: searchInfo,
    items: tableData,
    total,
    load,
    submit,
    reset,
    changePage,
    changePageSize,
    reloadAfterRemoval
  } = usePagedList({
    defaults: {
      page: 1,
      pageSize: 10,
      orderKey: 'id',
      desc: true,
      username: '',
      nickname: '',
      phone: '',
      email: ''
    },
    request: getUserList
  })
  const page = computed(() => searchInfo.page)
  const pageSize = computed(() => searchInfo.pageSize)

  const onSubmit = () => submit()

  const onReset = () => reset()

  const sortChange = ({ prop, order }) => {
    if (prop) {
      searchInfo.orderKey = prop === 'ID' ? 'id' : toSQLLine(prop)
      searchInfo.desc = order === 'descending'
    }
    load()
  }

  const handleSizeChange = (val) => changePageSize(val)

  const handleCurrentChange = (val) => changePage(val)

  // 监听表格数据，更新 authorityIds 并自动将所有头像缓存到本地离线存储
  watch(
    () => tableData.value,
    (list) => {
      setAuthorityIds()
      if (list && list.length) {
        list.forEach((u) => {
          const pic = u.headerImgPreviewUrl || u.headerImg
          if (pic) fetchAndCacheAvatar(pic)
        })
      }
    },
    { immediate: true }
  )

  const authOptions = ref([])
  const setOptions = (authData) => {
    authOptions.value = []
    setAuthorityOptions(authData, authOptions.value)
  }

  const initPage = async () => {
    const [, res] = await Promise.all([load(), getAuthorityList()])
    if (res.code === 0) setOptions(res.data)
  }

  initPage()

  // 重置密码对话框相关
  const resetPwdDialog = ref(false)
  const resetPwdForm = ref(null)
  const resetPwdSubmitting = ref(false)
  const resetPwdInfo = ref({
    ID: '',
    userName: '',
    nickName: '',
    password: ''
  })

  // 生成随机高强度密码
  const generateRandomPassword = () => {
    const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*'
    let password = ''
    for (let i = 0; i < 12; i++) {
      password += chars.charAt(Math.floor(Math.random() * chars.length))
    }
    resetPwdInfo.value.password = password
    navigator.clipboard.writeText(password).then(() => {
      ElMessage({
        type: 'success',
        message: '密码已自动复制到剪贴板'
      })
    }).catch(() => {
      ElMessage({
        type: 'info',
        message: '密码已生成，请复制使用'
      })
    })
  }

  const resetPasswordFunc = (row) => {
    resetPwdInfo.value.ID = row.ID
    resetPwdInfo.value.userName = row.userName
    resetPwdInfo.value.nickName = row.nickName
    resetPwdInfo.value.password = ''
    resetPwdDialog.value = true
  }

  const confirmResetPassword = async () => {
    if (!resetPwdInfo.value.password) {
      ElMessage({
        type: 'warning',
        message: '请输入或生成新密码'
      })
      return
    }

    resetPwdSubmitting.value = true
    try {
      const res = await resetPassword({
        ID: resetPwdInfo.value.ID,
        password: resetPwdInfo.value.password
      })

      if (res.code === 0) {
        ElMessage({
          type: 'success',
          message: res.msg || '密码重置成功'
        })
        resetPwdDialog.value = false
      } else {
        ElMessage({
          type: 'error',
          message: res.msg || '密码重置失败'
        })
      }
    } finally {
      resetPwdSubmitting.value = false
    }
  }

  const closeResetPwdDialog = () => {
    resetPwdInfo.value.password = ''
    resetPwdDialog.value = false
  }

  const setAuthorityIds = () => {
    tableData.value &&
      tableData.value.forEach((user) => {
        user.authorityIds =
          user.authorities &&
          user.authorities.map((i) => {
            return i.authorityId
          })
      })
  }

  const deletingUserId = ref(null)
  const deleteUserFunc = async (row) => {
    if (!row?.ID || deletingUserId.value !== null) return
    deletingUserId.value = row.ID
    try {
      const confirmed = await ElMessageBox.confirm(`确定要删除用户「${row.nickName || row.userName}」吗？此操作无法撤销。`, '删除确认', {
        confirmButtonText: '确定删除',
        cancelButtonText: '取消',
        type: 'warning',
        confirmButtonClass: 'el-button--danger'
      }).catch(() => false)
      if (!confirmed) return
      const res = await deleteUser({ id: row.ID })
      if (res.code === 0) {
        ElMessage.success('用户删除成功')
        await reloadAfterRemoval()
      }
    } finally {
      deletingUserId.value = null
    }
  }

  // 用户表单状态
  const userInfo = ref({
    userName: '',
    password: '',
    nickName: '',
    headerImg: '',
    authorityId: '',
    authorityIds: [],
    phone: '',
    email: '',
    enable: 1
  })

  // 当头像被选择或更新时，即刻缓存到本地持久化
  watch(
    () => userInfo.value.headerImg,
    (imgUrl) => {
      if (imgUrl) fetchAndCacheAvatar(imgUrl)
    }
  )

  const rules = ref({
    userName: [
      { required: true, message: '请输入登录用户名', trigger: 'blur' },
      { min: 5, message: '用户名最低5位字符', trigger: 'blur' }
    ],
    password: [
      {
        validator: (rule, value, callback) => {
          if (dialogFlag.value === 'add') {
            if (!value) {
              callback(new Error('请输入初始登录密码'))
            } else if (value.length < 6) {
              callback(new Error('密码最低6位字符'))
            } else {
              callback()
            }
          } else {
            callback()
          }
        },
        trigger: 'blur'
      }
    ],
    nickName: [{ required: true, message: '请输入用户昵称', trigger: 'blur' }],
    phone: [
      {
        validator: (rule, value, callback) => {
          if (!value) {
            callback()
          } else if (!/^1[3-9]\d{9}$/.test(value)) {
            callback(new Error('请输入合法的11位手机号码'))
          } else {
            callback()
          }
        },
        trigger: 'blur'
      }
    ],
    email: [
      {
        validator: (rule, value, callback) => {
          if (!value) {
            callback()
          } else if (!/^([0-9A-Za-z\-_.]+)@([0-9a-z]+\.[a-z]{2,3}(\.[a-z]{2})?)$/.test(value)) {
            callback(new Error('请输入合法的电子邮箱地址'))
          } else {
            callback()
          }
        },
        trigger: 'blur'
      }
    ],
    authorityIds: [
      {
        validator: (rule, value, callback) => {
          if (!value || !value.length) {
            callback(new Error('请至少选择一个用户角色'))
          } else {
            callback()
          }
        },
        trigger: 'change'
      }
    ]
  })

  const handleAuthorityChange = () => {
    if (userInfo.value.authorityIds?.length) {
      userInfo.value.authorityId = userInfo.value.authorityIds[0]
    } else {
      userInfo.value.authorityId = ''
    }
  }

  const userForm = ref(null)
  const formSubmitting = ref(false)

  const enterAddUserDialog = async () => {
    if (!userForm.value) return
    userInfo.value.authorityId = userInfo.value.authorityIds?.[0] || ''
    userForm.value.validate(async (valid) => {
      if (valid) {
        formSubmitting.value = true
        try {
          const req = {
            ...userInfo.value
          }
          if (dialogFlag.value === 'add') {
            const res = await register(req)
            if (res.code === 0) {
              if (req.headerImg) fetchAndCacheAvatar(req.headerImg)
              ElMessage({ type: 'success', message: '用户创建成功' })
              await load()
              closeAddUserDialog()
            }
          }
          if (dialogFlag.value === 'edit') {
            const res = await setUserInfo(req)
            if (res.code === 0) {
              if (req.headerImg) fetchAndCacheAvatar(req.headerImg)
              ElMessage({ type: 'success', message: '用户信息更新成功' })
              await load()
              closeAddUserDialog()
            }
          }
        } finally {
          formSubmitting.value = false
        }
      }
    })
  }

  const addUserDialog = ref(false)
  const closeAddUserDialog = () => {
    if (userForm.value) {
      userForm.value.resetFields()
    }
    userInfo.value = {
      userName: '',
      password: '',
      nickName: '',
      headerImg: '',
      authorityId: '',
      authorityIds: [],
      phone: '',
      email: '',
      enable: 1
    }
    addUserDialog.value = false
  }

  const dialogFlag = ref('add')

  const addUser = () => {
    dialogFlag.value = 'add'
    userInfo.value = {
      userName: '',
      password: '',
      nickName: '',
      headerImg: '',
      authorityId: '',
      authorityIds: [],
      phone: '',
      email: '',
      enable: 1
    }
    addUserDialog.value = true
  }

  const tempAuth = {}
  const changeAuthority = async (row, flag, removeAuth) => {
    if (flag) {
      if (!removeAuth) {
        tempAuth[row.ID] = [...row.authorityIds]
      }
      return
    }
    await nextTick()
    const res = await setUserAuthorities({
      ID: row.ID,
      authorityIds: row.authorityIds
    })
    if (res.code === 0) {
      ElMessage({ type: 'success', message: '角色设置成功' })
    } else {
      if (!removeAuth) {
        row.authorityIds = [...tempAuth[row.ID]]
        delete tempAuth[row.ID]
      } else {
        row.authorityIds = [removeAuth, ...row.authorityIds]
      }
    }
  }

  const openEdit = (row) => {
    dialogFlag.value = 'edit'
    userInfo.value = JSON.parse(JSON.stringify(row))
    if (!userInfo.value.authorityIds && userInfo.value.authorities) {
      userInfo.value.authorityIds = userInfo.value.authorities.map(i => i.authorityId)
    }
    const pic = userInfo.value.headerImgPreviewUrl || userInfo.value.headerImg
    if (pic) fetchAndCacheAvatar(pic)
    addUserDialog.value = true
  }

  const switchEnable = async (row) => {
    userInfo.value = JSON.parse(JSON.stringify(row))
    await nextTick()
    const req = {
      ...userInfo.value
    }
    const res = await setUserInfo(req)
    if (res.code === 0) {
      ElMessage({
        type: 'success',
        message: `${req.enable === 2 ? '禁用' : '启用'}成功`
      })
      await load()
      userInfo.value.headerImg = ''
      userInfo.value.authorityIds = []
    }
  }
</script>

<style lang="scss" scoped>
  /* 列表检索输入框紧凑尺寸 */
  .compact-search-input {
    width: 155px !important;
  }

  /* 列表头像单元格 */
  .table-avatar-cell {
    display: flex;
    align-items: center;
    justify-content: center;

    :deep(.headerAvatar) {
      margin-right: 0;
    }
  }

  /* 表格操作栏 */
  .table-actions {
    display: flex;
    align-items: center;
    gap: 4px;
    white-space: nowrap;

    :deep(.el-button) {
      padding: 4px 6px;
      margin: 0;
      font-size: 13px;
    }
  }

  /* 抽屉头部（轻量简约） */
  .drawer-header-clean {
    display: flex;
    flex-direction: column;
    gap: 4px;

    .header-main {
      display: flex;
      align-items: center;
      gap: 8px;

      .header-title {
        margin: 0;
        font-size: 16px;
        font-weight: 700;
        color: var(--na-foreground);
      }

      .header-uid-badge {
        font-size: 11px;
        color: var(--na-muted-foreground);
        background: var(--na-muted);
        padding: 1px 6px;
        border-radius: 4px;
      }
    }

    .header-sub {
      margin: 0;
      font-size: 12px;
      color: var(--na-muted-foreground);
    }
  }

  /* 抽屉主体 */
  .drawer-body-clean {
    padding: 4px 0 20px;
  }

  /* 顶部轻量头像设置区 */
  .avatar-hero-clean {
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 12px 16px;
    margin-bottom: 12px;
    border-radius: var(--na-radius-sm);
    background: var(--na-muted);
    border: 1px solid var(--na-border);

    .avatar-wrapper {
      flex-shrink: 0;

      :deep(.w-40) {
        width: 64px !important;
        height: 64px !important;
        border: 2px solid var(--na-card);
        border-radius: 50%;
        background: var(--na-card);
        overflow: hidden;
      }
    }

    .avatar-meta {
      display: flex;
      flex-direction: column;
      gap: 3px;

      .avatar-label {
        font-size: 14px;
        font-weight: 600;
        color: var(--na-foreground);
      }

      .avatar-help {
        font-size: 11.5px;
        color: var(--na-muted-foreground);
        line-height: 1.4;
      }
    }
  }

  /* 表单容器：限制适度宽度，彻底解决输入框过长拉伸问题 */
  .user-form-container {
    width: 100%;
    max-width: 360px;
    margin: 0 auto;

    :deep(.el-form-item) {
      margin-bottom: 15px;

      .el-form-item__label {
        font-size: 13px;
        font-weight: 500;
        color: var(--na-foreground);
        padding-bottom: 4px;
        line-height: 1.3;
      }

      .el-input,
      .el-cascader {
        width: 100%;
        max-width: 340px;
      }
    }
  }

  /* 优雅极简的分组标线 */
  .form-sub-header {
    display: flex;
    align-items: center;
    font-size: 11.5px;
    font-weight: 600;
    color: var(--na-muted-foreground);
    letter-spacing: 0.04em;
    margin: 18px 0 10px;

    &::after {
      content: "";
      flex: 1;
      margin-left: 10px;
      height: 1px;
      background: var(--na-border);
      opacity: 0.8;
    }
  }

  .field-hint {
    font-size: 11px;
    color: var(--na-muted-foreground);
    margin-top: 3px;
    display: block;
  }

  /* 编辑模式密码提示 */
  .edit-pwd-hint-row {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 12px;
    color: var(--na-muted-foreground);
    padding: 8px 12px;
    border-radius: 6px;
    background: var(--na-muted);
    border: 1px dashed var(--na-border);
    margin-bottom: 15px;
    max-width: 340px;

    .el-icon {
      font-size: 14px;
      color: var(--na-primary);
      flex-shrink: 0;
    }
  }

  /* 状态开关行 */
  .status-clean-row {
    display: flex;
    align-items: center;
    gap: 10px;
    padding-top: 2px;

    .status-clean-tip {
      font-size: 12px;
      color: #ef4444;

      &.is-active {
        color: #10b981;
      }
    }
  }

  /* 抽屉底部操作栏 */
  .drawer-footer-clean {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 10px;
    width: 100%;

    .el-button {
      min-width: 80px;
    }
  }

  /* 重置密码对话框 */
  .reset-pwd-dialog {
    .reset-pwd-header-card {
      display: flex;
      align-items: center;
      gap: 12px;
      padding: 10px 14px;
      border-radius: var(--na-radius-sm);
      background: var(--na-primary-soft);
      border: 1px solid color-mix(in srgb, var(--na-primary) 18%, transparent);
      margin-bottom: 16px;

      .reset-icon-pill {
        font-size: 18px;
        color: var(--na-primary);
      }

      .reset-info {
        display: flex;
        flex-direction: column;
        gap: 2px;

        .reset-title {
          font-size: 13px;
          font-weight: 600;
          color: var(--na-foreground);
        }

        .reset-desc {
          font-size: 11px;
          color: var(--na-muted-foreground);
        }
      }
    }

    .form-row-2 {
      display: grid;
      grid-template-columns: 1fr 1fr;
      gap: 12px;
    }
  }
</style>
