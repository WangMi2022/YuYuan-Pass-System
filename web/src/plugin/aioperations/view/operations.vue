<template>
  <main class="na-page na-page--list smart-settings-page">
    <AppPageHeader
      title-id="smart-settings-title"
      title="智能能力配置"
      description="集中管理多模态模型接入、发票与文档智能识别、统一网关安全、多维度用量配额与全链路调用审计。"
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
      <!-- 左侧分类导航 -->
      <nav class="settings-nav" aria-label="智能能力配置分类">
        <div class="settings-nav__header">
          <span class="settings-nav__badge">AI Platform</span>
          <span class="settings-nav__caption">能力与治理中心</span>
        </div>
        <button
          v-for="section in settingsSections"
          :key="section.name"
          type="button"
          class="settings-nav__item"
          :class="{ 'is-active': activeSectionName === section.name }"
          :aria-current="activeSectionName === section.name ? 'page' : undefined"
          @click="activeSectionName = section.name"
        >
          <div class="nav-icon-box" :class="'nav-icon-box--' + section.name">
            <el-icon><component :is="section.icon" /></el-icon>
          </div>
          <div class="nav-label-box">
            <strong>{{ section.label }}</strong>
            <small>{{ section.hint }}</small>
          </div>
          <el-icon class="nav-arrow"><ArrowRight /></el-icon>
        </button>
      </nav>

      <!-- 右侧配置区域 -->
      <section class="settings-content" :aria-labelledby="'settings-section-' + activeSectionName">
        <header class="settings-section-header">
          <div class="header-main">
            <div class="header-title-row">
              <h2 :id="'settings-section-' + activeSectionName">{{ activeSection.label }}</h2>
              <el-tag v-if="sectionBadge" :type="sectionBadge.type" effect="light" class="header-pill">
                <span class="status-indicator-dot" :class="'status-indicator-dot--' + (sectionBadge.type || 'info')" />
                {{ sectionBadge.label }}
              </el-tag>
            </div>
            <p>{{ activeSection.description }}</p>
          </div>
          <div v-if="currentSaveTime" class="save-state-chip">
            <el-icon><CircleCheck /></el-icon>
            <span>服务端已同步 · {{ currentSaveTime }}</span>
          </div>
        </header>

        <!-- 1. 模型接入 -->
        <template v-if="activeSectionName === 'models'">
          <div v-if="!providers.enabled" class="gateway-alert-card">
            <div class="gateway-alert-icon">
              <el-icon><WarningFilled /></el-icon>
            </div>
            <div class="gateway-alert-body">
              <strong>统一 AI Gateway 当前处于关闭状态</strong>
              <p>所有配置的大模型接入点将被业务调用隔离。请在“网关安全”模块中启用统一 Gateway 以激活智能应用。</p>
            </div>
            <el-button type="primary" size="small" plain @click="activeSectionName = 'security'">
              去安全配置启用
            </el-button>
          </div>

          <el-form label-position="top" class="settings-form">
            <div class="provider-grid">
              <section
                v-for="provider in providerList"
                :key="provider.key"
                class="provider-card"
                :class="{ 'is-card-enabled': providers[provider.key].enabled }"
              >
                <header class="provider-card__header">
                  <div class="provider-identity">
                    <div class="provider-brand-badge" :class="'brand--' + provider.key">
                      <el-icon v-if="provider.key === 'openai-compatible'"><Cpu /></el-icon>
                      <el-icon v-else><Promotion /></el-icon>
                    </div>
                    <div>
                      <div class="provider-title-row">
                        <h3>{{ provider.label }}</h3>
                        <span class="provider-active-chip" :class="providers[provider.key].enabled ? 'is-active' : 'is-idle'">
                          {{ providers[provider.key].enabled ? '已就绪' : '未接入' }}
                        </span>
                      </div>
                      <p>{{ provider.hint }}</p>
                    </div>
                  </div>
                  <el-switch
                    v-model="providers[provider.key].enabled"
                    inline-prompt
                    active-text="开"
                    inactive-text="关"
                  />
                </header>

                <div class="provider-card__body">
                  <div class="field-grid">
                    <el-form-item label="Base URL" class="field--wide">
                      <el-input
                        v-model.trim="providers[provider.key]['base-url']"
                        placeholder="https://api.example.com/v1"
                        clearable
                      >
                        <template #prefix>
                          <el-icon class="input-icon"><Connection /></el-icon>
                        </template>
                      </el-input>
                    </el-form-item>
                    <el-form-item label="模型名称">
                      <el-input
                        v-model.trim="providers[provider.key].model"
                        placeholder="例如 gpt-4o, claude-3-5-sonnet"
                        clearable
                      />
                    </el-form-item>
                    <el-form-item label="请求超时（秒）">
                      <el-input-number
                        v-model="providers[provider.key]['timeout-seconds']"
                        :min="1"
                        :max="120"
                        controls-position="right"
                        class="full-width-number"
                      />
                    </el-form-item>
                    <el-form-item
                      class="field--wide"
                      :label="providers[provider.key]['api-key-configured'] ? 'API Key (已脱敏保护)' : 'API Key'"
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
                </div>

                <footer class="provider-card__footer">
                  <div class="footer-secret-info">
                    <template v-if="providers[provider.key]['api-key-configured']">
                      <el-icon class="secret-shield-icon"><CircleCheck /></el-icon>
                      <span class="secret-state">凭据已在密钥库中加密存储</span>
                    </template>
                    <span v-else class="secret-empty-state">尚未配置凭据</span>
                  </div>
                  <el-checkbox
                    v-if="providers[provider.key]['api-key-configured']"
                    v-model="providers[provider.key]['clear-api-key']"
                    class="clear-key-checkbox"
                  >
                    清除已存密钥
                  </el-checkbox>
                </footer>
              </section>
            </div>
          </el-form>
        </template>

        <!-- 2. 识别服务 -->
        <template v-else-if="activeSectionName === 'recognition'">
          <div class="security-tip-strip">
            <el-icon><InfoFilled /></el-icon>
            <span>
              凭据保存后保持服务端硬件级脱敏；超级管理员可点击图标临时查看明文。测试连接仅校验连通性，不会留存任何业务图片或发票票据。
            </span>
          </div>

          <el-form label-position="top" class="settings-form">
            <!-- 识别策略编排卡片 -->
            <section class="strategy-pipeline-card">
              <header class="group-heading">
                <div>
                  <h3>智能流水线识别策略</h3>
                  <p>系统按策略梯次调度：优先解析二维码，其次进入 OCR 主识别；低于置信度阈值时触发视觉模型兜底，必要时转入人工审核。</p>
                </div>
              </header>

              <div class="pipeline-steps-visual">
                <div class="pipeline-step">
                  <span class="pipeline-step__num">1</span>
                  <div>
                    <strong>二维码解析</strong>
                    <small>税局核心字段零误差提取</small>
                  </div>
                </div>
                <div class="pipeline-divider">➔</div>
                <div class="pipeline-step">
                  <span class="pipeline-step__num">2</span>
                  <div>
                    <strong>权威 OCR 识别</strong>
                    <small>版面还原与票据全文提取</small>
                  </div>
                </div>
                <div class="pipeline-divider">➔</div>
                <div class="pipeline-step">
                  <span class="pipeline-step__num">3</span>
                  <div>
                    <strong>视觉大模型兜底</strong>
                    <small>低置信度深度语义推理</small>
                  </div>
                </div>
                <div class="pipeline-divider">➔</div>
                <div class="pipeline-step">
                  <span class="pipeline-step__num">4</span>
                  <div>
                    <strong>人工复核归档</strong>
                    <small>异常票据与合规验真</small>
                  </div>
                </div>
              </div>

              <div class="policy-controls-row">
                <div class="threshold-field-box">
                  <label class="custom-field-label">视觉模型兜底阈值</label>
                  <div class="threshold-input-group">
                    <el-input-number
                      v-model="invoice['fallback-threshold']"
                      :min="0.1"
                      :max="1"
                      :step="0.01"
                      :precision="2"
                      controls-position="right"
                    />
                    <span class="threshold-badge">
                      置信度 &lt; {{ Math.round(invoice['fallback-threshold'] * 100) }}% 自动激活视觉大模型
                    </span>
                  </div>
                </div>

                <div class="setting-toggle setting-toggle--inline">
                  <div class="setting-toggle__copy">
                    <strong>允许识别服务使用私有内网地址</strong>
                    <small>仅在私有化部署的企业内部 OCR、验真网关或离线模型时开启。</small>
                  </div>
                  <el-switch
                    v-model="invoice['allow-private-endpoints']"
                    active-text="允许"
                    inactive-text="拒绝"
                    inline-prompt
                  />
                </div>
              </div>
            </section>

            <!-- OCR 服务分类 -->
            <section class="settings-group">
              <header class="group-heading">
                <div>
                  <h3>OCR 识别引擎</h3>
                  <p>标准增值税发票推荐使用百度云 OCR；私有化或企业自建服务可接入兼容 HTTP 网关。</p>
                </div>
              </header>

              <div class="provider-grid">
                <!-- 百度 OCR -->
                <section class="provider-card" :class="{ 'is-card-enabled': invoice.baidu.enabled }">
                  <header class="provider-card__header">
                    <div class="provider-identity">
                      <div class="provider-brand-badge brand--baidu">
                        <el-icon><Picture /></el-icon>
                      </div>
                      <div>
                        <div class="provider-title-row">
                          <h3>百度发票 OCR</h3>
                          <span class="provider-active-chip" :class="invoice.baidu.enabled ? 'is-active' : 'is-idle'">
                            {{ invoice.baidu.enabled ? '已启用' : '未启用' }}
                          </span>
                        </div>
                        <p>百度智能云 VAT 发票专用识别引擎。</p>
                      </div>
                    </div>
                    <el-switch v-model="invoice.baidu.enabled" inline-prompt active-text="开" inactive-text="关" />
                  </header>

                  <div class="provider-card__body">
                    <div class="field-grid">
                      <el-form-item label="API Key" class="field--wide">
                        <SecretInput
                          v-model.trim="invoice.baidu['api-key']"
                          secret-path="ai.invoice.baidu.api-key"
                          :configured="invoice.baidu['api-key-configured']"
                          :can-reveal="canRevealProviderKeys"
                          :disabled="invoice.baidu['clear-api-key']"
                          placeholder="输入百度 API Key"
                        />
                      </el-form-item>
                      <el-form-item label="Secret Key" class="field--wide">
                        <SecretInput
                          v-model.trim="invoice.baidu['secret-key']"
                          secret-path="ai.invoice.baidu.secret-key"
                          :configured="invoice.baidu['secret-key-configured']"
                          :can-reveal="canRevealProviderKeys"
                          :disabled="invoice.baidu['clear-secret-key']"
                          placeholder="输入百度 Secret Key"
                        />
                      </el-form-item>
                      <el-form-item label="超时时间（秒）" class="field--wide">
                        <el-input-number
                          v-model="invoice.baidu['timeout-seconds']"
                          :min="1"
                          :max="120"
                          controls-position="right"
                          class="full-width-number"
                        />
                      </el-form-item>
                    </div>
                  </div>

                  <footer class="provider-card__footer">
                    <div class="footer-actions-left">
                      <el-checkbox v-if="invoice.baidu['api-key-configured']" v-model="invoice.baidu['clear-api-key']">
                        清除 API Key
                      </el-checkbox>
                      <el-checkbox v-if="invoice.baidu['secret-key-configured']" v-model="invoice.baidu['clear-secret-key']">
                        清除 Secret Key
                      </el-checkbox>
                    </div>
                    <el-button
                      type="primary"
                      plain
                      size="small"
                      :loading="testingInvoice === 'baidu'"
                      @click="testInvoice('baidu')"
                    >
                      测试连通性
                    </el-button>
                  </footer>
                </section>

                <!-- 企业 OCR 网关 -->
                <section class="provider-card" :class="{ 'is-card-enabled': invoice['public-ocr'].enabled }">
                  <header class="provider-card__header">
                    <div class="provider-identity">
                      <div class="provider-brand-badge brand--gateway">
                        <el-icon><Connection /></el-icon>
                      </div>
                      <div>
                        <div class="provider-title-row">
                          <h3>企业 OCR 网关</h3>
                          <span class="provider-active-chip" :class="invoice['public-ocr'].enabled ? 'is-active' : 'is-idle'">
                            {{ invoice['public-ocr'].enabled ? '已启用' : '未启用' }}
                          </span>
                        </div>
                        <p>兼容 multipart JSON 标准协议的企业自建 OCR 服务。</p>
                      </div>
                    </div>
                    <el-switch v-model="invoice['public-ocr'].enabled" inline-prompt active-text="开" inactive-text="关" />
                  </header>

                  <div class="provider-card__body">
                    <div class="field-grid">
                      <el-form-item label="网关接口地址" class="field--wide">
                        <el-input
                          v-model.trim="invoice['public-ocr'].endpoint"
                          placeholder="https://ocr.example.com/recognize"
                          clearable
                        />
                      </el-form-item>
                      <el-form-item label="API Key" class="field--wide">
                        <SecretInput
                          v-model.trim="invoice['public-ocr']['api-key']"
                          secret-path="ai.invoice.public-ocr.api-key"
                          :configured="invoice['public-ocr']['api-key-configured']"
                          :can-reveal="canRevealProviderKeys"
                          :disabled="invoice['public-ocr']['clear-api-key']"
                          placeholder="输入接口访问 API Key"
                        />
                      </el-form-item>
                      <el-form-item label="超时时间（秒）" class="field--wide">
                        <el-input-number
                          v-model="invoice['public-ocr']['timeout-seconds']"
                          :min="1"
                          :max="120"
                          controls-position="right"
                          class="full-width-number"
                        />
                      </el-form-item>
                    </div>
                  </div>

                  <footer class="provider-card__footer">
                    <div class="footer-actions-left">
                      <span v-if="invoice['public-ocr'].protocol" class="detected-label">
                        {{ invoice['public-ocr'].provider }} · {{ invoice['public-ocr'].protocol }}
                      </span>
                      <el-checkbox v-if="invoice['public-ocr']['api-key-configured']" v-model="invoice['public-ocr']['clear-api-key']">
                        清除 API Key
                      </el-checkbox>
                    </div>
                    <el-button
                      type="primary"
                      plain
                      size="small"
                      :loading="testingInvoice === 'public-ocr'"
                      @click="testInvoice('public-ocr')"
                    >
                      测试连通性
                    </el-button>
                  </footer>
                </section>
              </div>
            </section>

            <!-- 验真与模型兜底 -->
            <section class="settings-group service-stack">
              <header class="group-heading">
                <div>
                  <h3>权威验真与多模态兜底</h3>
                  <p>验真接口负责发票全国统一发票查验平台校验；视觉大模型则在低置信度时从复杂票据图像中补充提取结构化数据。</p>
                </div>
              </header>

              <!-- 权威验真服务 -->
              <section class="provider-card provider-card--wide" :class="{ 'is-card-enabled': invoice.verification.enabled }">
                <header class="provider-card__header">
                  <div class="provider-identity">
                    <div class="provider-brand-badge brand--verify">
                      <el-icon><Lock /></el-icon>
                    </div>
                    <div>
                      <div class="provider-title-row">
                        <h3>权威发票验真服务</h3>
                        <span class="provider-active-chip" :class="invoice.verification.enabled ? 'is-active' : 'is-idle'">
                          {{ invoice.verification.enabled ? '已启用' : '未启用' }}
                        </span>
                      </div>
                      <p>支持百度发票验真接口或企业自建兼容验真 HTTP 网关，测试后自动探测协议版本。</p>
                    </div>
                  </div>
                  <el-switch v-model="invoice.verification.enabled" inline-prompt active-text="开" inactive-text="关" />
                </header>

                <div class="provider-card__body">
                  <div class="field-grid field-grid--four">
                    <el-form-item label="接口地址 (百度官方留空)" class="field--double">
                      <el-input v-model.trim="invoice.verification.endpoint" placeholder="私有 HTTP 网关地址，百度验真请留空" clearable />
                    </el-form-item>
                    <el-form-item label="API Key">
                      <SecretInput
                        v-model.trim="invoice.verification['api-key']"
                        secret-path="ai.invoice.verification.api-key"
                        :configured="invoice.verification['api-key-configured']"
                        :can-reveal="canRevealProviderKeys"
                        :disabled="invoice.verification['clear-api-key']"
                        placeholder="输入 API Key"
                      />
                    </el-form-item>
                    <el-form-item label="Secret Key (百度专用)">
                      <SecretInput
                        v-model.trim="invoice.verification['secret-key']"
                        secret-path="ai.invoice.verification.secret-key"
                        :configured="invoice.verification['secret-key-configured']"
                        :can-reveal="canRevealProviderKeys"
                        :disabled="invoice.verification['clear-secret-key']"
                        placeholder="百度验真需要"
                      />
                    </el-form-item>
                    <el-form-item label="超时时间（秒）">
                      <el-input-number
                        v-model="invoice.verification['timeout-seconds']"
                        :min="1"
                        :max="120"
                        controls-position="right"
                        class="full-width-number"
                      />
                    </el-form-item>
                  </div>
                </div>

                <footer class="provider-card__footer">
                  <div class="footer-actions-left">
                    <span v-if="invoice.verification.protocol" class="detected-label">
                      协议：{{ invoice.verification.provider }} / {{ invoice.verification.protocol }}
                    </span>
                    <el-checkbox v-if="invoice.verification['api-key-configured']" v-model="invoice.verification['clear-api-key']">
                      清除 API Key
                    </el-checkbox>
                    <el-checkbox v-if="invoice.verification['secret-key-configured']" v-model="invoice.verification['clear-secret-key']">
                      清除 Secret Key
                    </el-checkbox>
                  </div>
                  <el-button
                    type="primary"
                    plain
                    size="small"
                    :loading="testingInvoice === 'verification'"
                    @click="testInvoice('verification')"
                  >
                    测试连通性
                  </el-button>
                </footer>
              </section>

              <!-- 多模态视觉模型 -->
              <section class="provider-card provider-card--wide" :class="{ 'is-card-enabled': invoice.multimodal.enabled }">
                <header class="provider-card__header">
                  <div class="provider-identity">
                    <div class="provider-brand-badge brand--multimodal">
                      <el-icon><Cpu /></el-icon>
                    </div>
                    <div>
                      <div class="provider-title-row">
                        <h3>多模态视觉模型兜底 (Vision AI)</h3>
                        <span class="provider-active-chip" :class="invoice.multimodal.enabled ? 'is-active' : 'is-idle'">
                          {{ invoice.multimodal.enabled ? '已启用' : '未启用' }}
                        </span>
                      </div>
                      <p>支持 OpenAI Compatible 视觉模型与 Anthropic Claude Vision，测试连接后自动识别通信协议。</p>
                    </div>
                  </div>
                  <el-switch v-model="invoice.multimodal.enabled" inline-prompt active-text="开" inactive-text="关" />
                </header>

                <div class="provider-card__body">
                  <div class="field-grid field-grid--four">
                    <el-form-item label="Base URL" class="field--double">
                      <el-input v-model.trim="invoice.multimodal['base-url']" placeholder="https://api.example.com/v1" clearable />
                    </el-form-item>
                    <el-form-item label="模型名称">
                      <el-input v-model.trim="invoice.multimodal.model" placeholder="例如 qwen-vl-max, gpt-4o" clearable />
                    </el-form-item>
                    <el-form-item label="API Key">
                      <SecretInput
                        v-model.trim="invoice.multimodal['api-key']"
                        secret-path="ai.invoice.multimodal.api-key"
                        :configured="invoice.multimodal['api-key-configured']"
                        :can-reveal="canRevealProviderKeys"
                        :disabled="invoice.multimodal['clear-api-key']"
                        placeholder="输入 API Key"
                      />
                    </el-form-item>
                    <el-form-item label="超时时间（秒）">
                      <el-input-number
                        v-model="invoice.multimodal['timeout-seconds']"
                        :min="1"
                        :max="120"
                        controls-position="right"
                        class="full-width-number"
                      />
                    </el-form-item>
                  </div>
                </div>

                <footer class="provider-card__footer">
                  <div class="footer-actions-left">
                    <span v-if="invoice.multimodal.protocol" class="detected-label">
                      协议识别：{{ invoice.multimodal.protocol }}
                    </span>
                    <el-checkbox v-if="invoice.multimodal['api-key-configured']" v-model="invoice.multimodal['clear-api-key']">
                      清除 API Key
                    </el-checkbox>
                  </div>
                  <el-button
                    type="primary"
                    plain
                    size="small"
                    :loading="testingInvoice === 'multimodal'"
                    @click="testInvoice('multimodal')"
                  >
                    测试连通性
                  </el-button>
                </footer>
              </section>
            </section>
          </el-form>
        </template>

        <!-- 3. 网关安全 -->
        <template v-else-if="activeSectionName === 'security'">
          <div class="security-module-container">
            <section class="security-card">
              <header class="security-card__header">
                <div class="sec-header-left">
                  <div class="sec-icon-circle"><el-icon><Lock /></el-icon></div>
                  <div>
                    <h3>网关核心治理与网络边界</h3>
                    <p>管理全局 AI 流量进出策略，强制实施鉴权、配额、数据脱敏与内网防火墙保护。</p>
                  </div>
                </div>
              </header>

              <div class="setting-list">
                <div class="setting-toggle">
                  <div class="setting-toggle__copy">
                    <strong>统一 AI Gateway 调度核心</strong>
                    <small>全系统所有业务模块（资产录入、发票识别、智能助手）均必须通过统一网关执行调用审计、Token 配额与熔断。</small>
                  </div>
                  <el-switch v-model="providers.enabled" active-text="启用" inactive-text="关闭" />
                </div>
                <div class="setting-toggle">
                  <div class="setting-toggle__copy">
                    <strong>允许模型服务访问私有企业内网</strong>
                    <small>仅在私有化部署的大模型集群处于 10.x / 172.16.x / 192.168.x 等企业私网时开启，避免内部探测风险。</small>
                  </div>
                  <el-switch v-model="providers['allow-private-endpoints']" active-text="允许" inactive-text="拒绝" />
                </div>
              </div>
            </section>

            <section class="security-card">
              <header class="security-card__header">
                <div class="sec-header-left">
                  <div class="sec-icon-circle sec-icon-circle--accent"><el-icon><Setting /></el-icon></div>
                  <div>
                    <h3>内容合规与数据防泄露拦截 (DLP)</h3>
                    <p>在 Prompt 发送与图片外发至第三方大模型服务前实施深度拦截。</p>
                  </div>
                </div>
              </header>

              <el-form label-position="top" class="settings-form security-form">
                <div class="security-grid">
                  <div class="security-field-card">
                    <el-form-item label="业务敏感词阻断过滤">
                      <el-select
                        v-model="providers['sensitive-words']"
                        multiple
                        filterable
                        allow-create
                        default-first-option
                        placeholder="输入敏感词后按回车添加"
                        class="full-width-select"
                      />
                      <span class="field-help">
                        <el-icon><InfoFilled /></el-icon>
                        Prompt 中若命中上述词汇，请求将在进入外发网络前被立即阻断并记录审计事件。
                      </span>
                    </el-form-item>
                  </div>

                  <div class="security-field-card">
                    <el-form-item label="允许外发图片的业务模块白名单">
                      <el-select
                        v-model="providers['allow-vision-modules']"
                        multiple
                        filterable
                        allow-create
                        default-first-option
                        placeholder="输入模块标识后按回车添加 (如 asset, invoice)"
                        class="full-width-select"
                      />
                      <span class="field-help">
                        <el-icon><InfoFilled /></el-icon>
                        未在白名单登记的系统模块，严禁向外部第三方大模型服务发送任何图片或二进制文件。
                      </span>
                    </el-form-item>
                  </div>
                </div>
              </el-form>
            </section>

            <!-- 安全防护态势摘要条 -->
            <div class="security-posture-banner">
              <div class="posture-item">
                <el-icon class="posture-icon is-good"><CircleCheck /></el-icon>
                <div>
                  <strong>传输层 TLS 加密</strong>
                  <small>全部外部调用强制 HTTPS/TLS 1.3</small>
                </div>
              </div>
              <div class="posture-item">
                <el-icon class="posture-icon is-good"><CircleCheck /></el-icon>
                <div>
                  <strong>凭据硬件脱敏存储</strong>
                  <small>数据库仅存放加密密文与指纹</small>
                </div>
              </div>
              <div class="posture-item">
                <el-icon class="posture-icon is-good"><CircleCheck /></el-icon>
                <div>
                  <strong>全量调用审计追踪</strong>
                  <small>严格留存用量与调用耗时元数据</small>
                </div>
              </div>
              <div class="posture-item">
                <el-icon class="posture-icon is-good"><CircleCheck /></el-icon>
                <div>
                  <strong>零数据留存政策</strong>
                  <small>不向第三方模型留存用户业务原文</small>
                </div>
              </div>
            </div>
          </div>
        </template>

        <!-- 4. 配额计费 -->
        <template v-else-if="activeSectionName === 'billing'">
          <section class="settings-group settings-group--first">
            <header class="group-heading">
              <div>
                <h3>模型调用单价配置 (Pricing Matrix)</h3>
                <p>根据各供应商账单价格录入单价，系统将依据每次调用的输入/输出 Token 数量精确估算模型开销。</p>
              </div>
            </header>

            <div class="pricing-card-wrapper">
              <div class="pricing-table">
                <div class="pricing-row pricing-row--header">
                  <span>Provider 供应商</span>
                  <span>输入费用 / 百万 Token</span>
                  <span>输出费用 / 百万 Token</span>
                  <span class="align-center">接入状态</span>
                </div>
                <div v-for="provider in providerList" :key="provider.key" class="pricing-row">
                  <div class="pricing-provider-cell">
                    <div class="provider-brand-badge brand--sm" :class="'brand--' + provider.key">
                      <el-icon v-if="provider.key === 'openai-compatible'"><Cpu /></el-icon>
                      <el-icon v-else><Promotion /></el-icon>
                    </div>
                    <strong>{{ provider.label }}</strong>
                  </div>
                  <div class="pricing-input-cell">
                    <el-input
                      :model-value="providers[provider.key]['input-cost-per-million']"
                      inputmode="decimal"
                      placeholder="0.000000"
                      @update:model-value="providers[provider.key]['input-cost-per-million'] = decimalValue($event)"
                    >
                      <template #prefix><span class="currency-prefix">¥</span></template>
                      <template #suffix><span class="unit-suffix">/ 1M</span></template>
                    </el-input>
                  </div>
                  <div class="pricing-input-cell">
                    <el-input
                      :model-value="providers[provider.key]['output-cost-per-million']"
                      inputmode="decimal"
                      placeholder="0.000000"
                      @update:model-value="providers[provider.key]['output-cost-per-million'] = decimalValue($event)"
                    >
                      <template #prefix><span class="currency-prefix">¥</span></template>
                      <template #suffix><span class="unit-suffix">/ 1M</span></template>
                    </el-input>
                  </div>
                  <div class="align-center">
                    <el-tag :type="providers[provider.key].enabled ? 'success' : 'info'" effect="light">
                      {{ providers[provider.key].enabled ? '已启用' : '未启用' }}
                    </el-tag>
                  </div>
                </div>
              </div>
            </div>
          </section>

          <section class="settings-group quota-section">
            <header class="group-heading group-heading--action">
              <div>
                <h3>用量与并发配额管理</h3>
                <p>支持按全局、业务模块、用户角色以及指定用户设置调用上限。所有限制同时生效，填 0 表示不设上限。</p>
              </div>
              <el-button type="primary" :icon="Plus" @click="openQuota()">新增配额</el-button>
            </header>

            <el-table
              v-if="quotaLoading || quotas.length > 0"
              v-loading="quotaLoading && !quotaLoaded"
              :data="quotas"
              row-key="ID"
              class="modern-table"
            >
              <el-table-column prop="scopeType" label="控制范围" width="120">
                <template #default="{ row }">
                  <el-tag :type="getScopeTypeMeta(row.scopeType).type" effect="light">
                    {{ getScopeTypeMeta(row.scopeType).label }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="scopeId" label="范围标识" min-width="170">
                <template #default="{ row }">
                  <span class="code-badge">{{ row.scopeId }}</span>
                </template>
              </el-table-column>
              <el-table-column prop="dailyRequests" label="每日请求上限" width="130" align="right">
                <template #default="{ row }">
                  <span class="mono-num">{{ row.dailyRequests ? number(row.dailyRequests) : '无限制' }}</span>
                </template>
              </el-table-column>
              <el-table-column prop="dailyTokens" label="每日 Token 上限" width="140" align="right">
                <template #default="{ row }">
                  <span class="mono-num">{{ row.dailyTokens ? number(row.dailyTokens) : '无限制' }}</span>
                </template>
              </el-table-column>
              <el-table-column prop="monthlyCostMicros" label="月度预算上限" width="140" align="right">
                <template #default="{ row }">
                  <span class="mono-num money-text">{{ money(row.monthlyCostMicros) }}</span>
                </template>
              </el-table-column>
              <el-table-column prop="maxConcurrency" label="最大并发" width="110" align="right">
                <template #default="{ row }">
                  <span class="mono-num">{{ row.maxConcurrency || '无限制' }}</span>
                </template>
              </el-table-column>
              <el-table-column label="生效状态" width="100" align="center">
                <template #default="{ row }">
                  <span class="table-status-pill" :class="row.enabled ? 'is-active' : 'is-idle'">
                    {{ row.enabled ? '已生效' : '已停用' }}
                  </span>
                </template>
              </el-table-column>
              <el-table-column label="操作" width="90" align="center">
                <template #default="{ row }">
                  <el-button text type="primary" :icon="Edit" @click="openQuota(row)">编辑</el-button>
                </template>
              </el-table-column>
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
            <AppEmptyState
              v-if="!quotaLoading && !quotas.length"
              class="quota-empty-state"
              compact
              title="尚未设置智能服务配额"
              description="可按全局、模块、角色或用户限制每日请求量、Token 上限、月度预算与并发数。"
            />
          </section>
        </template>

        <!-- 5. Prompt 模板 -->
        <template v-else-if="activeSectionName === 'prompts'">
          <div class="section-toolbar">
            <div class="toolbar-hint">
              <el-icon><InfoFilled /></el-icon>
              <span>每个 Prompt 标识仅允许激活一个版本。创建草稿后，经人工确认激活才会生效到 Gateway。</span>
            </div>
            <el-button type="primary" :icon="Plus" @click="openPrompt">创建新版本</el-button>
          </div>

          <el-table
            v-loading="promptLoading && !promptLoaded"
            :data="prompts"
            row-key="ID"
            class="modern-table"
          >
            <el-table-column prop="promptKey" label="模板唯一标识" min-width="180">
              <template #default="{ row }">
                <span class="code-badge code-badge--primary">{{ row.promptKey }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="version" label="版本" width="90" align="center">
              <template #default="{ row }">
                <span class="version-badge">v{{ row.version }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="status" label="状态" width="110" align="center">
              <template #default="{ row }">
                <el-tag :type="promptStatus(row.status).type" effect="light">
                  {{ promptStatus(row.status).label }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="CreatedAt" label="创建时间" min-width="170">
              <template #default="{ row }">{{ dateTime(row.CreatedAt) }}</template>
            </el-table-column>
            <el-table-column label="Prompt 提示词内容" min-width="300" show-overflow-tooltip>
              <template #default="{ row }">
                <span class="prompt-content-snippet">{{ row.content }}</span>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="110" align="center">
              <template #default="{ row }">
                <el-button
                  v-if="row.status !== 'active'"
                  text
                  type="primary"
                  :icon="Check"
                  :loading="activatingPromptId === row.ID"
                  @click="activatePrompt(row)"
                >
                  激活上线
                </el-button>
                <span v-else class="active-mark-text">当前活跃</span>
              </template>
            </el-table-column>
            <template #empty>
              <AppEmptyState compact title="暂无 Prompt 版本" description="创建 Prompt 草稿并人工激活后，智能服务网关将调用对应版本。">
                <template #actions>
                  <el-button type="primary" :icon="Plus" @click="openPrompt">创建首个版本</el-button>
                </template>
              </AppEmptyState>
            </template>
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

        <!-- 6. 运行监控 -->
        <template v-else>
          <!-- 统计指标卡片组 -->
          <section class="summary-band" aria-label="智能服务用量摘要">
            <div class="kpi-card kpi--blue">
              <div class="kpi-icon-box">
                <el-icon><Cpu /></el-icon>
              </div>
              <div class="kpi-meta">
                <span>今日调用次数</span>
                <strong>{{ usage.todayRequests || 0 }}</strong>
                <small>当前账户成功处理的推理请求</small>
              </div>
            </div>

            <div class="kpi-card kpi--violet">
              <div class="kpi-icon-box">
                <el-icon><DataAnalysis /></el-icon>
              </div>
              <div class="kpi-meta">
                <span>今日消耗 Token</span>
                <strong>{{ number(usage.todayTokens) }}</strong>
                <small>包含输入 Prompt 与输出 Completion</small>
              </div>
            </div>

            <div class="kpi-card kpi--emerald">
              <div class="kpi-icon-box">
                <el-icon><Coin /></el-icon>
              </div>
              <div class="kpi-meta">
                <span>本月估算费用</span>
                <strong>{{ money(usage.monthCostMicros) }}</strong>
                <small>按供应商实际配置单价加权测算</small>
              </div>
            </div>

            <div class="kpi-card kpi--amber">
              <div class="kpi-icon-box">
                <el-icon><Document /></el-icon>
              </div>
              <div class="kpi-meta">
                <span>累计调用总计</span>
                <strong>{{ number(usage.totalRequests) }}</strong>
                <small>系统上线以来的总调用次数</small>
              </div>
            </div>
          </section>

          <!-- 筛选过滤栏 -->
          <div class="filter-card">
            <div class="filter-row">
              <el-select v-model="invocationSearch.status" clearable placeholder="全部状态" class="filter-select">
                <el-option label="成功" value="success" />
                <el-option label="失败" value="failed" />
                <el-option label="已阻断" value="blocked" />
              </el-select>
              <el-input v-model="invocationSearch.module" clearable placeholder="业务模块 (如 asset)" />
              <el-input v-model="invocationSearch.provider" clearable placeholder="Provider 供应商" />
              <el-input-number v-model="invocationSearch.userId" :min="1" :controls="false" placeholder="用户 ID" />
              <el-button type="primary" :icon="Search" :loading="invocationLoading" @click="searchInvocations">
                查询
              </el-button>
            </div>
          </div>

          <!-- 调用明细表格 -->
          <el-table
            v-loading="invocationLoading && !invocationLoaded"
            :data="invocations"
            row-key="ID"
            class="modern-table"
          >
            <el-table-column prop="CreatedAt" label="请求时间" min-width="165">
              <template #default="{ row }">
                <span class="mono-num">{{ dateTime(row.CreatedAt) }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="userId" label="用户 ID" width="90" align="center">
              <template #default="{ row }">
                <span class="code-badge">#{{ row.userId }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="module" label="业务模块" min-width="110">
              <template #default="{ row }">
                <span class="table-module-tag">{{ row.module }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="operation" label="操作名称" min-width="150">
              <template #default="{ row }">
                <span class="table-op-name">{{ row.operation }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="provider" label="Provider" min-width="140">
              <template #default="{ row }">
                <span class="table-provider-name">{{ row.provider }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="model" label="使用模型" min-width="150" show-overflow-tooltip>
              <template #default="{ row }">
                <span class="table-model-name">{{ row.model }}</span>
              </template>
            </el-table-column>
            <el-table-column label="总 Token" width="120" align="right">
              <template #default="{ row }">
                <span class="mono-num">{{ number(Number(row.inputTokens) + Number(row.outputTokens)) }}</span>
              </template>
            </el-table-column>
            <el-table-column label="估算费用" width="120" align="right">
              <template #default="{ row }">
                <span class="mono-num money-text">{{ money(row.estimatedCostMicros) }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="durationMs" label="耗时" width="110" align="right">
              <template #default="{ row }">
                <span
                  class="latency-badge"
                  :class="{
                    'latency--fast': row.durationMs < 800,
                    'latency--normal': row.durationMs >= 800 && row.durationMs <= 2500,
                    'latency--slow': row.durationMs > 2500
                  }"
                >
                  {{ row.durationMs }} ms
                </span>
              </template>
            </el-table-column>
            <el-table-column prop="errorType" label="错误类型" width="110">
              <template #default="{ row }">
                <span :class="{ 'error-type-text': row.errorType }">{{ row.errorType || '—' }}</span>
              </template>
            </el-table-column>
            <el-table-column label="状态" width="100" align="center">
              <template #default="{ row }">
                <el-tag :type="statusMeta(row.status).type" effect="light">
                  {{ statusMeta(row.status).label }}
                </el-tag>
              </template>
            </el-table-column>
            <template #empty>
              <AppEmptyState
                compact
                title="暂无模型调用记录"
                description="智能业务通过统一 Gateway 调用模型后，此处将全景呈现状态、用量、耗时与错误分析。"
                :highlights="['不保存 Prompt 原文以确保隐私安全', '不留存原始图片与模型输出明细']"
              />
            </template>
          </el-table>

          <div class="na-pagination">
            <el-pagination
              v-model:current-page="invocationSearch.page"
              v-model:page-size="invocationSearch.pageSize"
              :total="invocationTotal"
              layout="total, sizes, prev, pager, next"
              :page-sizes="[10, 20, 50, 100]"
              @change="loadInvocations"
              @size-change="resetInvocationPage"
            />
          </div>
        </template>
      </section>
    </div>

    <!-- 配额弹窗 -->
    <el-dialog
      v-model="quotaDialogVisible"
      :title="quotaForm.ID ? '编辑智能服务配额' : '新增智能服务配额'"
      width="min(600px, calc(100vw - 32px))"
      destroy-on-close
      class="smart-dialog"
    >
      <el-form label-position="top">
        <div class="dialog-grid">
          <el-form-item label="控制范围">
            <el-select v-model="quotaForm.scopeType" class="full-width-select">
              <el-option label="全局 (Global)" value="global" />
              <el-option label="业务模块 (Module)" value="module" />
              <el-option label="角色权限 (Authority)" value="authority" />
              <el-option label="指定用户 (User)" value="user" />
            </el-select>
          </el-form-item>
          <el-form-item label="范围标识">
            <el-input
              v-model="quotaForm.scopeId"
              :placeholder="quotaForm.scopeType === 'global' ? 'global' : '模块标识、角色ID或用户ID'"
            />
          </el-form-item>
        </div>
        <div class="dialog-grid">
          <el-form-item label="每日请求上限 (次)">
            <el-input-number v-model="quotaForm.dailyRequests" :min="0" class="full-width-number" />
          </el-form-item>
          <el-form-item label="每日 Token 上限">
            <el-input-number v-model="quotaForm.dailyTokens" :min="0" class="full-width-number" />
          </el-form-item>
          <el-form-item label="月度预算上限（元）">
            <el-input
              :model-value="quotaForm.monthlyBudgetYuan"
              inputmode="decimal"
              placeholder="0.00"
              @update:model-value="quotaForm.monthlyBudgetYuan = decimalValue($event)"
            >
              <template #prefix><span class="currency-prefix">¥</span></template>
            </el-input>
          </el-form-item>
          <el-form-item label="最大并发连接数">
            <el-input-number v-model="quotaForm.maxConcurrency" :min="0" class="full-width-number" />
          </el-form-item>
        </div>
        <div class="dialog-switch-row">
          <el-switch v-model="quotaForm.enabled" active-text="立即启用此配额策略" />
        </div>
      </el-form>
      <template #footer>
        <el-button @click="quotaDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="savingQuota" @click="submitQuota">保存配额</el-button>
      </template>
    </el-dialog>

    <!-- Prompt 模板弹窗 -->
    <el-dialog
      v-model="promptDialogVisible"
      title="创建 Prompt 模板版本"
      width="min(760px, calc(100vw - 32px))"
      destroy-on-close
      class="smart-dialog"
    >
      <el-form label-position="top">
        <el-form-item label="Prompt 模板标识 (Key)">
          <el-input v-model="promptForm.promptKey" placeholder="例如 asset-draft-v1" />
        </el-form-item>
        <el-form-item label="Prompt 提示词内容">
          <el-input
            v-model="promptForm.content"
            type="textarea"
            :rows="9"
            maxlength="131072"
            show-word-limit
            placeholder="输入针对大模型的系统提示词与指引指令..."
          />
        </el-form-item>
        <el-form-item label="输出 JSON Schema 约束（可选）">
          <el-input
            v-model="promptForm.outputSchema"
            type="textarea"
            :rows="4"
            placeholder="规范模型结构化输出的 JSON Schema 格式..."
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="promptDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="savingPrompt" @click="submitPrompt">创建并提交草稿</el-button>
      </template>
    </el-dialog>
  </main>
</template>

<script setup>
import { computed, onMounted, reactive, ref, watch } from 'vue'
import {
  ArrowRight,
  Check,
  CircleCheck,
  Coin,
  Connection,
  Cpu,
  DataAnalysis,
  Document,
  Edit,
  InfoFilled,
  Lock,
  Picture,
  Plus,
  Promotion,
  Refresh,
  Search,
  Setting,
  WarningFilled
} from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import AppEmptyState from '@/components/page/AppEmptyState.vue'
import AppPageHeader from '@/components/page/AppPageHeader.vue'
import SecretInput from '@/components/secretInput/index.vue'
import { useUserStore } from '@/pinia/modules/user'
import {
  activateAIPrompt,
  createAIPrompt,
  getAIInvocations,
  getAIInvoiceRecognition,
  getAIProviders,
  getAIPrompts,
  getAIQuotas,
  getAIUsageSummary,
  saveAIQuota,
  testAIInvoiceRecognition,
  updateAIInvoiceRecognition,
  updateAIProviders
} from '@/plugin/aioperations/api/operations'
import { providerFormValue, providerPayloadValue, providerSecretPath } from '@/plugin/aioperations/utils/provider'
import { defaultInvoiceRecognition, invoiceRecognitionFormValue, invoiceRecognitionPayload } from '@/plugin/aioperations/utils/invoiceRecognition'
import { defaultQuota, quotaFormValue, quotaPayloadValue } from '@/plugin/aioperations/utils/quota'

defineOptions({ name: 'SmartCapabilitySettings' })

const settingsSections = [
  { name: 'models', label: '模型接入', hint: '大模型接入点与凭据', description: '配置通用大模型推理服务的端点、模型标识、访问密钥与超时时间。', icon: Cpu },
  { name: 'recognition', label: '识别服务', hint: 'OCR、验真与视觉兜底', description: '编排发票 OCR、权威税局验真接口与低置信度视觉大模型兜底识别流水线。', icon: Picture },
  { name: 'security', label: '网关安全', hint: '访问边界与敏感词管控', description: '控制统一智能 Gateway 入口、私网 IP 访问特权、业务敏感词与多模态数据外发范围。', icon: Lock },
  { name: 'billing', label: '配额计费', hint: '单价矩阵与用量边界', description: '维护各模型输入输出单价，并对全局、模块、角色和用户设定多层并发与消费配额。', icon: Coin },
  { name: 'prompts', label: 'Prompt 模板', hint: '版本迭代与激活回退', description: '管理各智能业务场景的 Prompt 提示词版本，支持草稿创建与即时热激活。', icon: Document },
  { name: 'monitoring', label: '运行监控', hint: '用量大盘与请求审计', description: '实时观测智能服务的调用量、Token 吞吐、费用支出、响应延迟与异常阻断记录。', icon: DataAnalysis }
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
  { key: 'openai-compatible', label: 'OpenAI Compatible', hint: '兼容 Chat Completions 标准协议的大模型服务。' },
  { key: 'anthropic', label: 'Anthropic', hint: '采用 Messages API 标准协议的 Claude 系列模型。' }
]
const invocationSearch = reactive({ page: 1, pageSize: 10, status: '', module: '', provider: '', userId: undefined })
const quotaSearch = reactive({ page: 1, pageSize: 10 })
const promptSearch = reactive({ page: 1, pageSize: 10 })
const quotaForm = reactive(defaultQuota())
const promptForm = reactive({ promptKey: '', content: '', outputSchema: '' })

const providerEnabledCount = computed(() => providerList.filter((item) => providers[item.key].enabled).length)
const recognitionEnabledCount = computed(() => ['baidu', 'public-ocr', 'verification', 'multimodal'].filter((key) => invoice[key].enabled).length)
const sectionBadge = computed(() => {
  if (activeSectionName.value === 'models') return { label: providerEnabledCount.value + ' 个 Provider 运行中', type: providerEnabledCount.value ? 'success' : 'info' }
  if (activeSectionName.value === 'recognition') return { label: recognitionEnabledCount.value + ' 项服务已接入', type: recognitionEnabledCount.value ? 'success' : 'info' }
  if (activeSectionName.value === 'security') return { label: providers.enabled ? 'Gateway 核心正常' : 'Gateway 已阻断', type: providers.enabled ? 'success' : 'warning' }
  if (activeSectionName.value === 'billing') return { label: quotaTotal.value + ' 条配额规则', type: 'info' }
  if (activeSectionName.value === 'prompts') return { label: promptTotal.value + ' 个模板版本', type: 'info' }
  return { label: number(usage.value.todayRequests) + ' 次今日调用', type: 'info' }
})
const saveActionLabel = computed(() => ({ models: '保存模型接入', recognition: '保存识别服务', security: '保存安全策略', billing: '保存模型单价' }[activeSectionName.value] || ''))
const activeSaving = computed(() => activeSectionName.value === 'recognition' ? savingInvoice.value : savingProviders.value)
const currentSaveTime = computed(() => activeSectionName.value === 'recognition' ? lastInvoiceSaveTime.value : ['models', 'security', 'billing'].includes(activeSectionName.value) ? lastProviderSaveTime.value : '')

function getScopeTypeMeta(scope) {
  const map = {
    global: { label: '全局', type: 'primary' },
    module: { label: '模块', type: 'success' },
    authority: { label: '角色', type: 'warning' },
    user: { label: '用户', type: 'info' }
  }
  return map[scope] || { label: scope, type: 'info' }
}

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
    } else ElMessage.error(response.msg || '无法读取用量配额')
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
    } else ElMessage.error(response.msg || '无法读取 Prompt 模板')
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
.smart-settings-page {
  min-width: 0;
}

// -----------------------------------------------------------------------------
// 主外壳容器：高质感双栏卡片分栏体系
// -----------------------------------------------------------------------------
.settings-shell {
  display: grid;
  grid-template-columns: 260px minmax(0, 1fr);
  min-height: 680px;
  overflow: hidden;
  border: 1px solid var(--na-border);
  border-radius: var(--na-radius, 12px);
  background: var(--na-card);
  box-shadow: var(--na-shadow-sm);
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
}

// -----------------------------------------------------------------------------
// 左侧导航栏：现代卡片式分组项
// -----------------------------------------------------------------------------
.settings-nav {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 16px 14px;
  border-right: 1px solid var(--na-border);
  background: var(--na-muted, #f8f9fa);

  &__header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 4px 10px 14px;
    border-bottom: 1px solid var(--na-border);
    margin-bottom: 6px;
  }

  &__badge {
    font-size: 0.7rem;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.06em;
    padding: 2px 8px;
    border-radius: 4px;
    background: var(--na-primary-soft);
    color: var(--na-primary);
  }

  &__caption {
    font-size: 0.72rem;
    color: var(--na-muted-foreground);
  }

  &__item {
    position: relative;
    display: flex;
    align-items: center;
    width: 100%;
    gap: 12px;
    padding: 10px 12px;
    border: 1px solid transparent;
    border-radius: 8px;
    color: var(--na-muted-foreground);
    background: transparent;
    text-align: left;
    cursor: pointer;
    transition: all 0.2s cubic-bezier(0.16, 1, 0.3, 1);

    &:hover {
      color: var(--na-foreground);
      background: var(--na-card);
      border-color: var(--na-border);
      transform: translateX(2px);

      .nav-arrow {
        opacity: 0.6;
        transform: translateX(0);
      }
    }

    &:focus-visible {
      outline: 2px solid var(--na-primary);
      outline-offset: 2px;
    }

    &.is-active {
      color: var(--na-primary);
      border-color: var(--na-border-strong, var(--na-border));
      background: var(--na-card);
      box-shadow: var(--na-shadow-sm);

      .nav-arrow {
        opacity: 1;
        transform: translateX(0);
        color: var(--na-primary);
      }

      .nav-icon-box {
        background: var(--na-primary-soft);
        color: var(--na-primary);
        box-shadow: 0 0 0 1px var(--na-primary-soft);
      }
    }
  }
}

.nav-icon-box {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 34px;
  height: 34px;
  border-radius: 8px;
  background: var(--na-card);
  color: var(--na-muted-foreground);
  border: 1px solid var(--na-border);
  font-size: 16px;
  flex-shrink: 0;
  transition: all 0.2s ease;
}

.nav-label-box {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
  flex: 1;

  strong {
    color: inherit;
    font-size: 0.86rem;
    font-weight: 600;
  }

  small {
    font-size: 0.72rem;
    color: var(--na-muted-foreground);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.nav-arrow {
  font-size: 12px;
  opacity: 0;
  transform: translateX(-4px);
  transition: all 0.2s ease;
  color: var(--na-muted-foreground);
}

// -----------------------------------------------------------------------------
// 右侧内容区与标题头
// -----------------------------------------------------------------------------
.settings-content {
  min-width: 0;
  padding: 24px 28px;
  background: var(--na-card);
}

.settings-section-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 20px;
  padding-bottom: 18px;
  margin-bottom: 20px;
  border-bottom: 1px solid var(--na-border);

  .header-main {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .header-title-row {
    display: flex;
    align-items: center;
    gap: 12px;
    flex-wrap: wrap;

    h2 {
      margin: 0;
      color: var(--na-foreground);
      font-size: 1.18rem;
      font-weight: 700;
      letter-spacing: -0.01em;
    }
  }

  p {
    margin: 0;
    color: var(--na-muted-foreground);
    font-size: 0.8rem;
    line-height: 1.5;
  }
}

.status-indicator-dot {
  display: inline-block;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  margin-right: 5px;

  &--success {
    background: var(--na-success);
    box-shadow: 0 0 0 2px var(--na-success-soft);
  }
  &--warning {
    background: var(--na-warning);
    box-shadow: 0 0 0 2px var(--na-warning-soft);
  }
  &--info {
    background: var(--na-info);
    box-shadow: 0 0 0 2px var(--na-info-soft);
  }
}

.save-state-chip {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 5px 12px;
  border-radius: 20px;
  background: var(--na-success-soft);
  color: var(--na-success);
  font-size: 0.74rem;
  font-weight: 500;
  white-space: nowrap;
}

// -----------------------------------------------------------------------------
// 网关预警 Banner
// -----------------------------------------------------------------------------
.gateway-alert-card {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 14px 18px;
  margin-bottom: 20px;
  border-radius: 10px;
  background: var(--na-warning-soft);
  border: 1px solid rgba(217, 119, 6, 0.25);
  color: var(--na-foreground);

  .gateway-alert-icon {
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 24px;
    color: var(--na-warning);
    flex-shrink: 0;
  }

  .gateway-alert-body {
    flex: 1;
    min-width: 0;

    strong {
      display: block;
      font-size: 0.86rem;
      font-weight: 600;
      color: var(--na-action-warning, var(--na-warning));
    }

    p {
      margin: 3px 0 0;
      font-size: 0.75rem;
      color: var(--na-muted-foreground);
      line-height: 1.45;
    }
  }
}

.security-tip-strip {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 16px;
  margin-bottom: 18px;
  border-radius: 8px;
  background: var(--na-info-soft);
  border: 1px solid rgba(2, 132, 199, 0.2);
  color: var(--na-action-info, var(--na-info));
  font-size: 0.76rem;
  line-height: 1.45;

  .el-icon {
    font-size: 16px;
    flex-shrink: 0;
  }
}

// -----------------------------------------------------------------------------
// Provider 卡片系统 (OpenAI, Claude, OCR, etc.)
// -----------------------------------------------------------------------------
.provider-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 18px;
}

.provider-card {
  display: flex;
  flex-direction: column;
  min-width: 0;
  border: 1px solid var(--na-border);
  border-radius: 10px;
  background: var(--na-surface-muted, #fafafa);
  transition: all 0.24s cubic-bezier(0.16, 1, 0.3, 1);

  &:hover {
    border-color: var(--na-border-strong, var(--na-primary-soft));
    box-shadow: var(--na-shadow-sm);
    transform: translateY(-1px);
  }

  &.is-card-enabled {
    border-color: var(--na-primary-soft);
    background: var(--na-card);
    box-shadow: 0 4px 12px rgb(109 93 251 / 4%);
  }

  &--wide {
    grid-column: 1 / -1;
    background: var(--na-card);
  }

  &__header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    padding: 16px;
    border-bottom: 1px solid var(--na-border);
  }

  &__body {
    padding: 16px;
    flex: 1;
  }

  &__footer {
    display: flex;
    min-height: 44px;
    flex-wrap: wrap;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    padding: 10px 16px;
    border-top: 1px dashed var(--na-border);
    background: var(--na-surface-muted, #fcfcfc);
    border-radius: 0 0 10px 10px;
  }
}

.provider-identity {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
}

.provider-brand-badge {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border-radius: 10px;
  background: var(--na-primary-soft);
  color: var(--na-primary);
  font-size: 20px;
  flex-shrink: 0;

  &.brand--anthropic {
    background: rgb(217 119 6 / 12%);
    color: #d97706;
  }

  &.brand--baidu {
    background: rgb(2 132 199 / 12%);
    color: #0284c7;
  }

  &.brand--gateway {
    background: rgb(5 150 105 / 12%);
    color: #059669;
  }

  &.brand--verify {
    background: rgb(139 92 246 / 12%);
    color: #8b5cf6;
  }

  &.brand--multimodal {
    background: rgb(244 63 94 / 12%);
    color: #f43f5e;
  }

  &.brand--sm {
    width: 28px;
    height: 28px;
    font-size: 15px;
    border-radius: 6px;
  }
}

.provider-title-row {
  display: flex;
  align-items: center;
  gap: 8px;

  h3 {
    margin: 0;
    color: var(--na-foreground);
    font-size: 0.94rem;
    font-weight: 600;
  }
}

.provider-identity p {
  margin: 3px 0 0;
  color: var(--na-muted-foreground);
  font-size: 0.74rem;
  line-height: 1.4;
}

.provider-active-chip {
  font-size: 0.68rem;
  font-weight: 500;
  padding: 1px 7px;
  border-radius: 10px;

  &.is-active {
    background: var(--na-success-soft);
    color: var(--na-success);
  }

  &.is-idle {
    background: var(--na-muted);
    color: var(--na-muted-foreground);
  }
}

// -----------------------------------------------------------------------------
// 表单网格系统与输入微调
// -----------------------------------------------------------------------------
.field-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px 14px;

  &--four {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }
}

.field--wide {
  grid-column: 1 / -1;
}

.field--double {
  grid-column: span 2;
}

.full-width-number {
  width: 100% !important;
}

.full-width-select {
  width: 100% !important;
}

.input-icon {
  color: var(--na-muted-foreground);
  font-size: 14px;
}

.footer-secret-info {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 0.74rem;

  .secret-shield-icon {
    color: var(--na-success);
    font-size: 14px;
  }

  .secret-state {
    color: var(--na-success);
    font-weight: 500;
  }

  .secret-empty-state {
    color: var(--na-muted-foreground);
  }
}

.footer-actions-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.detected-label {
  display: inline-flex;
  align-items: center;
  padding: 2px 8px;
  border-radius: 4px;
  background: var(--na-info-soft);
  color: var(--na-action-info, var(--na-info));
  font-size: 0.72rem;
  font-weight: 500;
}

// -----------------------------------------------------------------------------
// 识别流水线编排策略卡片
// -----------------------------------------------------------------------------
.strategy-pipeline-card {
  padding: 18px 20px;
  margin-bottom: 22px;
  border: 1px solid var(--na-border);
  border-radius: 10px;
  background: var(--na-card);
  box-shadow: var(--na-shadow-sm);
}

.pipeline-steps-visual {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 16px 0;
  margin: 10px 0 16px;
  border-top: 1px solid var(--na-border);
  border-bottom: 1px solid var(--na-border);
  overflow-x: auto;
}

.pipeline-step {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 14px;
  border-radius: 8px;
  background: var(--na-muted);
  flex-shrink: 0;

  &__num {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 22px;
    height: 22px;
    border-radius: 50%;
    background: var(--na-primary);
    color: var(--na-on-primary, #fff);
    font-size: 0.72rem;
    font-weight: 700;
  }

  strong {
    display: block;
    color: var(--na-foreground);
    font-size: 0.82rem;
  }

  small {
    display: block;
    color: var(--na-muted-foreground);
    font-size: 0.7rem;
  }
}

.pipeline-divider {
  color: var(--na-border-strong, #ccc);
  font-size: 0.8rem;
  user-select: none;
}

.policy-controls-row {
  display: grid;
  grid-template-columns: 340px minmax(0, 1fr);
  gap: 20px;
  align-items: center;
}

.threshold-field-box {
  display: flex;
  flex-direction: column;
  gap: 6px;

  .custom-field-label {
    color: var(--na-foreground);
    font-size: 0.82rem;
    font-weight: 600;
  }

  .threshold-input-group {
    display: flex;
    align-items: center;
    gap: 10px;

    :deep(.el-input-number) {
      width: 120px;
    }
  }

  .threshold-badge {
    color: var(--na-muted-foreground);
    font-size: 0.74rem;
  }
}

.setting-toggle {
  display: flex;
  min-width: 0;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  padding: 14px 0;
  border-bottom: 1px solid var(--na-border);

  &--inline {
    padding: 0;
    border-bottom: 0;
  }
}

.setting-toggle__copy {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 4px;

  strong {
    color: var(--na-foreground);
    font-size: 0.86rem;
    font-weight: 600;
  }

  small {
    color: var(--na-muted-foreground);
    font-size: 0.74rem;
    line-height: 1.45;
  }
}

// -----------------------------------------------------------------------------
// 网关安全模块
// -----------------------------------------------------------------------------
.security-module-container {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.security-card {
  padding: 20px;
  border: 1px solid var(--na-border);
  border-radius: 10px;
  background: var(--na-card);
  box-shadow: var(--na-shadow-sm);

  &__header {
    margin-bottom: 16px;

    .sec-header-left {
      display: flex;
      align-items: center;
      gap: 12px;

      h3 {
        margin: 0;
        color: var(--na-foreground);
        font-size: 0.96rem;
        font-weight: 600;
      }

      p {
        margin: 3px 0 0;
        color: var(--na-muted-foreground);
        font-size: 0.74rem;
      }
    }
  }
}

.sec-icon-circle {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: var(--na-primary-soft);
  color: var(--na-primary);
  font-size: 18px;
  flex-shrink: 0;

  &--accent {
    background: rgb(2 132 199 / 12%);
    color: #0284c7;
  }
}

.security-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
}

.security-field-card {
  padding: 16px;
  border-radius: 8px;
  background: var(--na-muted);
  border: 1px solid var(--na-border);

  .field-help {
    display: flex;
    align-items: flex-start;
    gap: 6px;
    margin-top: 8px;
    color: var(--na-muted-foreground);
    font-size: 0.72rem;
    line-height: 1.45;

    .el-icon {
      margin-top: 2px;
      flex-shrink: 0;
    }
  }
}

.security-posture-banner {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 14px;
  padding: 16px 20px;
  border-radius: 10px;
  background: var(--na-surface-muted, #fafafa);
  border: 1px solid var(--na-border);

  .posture-item {
    display: flex;
    align-items: flex-start;
    gap: 10px;

    .posture-icon {
      font-size: 18px;
      flex-shrink: 0;

      &.is-good {
        color: var(--na-success);
      }
    }

    strong {
      display: block;
      color: var(--na-foreground);
      font-size: 0.8rem;
      font-weight: 600;
    }

    small {
      display: block;
      margin-top: 2px;
      color: var(--na-muted-foreground);
      font-size: 0.7rem;
      line-height: 1.35;
    }
  }
}

// -----------------------------------------------------------------------------
// 配额与模型计费矩阵
// -----------------------------------------------------------------------------
.pricing-card-wrapper {
  overflow-x: auto;
  border: 1px solid var(--na-border);
  border-radius: 10px;
  box-shadow: var(--na-shadow-sm);
  background: var(--na-card);
}

.pricing-table {
  min-width: 720px;
}

.pricing-row {
  display: grid;
  grid-template-columns: minmax(220px, 1fr) 230px 230px 100px;
  align-items: center;
  gap: 16px;
  padding: 14px 18px;
  border-bottom: 1px solid var(--na-border);
  transition: background-color 0.16s ease;

  &:last-child {
    border-bottom: 0;
  }

  &:not(&--header):hover {
    background: var(--na-muted);
  }

  &--header {
    color: var(--na-muted-foreground);
    background: var(--na-muted);
    font-size: 0.76rem;
    font-weight: 600;
    letter-spacing: 0.02em;
  }
}

.pricing-provider-cell {
  display: flex;
  align-items: center;
  gap: 10px;

  strong {
    color: var(--na-foreground);
    font-size: 0.88rem;
  }
}

.currency-prefix {
  color: var(--na-muted-foreground);
  font-weight: 600;
  font-size: 0.82rem;
}

.unit-suffix {
  color: var(--na-muted-foreground);
  font-size: 0.72rem;
}

.align-center {
  text-align: center;
}

.settings-group {
  padding-top: 22px;
  margin-top: 22px;
  border-top: 1px solid var(--na-border);

  &--first {
    padding-top: 0;
    margin-top: 0;
    border-top: 0;
  }
}

.group-heading {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 14px;

  h3 {
    margin: 0;
    color: var(--na-foreground);
    font-size: 0.98rem;
    font-weight: 600;
  }

  p {
    margin: 4px 0 0;
    color: var(--na-muted-foreground);
    font-size: 0.76rem;
    line-height: 1.45;
  }

  &--action {
    align-items: center;
  }
}

// -----------------------------------------------------------------------------
// 运行监控大盘与 KPI 指标卡
// -----------------------------------------------------------------------------
.summary-band {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 14px;
  margin-bottom: 20px;
}

.kpi-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 18px 20px;
  border: 1px solid var(--na-border);
  border-radius: 10px;
  background: var(--na-card);
  box-shadow: var(--na-shadow-sm);
  transition: all 0.2s cubic-bezier(0.16, 1, 0.3, 1);

  &:hover {
    transform: translateY(-2px);
    box-shadow: var(--na-shadow-md);
  }

  .kpi-icon-box {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 44px;
    height: 44px;
    border-radius: 10px;
    font-size: 22px;
    flex-shrink: 0;
  }

  .kpi-meta {
    display: flex;
    flex-direction: column;
    gap: 3px;
    min-width: 0;

    span {
      color: var(--na-muted-foreground);
      font-size: 0.74rem;
    }

    strong {
      color: var(--na-foreground);
      font-size: 1.35rem;
      font-weight: 700;
      font-variant-numeric: tabular-nums;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    small {
      color: var(--na-muted-foreground);
      font-size: 0.68rem;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
  }

  &.kpi--blue {
    .kpi-icon-box {
      background: rgb(14 165 233 / 12%);
      color: #0ea5e9;
    }
  }

  &.kpi--violet {
    .kpi-icon-box {
      background: rgb(109 93 251 / 12%);
      color: var(--na-primary);
    }
  }

  &.kpi--emerald {
    .kpi-icon-box {
      background: rgb(5 150 105 / 12%);
      color: #059669;
    }
  }

  &.kpi--amber {
    .kpi-icon-box {
      background: rgb(217 119 6 / 12%);
      color: #d97706;
    }
  }
}

// -----------------------------------------------------------------------------
// 过滤器与表格通用强化
// -----------------------------------------------------------------------------
.filter-card {
  padding: 14px 16px;
  margin-bottom: 16px;
  border: 1px solid var(--na-border);
  border-radius: 8px;
  background: var(--na-surface-muted, #fafafa);
}

.filter-row {
  display: grid;
  grid-template-columns: 140px repeat(2, minmax(0, 1fr)) 140px auto;
  gap: 12px;
  align-items: center;

  :deep(.el-input-number) {
    width: 100%;
  }
}

.modern-table {
  width: 100%;
  border: 1px solid var(--na-border);
  border-radius: 8px;
  overflow: hidden;

  :deep(.el-table__header) th {
    background: var(--na-muted);
    color: var(--na-muted-foreground);
    font-size: 0.76rem;
    font-weight: 600;
  }
}

.code-badge {
  display: inline-block;
  padding: 2px 7px;
  border-radius: 4px;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 0.75rem;
  background: var(--na-muted);
  color: var(--na-foreground);
  border: 1px solid var(--na-border);

  &--primary {
    background: var(--na-primary-soft);
    color: var(--na-primary);
    border-color: rgba(109, 93, 251, 0.2);
  }
}

.version-badge {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 12px;
  font-size: 0.72rem;
  font-weight: 700;
  background: var(--na-primary-soft);
  color: var(--na-primary);
}

.mono-num {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-variant-numeric: tabular-nums;
  font-size: 0.8rem;
}

.money-text {
  color: var(--na-success);
  font-weight: 600;
}

.latency-badge {
  display: inline-block;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 0.72rem;
  font-family: ui-monospace, SFMono-Regular, monospace;
  font-weight: 600;

  &.latency--fast {
    background: var(--na-success-soft);
    color: var(--na-success);
  }

  &.latency--normal {
    background: var(--na-warning-soft);
    color: var(--na-warning);
  }

  &.latency--slow {
    background: var(--na-danger-soft);
    color: var(--na-danger);
  }
}

.table-status-pill {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 10px;
  font-size: 0.72rem;
  font-weight: 500;

  &.is-active {
    background: var(--na-success-soft);
    color: var(--na-success);
  }

  &.is-idle {
    background: var(--na-muted);
    color: var(--na-muted-foreground);
  }
}

.table-module-tag {
  font-size: 0.78rem;
  font-weight: 500;
  color: var(--na-foreground);
}

.table-op-name {
  font-size: 0.78rem;
  color: var(--na-muted-foreground);
}

.table-provider-name {
  font-size: 0.78rem;
  color: var(--na-foreground);
  font-weight: 500;
}

.table-model-name {
  font-size: 0.76rem;
  font-family: ui-monospace, SFMono-Regular, monospace;
  color: var(--na-muted-foreground);
}

.error-type-text {
  color: var(--na-danger);
  font-weight: 500;
  font-size: 0.74rem;
}

.prompt-content-snippet {
  font-family: ui-monospace, SFMono-Regular, monospace;
  font-size: 0.76rem;
  color: var(--na-muted-foreground);
}

.active-mark-text {
  color: var(--na-success);
  font-size: 0.76rem;
  font-weight: 500;
}

.section-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 16px;

  .toolbar-hint {
    display: flex;
    align-items: center;
    gap: 8px;
    color: var(--na-muted-foreground);
    font-size: 0.78rem;

    .el-icon {
      color: var(--na-primary);
    }
  }
}

.na-pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}

.quota-empty-state {
  min-height: 200px;
  border: 1px solid var(--na-border);
  border-radius: 8px;
}

// -----------------------------------------------------------------------------
// 弹窗表单布局
// -----------------------------------------------------------------------------
.dialog-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0 16px;
}

.dialog-switch-row {
  padding-top: 10px;
  margin-top: 10px;
  border-top: 1px solid var(--na-border);
}

// -----------------------------------------------------------------------------
// 响应式微调
// -----------------------------------------------------------------------------
@media (max-width: 1200px) {
  .security-posture-banner {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
  .summary-band {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
  .policy-controls-row {
    grid-template-columns: 1fr;
    gap: 14px;
  }
}

@media (max-width: 1024px) {
  .provider-grid,
  .security-grid {
    grid-template-columns: 1fr;
  }

  .field-grid--four {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 900px) {
  .settings-shell {
    grid-template-columns: 1fr;
  }

  .settings-nav {
    flex-direction: row;
    overflow-x: auto;
    border-right: 0;
    border-bottom: 1px solid var(--na-border);
    padding: 10px;

    &__header {
      display: none;
    }

    &__item {
      min-width: 180px;
      flex-shrink: 0;

      .nav-arrow {
        display: none;
      }
    }
  }

  .filter-row {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 680px) {
  .settings-content {
    padding: 16px;
  }

  .settings-section-header,
  .group-heading--action,
  .section-toolbar {
    align-items: flex-start;
    flex-direction: column;
  }

  .field-grid,
  .field-grid--four,
  .dialog-grid,
  .filter-row,
  .summary-band,
  .security-posture-banner {
    grid-template-columns: 1fr;
  }

  .field--double {
    grid-column: auto;
  }

  .provider-card__footer {
    flex-direction: column;
    align-items: flex-start;

    .el-button {
      width: 100%;
    }
  }
}

@media (prefers-reduced-motion: reduce) {
  .settings-nav__item,
  .provider-card,
  .kpi-card {
    transition: none;
  }
}
</style>
