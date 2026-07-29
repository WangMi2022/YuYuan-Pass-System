<template>
  <el-tag :type="meta.type" effect="light" class="invoice-status-tag">
    <span class="status-dot" aria-hidden="true" />
    {{ meta.label }}
  </el-tag>
</template>

<script setup>
import { computed } from 'vue'
import { invoiceStatusMeta } from '@/plugin/invoice/utils/invoice'

defineOptions({ name: 'InvoiceStatusTag' })

const props = defineProps({
  status: { type: String, default: '' }
})

const meta = computed(() => invoiceStatusMeta(props.status))
</script>

<style scoped>
.invoice-status-tag { gap: 5px; }
.status-dot { width: 6px; height: 6px; border-radius: 50%; background: currentColor; }
</style>
