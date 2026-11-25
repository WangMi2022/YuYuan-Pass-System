<template>
  <div
      class="w-40 h-40 relative rounded border border-dashed border-gray-300 cursor-pointer group"
      :class="[{ 'rounded-full': rounded, 'is-saving': loading }]"
  >
    <div class="w-full h-full overflow-hidden" :class="rounded ? 'rounded-full' : ''">
      <el-icon
          v-if="isVideoExt(model || '')"
          :size="32"
          class="absolute top-[calc(50%-16px)] left-[calc(50%-16px)]"
      >
        <VideoPlay />
      </el-icon>
      <video
          v-if="isVideoExt(model || '')"
          class="w-full h-full object-cover"
          muted
          preload="metadata"
      >
        <source :src="getUrl(model) + '#t=1'" />
      </video>

      <el-image
          v-if="model && !isVideoExt(model)"
          class="w-full h-full"
          :src="imgUrl"
          :preview-src-list="rounded ? [] : srcList"
          fit="cover"
          @click="handleImageClick"
      >
        <template #error>
          <img :src="fallbackImage" alt="默认用户头像" class="w-full h-full object-cover" />
        </template>
      </el-image>
      <div
          v-else
          class="text-gray-600 group-hover:bg-gray-200 group-hover:opacity-60 w-full h-full flex justify-center items-center"
          @click="chooseItem"
      >
        <el-icon>
          <plus />
        </el-icon>
        上传
      </div>
    </div>
    <!-- 删除按钮在外层容器中 -->
    <div
        v-if="model"
        class="right-0 top-0 hidden text-gray-400 group-hover:flex justify-center items-center absolute z-10"
        @click="deleteItem"
    >
      <el-icon :size="24">
        <CircleCloseFilled />
      </el-icon>
    </div>
    <span v-if="loading" class="select-image-saving" aria-label="头像保存中">
      <el-icon class="is-loading"><Loading /></el-icon>
    </span>
  </div>
</template>
<script setup>
  import fallbackImage from '@/assets/noBody.png'
  import { getUrl, isVideoExt } from '@/utils/image'
  import { CircleCloseFilled, Loading, Plus } from '@element-plus/icons-vue'
  import { computed } from 'vue'

  const props = defineProps({
    model: {
      default: '',
      type: String
    },
    rounded: {
      default: false,
      type: Boolean
    },
    previewUrl: {
      default: '',
      type: String
    },
    loading: {
      default: false,
      type: Boolean
    }
  })

  const emits = defineEmits(['chooseItem', 'deleteItem'])

  const chooseItem = () => {
    emits('chooseItem')
  }

  const deleteItem = () => {
    emits('deleteItem')
  }

  const handleImageClick = () => {
    if (props.rounded) chooseItem()
  }

  const imgUrl = computed(() => {
    return getUrl(props.previewUrl || props.model)
  })

  const srcList = computed(() => {
    return imgUrl.value ? [imgUrl.value] : []
  })
</script>

<style scoped>
.is-saving {
  pointer-events: none;
}

.select-image-saving {
  position: absolute;
  inset: 0;
  z-index: 20;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--na-primary);
  font-size: 24px;
  border-radius: inherit;
  background: color-mix(in srgb, var(--na-card) 72%, transparent);
}
</style>
