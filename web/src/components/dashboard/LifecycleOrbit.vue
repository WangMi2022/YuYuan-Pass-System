<template>
  <section class="lifecycle-panel na-panel" aria-labelledby="lifecycle-title">
    <header class="lifecycle-panel__header">
      <div>
        <span>资产生命周期</span>
        <strong id="lifecycle-title">{{ formattedTotal }} 件实物资产</strong>
      </div>
      <span class="lifecycle-panel__sync"><i />实时数据</span>
    </header>

    <div class="lifecycle-orbit">
      <span class="lifecycle-orbit__ring is-outer" />
      <span class="lifecycle-orbit__ring is-middle" />
      <span class="lifecycle-orbit__ring is-inner" />

      <div
        class="lifecycle-orbit__core"
        :style="{ '--retention-rate': `${retentionNumber}%` }"
      >
        <div>
          <span>价值保有率</span>
          <strong>{{ retentionRate }}</strong>
          <small>当前估值 / 资产原值</small>
        </div>
      </div>

      <button
        v-for="(item, index) in items"
        :key="item.status"
        type="button"
        class="lifecycle-node"
        :class="`lifecycle-node--${index + 1}`"
        :style="{ '--node-color': item.color }"
        :aria-label="`查看${item.label}资产，共 ${item.quantity} 件`"
        @click="$emit('select', item.status)"
      >
        <span class="lifecycle-node__icon">
          <el-icon><component :is="item.icon" /></el-icon>
        </span>
        <span class="lifecycle-node__body">
          <strong>{{ item.label }}</strong>
          <span>{{ item.formattedQuantity }}</span>
          <small>{{ item.assetKinds }} 份档案</small>
        </span>
      </button>
    </div>
  </section>
</template>

<script setup>
defineOptions({ name: 'LifecycleOrbit' })

defineProps({
  formattedTotal: { type: String, default: '0' },
  retentionRate: { type: String, default: '0.0%' },
  retentionNumber: { type: Number, default: 0 },
  items: { type: Array, default: () => [] }
})

defineEmits(['select'])
</script>

<style scoped lang="scss">
.lifecycle-panel {
  min-width: 0;
  overflow: hidden;
  border-radius: 16px;
}

.lifecycle-panel__header {
  display: flex;
  height: 66px;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  border-bottom: 1px solid var(--na-border);
}

.lifecycle-panel__header > div {
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.lifecycle-panel__header span {
  color: var(--na-muted-foreground);
  font-size: 11px;
}

.lifecycle-panel__header strong {
  color: var(--na-foreground);
  font-size: 15px;
  font-weight: 650;
}

.lifecycle-panel__sync {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  font-size: 10px !important;
}

.lifecycle-panel__sync i {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: var(--na-success);
  box-shadow: 0 0 0 4px var(--na-success-soft);
}

.lifecycle-orbit {
  position: relative;
  min-height: 470px;
  overflow: hidden;
  background:
    radial-gradient(circle at 50% 50%, var(--na-primary-soft), transparent 31%),
    linear-gradient(color-mix(in srgb, var(--na-border) 28%, transparent) 1px, transparent 1px),
    linear-gradient(90deg, color-mix(in srgb, var(--na-border) 28%, transparent) 1px, transparent 1px);
  background-size: auto, 32px 32px, 32px 32px;
}

.lifecycle-orbit::before,
.lifecycle-orbit::after {
  position: absolute;
  top: 50%;
  left: 50%;
  background: color-mix(in srgb, var(--na-border) 72%, transparent);
  content: '';
  translate: -50% -50%;
}

.lifecycle-orbit::before { width: min(580px, 72%); height: 1px; }
.lifecycle-orbit::after { width: 1px; height: 76%; }

.lifecycle-orbit__ring {
  position: absolute;
  z-index: 1;
  top: 50%;
  left: 50%;
  border: 1px solid color-mix(in srgb, var(--na-primary) 20%, var(--na-border));
  border-radius: 50%;
  translate: -50% -50%;
}

.lifecycle-orbit__ring.is-outer {
  width: 430px;
  height: 430px;
  border-color: color-mix(in srgb, var(--na-primary) 12%, var(--na-border));
}

.lifecycle-orbit__ring.is-middle {
  width: 320px;
  height: 320px;
  border-style: dashed;
}

.lifecycle-orbit__ring.is-inner { width: 210px; height: 210px; }

.lifecycle-orbit__core {
  position: absolute;
  z-index: 3;
  top: 50%;
  left: 50%;
  display: grid;
  width: 158px;
  height: 158px;
  padding: 10px;
  place-items: center;
  translate: -50% -50%;
  border-radius: 50%;
  background: conic-gradient(var(--na-primary) var(--retention-rate), var(--na-muted) 0);
  box-shadow: 0 12px 34px color-mix(in srgb, var(--na-primary) 16%, transparent);
}

.lifecycle-orbit__core::before {
  position: absolute;
  inset: 10px;
  border: 1px solid color-mix(in srgb, var(--na-primary) 18%, var(--na-border));
  border-radius: 50%;
  background: var(--na-card);
  content: '';
}

.lifecycle-orbit__core > div {
  position: relative;
  z-index: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.lifecycle-orbit__core span,
.lifecycle-orbit__core small {
  color: var(--na-muted-foreground);
  font-size: 10px;
}

.lifecycle-orbit__core strong {
  margin: 5px 0 3px;
  color: var(--na-primary);
  font-size: 28px;
  font-variant-numeric: tabular-nums;
  letter-spacing: -.04em;
}

.lifecycle-orbit__core small { font-size: 9px; }

.lifecycle-node {
  position: absolute;
  z-index: 4;
  display: flex;
  min-width: 146px;
  align-items: center;
  gap: 9px;
  padding: 8px 11px 8px 8px;
  border: 1px solid var(--na-border);
  border-radius: 12px;
  background: color-mix(in srgb, var(--na-card) 94%, transparent);
  color: var(--na-foreground);
  box-shadow: var(--na-shadow-sm);
  text-align: left;
  backdrop-filter: blur(10px);
  transition: border-color 150ms ease, box-shadow 150ms ease, transform 150ms ease;
}

.lifecycle-node:hover {
  border-color: color-mix(in srgb, var(--node-color) 44%, var(--na-border));
  box-shadow: 0 9px 25px color-mix(in srgb, var(--node-color) 12%, transparent);
  transform: translateY(-2px);
}

.lifecycle-node__icon {
  display: grid;
  width: 34px;
  height: 34px;
  place-items: center;
  flex: 0 0 34px;
  border-radius: 9px;
  background: color-mix(in srgb, var(--node-color) 12%, transparent);
  color: var(--node-color);
  font-size: 15px;
}

.lifecycle-node__body {
  display: grid;
  min-width: 0;
  grid-template-columns: 1fr auto;
  column-gap: 10px;
}

.lifecycle-node__body strong {
  grid-column: 1 / -1;
  font-size: 11px;
  font-weight: 620;
}

.lifecycle-node__body > span {
  margin-top: 3px;
  color: var(--na-foreground);
  font-size: 14px;
  font-weight: 700;
  font-variant-numeric: tabular-nums;
}

.lifecycle-node__body small {
  align-self: end;
  margin-bottom: 1px;
  color: var(--na-muted-foreground);
  font-size: 8px;
}

.lifecycle-node--1 { top: 7%; left: 50%; translate: -50% 0; }
.lifecycle-node--2 { top: 29%; right: 5%; }
.lifecycle-node--3 { right: 16%; bottom: 7%; }
.lifecycle-node--4 { bottom: 7%; left: 16%; }
.lifecycle-node--5 { top: 29%; left: 5%; }

@media (max-width: 1180px) {
  .lifecycle-orbit { min-height: 430px; }
  .lifecycle-orbit__ring.is-outer { width: 380px; height: 380px; }
  .lifecycle-node--2 { right: 2%; }
  .lifecycle-node--3 { right: 10%; }
  .lifecycle-node--4 { left: 10%; }
  .lifecycle-node--5 { left: 2%; }
}

@media (max-width: 720px) {
  .lifecycle-panel__header { height: 60px; padding: 0 14px; }
  .lifecycle-orbit {
    display: grid;
    min-height: 0;
    padding: 18px 14px;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 8px;
    background: var(--na-card);
  }
  .lifecycle-orbit::before,
  .lifecycle-orbit::after,
  .lifecycle-orbit__ring { display: none; }
  .lifecycle-orbit__core {
    position: relative;
    top: auto;
    left: auto;
    width: 138px;
    height: 138px;
    margin: 3px auto 12px;
    grid-column: 1 / -1;
    translate: 0;
  }
  .lifecycle-node {
    position: static;
    min-width: 0;
    width: 100%;
    translate: 0;
  }
  .lifecycle-node:last-child { grid-column: 1 / -1; }
}

@media (max-width: 430px) {
  .lifecycle-orbit { grid-template-columns: 1fr; }
  .lifecycle-node:last-child { grid-column: auto; }
}
</style>
