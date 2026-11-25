const MICROS_PER_YUAN = 1_000_000

const finiteNumber = (value) => {
  const number = Number(value)
  return Number.isFinite(number) ? number : 0
}

export const defaultQuota = () => ({
  ID: 0,
  scopeType: 'global',
  scopeId: 'global',
  dailyRequests: 0,
  dailyTokens: 0,
  monthlyBudgetYuan: 0,
  maxConcurrency: 0,
  enabled: true
})

export const quotaFormValue = (value = {}) => ({
  ...defaultQuota(),
  ...value,
  monthlyBudgetYuan: Math.max(0, finiteNumber(value.monthlyCostMicros) / MICROS_PER_YUAN)
})

export const quotaPayloadValue = (value = {}) => {
  const payload = {
    ...value,
    monthlyCostMicros: Math.max(0, Math.round(finiteNumber(value.monthlyBudgetYuan) * MICROS_PER_YUAN))
  }
  delete payload.monthlyBudgetYuan
  return payload
}
