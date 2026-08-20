<template>
  <main class="na-page na-page--list smart-settings-page">
    <AppPageHeader
      title-id="smart-settings-title"
      title="智能能力配置"
      description="管理模型接入、智能识别、网关安全、配额计费与运行审计。"
    >
      <template #actions>
        <el-button :icon="Refresh" :loading="loading" @click="loadActiveSection(true)">
          刷新
        </el-button>
        <el-button
          v-if="saveActionLabel"
          type="primary"
          :icon="Check"
          :loading="activeSaving"
          @click="saveActiveSettings"
        >
          {{ saveActionLabel }}
        </el-button>
      </template>
    </AppPageHeader>

    <div class="settings-shell">
      <nav class="settings-nav" aria-label="智能能力配置分类">
        <button
          v-for="section in settingsSections"
          :key="section.name"
          type="button"
          class="settings-nav__item"
          :class="{ 'is-active': activeSectionName === section.name }"
          :aria-current="activeSectionName === section.name ? 'page' : undefined"
          @click="activeSectionName = section.name"
        >
          <el-icon><component :is="section.icon" /></el-icon>
          <span>
            <strong>{{ section.label }}</strong>
            <small>{{ section.hint }}</small>
          </span>
        </button>
      </nav>

      <section class="settings-content" :aria-labelledby="'settings-section-' + activeSectionName">
        <header class="settings-section-header">
          <div>
            <h2 :id="'settings-section-' + activeSectionName">{{ activeSection.label }}</h2>
            <p>{{ activeSection.description }}</p>
            <small v-if="currentSaveTime" class="save-state">已从服务端回读 · {{ currentSaveTime }}</small>
          </div>
          <el-tag v-if="sectionBadge" :type="sectionBadge.type" effect="light">
            {{ sectionBadge.label }}
          </el-tag>
        </header>

        <template v-if="activeSectionName === 'models'">
          <el-alert
            v-if="!providers.enabled"
            type="warning"
            :closable="false"
            title="统一网关当前关闭，已配置的模型不会被智能业务调用。"
          />
          <el-form label-position="top" class="settings-form">
            <div class="provider-grid">
              <section v-for="provider in providerList" :key="provider.key" class="provider-card">
                <header class="provider-card__header">
                  <div>
                    <h3>{{ provider.label }}</h3>
                    <p>{{ provider.hint }}</p>
                  </div>
                  <el-switch v-model="providers[provider.key].enabled" />
                </header>
                <div class="field-grid">
                  <el-form-item label="Base URL" class="field--wide">
                    <el-input
                      v-model.trim="providers[provider.key]['base-url']"
                      placeholder="https://api.example.com/v1"
                    />
                  </el-form-item>
                  <el-form-item label="模型">
                    <el-input v-model.trim="providers[provider.key].model" placeholder="模型名称" />
                  </el-form-item>
                  <el-form-item label="超时（秒）">
                    <el-input-number
                      v-model="providers[provider.key]['timeout-seconds']"
                      :min="1"
                      :max="120"
                      controls-position="right"
                    />
                  </el-form-item>
                  <el-form-item
                    class="field--wide"
                    :label="providers[provider.key]['api-key-configured'] ? '替换或查看 API Key' : 'API Key'"
                  >
                    <SecretInput
                      v-model.trim="providers[provider.key]['api-key']"
                      :secret-path="providerSecretPath(provider.key)"
                      :configured="providers[provider.key]['api-key-configured']"
                      :can-reveal="canRevealProviderKeys"
                      :disabled="providers[provider.key]['clear-api-key']"
                      placeholder="输入 API Key"
                    />
                  </el-form-item>
                </div>
                <div class="provider-card__footer">
                  <span v-if="providers[provider.key]['api-key-configured']" class="secret-state">密钥已安全保存</span>
                  <el-checkbox
                    v-if="providers[provider.key]['api-key-configured']"
                    v-model="providers[provider.key]['clear-api-key']"
                  >
                    清除已保存密钥
                  </el-checkbox>
                </div>
              </section>
            </div>
          </el-form>
        </template>

        <template v-else-if="activeSectionName === 'recognition'">
          <el-alert
            type="info"
            :closable="false"
            title="凭据保存后保持脱敏；超级管理员可点击眼睛临时查看。测试连接不会保存业务图片。"
          />
          <el-form label-position="top" class="settings-form">
            <section class="settings-group">
              <header class="group-heading">
                <div><h3>识别策略</h3><p>二维码优先，OCR 主识别，低置信度时调用视觉模型，最后进入人工确认。</p></div>
              </header>
              <div class="policy-row">
                <el-form-item label="视觉模型兜底阈值">
                  <el-input-number
                    v-model="invoice['fallback-threshold']"
                    :min="0.1"
                    :max="1"
                    :step="0.01"
                    :precision="2"
                    controls-position="right"
                  />
                </el-form-item>
                <div class="setting-toggle">
                  <div class="setting-toggle__copy"><strong>允许识别服务使用内网地址</strong><small>仅限受控企业内网 OCR、验真或模型网关。</small></div>
                  <el-switch v-model="invoice['allow-private-endpoints']" active-text="允许" inactive-text="拒绝" />
                </div>
              </div>
            </section>

            <section class="settings-group">
              <header class="group-heading"><div><h3>OCR 服务</h3><p>标准发票可使用百度 OCR，企业服务可接入兼容 HTTP 网关。</p></div></header>
              <div class="provider-grid">
                <section class="provider-card">
                  <header class="provider-card__header"><div><h3>百度发票 OCR</h3><p>百度智能云 VAT 发票识别。</p></div><el-switch v-model="invoice.baidu.enabled" /></header>
                  <div class="field-grid">
                    <el-form-item label="API Key" class="field--wide"><SecretInput v-model.trim="invoice.baidu['api-key']" secret-path="ai.invoice.baidu.api-key" :configured="invoice.baidu['api-key-configured']" :can-reveal="canRevealProviderKeys" :disabled="invoice.baidu['clear-api-key']" placeholder="输入 API Key" /></el-form-item>
                    <el-form-item label="Secret Key" class="field--wide"><SecretInput v-model.trim="invoice.baidu['secret-key']" secret-path="ai.invoice.baidu.secret-key" :configured="invoice.baidu['secret-key-configured']" :can-reveal="canRevealProviderKeys" :disabled="invoice.baidu['clear-secret-key']" placeholder="输入 Secret Key" /></el-form-item>
                    <el-form-item label="超时（秒）"><el-input-number v-model="invoice.baidu['timeout-seconds']" :min="1" :max="120" controls-position="right" /></el-form-item>
                  </div>
                  <div class="provider-card__footer">
                    <el-checkbox v-if="invoice.baidu['api-key-configured']" v-model="invoice.baidu['clear-api-key']">清除 API Key</el-checkbox>
                    <el-checkbox v-if="invoice.baidu['secret-key-configured']" v-model="invoice.baidu['clear-secret-key']">清除 Secret Key</el-checkbox>
                    <el-button :loading="testingInvoice === 'baidu'" @click="testInvoice('baidu')">测试连接</el-button>
                  </div>
                </section>

                <section class="provider-card">
                  <header class="provider-card__header"><div><h3>企业 OCR 网关</h3><p>兼容 multipart JSON 的企业 OCR 服务。</p></div><el-switch v-model="invoice['public-ocr'].enabled" /></header>
                  <div class="field-grid">
                    <el-form-item label="接口地址" class="field--wide"><el-input v-model.trim="invoice['public-ocr'].endpoint" placeholder="https://ocr.example.com/recognize" /></el-form-item>
                    <el-form-item label="API Key" class="field--wide"><SecretInput v-model.trim="invoice['public-ocr']['api-key']" secret-path="ai.invoice.public-ocr.api-key" :configured="invoice['public-ocr']['api-key-configured']" :can-reveal="canRevealProviderKeys" :disabled="invoice['public-ocr']['clear-api-key']" placeholder="输入 API Key" /></el-form-item>
                    <el-form-item label="超时（秒）"><el-input-number v-model="invoice['public-ocr']['timeout-seconds']" :min="1" :max="120" controls-position="right" /></el-form-item>
                  </div>
                  <div class="provider-card__footer">
                    <span v-if="invoice['public-ocr'].protocol" class="detected-label">{{ invoice['public-ocr'].provider }} / {{ invoice['public-ocr'].protocol }}</span>
                    <el-checkbox v-if="invoice['public-ocr']['api-key-configured']" v-model="invoice['public-ocr']['clear-api-key']">清除 API Key</el-checkbox>
                    <el-button :loading="testingInvoice === 'public-ocr'" @click="testInvoice('public-ocr')">测试连接</el-button>
                  </div>
                </section>
              </div>
            </section>

            <section class="settings-group service-stack">
              <header class="group-heading"><div><h3>验真与模型兜底</h3><p>验真负责来源校验，视觉模型只在 OCR 结果不足时补充识别。</p></div></header>
              <section class="provider-card provider-card--wide">
                <header class="provider-card__header"><div><h3>权威验真服务</h3><p>支持百度验真或兼容 HTTP 验真网关，测试后自动识别协议。</p></div><el-switch v-model="invoice.verification.enabled" /></header>
                <div class="field-grid field-grid--four">
                  <el-form-item label="接口地址（百度留空）" class="field--double"><el-input v-model.trim="invoice.verification.endpoint" placeholder="HTTP 网关地址，可留空" /></el-form-item>
                  <el-form-item label="API Key"><SecretInput v-model.trim="invoice.verification['api-key']" secret-path="ai.invoice.verification.api-key" :configured="invoice.verification['api-key-configured']" :can-reveal="canRevealProviderKeys" :disabled="invoice.verification['clear-api-key']" placeholder="输入 API Key" /></el-form-item>
                  <el-form-item label="Secret Key"><SecretInput v-model.trim="invoice.verification['secret-key']" secret-path="ai.invoice.verification.secret-key" :configured="invoice.verification['secret-key-configured']" :can-reveal="canRevealProviderKeys" :disabled="invoice.verification['clear-secret-key']" placeholder="百度验真需要" /></el-form-item>
                  <el-form-item label="超时（秒）"><el-input-number v-model="invoice.verification['timeout-seconds']" :min="1" :max="120" controls-position="right" /></el-form-item>
                </div>
                <div class="provider-card__footer">
                  <span v-if="invoice.verification.protocol" class="detected-label">{{ invoice.verification.provider }} / {{ invoice.verification.protocol }}</span>
                  <el-checkbox v-if="invoice.verification['api-key-configured']" v-model="invoice.verification['clear-api-key']">清除 API Key</el-checkbox>
                  <el-checkbox v-if="invoice.verification['secret-key-configured']" v-model="invoice.verification['clear-secret-key']">清除 Secret Key</el-checkbox>
                  <el-button :loading="testingInvoice === 'verification'" @click="testInvoice('verification')">测试连接</el-button>
                </div>
              </section>

              <section class="provider-card provider-card--wide">
                <header class="provider-card__header"><div><h3>多模态视觉模型</h3><p>支持 OpenAI Compatible 和 Anthropic，测试后自动识别协议。</p></div><el-switch v-model="invoice.multimodal.enabled" /></header>
                <div class="field-grid field-grid--four">
                  <el-form-item label="Base URL" class="field--double"><el-input v-model.trim="invoice.multimodal['base-url']" placeholder="https://api.example.com/v1" /></el-form-item>
                  <el-form-item label="模型"><el-input v-model.trim="invoice.multimodal.model" placeholder="vision-model" /></el-form-item>
                  <el-form-item label="API Key"><SecretInput v-model.trim="invoice.multimodal['api-key']" secret-path="ai.invoice.multimodal.api-key" :configured="invoice.multimodal['api-key-configured']" :can-reveal="canRevealProviderKeys" :disabled="invoice.multimodal['clear-api-key']" placeholder="输入 API Key" /></el-form-item>
                  <el-form-item label="超时（秒）"><el-input-number v-model="invoice.multimodal['timeout-seconds']" :min="1" :max="120" controls-position="right" /></el-form-item>
                </div>
                <div class="provider-card__footer">
                  <span v-if="invoice.multimodal.protocol" class="detected-label">协议：{{ invoice.multimodal.protocol }}</span>
                  <el-checkbox v-if="invoice.multimodal['api-key-configured']" v-model="invoice.multimodal['clear-api-key']">清除 API Key</el-checkbox>
                  <el-button :loading="testingInvoice === 'multimodal'" @click="testInvoice('multimodal')">测试连接</el-button>
                </div>
              </section>
            </section>
          </el-form>
        </template>

        <template v-else-if="activeSectionName === 'security'">
          <div class="setting-list">
            <div class="setting-toggle">
              <div class="setting-toggle__copy"><strong>统一 AI Gateway</strong><small>所有智能业务通过统一入口执行策略、配额、脱敏和审计。</small></div>
              <el-switch v-model="providers.enabled" active-text="启用" inactive-text="关闭" />
            </div>
            <div class="setting-toggle">
              <div class="setting-toggle__copy"><strong>允许模型服务使用内网地址</strong><small>仅在模型部署于受控企业网络时启用。</small></div>
              <el-switch v-model="providers['allow-private-endpoints']" active-text="允许" inactive-text="拒绝" />
            </div>
          </div>
          <el-form label-position="top" class="settings-form security-form">
            <div class="security-grid">
              <el-form-item label="业务敏感词">
                <el-select v-model="providers['sensitive-words']" multiple filterable allow-create default-first-option placeholder="输入敏感词后回车" />
                <span class="field-help">命中后请求会在进入模型前被阻断。</span>
              </el-form-item>
              <el-form-item label="允许发送图片的业务模块">
                <el-select v-model="providers['allow-vision-modules']" multiple filterable allow-create default-first-option placeholder="输入模块标识后回车" />
                <span class="field-help">例如 asset；未列出的模块不能向第三方模型发送图片。</span>
              </el-form-item>
            </div>
          </el-form>
        </template>

        <template v-else-if="activeSectionName === 'billing'">
          <section class="settings-group settings-group--first">
            <header class="group-heading"><div><h3>模型单价</h3><p>按供应商账单填写人民币单价，系统据此估算每次调用费用。</p></div></header>
            <div class="pricing-table">
              <div class="pricing-row pricing-row--header"><span>Provider</span><span>输入费用 / 百万 Token</span><span>输出费用 / 百万 Token</span><span>状态</span></div>
              <div v-for="provider in providerList" :key="provider.key" class="pricing-row">
                <strong>{{ provider.label }}</strong>
                <el-input
                  :model-value="providers[provider.key]['input-cost-per-million']"
                  inputmode="decimal"
                  placeholder="0.000000"
                  @update:model-value="providers[provider.key]['input-cost-per-million'] = decimalValue($event)"
                />
                <el-input
                  :model-value="providers[provider.key]['output-cost-per-million']"
                  inputmode="decimal"
                  placeholder="0.000000"
                  @update:model-value="providers[provider.key]['output-cost-per-million'] = decimalValue($event)"
                />
                <el-tag :type="providers[provider.key].enabled ? 'success' : 'info'" effect="light">{{ providers[provider.key].enabled ? '已启用' : '未启用' }}</el-tag>
              </div>
            </div>
          </section>

          <section class="settings-group quota-section">
            <header class="group-heading group-heading--action">
              <div><h3>用量配额</h3><p>全局、模块、角色和用户配额同时生效；0 表示不限制。</p></div>
              <el-button type="primary" :icon="Plus" @click="openQuota()">新增配额</el-button>
            </header>
            <el-table v-if="quotaLoading || quotas.length > 0" v-loading="quotaLoading && !quotaLoaded" :data="quotas" row-key="ID">
              <el-table-column prop="scopeType" label="范围" width="110" />
              <el-table-column prop="scopeId" label="范围标识" min-width="170" />
              <el-table-column prop="dailyRequests" label="每日请求" width="110" align="right" />
              <el-table-column prop="dailyTokens" label="每日 Token" width="125" align="right" />
              <el-table-column prop="monthlyCostMicros" label="月预算" width="135" align="right"><template #default="{ row }">{{ money(row.monthlyCostMicros) }}</template></el-table-column>
              <el-table-column prop="maxConcurrency" label="最大并发" width="105" align="right" />
              <el-table-column label="状态" width="90" align="center"><template #default="{ row }"><el-tag :type="row.enabled ? 'success' : 'info'">{{ row.enabled ? '启用' : '停用' }}</el-tag></template></el-table-column>
              <el-table-column label="操作" width="90"><template #default="{ row }"><el-button text type="primary" :icon="Edit" @click="openQuota(row)">编辑</el-button></template></el-table-column>
            </el-table>
            <div v-if="quotaTotal > 10" class="na-pagination">
              <el-pagination
                v-model:current-page="quotaSearch.page"
                v-model:page-size="quotaSearch.pageSize"
                :page-sizes="[10, 20, 50]"
                :total="quotaTotal"
                layout="total, sizes, prev, pager, next"
                @change="loadQuotas"
                @size-change="resetQuotaPage"
              />
            </div>
            <AppEmptyState v-if="!quotaLoading && !quotas.length" class="quota-empty-state" compact title="尚未设置智能服务配额" description="可按全局、模块、角色或用户限制请求量、Token、预算和并发。" />
          </section>
        </template>

        <template v-else-if="activeSectionName === 'prompts'">
          <div class="section-toolbar"><span>每个 Prompt 标识只保留一个活跃版本。</span><el-button type="primary" :icon="Plus" @click="openPrompt">创建版本</el-button></div>
          <el-table v-loading="promptLoading && !promptLoaded" :data="prompts" row-key="ID">
            <el-table-column prop="promptKey" label="标识" min-width="180" />
            <el-table-column prop="version" label="版本" width="85" />
            <el-table-column prop="status" label="状态" width="105"><template #default="{ row }"><el-tag :type="promptStatus(row.status).type">{{ promptStatus(row.status).label }}</el-tag></template></el-table-column>
            <el-table-column prop="CreatedAt" label="创建时间" min-width="165"><template #default="{ row }">{{ dateTime(row.CreatedAt) }}</template></el-table-column>
            <el-table-column label="内容" min-width="260" show-overflow-tooltip><template #default="{ row }">{{ row.content }}</template></el-table-column>
            <el-table-column label="操作" width="100"><template #default="{ row }"><el-button v-if="row.status !== 'active'" text type="primary" :icon="Check" :loading="activatingPromptId === row.ID" @click="activatePrompt(row)">激活</el-button></template></el-table-column>
            <template #empty><AppEmptyState compact title="还没有 Prompt 版本" description="创建草稿并人工激活后，Gateway 才会使用对应版本。"><template #actions><el-button type="primary" :icon="Plus" @click="openPrompt">创建第一个版本</el-button></template></AppEmptyState></template>
          </el-table>
          <div v-if="promptTotal > 10" class="na-pagination">
            <el-pagination
              v-model:current-page="promptSearch.page"
              v-model:page-size="promptSearch.pageSize"
              :page-sizes="[10, 20, 50]"
              :total="promptTotal"
              layout="total, sizes, prev, pager, next"
              @change="loadPrompts"
              @size-change="resetPromptPage"
            />
          </div>
        </template>

        <template v-else>
          <section class="summary-band" aria-label="智能服务用量摘要">
            <div><span>今日调用</span><strong>{{ usage.todayRequests || 0 }}</strong><small>当前账户成功请求</small></div>
            <div><span>今日 Token</span><strong>{{ number(usage.todayTokens) }}</strong><small>输入与输出合计</small></div>
            <div><span>本月费用</span><strong>{{ money(usage.monthCostMicros) }}</strong><small>按模型单价估算</small></div>
            <div><span>累计调用</span><strong>{{ number(usage.totalRequests) }}</strong><small>当前账户成功请求</small></div>
          </section>
          <div class="filter-row">
            <el-select v-model="invocationSearch.status" clearable placeholder="全部状态"><el-option label="成功" value="success" /><el-option label="失败" value="failed" /><el-option label="已阻断" value="blocked" /></el-select>
            <el-input v-model="invocationSearch.module" clearable placeholder="业务模块" />
            <el-input v-model="invocationSearch.provider" clearable placeholder="Provider" />
            <el-input-number v-model="invocationSearch.userId" :min="1" :controls="false" placeholder="用户 ID" />
            <el-button :icon="Search" :loading="invocationLoading" @click="searchInvocations">查询</el-button>
          </div>
          <el-table v-loading="invocationLoading && !invocationLoaded" :data="invocations" row-key="ID">
            <el-table-column prop="CreatedAt" label="时间" min-width="165"><template #default="{ row }">{{ dateTime(row.CreatedAt) }}</template></el-table-column>
            <el-table-column prop="userId" label="用户" width="80" />
            <el-table-column prop="module" label="模块" min-width="110" />
            <el-table-column prop="operation" label="操作" min-width="140" />
            <el-table-column prop="provider" label="Provider" min-width="145" />
            <el-table-column prop="model" label="模型" min-width="145" show-overflow-tooltip />
            <el-table-column label="Token" width="120" align="right"><template #default="{ row }">{{ number(Number(row.inputTokens) + Number(row.outputTokens)) }}</template></el-table-column>
            <el-table-column label="费用" width="120" align="right"><template #default="{ row }">{{ money(row.estimatedCostMicros) }}</template></el-table-column>
            <el-table-column prop="durationMs" label="耗时" width="105" align="right"><template #default="{ row }">{{ row.durationMs }} ms</template></el-table-column>
            <el-table-column prop="errorType" label="错误类型" width="105"><template #default="{ row }">{{ row.errorType || '—' }}</template></el-table-column>
            <el-table-column label="状态" width="100" align="center"><template #default="{ row }"><el-tag :type="statusMeta(row.status).type" effect="light">{{ statusMeta(row.status).label }}</el-tag></template></el-table-column>
            <template #empty><AppEmptyState compact title="暂无模型调用记录" description="智能业务通过统一 Gateway 调用模型后，这里会展示状态、用量、费用和耗时。" :highlights="['不保存 Prompt 原文', '不保存图片和模型输出']" /></template>
          </el-table>
          <div class="na-pagination"><el-pagination v-model:current-page="invocationSearch.page" v-model:page-size="invocationSearch.pageSize" :total="invocationTotal" layout="total, sizes, prev, pager, next" :page-sizes="[10, 20, 50, 100]" @change="loadInvocations" @size-change="resetInvocationPage" /></div>
        </template>
      </section>
    </div>

    <el-dialog v-model="quotaDialogVisible" :title="quotaForm.ID ? '编辑智能服务配额' : '新增智能服务配额'" width="min(560px, calc(100vw - 32px))" destroy-on-close>
      <el-form label-position="top">
        <div class="dialog-grid">
          <el-form-item label="范围"><el-select v-model="quotaForm.scopeType"><el-option label="全局" value="global" /><el-option label="模块" value="module" /><el-option label="角色" value="authority" /><el-option label="用户" value="user" /></el-select></el-form-item>
          <el-form-item label="范围标识"><el-input v-model="quotaForm.scopeId" :placeholder="quotaForm.scopeType === 'global' ? 'global' : '模块名、角色 ID 或用户 ID'" /></el-form-item>
        </div>
        <div class="dialog-grid">
          <el-form-item label="每日请求"><el-input-number v-model="quotaForm.dailyRequests" :min="0" /></el-form-item>
          <el-form-item label="每日 Token"><el-input-number v-model="quotaForm.dailyTokens" :min="0" /></el-form-item>
          <el-form-item label="月预算（元）"><el-input :model-value="quotaForm.monthlyBudgetYuan" inputmode="decimal" placeholder="0.00" @update:model-value="quotaForm.monthlyBudgetYuan = decimalValue($event)" /></el-form-item>
          <el-form-item label="最大并发"><el-input-number v-model="quotaForm.maxConcurrency" :min="0" /></el-form-item>
        </div>
        <el-switch v-model="quotaForm.enabled" active-text="启用配额" />
      </el-form>
      <template #footer><el-button @click="quotaDialogVisible = false">取消</el-button><el-button type="primary" :loading="savingQuota" @click="submitQuota">保存</el-button></template>
    </el-dialog>

    <el-dialog v-model="promptDialogVisible" title="创建 Prompt 版本" width="min(720px, calc(100vw - 32px))" destroy-on-close>
      <el-form label-position="top">
        <el-form-item label="Prompt 标识"><el-input v-model="promptForm.promptKey" placeholder="例如 asset-draft-v1" /></el-form-item>
        <el-form-item label="Prompt 内容"><el-input v-model="promptForm.content" type="textarea" :rows="8" maxlength="131072" show-word-limit /></el-form-item>
        <el-form-item label="输出 JSON Schema（可选）"><el-input v-model="promptForm.outputSchema" type="textarea" :rows="4" /></el-form-item>
      </el-form>
      <template #footer><el-button @click="promptDialogVisible = false">取消</el-button><el-button type="primary" :loading="savingPrompt" @click="submitPrompt">创建草稿</el-button></template>
    </el-dialog>
  </main>
</template>

<script setup>
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { Check, Coin, Cpu, DataAnalysis, Document, Edit, Lock, Picture, Plus, Refresh, Search } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import AppEmptyState from '@/components/page/AppEmptyState.vue'
import AppPageHeader from '@/components/page/AppPageHeader.vue'
import SecretInput from '@/components/secretInput/index.vue'
import { useUserStore } from '@/pinia/modules/user'
import { activateAIPrompt, createAIPrompt, getAIInvocations, getAIInvoiceRecognition, getAIProviders, getAIPrompts, getAIQuotas, getAIUsageSummary, saveAIQuota, testAIInvoiceRecognition, updateAIInvoiceRecognition, updateAIProviders } from '@/plugin/aioperations/api/operations'
import { providerFormValue, providerPayloadValue, providerSecretPath } from '@/plugin/aioperations/utils/provider'
import { defaultInvoiceRecognition, invoiceRecognitionFormValue, invoiceRecognitionPayload } from '@/plugin/aioperations/utils/invoiceRecognition'
import { defaultQuota, quotaFormValue, quotaPayloadValue } from '@/plugin/aioperations/utils/quota'

defineOptions({ name: 'SmartCapabilitySettings' })

const settingsSections = [
  { name: 'models', label: '模型接入', hint: 'Provider 与密钥', description: '配置通用模型服务的地址、模型、凭据与调用超时。', icon: Cpu },
  { name: 'recognition', label: '识别服务', hint: 'OCR、验真与视觉模型', description: '配置发票 OCR、权威验真和低置信度视觉模型兜底。', icon: Picture },
  { name: 'security', label: '网关安全', hint: '策略与数据边界', description: '控制统一网关、内网端点、敏感词和图片外发范围。', icon: Lock },
  { name: 'billing', label: '配额计费', hint: '单价、预算与并发', description: '维护模型单价，以及全局、模块、角色和用户的用量边界。', icon: Coin },
  { name: 'prompts', label: 'Prompt 模板', hint: '版本与激活', description: '维护可审计、可回退的 Prompt 模板版本。', icon: Document },
  { name: 'monitoring', label: '运行监控', hint: '用量与调用日志', description: '查看智能服务的请求量、Token、费用、耗时和失败原因。', icon: DataAnalysis }
]
const activeSectionName = ref('models')
const activeSection = computed(() => settingsSections.find((item) => item.name === activeSectionName.value) || settingsSections[0])
const loadedResources = new Set()
const loading = ref(false)
const savingProviders = ref(false)
const savingInvoice = ref(false)
const lastProviderSaveTime = ref('')
const lastInvoiceSaveTime = ref('')
const userStore = useUserStore()
const canRevealProviderKeys = computed(() => Number(userStore.userInfo.authorityId) === 888)
const usage = ref({})
const invocations = ref([])
const invocationTotal = ref(0)
const invocationLoading = ref(false)
const invocationLoaded = ref(false)
const quotas = ref([])
const quotaTotal = ref(0)
const quotaLoading = ref(false)
const quotaLoaded = ref(false)
const prompts = ref([])
const promptTotal = ref(0)
const promptLoading = ref(false)
const promptLoaded = ref(false)
const quotaDialogVisible = ref(false)
const promptDialogVisible = ref(false)
const savingQuota = ref(false)
const savingPrompt = ref(false)
const activatingPromptId = ref(0)
const testingInvoice = ref('')
const providers = reactive(defaultProviders())
const invoice = reactive(defaultInvoiceRecognition())
const providerList = [
  { key: 'openai-compatible', label: 'OpenAI Compatible', hint: '兼容 Chat Completions 的模型服务。' },
  { key: 'anthropic', label: 'Anthropic', hint: '使用 Anthropic Messages API。' }
]
const invocationSearch = reactive({ page: 1, pageSize: 10, status: '', module: '', provider: '', userId: undefined })
const quotaSearch = reactive({ page: 1, pageSize: 10 })
const promptSearch = reactive({ page: 1, pageSize: 10 })
const quotaForm = reactive(defaultQuota())
const promptForm = reactive({ promptKey: '', content: '', outputSchema: '' })

const providerEnabledCount = computed(() => providerList.filter((item) => providers[item.key].enabled).length)
const recognitionEnabledCount = computed(() => ['baidu', 'public-ocr', 'verification', 'multimodal'].filter((key) => invoice[key].enabled).length)
const sectionBadge = computed(() => {
  if (activeSectionName.value === 'models') return { label: providerEnabledCount.value + ' 个 Provider 已启用', type: providerEnabledCount.value ? 'success' : 'info' }
  if (activeSectionName.value === 'recognition') return { label: recognitionEnabledCount.value + ' 项服务已启用', type: recognitionEnabledCount.value ? 'success' : 'info' }
  if (activeSectionName.value === 'security') return { label: providers.enabled ? 'Gateway 已启用' : 'Gateway 已关闭', type: providers.enabled ? 'success' : 'warning' }
  if (activeSectionName.value === 'billing') return { label: quotaTotal.value + ' 条配额', type: 'info' }
  if (activeSectionName.value === 'prompts') return { label: promptTotal.value + ' 个版本', type: 'info' }
  return { label: number(usage.value.todayRequests) + ' 次今日调用', type: 'info' }
})
const saveActionLabel = computed(() => ({ models: '保存模型接入', recognition: '保存识别服务', security: '保存安全策略', billing: '保存模型单价' }[activeSectionName.value] || ''))
const activeSaving = computed(() => activeSectionName.value === 'recognition' ? savingInvoice.value : savingProviders.value)
const currentSaveTime = computed(() => activeSectionName.value === 'recognition' ? lastInvoiceSaveTime.value : ['models', 'security', 'billing'].includes(activeSectionName.value) ? lastProviderSaveTime.value : '')

function defaultProvider() {
  return { enabled: false, 'base-url': '', 'api-key': '', 'api-key-configured': false, 'clear-api-key': false, model: '', 'timeout-seconds': 60, 'input-cost-per-million': 0, 'output-cost-per-million': 0 }
}

function defaultProviders() {
  return { enabled: false, 'allow-private-endpoints': false, 'sensitive-words': [], 'allow-vision-modules': [], 'openai-compatible': defaultProvider(), anthropic: { ...defaultProvider(), 'base-url': 'https://api.anthropic.com' } }
}

function applyProviders(value = {}) {
  const defaults = defaultProviders()
  providers.enabled = Boolean(value.enabled)
  providers['allow-private-endpoints'] = Boolean(value['allow-private-endpoints'])
  providers['sensitive-words'] = Array.isArray(value['sensitive-words']) ? [...value['sensitive-words']] : []
  providers['allow-vision-modules'] = Array.isArray(value['allow-vision-modules']) ? [...value['allow-vision-modules']] : []
  for (const { key } of providerList) Object.assign(providers[key], providerFormValue(value[key], defaults[key]))
}

function applyInvoice(value = {}) {
  Object.assign(invoice, invoiceRecognitionFormValue(value))
}

function providerPayload() {
  const payload = {
    enabled: providers.enabled,
    'allow-private-endpoints': providers['allow-private-endpoints'],
    'sensitive-words': [...providers['sensitive-words']],
    'allow-vision-modules': [...providers['allow-vision-modules']]
  }
  for (const { key } of providerList) payload[key] = providerPayloadValue(providers[key])
  return payload
}

function number(value) {
  return new Intl.NumberFormat('zh-CN').format(Number(value || 0))
}

function money(value) {
  return '¥' + (Number(value || 0) / 1000000).toFixed(4)
}

function decimalValue(value) {
  const source = String(value ?? '').replace(/[^\d.]/g, '')
  const dot = source.indexOf('.')
  if (dot < 0) return source
  return source.slice(0, dot + 1) + source.slice(dot + 1).replace(/\./g, '')
}

function dateTime(value) {
  return value ? new Date(value).toLocaleString('zh-CN', { hour12: false }) : '—'
}

function statusMeta(status) {
  return ({ success: { label: '成功', type: 'success' }, failed: { label: '失败', type: 'danger' }, blocked: { label: '已阻断', type: 'warning' } }[status] || { label: status || '未知', type: 'info' })
}

function promptStatus(status) {
  return ({ active: { label: '已激活', type: 'success' }, draft: { label: '草稿', type: 'info' }, retired: { label: '已停用', type: 'warning' } }[status] || { label: status, type: 'info' })
}

async function loadProviders() {
  const response = await getAIProviders()
  if (response.code === 0) applyProviders(response.data)
  else ElMessage.error(response.msg || '无法读取模型接入配置')
}

async function loadInvoice() {
  const response = await getAIInvoiceRecognition()
  if (response.code === 0) applyInvoice(response.data)
  else ElMessage.error(response.msg || '无法读取识别服务配置')
}

async function loadUsage() {
  const response = await getAIUsageSummary()
  if (response.code === 0) usage.value = response.data || {}
}

async function loadInvocations() {
  invocationLoading.value = true
  try {
    const response = await getAIInvocations(invocationSearch)
    if (response.code === 0) {
      invocations.value = response.data?.list || []
      invocationTotal.value = response.data?.total || 0
    } else ElMessage.error(response.msg || '无法读取调用日志')
  } finally {
    invocationLoading.value = false
    invocationLoaded.value = true
  }
}

function searchInvocations() {
  invocationSearch.page = 1
  return loadInvocations()
}

function resetInvocationPage() {
  invocationSearch.page = 1
}

async function loadQuotas() {
  quotaLoading.value = true
  try {
    const response = await getAIQuotas({ paged: true, ...quotaSearch })
    if (response.code === 0) {
      quotas.value = response.data?.list || []
      quotaTotal.value = Number(response.data?.total || 0)
    }
    else ElMessage.error(response.msg || '无法读取用量配额')
  } finally {
    quotaLoading.value = false
    quotaLoaded.value = true
  }
}

function resetQuotaPage() {
  quotaSearch.page = 1
}

async function loadPrompts() {
  promptLoading.value = true
  try {
    const response = await getAIPrompts({ paged: true, ...promptSearch })
    if (response.code === 0) {
      prompts.value = response.data?.list || []
      promptTotal.value = Number(response.data?.total || 0)
    }
    else ElMessage.error(response.msg || '无法读取 Prompt 模板')
  } finally {
    promptLoading.value = false
    promptLoaded.value = true
  }
}

function resetPromptPage() {
  promptSearch.page = 1
}

async function loadActiveSection(force = false) {
  const section = activeSectionName.value
  loading.value = true
  try {
    if (section === 'models' || section === 'security') {
      if (force || !loadedResources.has('providers')) {
        await loadProviders()
        loadedResources.add('providers')
      }
    } else if (section === 'recognition') {
      if (force || !loadedResources.has('invoice')) {
        await loadInvoice()
        loadedResources.add('invoice')
      }
    } else if (section === 'billing') {
      const tasks = []
      if (force || !loadedResources.has('providers')) tasks.push(loadProviders().then(() => loadedResources.add('providers')))
      if (force || !loadedResources.has('quotas')) tasks.push(loadQuotas().then(() => loadedResources.add('quotas')))
      await Promise.all(tasks)
    } else if (section === 'prompts') {
      if (force || !loadedResources.has('prompts')) {
        await loadPrompts()
        loadedResources.add('prompts')
      }
    } else {
      const tasks = []
      if (force || !loadedResources.has('usage')) tasks.push(loadUsage().then(() => loadedResources.add('usage')))
      if (force || !loadedResources.has('invocations')) tasks.push(loadInvocations().then(() => loadedResources.add('invocations')))
      await Promise.all(tasks)
    }
  } finally {
    loading.value = false
  }
}

async function persistProviders(successMessage) {
  savingProviders.value = true
  try {
    const response = await updateAIProviders(providerPayload())
    if (response.code === 0) {
      applyProviders(response.data)
      lastProviderSaveTime.value = new Date().toLocaleTimeString('zh-CN', { hour12: false })
      ElMessage.success(response.msg || successMessage)
    } else ElMessage.error(response.msg || '保存失败')
  } finally {
    savingProviders.value = false
  }
}

async function saveInvoice() {
  savingInvoice.value = true
  try {
    const response = await updateAIInvoiceRecognition(invoiceRecognitionPayload(invoice))
    if (response.code === 0) {
      await loadInvoice()
      lastInvoiceSaveTime.value = new Date().toLocaleTimeString('zh-CN', { hour12: false })
      ElMessage.success(response.msg || '识别服务配置已保存')
    } else ElMessage.error(response.msg || '保存失败')
  } finally {
    savingInvoice.value = false
  }
}

function saveActiveSettings() {
  if (activeSectionName.value === 'recognition') return saveInvoice()
  const messages = { models: '模型接入配置已保存', security: '网关安全策略已保存', billing: '模型单价已保存' }
  return persistProviders(messages[activeSectionName.value] || '智能能力配置已保存')
}

async function testInvoice(target) {
  if (testingInvoice.value) return
  testingInvoice.value = target
  try {
    const response = await testAIInvoiceRecognition({ target, config: invoiceRecognitionPayload(invoice) })
    if (response.code !== 0) return
    const detection = response.data || {}
    if (target === 'public-ocr') {
      invoice['public-ocr'].provider = detection.provider || ''
      invoice['public-ocr'].protocol = detection.protocol || ''
    } else if (target === 'verification') {
      invoice.verification.provider = detection.provider || ''
      invoice.verification.protocol = detection.protocol || ''
    } else if (target === 'multimodal') invoice.multimodal.protocol = detection.protocol || ''
    ElMessage.success('连接测试成功，服务协议已自动识别')
  } finally {
    testingInvoice.value = ''
  }
}

function openQuota(row) {
  Object.assign(quotaForm, quotaFormValue(row))
  quotaDialogVisible.value = true
}

async function submitQuota() {
  if (quotaForm.scopeType === 'global' && !quotaForm.scopeId) quotaForm.scopeId = 'global'
  savingQuota.value = true
  try {
    const response = await saveAIQuota(quotaPayloadValue(quotaForm))
    if (response.code === 0) {
      ElMessage.success(response.msg || '智能服务配额已保存')
      quotaDialogVisible.value = false
      quotaSearch.page = 1
      await loadQuotas()
    } else ElMessage.error(response.msg || '保存失败')
  } finally {
    savingQuota.value = false
  }
}

function openPrompt() {
  Object.assign(promptForm, { promptKey: '', content: '', outputSchema: '' })
  promptDialogVisible.value = true
}

async function submitPrompt() {
  savingPrompt.value = true
  try {
    const response = await createAIPrompt(promptForm)
    if (response.code === 0) {
      ElMessage.success(response.msg || 'Prompt 草稿已创建')
      promptDialogVisible.value = false
      promptSearch.page = 1
      await loadPrompts()
    } else ElMessage.error(response.msg || '创建失败')
  } finally {
    savingPrompt.value = false
  }
}

async function activatePrompt(row) {
  try {
    await ElMessageBox.confirm('确认激活 ' + row.promptKey + ' 的 V' + row.version + '？当前活跃版本将退役。', '激活 Prompt', { type: 'warning' })
    activatingPromptId.value = row.ID
    const response = await activateAIPrompt({ promptKey: row.promptKey, version: row.version })
    if (response.code === 0) {
      ElMessage.success(response.msg || 'Prompt 已激活')
      await loadPrompts()
    } else ElMessage.error(response.msg || '激活失败')
  } catch (error) {
    if (error !== 'cancel' && error !== 'close') ElMessage.error('激活失败')
  } finally {
    activatingPromptId.value = 0
  }
}

watch(activeSectionName, () => loadActiveSection())
onMounted(() => loadActiveSection(true))
</script>

<style scoped lang="scss">
.smart-settings-page { min-width: 0; }
.settings-shell { display: grid; grid-template-columns: 220px minmax(0, 1fr); min-height: 620px; overflow: hidden; border: 1px solid var(--na-border); border-radius: 8px; background: var(--na-card); box-shadow: var(--na-shadow-sm); }
.settings-nav { display: flex; flex-direction: column; gap: 4px; padding: 12px; border-right: 1px solid var(--na-border); background: var(--na-surface-muted); }
.settings-nav__item { display: grid; grid-template-columns: 30px minmax(0, 1fr); align-items: center; width: 100%; gap: 10px; padding: 10px; border: 1px solid transparent; border-radius: 6px; color: var(--na-muted-foreground); background: transparent; text-align: left; cursor: pointer; transition: color .18s ease, background-color .18s ease, border-color .18s ease; }
.settings-nav__item:hover { color: var(--na-foreground); background: var(--na-card); }
.settings-nav__item:focus-visible { outline: 2px solid var(--na-primary); outline-offset: 2px; }
.settings-nav__item.is-active { color: var(--na-primary); border-color: var(--na-border); background: var(--na-card); }
.settings-nav__item .el-icon { font-size: 18px; }
.settings-nav__item span { display: flex; min-width: 0; flex-direction: column; gap: 2px; }
.settings-nav__item strong { color: inherit; font-size: .84rem; font-weight: 600; }
.settings-nav__item small { overflow: hidden; color: var(--na-muted-foreground); font-size: .7rem; text-overflow: ellipsis; white-space: nowrap; }
.settings-content { min-width: 0; padding: 20px; }
.settings-section-header { display: flex; align-items: flex-start; justify-content: space-between; gap: 16px; margin-bottom: 18px; }
.settings-section-header h2 { margin: 0; color: var(--na-foreground); font-size: 1.05rem; }
.settings-section-header p { margin: 5px 0 0; color: var(--na-muted-foreground); font-size: .78rem; }
.save-state, .secret-state, .detected-label { display: inline-block; margin-top: 6px; color: var(--el-color-success); font-size: .73rem; }
.settings-form { margin-top: 16px; }
.provider-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 14px; }
.provider-card { min-width: 0; padding: 14px; border: 1px solid var(--na-border); border-radius: 8px; background: var(--na-surface-muted); }
.provider-card--wide { background: var(--na-card); }
.provider-card__header { display: flex; align-items: flex-start; justify-content: space-between; gap: 12px; margin-bottom: 14px; }
.provider-card h3, .group-heading h3 { margin: 0; color: var(--na-foreground); font-size: .9rem; }
.provider-card p, .group-heading p { margin: 4px 0 0; color: var(--na-muted-foreground); font-size: .73rem; line-height: 1.45; }
.field-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 0 12px; }
.field-grid--four { grid-template-columns: repeat(4, minmax(0, 1fr)); }
.field--wide { grid-column: 1 / -1; }
.field--double { grid-column: span 2; }
.provider-card :deep(.el-input-number) { width: 100%; }
.provider-card__footer { display: flex; min-height: 28px; flex-wrap: wrap; align-items: center; gap: 10px; }
.provider-card__footer .el-button { margin-left: auto; }
.provider-card__footer .secret-state { margin: 0 auto 0 0; }
.settings-group { padding-top: 18px; margin-top: 18px; border-top: 1px solid var(--na-border); }
.settings-group--first { padding-top: 0; margin-top: 0; border-top: 0; }
.group-heading { display: flex; align-items: flex-start; justify-content: space-between; gap: 16px; margin-bottom: 12px; }
.group-heading--action { align-items: center; }
.policy-row { display: grid; grid-template-columns: 190px minmax(0, 1fr); gap: 16px; align-items: end; }
.policy-row :deep(.el-input-number) { width: 100%; }
.setting-list { border-top: 1px solid var(--na-border); }
.setting-toggle { display: flex; min-width: 0; align-items: center; justify-content: space-between; gap: 20px; padding: 14px 0; border-bottom: 1px solid var(--na-border); }
.setting-toggle__copy { display: flex; min-width: 0; flex-direction: column; gap: 4px; }
.setting-toggle strong { color: var(--na-foreground); font-size: .84rem; }
.setting-toggle small { color: var(--na-muted-foreground); font-size: .73rem; line-height: 1.45; }
.security-form { margin-top: 18px; }
.security-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 16px; }
.security-grid :deep(.el-select) { width: 100%; }
.field-help { display: block; margin-top: 6px; color: var(--na-muted-foreground); font-size: .7rem; }
.service-stack { display: grid; gap: 14px; }
.service-stack .group-heading { margin-bottom: 0; }
.pricing-table { overflow-x: auto; border: 1px solid var(--na-border); border-radius: 8px; }
.quota-empty-state { min-height: 200px; border: 1px solid var(--na-border); border-radius: 8px; }
.pricing-row { display: grid; min-width: 690px; grid-template-columns: minmax(180px, 1fr) 210px 210px 90px; align-items: center; gap: 12px; padding: 12px 14px; border-bottom: 1px solid var(--na-border); }
.pricing-row:last-child { border-bottom: 0; }
.pricing-row--header { color: var(--na-muted-foreground); background: var(--na-surface-muted); font-size: .72rem; }
.pricing-row > strong { color: var(--na-foreground); font-size: .82rem; }
.pricing-row :deep(.el-input-number) { width: 100%; }
.section-toolbar { display: flex; align-items: center; justify-content: space-between; gap: 16px; margin-bottom: 14px; color: var(--na-muted-foreground); font-size: .76rem; }
.summary-band { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); margin-bottom: 16px; overflow: hidden; border: 1px solid var(--na-border); border-radius: 8px; }
.summary-band > div { display: flex; min-width: 0; flex-direction: column; gap: 5px; padding: 16px 18px; border-right: 1px solid var(--na-border); }
.summary-band > div:last-child { border-right: 0; }
.summary-band span, .summary-band small { color: var(--na-muted-foreground); font-size: .72rem; }
.summary-band strong { overflow: hidden; color: var(--na-foreground); font-size: 1.25rem; font-variant-numeric: tabular-nums; text-overflow: ellipsis; white-space: nowrap; }
.filter-row { display: grid; grid-template-columns: 140px repeat(2, minmax(0, 1fr)) 130px auto; gap: 10px; margin-bottom: 14px; }
.filter-row :deep(.el-input-number) { width: 100%; }
.na-pagination { display: flex; justify-content: flex-end; margin-top: 14px; }
.dialog-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 0 12px; }
.dialog-grid :deep(.el-input-number), .dialog-grid :deep(.el-select) { width: 100%; }
@media (max-width: 1100px) {
  .provider-grid, .security-grid { grid-template-columns: 1fr; }
  .field-grid--four { grid-template-columns: repeat(2, minmax(0, 1fr)); }
}
@media (max-width: 900px) {
  .settings-shell { grid-template-columns: 1fr; }
  .settings-nav { flex-direction: row; overflow-x: auto; border-right: 0; border-bottom: 1px solid var(--na-border); }
  .settings-nav__item { min-width: 170px; }
  .summary-band { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .summary-band > div:nth-child(2) { border-right: 0; }
  .summary-band > div:nth-child(-n + 2) { border-bottom: 1px solid var(--na-border); }
  .filter-row { grid-template-columns: repeat(2, minmax(0, 1fr)); }
}
@media (max-width: 680px) {
  .settings-content { padding: 16px; }
  .settings-section-header, .group-heading--action, .section-toolbar, .setting-toggle { align-items: flex-start; flex-direction: column; }
  .field-grid, .field-grid--four, .policy-row, .dialog-grid, .filter-row, .summary-band { grid-template-columns: 1fr; }
  .field--double { grid-column: auto; }
  .summary-band > div { border-right: 0; border-bottom: 1px solid var(--na-border); }
  .summary-band > div:last-child { border-bottom: 0; }
  .provider-card__footer .el-button { margin-left: 0; }
}
@media (prefers-reduced-motion: reduce) {
  .settings-nav__item { transition: none; }
}
</style>
