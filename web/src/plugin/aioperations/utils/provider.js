const MICROS_PER_UNIT = 1_000_000

const finiteNumber = (value) => {
  const number = Number(value)
  return Number.isFinite(number) ? number : 0
}

const microsToUnit = (value) => finiteNumber(value) / MICROS_PER_UNIT
const unitToMicros = (value) => Math.max(0, Math.round(finiteNumber(value) * MICROS_PER_UNIT))

export const providerFormValue = (value = {}, defaults = {}) => ({
  ...defaults,
  ...value,
  'input-cost-per-million': microsToUnit(value['input-cost-micros-per-million']),
  'output-cost-per-million': microsToUnit(value['output-cost-micros-per-million'])
})

export const providerPayloadValue = (value = {}) => {
  const payload = {
    ...value,
    'input-cost-micros-per-million': unitToMicros(value['input-cost-per-million']),
    'output-cost-micros-per-million': unitToMicros(value['output-cost-per-million'])
  }
  delete payload['input-cost-per-million']
  delete payload['output-cost-per-million']
  delete payload['api-key-configured']
  return payload
}
