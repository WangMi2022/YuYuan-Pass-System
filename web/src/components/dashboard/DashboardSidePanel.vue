<template>
  <aside class="dashboard-side-stack">
    <section class="attention-panel na-panel" aria-labelledby="attention-title">
      <header>
        <div>
          <span>需要你的关注</span>
          <strong id="attention-title">{{ attentionTitle }}</strong>
        </div>
        <b v-if="attentionCount">{{ attentionCount }}</b>
      </header>
      <div class="attention-list">
        <article v-for="(item, index) in attentions" :key="item.key || index">
          <i :class="`tone-${item.tone || index}`" />
          <span>
            <strong>{{ item.title }}</strong>
            <small>{{ item.meta }}</small>
          </span>
          <el-icon><ArrowRight /></el-icon>
        </article>
        <el-empty v-if="!attentions.length" description="暂无关注事项" :image-size="58" />
      </div>
    </section>

    <section class="quick-panel na-panel" aria-labelledby="quick-title">
      <header>
        <span id="quick-title">快速发起</span>
        <small>按权限显示</small>
      </header>
      <div class="quick-grid">
        <button
          v-for="item in actions"
          :key="item.route"
          type="button"
          @click="$emit('action', item.route)"
        >
          <span><el-icon><component :is="item.icon" /></el-icon></span>
          <strong>{{ item.title }}</strong>
          <small>{{ item.desc }}</small>
        </button>
      </div>
    </section>
  </aside>
</template>

<script setup>
import { ArrowRight } from '@element-plus/icons-vue'

defineOptions({ name: 'DashboardSidePanel' })

defineProps({
  attentionTitle: { type: String, default: '最新动态' },
  attentionCount: { type: Number, default: 0 },
  attentions: { type: Array, default: () => [] },
  actions: { type: Array, default: () => [] }
})

defineEmits(['action'])
</script>

<style scoped lang="scss">
.dashboard-side-stack {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 14px;
}

.dashboard-side-stack > section { border-radius: 16px; }

.attention-panel > header,
.quick-panel > header {
  display: flex;
  min-height: 58px;
  align-items: center;
  justify-content: space-between;
  padding: 0 16px;
  border-bottom: 1px solid var(--na-border);
}

.attention-panel > header > div {
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.attention-panel header span,
.quick-panel header small {
  color: var(--na-muted-foreground);
  font-size: 10px;
}

.attention-panel header strong,
.quick-panel header span {
  font-size: 13px;
  font-weight: 650;
}

.attention-panel header b {
  display: grid;
  width: 27px;
  height: 27px;
  place-items: center;
  border-radius: 50%;
  background: var(--na-danger-soft);
  color: var(--na-danger);
  font-size: 10px;
}

.attention-list article {
  display: grid;
  min-height: 66px;
  padding: 12px 15px;
  grid-template-columns: 4px 1fr auto;
  align-items: center;
  gap: 11px;
  border-bottom: 1px solid var(--na-border);
}

.attention-list article:last-child { border-bottom: 0; }

.attention-list article > i {
  width: 4px;
  height: 34px;
  border-radius: 99px;
  background: var(--na-danger);
}

.attention-list article > i.tone-1 { background: var(--na-warning); }
.attention-list article > i.tone-2 { background: var(--na-info); }
.attention-list article > i.tone-success { background: var(--na-success); }

.attention-list article > span {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 4px;
}

.attention-list article strong {
  overflow: hidden;
  font-size: 11px;
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.attention-list article small {
  overflow: hidden;
  color: var(--na-muted-foreground);
  font-size: 9px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.attention-list article > .el-icon {
  color: var(--na-muted-foreground);
  font-size: 12px;
}

.quick-grid {
  display: grid;
  padding: 12px;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px;
}

.quick-grid button {
  display: flex;
  min-width: 0;
  min-height: 88px;
  padding: 11px;
  flex-direction: column;
  align-items: flex-start;
  border: 1px solid var(--na-border);
  border-radius: 11px;
  background: color-mix(in srgb, var(--na-muted) 58%, var(--na-card));
  color: var(--na-foreground);
  text-align: left;
  transition: border-color 150ms ease, background-color 150ms ease, transform 150ms ease;
}

.quick-grid button:hover {
  border-color: color-mix(in srgb, var(--na-primary) 38%, var(--na-border));
  background: var(--na-primary-soft);
  transform: translateY(-1px);
}

.quick-grid button > span {
  display: grid;
  width: 29px;
  height: 29px;
  place-items: center;
  border-radius: 8px;
  background: var(--na-primary-soft);
  color: var(--na-primary);
  font-size: 14px;
}

.quick-grid button strong { margin-top: 8px; font-size: 10px; font-weight: 650; }
.quick-grid button small {
  overflow: hidden;
  width: 100%;
  margin-top: 2px;
  color: var(--na-muted-foreground);
  font-size: 8px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

@media (max-width: 980px) {
  .dashboard-side-stack { display: grid; grid-template-columns: 1fr 1fr; }
}

@media (max-width: 650px) {
  .dashboard-side-stack { display: flex; }
}
</style>
