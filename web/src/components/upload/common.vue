<template>
  <div>
    <el-upload
      :action="`${getBaseUrl()}/fileUploadAndDownload/upload`"
      accept="image/*"
      :before-upload="checkFile"
      :on-error="uploadError"
      :on-success="uploadSuccess"
      :show-file-list="false"
      :data="{'classId': props.classId}"
      :headers="{'x-token': token}"
      multiple
      class="upload-btn"
    >
      <el-button type="primary" :icon="Upload">普通上传</el-button>
    </el-upload>
  </div>
</template>

<script setup>
  import { ref } from 'vue'
  import { ElMessage } from 'element-plus'
  import { isImageExt, isImageMime } from '@/utils/image'
  import { getBaseUrl } from '@/utils/format'
  import { Upload } from "@element-plus/icons-vue";
  import { useUserStore } from "@/pinia";

  defineOptions({
    name: 'UploadCommon'
  })

  const userStore = useUserStore()

  const token = userStore.token

  const props = defineProps({
    classId: {
      type: Number,
      default: 0
    }
  })

  const emit = defineEmits(['on-success'])

  const fullscreenLoading = ref(false)

  const checkFile = (file) => {
    fullscreenLoading.value = true
    const isLt500K = file.size / 1024 / 1024 < 0.5 // 500K, @todo 应支持在项目中设置
    const isImage = isImageMime(file.type) || isImageExt(file.name)
    let pass = true
    if (!isImage) {
      ElMessage.error('媒体库仅允许上传图片文件，支持 jpg、png、gif、webp、svg、bmp、avif')
      fullscreenLoading.value = false
      pass = false
    }
    if (!isLt500K && isImage) {
      ElMessage.error('未压缩的上传图片大小不能超过 500KB，请使用压缩上传')
      fullscreenLoading.value = false
      pass = false
    }

    console.log('upload file check result: ', pass)

    return pass
  }

  const uploadSuccess = (res) => {
    const { data } = res
    if (data.file) {
      emit('on-success', data.file.url)
    }
  }

  const uploadError = () => {
    ElMessage({
      type: 'error',
      message: '上传失败'
    })
    fullscreenLoading.value = false
  }
</script>
