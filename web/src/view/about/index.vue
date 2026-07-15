<template>
  <main class="about-page">
    <section class="about-hero">
      <div>
        <p class="eyebrow">ABOUT SYSTEM</p>
        <h1>资产管理中心</h1>
        <p>
          面向二次开发的资产管理后台，聚合资产档案、分类统计、数据大屏、Excel 导入导出与统一 OSS 存储能力。
        </p>
      </div>
      <el-tag type="success" effect="light">内部系统</el-tag>
    </section>

    <section class="about-grid">
      <el-card shadow="never" class="about-card">
        <template #header>核心能力</template>
        <ul class="feature-list">
          <li>资产分类、资产档案和责任人统一维护</li>
          <li>资产数量、原值、当前估值自动汇总</li>
          <li>资产可视化大屏已接入仪表盘首页</li>
          <li>Excel 导入、导出和模板文件统一写入 OSS</li>
        </ul>
      </el-card>

      <el-card shadow="never" class="about-card">
        <template #header>运行环境</template>
        <div class="info-list">
          <div><span>前端</span><strong>Vue 3 / Element Plus</strong></div>
          <div><span>后端</span><strong>Gin / GORM</strong></div>
          <div><span>数据库</span><strong>PostgreSQL</strong></div>
          <div><span>对象存储</span><strong>MinIO / S3 兼容 OSS</strong></div>
        </div>
      </el-card>
    </section>
  </main>
</template>

<script setup>
defineOptions({ name: 'About' })
</script>

<style scoped>
.about-page {
  display: flex;
  flex-direction: column;
  gap: 18px;
  padding: 18px;
}

.about-hero {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 20px;
  padding: 26px;
  border: 1px solid var(--el-border-color-light);
  border-radius: 18px;
  background: var(--el-bg-color);
}

.eyebrow {
  margin: 0 0 8px;
  color: var(--el-color-primary);
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.12em;
}

.about-hero h1 {
  margin: 0 0 10px;
  color: var(--el-text-color-primary);
  font-size: 28px;
  font-weight: 700;
}

.about-hero p {
  max-width: 760px;
  margin: 0;
  color: var(--el-text-color-regular);
  line-height: 1.8;
}

.about-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 18px;
}

.about-card {
  border-radius: 16px;
}

.feature-list {
  display: grid;
  gap: 12px;
  margin: 0;
  padding-left: 18px;
  color: var(--el-text-color-regular);
  line-height: 1.7;
}

.info-list {
  display: grid;
  gap: 12px;
}

.info-list div {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 12px 0;
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.info-list div:last-child {
  border-bottom: 0;
}

.info-list span {
  color: var(--el-text-color-secondary);
}

.info-list strong {
  color: var(--el-text-color-primary);
}

@media (max-width: 768px) {
  .about-page {
    padding: 12px;
  }

  .about-hero,
  .about-grid {
    grid-template-columns: 1fr;
  }

  .about-hero {
    flex-direction: column;
    padding: 20px;
  }
}
</style>
