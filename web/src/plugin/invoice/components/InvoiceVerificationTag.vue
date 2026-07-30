<template>
  <el-tag :type="meta.type" effect="plain" class="invoice-verification-tag">
    <span class="verification-dot" aria-hidden="true" />
    {{ meta.label }}
  </el-tag>
</template>

<script setup>
import { computed } from 'vue'
import { invoiceVerificationStatusMeta } from '@/plugin/invoice/utils/invoice'

defineOptions({ name: 'InvoiceVerificationTag' })

const props = defineProps({
  status: { type: String, default: 'unverified' }
})

const meta = computed(() => invoiceVerificationStatusMeta(props.status))
</script>

<style scoped>
.invoice-verification-tag { gap: 5px; }
.verification-dot { width: 6px; height: 6px; border-radius: 50%; background: currentColor; }
</style>
