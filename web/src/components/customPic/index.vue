<template>
  <span class="headerAvatar">
    <template v-if="picType === 'avatar'">
      <el-avatar v-if="hasAvatar" :size="30" :src="avatar" @error="handleAvatarError" />
      <el-avatar v-else :size="30" :src="noAvatar" />
    </template>
    <template v-if="picType === 'img'">
      <img v-if="hasAvatar" :src="avatar" alt="用户头像" class="avatar" @error="handleAvatarError" />
      <img v-else :src="noAvatar" alt="默认用户头像" class="avatar" />
    </template>
    <template v-if="picType === 'file'">
      <el-image
        :src="file"
        alt="文件预览"
        class="file"
        :preview-src-list="previewSrcList"
        :preview-teleported="true"
      />
    </template>
  </span>
</template>

<script setup>
  import noAvatarPng from '@/assets/noBody.png'
  import { useUserStore } from '@/pinia/modules/user'
  import { resolveAvatarUrl } from '@/utils/avatar'
  import { computed, ref, watch } from 'vue'

  defineOptions({
    name: 'CustomPic'
  })

  const props = defineProps({
    picType: {
      type: String,
      required: false,
      default: 'avatar'
    },
    picSrc: {
      type: String,
      required: false,
      default: ''
    },
    preview: {
      type: Boolean,
      default: false
    }
  })

  const noAvatar = ref(noAvatarPng)
  const avatarFailed = ref(false)

  const userStore = useUserStore()

  const fileBaseUrl = import.meta.env.VITE_FILE_API || import.meta.env.VITE_BASE_API || '/'
  const avatar = computed(() => resolveAvatarUrl({
    picSrc: props.picSrc,
    headerImgPreviewUrl: userStore.userInfo.headerImgPreviewUrl,
    headerImg: userStore.userInfo.headerImg,
    fileBaseUrl
  }))
  const hasAvatar = computed(() => Boolean(avatar.value) && !avatarFailed.value)
  const file = computed(() => resolveAvatarUrl({ picSrc: props.picSrc, fileBaseUrl }))
  const previewSrcList = computed(() => (props.preview && file.value ? [file.value] : []))

  const handleAvatarError = () => {
    avatarFailed.value = true
  }

  watch(avatar, () => {
    avatarFailed.value = false
  })
</script>

<style scoped>
  .headerAvatar {
    display: flex;
    justify-content: center;
    align-items: center;
    margin-right: 8px;
  }
  .file {
    display: block;
    overflow: hidden;
    width: 80px;
    height: 80px;
    position: relative;
    border: 1px solid var(--na-border);
    border-radius: 8px;
    background: var(--na-muted);
  }
  .file :deep(.el-image__inner) {
    display: block;
    width: 100%;
    height: 100%;
    object-fit: contain;
  }
</style>
