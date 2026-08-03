<template>
  <el-input
    class="secret-input"
    :model-value="displayValue"
    :type="visible ? 'text' : 'password'"
    :placeholder="resolvedPlaceholder"
    :disabled="disabled"
    :readonly="hasRevealedValue"
    autocomplete="new-password"
    @update:model-value="handleInput"
  >
    <template #suffix>
      <el-tooltip :content="toggleHint" placement="top">
        <span class="secret-input__toggle-wrap">
          <el-button
            class="secret-input__toggle"
            text
            circle
            :icon="visible ? Hide : View"
            :loading="loading"
            :disabled="!canToggle"
            :aria-label="visible ? '隐藏密钥' : '显示密钥'"
            @click.stop="toggleVisibility"
          />
        </span>
      </el-tooltip>
    </template>
    <template v-if="$slots.append" #append>
      <slot name="append" />
    </template>
  </el-input>
</template>

<script setup>
  import { computed, onBeforeUnmount, ref, watch } from 'vue'
  import { Hide, View } from '@element-plus/icons-vue'
  import { revealSystemConfigSecret } from '@/api/system'

  defineOptions({
    name: 'SecretInput'
  })

  const props = defineProps({
    modelValue: {
      type: String,
      default: ''
    },
    modelModifiers: {
      type: Object,
      default: () => ({})
    },
    secretPath: {
      type: String,
      required: true
    },
    configured: {
      type: Boolean,
      default: false
    },
    canReveal: {
      type: Boolean,
      default: false
    },
    disabled: {
      type: Boolean,
      default: false
    },
    placeholder: {
      type: String,
      default: '请输入密钥'
    }
  })

  const emit = defineEmits(['update:modelValue', 'input'])

  const visible = ref(false)
  const loading = ref(false)
  const revealedValue = ref('')
  const hasRevealedValue = ref(false)
  let revealAttempt = 0

  const hasLocalValue = computed(() => Boolean(String(props.modelValue || '')))
  const displayValue = computed(() => hasRevealedValue.value
    ? revealedValue.value
    : props.modelValue
  )
  const canToggle = computed(() => !props.disabled && !loading.value && (
    hasLocalValue.value || (props.configured && props.canReveal)
  ))
  const resolvedPlaceholder = computed(() => {
    if (!props.configured) return props.placeholder
    return props.canReveal
      ? '已配置，点击眼睛查看'
      : '已配置，仅超级管理员可查看'
  })
  const toggleHint = computed(() => {
    if (visible.value) return '隐藏密钥'
    if (hasLocalValue.value) return '显示正在编辑的密钥'
    if (!props.configured) return '尚未配置密钥'
    if (!props.canReveal) return '仅超级管理员可查看'
    return '查看已保存密钥'
  })

  const clearRevealedValue = () => {
    revealedValue.value = ''
    hasRevealedValue.value = false
  }

  const invalidateReveal = () => {
    revealAttempt += 1
    loading.value = false
    clearRevealedValue()
  }

  const hide = () => {
    visible.value = false
    invalidateReveal()
  }

  const handleInput = (value) => {
    invalidateReveal()
    const nextValue = props.modelModifiers.trim ? value.trim() : value
    emit('update:modelValue', nextValue)
    emit('input', nextValue)
  }

  const toggleVisibility = async () => {
    if (!canToggle.value) return
    if (visible.value) {
      hide()
      return
    }
    if (hasLocalValue.value) {
      visible.value = true
      return
    }

    const currentAttempt = ++revealAttempt
    loading.value = true
    try {
      const response = await revealSystemConfigSecret(props.secretPath)
      if (
        currentAttempt !== revealAttempt ||
        response?.code !== 0 ||
        props.disabled ||
        !props.canReveal
      ) return
      revealedValue.value = response.data?.value || ''
      hasRevealedValue.value = true
      visible.value = true
    } catch {
      if (currentAttempt === revealAttempt) hide()
    } finally {
      if (currentAttempt === revealAttempt) loading.value = false
    }
  }

  watch(
    () => [props.secretPath, props.configured, props.canReveal, props.disabled],
    () => hide()
  )

  onBeforeUnmount(hide)
</script>

<style scoped>
  .secret-input__toggle-wrap {
    display: inline-flex;
    align-items: center;
    justify-content: center;
  }

  .secret-input__toggle {
    width: 28px;
    height: 28px;
    min-height: 28px;
    padding: 0;
    color: var(--na-muted-foreground);
  }

  .secret-input__toggle:not(.is-disabled):hover,
  .secret-input__toggle:not(.is-disabled):focus-visible {
    color: var(--na-primary);
    background: var(--na-primary-soft);
  }

  .secret-input :deep(.el-input__suffix-inner) {
    min-width: 28px;
  }
</style>
