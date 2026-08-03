<template>
  <main class="na-page na-page--list system-config-page">
    <AppPageHeader
      title-id="runtime-config-title"
      title="运行配置"
      description="服务、存储、安全与智能识别"
    >
      <template #actions>
        <el-tooltip
          :disabled="!isDirty"
          content="当前配置尚未保存"
          placement="bottom"
        >
          <span class="header-action-wrap">
            <el-button
              :icon="Refresh"
              :loading="reloading"
              :disabled="reloading || saving || isDirty || !configReady"
              @click="reload"
            >
              重载服务
            </el-button>
          </span>
        </el-tooltip>
        <el-button
          type="primary"
          :icon="Check"
          :loading="saving"
          :disabled="reloading || saving || !isDirty || !configReady"
          @click="update"
        >
          保存更改
        </el-button>
      </template>
    </AppPageHeader>

    <section
      v-loading="configLoading"
      class="runtime-summary"
      :class="{ unavailable: !configReady }"
      aria-label="运行配置总览"
    >
      <div class="runtime-summary-lead">
        <span class="signal-dot" :class="{ waiting: !configReady, failed: configLoadError }" aria-hidden="true" />
        <div>
          <span>运行状态</span>
          <strong>{{ configReady ? '配置中心在线' : '正在连接配置中心' }}</strong>
        </div>
      </div>
      <dl class="runtime-summary-facts">
        <div>
          <dt>主数据库</dt>
          <dd>{{ databaseTypeLabel }}</dd>
        </div>
        <div>
          <dt>文件存储</dt>
          <dd>{{ storageTypeLabel }}</dd>
        </div>
        <div>
          <dt>数据服务</dt>
          <dd>{{ dataServiceLabel }}</dd>
        </div>
        <div>
          <dt>识别通道</dt>
          <dd>{{ recognitionProviderCount }} / 4</dd>
        </div>
      </dl>
      <div
        class="config-sync-state"
        :class="{ dirty: isDirty, failed: configLoadError, waiting: !configReady && !configLoadError }"
        role="status"
      >
        <el-icon>
          <WarningFilled v-if="isDirty || configLoadError" />
          <CircleCheckFilled v-else-if="configReady" />
          <Refresh v-else />
        </el-icon>
        <span>{{ configLoadError ? '配置读取失败' : (!configReady ? '正在读取配置' : (isDirty ? '有未保存更改' : '配置已同步')) }}</span>
      </div>
    </section>

    <div class="config-console na-panel">
      <aside class="config-sidebar" aria-label="配置导航">
        <header class="config-sidebar-header">
          <span>配置导航</span>
          <strong>{{ visibleSections.length }} 个配置项</strong>
        </header>
        <nav class="config-sidebar-nav">
          <section
            v-for="group in visibleConfigGroups"
            :key="group.key"
            class="config-nav-group"
          >
            <div class="config-nav-group-title">
              <el-icon><component :is="group.icon" /></el-icon>
              <span>{{ group.label }}</span>
              <small>{{ group.sections.length }}</small>
            </div>
            <div class="config-nav-items">
              <button
                v-for="section in group.sections"
                :key="section.name"
                type="button"
                :disabled="saving || !configReady"
                :class="{ active: activeNames === section.name }"
                :aria-current="activeNames === section.name ? 'page' : undefined"
                @click="activeNames = section.name"
              >
                <span>{{ section.label }}</span>
                <el-icon class="config-nav-arrow" aria-hidden="true"><ArrowRight /></el-icon>
              </button>
            </div>
          </section>
        </nav>
      </aside>

      <div class="config-main">
        <div class="config-mobile-nav">
          <span>配置区域</span>
          <el-select v-model="activeNames" :disabled="saving || !configReady" aria-label="选择配置区域">
            <el-option-group
              v-for="group in visibleConfigGroups"
              :key="group.key"
              :label="group.label"
            >
              <el-option
                v-for="section in group.sections"
                :key="section.name"
                :label="section.label"
                :value="section.name"
              />
            </el-option-group>
          </el-select>
        </div>

        <section class="config-workbench" :aria-labelledby="`config-section-${activeNames}`">
          <header class="config-editor-header">
            <div class="config-editor-heading">
              <span class="config-editor-icon" aria-hidden="true">
                <el-icon><component :is="activeSection.icon" /></el-icon>
              </span>
              <div>
                <span>{{ activeSection.groupLabel }}</span>
                <h2 :id="`config-section-${activeNames}`">{{ activeSection.label }}</h2>
                <p>{{ activeSection.description }}</p>
              </div>
            </div>
          </header>

      <el-form
        v-if="configReady"
        ref="form"
        :model="config"
        :disabled="saving || configLoading"
        label-position="top"
        class="config-form"
      >
      <!--  System start  -->
      <el-tabs
        v-model="activeNames"
        class="config-tabs"
      >
        <el-tab-pane label="基础设置" name="1" lazy>
          <el-form-item label="端口值">
            <el-input-number
              v-model="config.system.addr"
              placeholder="请输入端口值"
            />
          </el-form-item>
          <el-form-item label="数据库类型">
            <el-select v-model="config.system['db-type']" class="w-full">
              <el-option value="mysql" />
              <el-option value="pgsql" />
              <el-option value="mssql" />
              <el-option value="sqlite" />
              <el-option value="oracle" />
            </el-select>
          </el-form-item>
          <el-form-item label="文件存储类型">
            <el-select v-model="config.system['oss-type']" class="w-full">
              <el-option value="local" label="本地" />
              <el-option value="qiniu" label="七牛" />
              <el-option value="tencent-cos" label="腾讯云COS" />
              <el-option value="aliyun-oss" label="阿里云OSS" />
              <el-option value="huawei-obs" label="华为云OBS" />
              <el-option value="cloudflare-r2" label="cloudflare R2" />
              <el-option value="minio">MinIO</el-option>
            </el-select>
          </el-form-item>
          <el-form-item label="多点登录拦截">
            <el-switch v-model="config.system['use-multipoint']" />
          </el-form-item>
          <el-form-item label="启用 Redis">
            <el-switch v-model="config.system['use-redis']" />
          </el-form-item>
          <el-form-item label="启用 MongoDB">
            <el-switch v-model="config.system['use-mongo']" />
          </el-form-item>
          <el-form-item label="严格角色模式">
            <el-switch v-model="config.system['use-strict-auth']" />
          </el-form-item>
          <el-form-item label="限流次数">
            <el-input-number v-model.number="config.system['iplimit-count']" />
          </el-form-item>
          <el-form-item label="限流时间">
            <el-input-number v-model.number="config.system['iplimit-time']" />
          </el-form-item>
          <el-form-item label="禁用自动迁移数据库表结构">
            <el-switch v-model="config.system['disable-auto-migrate']" />
          </el-form-item>
          <el-tooltip
            content="请修改完成后，注意一并修改前端env环境下的VITE_BASE_PATH"
            placement="top-start"
          >
            <el-form-item label="全局路由前缀">
              <el-input
                v-model.trim="config.system['router-prefix']"
                placeholder="请输入全局路由前缀"
              />
            </el-form-item>
          </el-tooltip>
        </el-tab-pane>
        <el-tab-pane label="JWT 签名" name="2" lazy>
          <el-form-item label="JWT 签名密钥">
            <SecretInput
              v-model.trim="config.jwt['signing-key']"
              secret-path="jwt.signing-key"
              :configured="isSecretConfigured('jwt.signing-key')"
              :can-reveal="canManageSystemSecrets"
              placeholder="请输入 JWT 签名密钥"
            >
              <template #append>
                <el-button @click="getUUID">生成</el-button>
              </template>
            </SecretInput>
          </el-form-item>
          <el-form-item label="有效期">
            <el-input
              v-model.trim="config.jwt['expires-time']"
              placeholder="请输入有效期"
            />
          </el-form-item>
          <el-form-item label="缓冲期">
            <el-input
              v-model.trim="config.jwt['buffer-time']"
              placeholder="请输入缓冲期"
            />
          </el-form-item>
          <el-form-item label="签发者">
            <el-input
              v-model.trim="config.jwt.issuer"
              placeholder="请输入签发者"
            />
          </el-form-item>
        </el-tab-pane>
        <el-tab-pane label="运行日志" name="3" lazy>
          <el-form-item label="级别">
            <el-select v-model="config.zap.level">
              <el-option value="off" label="关闭" />
              <el-option value="fatal" label="致命" />
              <el-option value="error" label="错误" />
              <el-option value="warn" label="警告" />
              <el-option value="info" label="信息" />
              <el-option value="debug" label="调试" />
              <el-option value="trace" label="跟踪" />
            </el-select>
          </el-form-item>
          <el-form-item label="输出">
            <el-select v-model="config.zap.format">
              <el-option value="console" label="console" />
              <el-option value="json" label="json" />
            </el-select>
          </el-form-item>
          <el-form-item label="日志前缀">
            <el-input
              v-model.trim="config.zap.prefix"
              placeholder="请输入日志前缀"
            />
          </el-form-item>
          <el-form-item label="日志文件夹">
            <el-input
              v-model.trim="config.zap.director"
              placeholder="请输入日志文件夹"
            />
          </el-form-item>
          <el-form-item label="编码级">
            <el-select v-model="config.zap['encode-level']" class="w-6/12">
              <el-option
                value="LowercaseLevelEncoder"
                label="LowercaseLevelEncoder"
              />
              <el-option
                value="LowercaseColorLevelEncoder"
                label="LowercaseColorLevelEncoder"
              />
              <el-option
                value="CapitalLevelEncoder"
                label="CapitalLevelEncoder"
              />
              <el-option
                value="CapitalColorLevelEncoder"
                label="CapitalColorLevelEncoder"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="栈名">
            <el-input
              v-model.trim="config.zap['stacktrace-key']"
              placeholder="请输入栈名"
            />
          </el-form-item>
          <el-form-item label="日志留存时间(默认以天为单位)">
            <el-input-number v-model="config.zap['retention-day']" />
          </el-form-item>
          <el-form-item label="显示行">
            <el-switch v-model="config.zap['show-line']" />
          </el-form-item>
          <el-form-item label="输出控制台">
            <el-switch v-model="config.zap['log-in-console']" />
          </el-form-item>
        </el-tab-pane>
        <el-tab-pane
          label="Redis"
          name="4"
          lazy
          v-if="config.system['use-redis']"
        >
          <el-form-item label="库">
            <el-input-number v-model="config.redis.db" min="0" max="16" />
          </el-form-item>
          <el-form-item label="地址">
            <el-input
              v-model.trim="config.redis.addr"
              placeholder="请输入地址"
            />
          </el-form-item>
          <el-form-item label="密码">
            <el-input
              v-model.trim="config.redis.password"
              placeholder="请输入密码"
            />
          </el-form-item>
        </el-tab-pane>
        <el-tab-pane label="邮件服务" name="5" lazy>
          <el-form-item label="接收者邮箱">
            <el-input
              v-model="config.email.to"
              placeholder="可多个，以逗号分隔"
            />
          </el-form-item>
          <el-form-item label="端口">
            <el-input-number v-model="config.email.port" />
          </el-form-item>
          <el-form-item label="发送者邮箱">
            <el-input
              v-model.trim="config.email.from"
              placeholder="请输入发送者邮箱"
            />
          </el-form-item>
          <el-form-item label="host">
            <el-input
              v-model.trim="config.email.host"
              placeholder="请输入host"
            />
          </el-form-item>
          <el-form-item label="是否为ssl">
            <el-switch v-model="config.email['is-ssl']" />
          </el-form-item>
          <el-form-item label="是否LoginAuth认证">
            <el-switch v-model="config.email['is-loginauth']" />
          </el-form-item>
          <el-form-item label="secret">
            <SecretInput
              v-model.trim="config.email.secret"
              secret-path="email.secret"
              :configured="isSecretConfigured('email.secret')"
              :can-reveal="canManageSystemSecrets"
              placeholder="请输入邮件服务密钥"
            />
          </el-form-item>
          <el-form-item label="测试邮件">
            <el-tooltip :disabled="!isDirty" content="请先保存当前配置" placement="top">
              <span class="form-action-wrap">
                <el-button
                  :loading="testingEmail"
                  :disabled="isDirty || testingEmail"
                  @click="email"
                >
                  测试邮件
                </el-button>
              </span>
            </el-tooltip>
          </el-form-item>
        </el-tab-pane>
        <el-tab-pane
          label="Mongo 数据库配置"
          name="14"
          lazy
          v-if="config.system['use-mongo']"
        >
          <el-form-item label="collection name(表名,一般不写)">
            <el-input
              v-model.trim="config.mongo.coll"
              placeholder="请输入collection name"
            />
          </el-form-item>
          <el-form-item label="mongodb 选项">
            <el-input
              v-model.trim="config.mongo.options"
              placeholder="请输入mongodb 选项"
            />
          </el-form-item>
          <el-form-item label="database name(数据库名)">
            <el-input
              v-model.trim="config.mongo.database"
              placeholder="请输入数据库名"
            />
          </el-form-item>
          <el-form-item label="用户名">
            <el-input
              v-model.trim="config.mongo.username"
              placeholder="请输入用户名"
            />
          </el-form-item>
          <el-form-item label="密码">
            <el-input
              v-model.trim="config.mongo.password"
              placeholder="请输入密码"
            />
          </el-form-item>
          <el-form-item label="最小连接池">
            <el-input-number v-model="config.mongo['min-pool-size']" min="0" />
          </el-form-item>
          <el-form-item label="最大连接池">
            <el-input-number
              v-model="config.mongo['max-pool-size']"
              min="100"
            />
          </el-form-item>
          <el-form-item label="socket超时时间">
            <el-input-number
              v-model="config.mongo['socket-timeout-ms']"
              min="0"
            />
          </el-form-item>
          <el-form-item label="连接超时时间">
            <el-input-number
              v-model="config.mongo['socket-timeout-ms']"
              min="0"
            />
          </el-form-item>
          <el-form-item label="是否开启zap日志">
            <el-switch v-model="config.mongo['is-zap']" />
          </el-form-item>
          <el-form-item
            v-for="(item, k) in config.mongo.hosts"
            :key="k"
            :label="`节点 ${k + 1}`"
          >
            <div v-for="(_, k2) in item" :key="k2">
              <el-form-item :key="k + k2" :label="k2" label-width="60">
                <el-input
                  v-model.trim="item[k2]"
                  :placeholder="k2 === 'host' ? '请输入地址' : '请输入端口'"
                />
              </el-form-item>
            </div>
            <el-form-item v-if="k > 0">
              <el-button
                type="danger"
                size="small"
                plain
                :icon="Minus"
                @click="removeNode(k)"
                class="ml-3"
              />
            </el-form-item>
          </el-form-item>
          <el-form-item>
            <el-button
              type="primary"
              size="small"
              plain
              :icon="Plus"
              @click="addNode"
            />
          </el-form-item>
        </el-tab-pane>
        <el-tab-pane label="验证码" name="7" lazy>
          <el-form-item label="字符长度">
            <el-input-number
              v-model="config.captcha['key-long']"
              :min="4"
              :max="6"
            />
          </el-form-item>
          <el-form-item label="图片宽度">
            <el-input-number v-model.number="config.captcha['img-width']" />
          </el-form-item>
          <el-form-item label="图片高度">
            <el-input-number v-model.number="config.captcha['img-height']" />
          </el-form-item>
        </el-tab-pane>
        <el-tab-pane label="数据库" name="9" lazy>
          <template v-if="config.system['db-type'] === 'mysql'">
            <el-form-item label="">
              <h3>MySQL</h3>
            </el-form-item>
            <el-form-item label="用户名">
              <el-input
                v-model.trim="config.mysql.username"
                placeholder="请输入用户名"
              />
            </el-form-item>
            <el-form-item label="密码">
              <el-input
                v-model.trim="config.mysql.password"
                placeholder="请输入密码"
              />
            </el-form-item>
            <el-form-item label="地址">
              <el-input
                v-model.trim="config.mysql.path"
                placeholder="请输入地址"
              />
            </el-form-item>
            <el-form-item label="数据库名称">
              <el-input
                v-model.trim="config.mysql['db-name']"
                placeholder="请输入数据库名称"
              />
            </el-form-item>
            <el-form-item label="前缀">
              <el-input
                v-model.trim="config.mysql['prefix']"
                placeholder="默认为空"
              />
            </el-form-item>
            <el-form-item label="复数表">
              <el-switch v-model="config.mysql['singular']" />
            </el-form-item>
            <el-form-item label="引擎">
              <el-input
                v-model.trim="config.mysql['engine']"
                placeholder="默认为InnoDB"
              />
            </el-form-item>
            <el-form-item label="maxIdleConns">
              <el-input-number
                v-model="config.mysql['max-idle-conns']"
                :min="1"
              />
            </el-form-item>
            <el-form-item label="maxOpenConns">
              <el-input-number
                v-model="config.mysql['max-open-conns']"
                :min="1"
              />
            </el-form-item>
            <el-form-item label="写入日志">
              <el-switch v-model="config.mysql['log-zap']" />
            </el-form-item>
            <el-form-item label="日志模式">
              <el-select v-model="config.mysql['log-mode']">
                <el-option value="off" label="关闭" />
                <el-option value="fatal" label="致命" />
                <el-option value="error" label="错误" />
                <el-option value="warn" label="警告" />
                <el-option value="info" label="信息" />
                <el-option value="debug" label="调试" />
                <el-option value="trace" label="跟踪" />
              </el-select>
            </el-form-item>
          </template>
          <template v-if="config.system['db-type'] === 'pgsql'">
            <el-form-item label="">
              <h3>PostgreSQL</h3>
            </el-form-item>
            <el-form-item label="用户名">
              <el-input
                v-model="config.pgsql.username"
                placeholder="请输入用户名"
              />
            </el-form-item>
            <el-form-item label="密码">
              <el-input
                v-model="config.pgsql.password"
                placeholder="请输入密码"
              />
            </el-form-item>
            <el-form-item label="地址">
              <el-input
                v-model.trim="config.pgsql.path"
                placeholder="请输入地址"
              />
            </el-form-item>
            <el-form-item label="数据库">
              <el-input
                v-model.trim="config.pgsql['db-name']"
                placeholder="请输入数据库"
              />
            </el-form-item>
            <el-form-item label="前缀">
              <el-input
                v-model.trim="config.pgsql['prefix']"
                placeholder="请输入前缀"
              />
            </el-form-item>
            <el-form-item label="复数表">
              <el-switch v-model="config.pgsql['singular']" />
            </el-form-item>
            <el-form-item label="引擎">
              <el-input
                v-model.trim="config.pgsql['engine']"
                placeholder="请输入引擎"
              />
            </el-form-item>
            <el-form-item label="maxIdleConns">
              <el-input-number v-model="config.pgsql['max-idle-conns']" />
            </el-form-item>
            <el-form-item label="maxOpenConns">
              <el-input-number v-model="config.pgsql['max-open-conns']" />
            </el-form-item>
            <el-form-item label="写入日志">
              <el-switch v-model="config.pgsql['log-zap']" />
            </el-form-item>
            <el-form-item label="日志模式">
              <el-select v-model="config.pgsql['log-mode']">
                <el-option value="off" label="关闭" />
                <el-option value="fatal" label="致命" />
                <el-option value="error" label="错误" />
                <el-option value="warn" label="警告" />
                <el-option value="info" label="信息" />
                <el-option value="debug" label="调试" />
                <el-option value="trace" label="跟踪" />
              </el-select>
            </el-form-item>
          </template>
          <template v-if="config.system['db-type'] === 'mssql'">
            <el-form-item label="">
              <h3>MsSQL</h3>
            </el-form-item>
            <el-form-item label="用户名">
              <el-input
                v-model.trim="config.mssql.username"
                placeholder="请输入用户名"
              />
            </el-form-item>
            <el-form-item label.trim="密码">
              <el-input
                v-model.trim="config.mssql.password"
                placeholder="请输入密码"
              />
            </el-form-item>
            <el-form-item label="地址">
              <el-input
                v-model.trim="config.mssql.path"
                placeholder="请输入地址"
              />
            </el-form-item>
            <el-form-item label="端口">
              <el-input
                v-model.trim="config.mssql.port"
                placeholder="请输入端口"
              />
            </el-form-item>
            <el-form-item label="数据库">
              <el-input
                v-model.trim="config.mssql['db-name']"
                placeholder="请输入数据库"
              />
            </el-form-item>
            <el-form-item label="前缀">
              <el-input
                v-model.trim="config.mssql['prefix']"
                placeholder="请输入前缀"
              />
            </el-form-item>
            <el-form-item label="复数表">
              <el-switch v-model="config.mssql['singular']" />
            </el-form-item>
            <el-form-item label="引擎">
              <el-input
                v-model.trim="config.mssql['engine']"
                placeholder="请输入引擎"
              />
            </el-form-item>
            <el-form-item label="maxIdleConns">
              <el-input-number v-model="config.mssql['max-idle-conns']" />
            </el-form-item>
            <el-form-item label="maxOpenConns">
              <el-input-number v-model="config.mssql['max-open-conns']" />
            </el-form-item>
            <el-form-item label="写入日志">
              <el-switch v-model="config.mssql['log-zap']" />
            </el-form-item>
            <el-form-item label="日志模式">
              <el-select v-model="config.mssql['log-mode']">
                <el-option value="off" label="关闭" />
                <el-option value="fatal" label="致命" />
                <el-option value="error" label="错误" />
                <el-option value="warn" label="警告" />
                <el-option value="info" label="信息" />
                <el-option value="debug" label="调试" />
                <el-option value="trace" label="跟踪" />
              </el-select>
            </el-form-item>
          </template>
          <template v-if="config.system['db-type'] === 'sqlite'">
            <el-form-item label="">
              <h3>sqlite</h3>
            </el-form-item>
            <el-form-item label="用户名">
              <el-input
                v-model.trim="config.sqlite.username"
                placeholder="请输入用户名"
              />
            </el-form-item>
            <el-form-item label="密码">
              <el-input
                v-model.trim="config.sqlite.password"
                placeholder="请输入密码"
              />
            </el-form-item>
            <el-form-item label="地址">
              <el-input
                v-model.trim="config.sqlite.path"
                placeholder="请输入地址"
              />
            </el-form-item>
            <el-form-item label="端口">
              <el-input
                v-model.trim="config.sqlite.port"
                placeholder="请输入端口"
              />
            </el-form-item>
            <el-form-item label="数据库">
              <el-input
                v-model.trim="config.sqlite['db-name']"
                placeholder="请输入数据库"
              />
            </el-form-item>
            <el-form-item label="maxIdleConns">
              <el-input-number v-model="config.sqlite['max-idle-conns']" />
            </el-form-item>
            <el-form-item label="maxOpenConns">
              <el-input-number v-model="config.sqlite['max-open-conns']" />
            </el-form-item>
            <el-form-item label="写入日志">
              <el-switch v-model="config.sqlite['log-zap']" />
            </el-form-item>
            <el-form-item label="日志模式">
              <el-select v-model="config.sqlite['log-mode']">
                <el-option value="off" label="关闭" />
                <el-option value="fatal" label="致命" />
                <el-option value="error" label="错误" />
                <el-option value="warn" label="警告" />
                <el-option value="info" label="信息" />
                <el-option value="debug" label="调试" />
                <el-option value="trace" label="跟踪" />
              </el-select>
            </el-form-item>
          </template>
          <template v-if="config.system['db-type'] === 'oracle'">
            <el-form-item label="">
              <h3>oracle</h3>
            </el-form-item>
            <el-form-item label="用户名">
              <el-input
                v-model.trim="config.oracle.username"
                placeholder="请输入用户名"
              />
            </el-form-item>
            <el-form-item label="密码">
              <el-input
                v-model.trim="config.oracle.password"
                placeholder="请输入密码"
              />
            </el-form-item>
            <el-form-item label="地址">
              <el-input
                v-model.trim="config.oracle.path"
                placeholder="请输入地址"
              />
            </el-form-item>
            <el-form-item label="数据库名称">
              <el-input
                v-model.trim="config.oracle['db-name']"
                placeholder="请输入数据库名称"
              />
            </el-form-item>
            <el-form-item label="前缀">
              <el-input
                v-model.trim="config.oracle['prefix']"
                placeholder="默认为空"
              />
            </el-form-item>
            <el-form-item label="复数表">
              <el-switch v-model="config.oracle['singular']" />
            </el-form-item>
            <el-form-item label="引擎">
              <el-input
                v-model.trim="config.oracle['engine']"
                placeholder="默认为InnoDB"
              />
            </el-form-item>
            <el-form-item label="maxIdleConns">
              <el-input-number
                v-model="config.oracle['max-idle-conns']"
                :min="1"
              />
            </el-form-item>
            <el-form-item label="maxOpenConns">
              <el-input-number
                v-model="config.oracle['max-open-conns']"
                :min="1"
              />
            </el-form-item>
            <el-form-item label="写入日志">
              <el-switch v-model="config.oracle['log-zap']" />
            </el-form-item>
            <el-form-item label="日志模式">
              <el-select v-model="config.oracle['log-mode']">
                <el-option value="off" label="关闭" />
                <el-option value="fatal" label="致命" />
                <el-option value="error" label="错误" />
                <el-option value="warn" label="警告" />
                <el-option value="info" label="信息" />
                <el-option value="debug" label="调试" />
                <el-option value="trace" label="跟踪" />
              </el-select>
            </el-form-item>
          </template>
        </el-tab-pane>
        <el-tab-pane label="文件存储" name="10" lazy>
          <template v-if="config.system['oss-type'] === 'local'">
            <h2>本地配置</h2>
            <el-form-item label="本地文件访问路径">
              <el-input
                v-model.trim="config.local.path"
                placeholder="请输入本地文件访问路径"
              />
            </el-form-item>
            <el-form-item label="本地文件存储路径">
              <el-input
                v-model.trim="config.local['store-path']"
                placeholder="请输入本地文件存储路径"
              />
            </el-form-item>
          </template>
          <template v-if="config.system['oss-type'] === 'qiniu'">
            <h2>七牛上传配置</h2>
            <el-form-item label="存储区域">
              <el-input
                v-model.trim="config.qiniu.zone"
                placeholder="请输入存储区域"
              />
            </el-form-item>
            <el-form-item label="空间名称">
              <el-input
                v-model.trim="config.qiniu.bucket"
                placeholder="请输入空间名称"
              />
            </el-form-item>
            <el-form-item label="CDN加速域名">
              <el-input
                v-model.trim="config.qiniu['img-path']"
                placeholder="请输入CDN加速域名"
              />
            </el-form-item>
            <el-form-item label="是否使用https">
              <el-switch v-model="config.qiniu['use-https']">开启</el-switch>
            </el-form-item>
            <el-form-item label="accessKey">
              <SecretInput
                v-model.trim="config.qiniu['access-key']"
                secret-path="qiniu.access-key"
                :configured="isSecretConfigured('qiniu.access-key')"
                :can-reveal="canManageSystemSecrets"
                placeholder="请输入 Access Key"
              />
            </el-form-item>
            <el-form-item label="secretKey">
              <SecretInput
                v-model.trim="config.qiniu['secret-key']"
                secret-path="qiniu.secret-key"
                :configured="isSecretConfigured('qiniu.secret-key')"
                :can-reveal="canManageSystemSecrets"
                placeholder="请输入 Secret Key"
              />
            </el-form-item>
            <el-form-item label="上传是否使用CDN上传加速">
              <el-switch v-model="config.qiniu['use-cdn-domains']" />
            </el-form-item>
          </template>
          <template v-if="config.system['oss-type'] === 'tencent-cos'">
            <h2>腾讯云COS上传配置</h2>
            <el-form-item label="存储桶名称">
              <el-input
                v-model.trim="config['tencent-cos']['bucket']"
                placeholder="请输入存储桶名称"
              />
            </el-form-item>
            <el-form-item label="所属地域">
              <el-input
                v-model.trim="config['tencent-cos'].region"
                placeholder="请输入所属地域"
              />
            </el-form-item>
            <el-form-item label="secretID">
              <SecretInput
                v-model.trim="config['tencent-cos']['secret-id']"
                secret-path="tencent-cos.secret-id"
                :configured="isSecretConfigured('tencent-cos.secret-id')"
                :can-reveal="canManageSystemSecrets"
                placeholder="请输入 Secret ID"
              />
            </el-form-item>
            <el-form-item label="secretKey">
              <SecretInput
                v-model.trim="config['tencent-cos']['secret-key']"
                secret-path="tencent-cos.secret-key"
                :configured="isSecretConfigured('tencent-cos.secret-key')"
                :can-reveal="canManageSystemSecrets"
                placeholder="请输入 Secret Key"
              />
            </el-form-item>
            <el-form-item label="路径前缀">
              <el-input
                v-model.trim="config['tencent-cos']['path-prefix']"
                placeholder="请输入路径前缀"
              />
            </el-form-item>
            <el-form-item label="访问域名">
              <el-input
                v-model.trim="config['tencent-cos']['base-url']"
                placeholder="请输入访问域名"
              />
            </el-form-item>
          </template>
          <template v-if="config.system['oss-type'] === 'aliyun-oss'">
            <h2>阿里云OSS上传配置</h2>
            <el-form-item label="区域">
              <el-input
                v-model.trim="config['aliyun-oss'].endpoint"
                placeholder="请输入区域"
              />
            </el-form-item>
            <el-form-item label="accessKeyId">
              <SecretInput
                v-model.trim="config['aliyun-oss']['access-key-id']"
                secret-path="aliyun-oss.access-key-id"
                :configured="isSecretConfigured('aliyun-oss.access-key-id')"
                :can-reveal="canManageSystemSecrets"
                placeholder="请输入 Access Key ID"
              />
            </el-form-item>
            <el-form-item label="accessKeySecret">
              <SecretInput
                v-model.trim="config['aliyun-oss']['access-key-secret']"
                secret-path="aliyun-oss.access-key-secret"
                :configured="isSecretConfigured('aliyun-oss.access-key-secret')"
                :can-reveal="canManageSystemSecrets"
                placeholder="请输入 Access Key Secret"
              />
            </el-form-item>
            <el-form-item label="存储桶名称">
              <el-input
                v-model.trim="config['aliyun-oss']['bucket-name']"
                placeholder="请输入存储桶名称"
              />
            </el-form-item>
            <el-form-item label="访问域名">
              <el-input
                v-model.trim="config['aliyun-oss']['bucket-url']"
                placeholder="请输入访问域名"
              />
            </el-form-item>
          </template>
          <template v-if="config.system['oss-type'] === 'huawei-obs'">
            <h2>华为云OBS上传配置</h2>
            <el-form-item label="路径">
              <el-input
                v-model.trim="config['hua-wei-obs'].path"
                placeholder="请输入路径"
              />
            </el-form-item>
            <el-form-item label="存储桶名称">
              <el-input
                v-model.trim="config['hua-wei-obs'].bucket"
                placeholder="请输入存储桶名称"
              />
            </el-form-item>
            <el-form-item label="区域">
              <el-input
                v-model.trim="config['hua-wei-obs'].endpoint"
                placeholder="请输入区域"
              />
            </el-form-item>
            <el-form-item label="accessKey">
              <SecretInput
                v-model.trim="config['hua-wei-obs']['access-key']"
                secret-path="hua-wei-obs.access-key"
                :configured="isSecretConfigured('hua-wei-obs.access-key')"
                :can-reveal="canManageSystemSecrets"
                placeholder="请输入 Access Key"
              />
            </el-form-item>
            <el-form-item label="secretKey">
              <SecretInput
                v-model.trim="config['hua-wei-obs']['secret-key']"
                secret-path="hua-wei-obs.secret-key"
                :configured="isSecretConfigured('hua-wei-obs.secret-key')"
                :can-reveal="canManageSystemSecrets"
                placeholder="请输入 Secret Key"
              />
            </el-form-item>
          </template>
          <template v-if="config.system['oss-type'] === 'cloudflare-r2'">
            <h2>Cloudflare R2上传配置</h2>
            <el-form-item label="路径">
              <el-input
                v-model.trim="config['cloudflare-r2'].path"
                placeholder="请输入路径"
              />
            </el-form-item>
            <el-form-item label="存储桶名称">
              <el-input
                v-model.trim="config['cloudflare-r2'].bucket"
                placeholder="请输入存储桶名称"
              />
            </el-form-item>
            <el-form-item label="Base URL">
              <el-input
                v-model.trim="config['cloudflare-r2']['base-url']"
                placeholder="请输入Base URL"
              />
            </el-form-item>
            <el-form-item label="Account ID">
              <el-input
                v-model.trim="config['cloudflare-r2']['account-id']"
                placeholder="请输入secretKey"
              />
            </el-form-item>
            <el-form-item label="Access Key ID">
              <SecretInput
                v-model.trim="config['cloudflare-r2']['access-key-id']"
                secret-path="cloudflare-r2.access-key-id"
                :configured="isSecretConfigured('cloudflare-r2.access-key-id')"
                :can-reveal="canManageSystemSecrets"
                placeholder="请输入 Access Key ID"
              />
            </el-form-item>
            <el-form-item label="Secret Access Key">
              <SecretInput
                v-model.trim="config['cloudflare-r2']['secret-access-key']"
                secret-path="cloudflare-r2.secret-access-key"
                :configured="isSecretConfigured('cloudflare-r2.secret-access-key')"
                :can-reveal="canManageSystemSecrets"
                placeholder="请输入 Secret Access Key"
              />
            </el-form-item>
          </template>
          <template v-if="config.system['oss-type'] === 'minio'">
            <h2>MinIO上传配置</h2>
            <el-form-item label="Endpoint">
              <el-input
                v-model.trim="config.minio.endpoint"
                placeholder="请输入Endpoint，如 127.0.0.1:9000"
              />
            </el-form-item>
            <el-form-item label="Access Key ID">
              <SecretInput
                v-model.trim="config.minio['access-key-id']"
                secret-path="minio.access-key-id"
                :configured="isSecretConfigured('minio.access-key-id')"
                :can-reveal="canManageSystemSecrets"
                placeholder="请输入Access Key ID"
              />
            </el-form-item>
            <el-form-item label="Access Key Secret">
              <SecretInput
                v-model.trim="config.minio['access-key-secret']"
                secret-path="minio.access-key-secret"
                :configured="isSecretConfigured('minio.access-key-secret')"
                :can-reveal="canManageSystemSecrets"
                placeholder="请输入Access Key Secret"
              />
            </el-form-item>
            <el-form-item label="存储桶名称">
              <el-input
                v-model.trim="config.minio['bucket-name']"
                placeholder="请输入存储桶名称"
              />
            </el-form-item>
            <el-form-item label="访问域名">
              <el-input
                v-model.trim="config.minio['bucket-url']"
                placeholder="请输入访问域名"
              />
            </el-form-item>
            <el-form-item label="Base Path">
              <el-input
                v-model.trim="config.minio['base-path']"
                placeholder="请输入Base Path"
              />
            </el-form-item>
            <el-form-item label="开启SSL">
              <el-switch v-model="config.minio['use-ssl']" />
            </el-form-item>
          </template>
        </el-tab-pane>
        <el-tab-pane label="发票识别" name="11" lazy>
          <div class="recognition-settings">
            <div class="recognition-overview">
              <div class="recognition-flow" aria-label="发票识别顺序">
                <span class="flow-node is-fixed">二维码</span>
                <span class="flow-step">
                  <span class="flow-arrow">→</span>
                  <span class="flow-node">百度 / 公网 OCR</span>
                </span>
                <span class="flow-step">
                  <span class="flow-arrow">→</span>
                  <span class="flow-node">多模态模型</span>
                </span>
                <span class="flow-step">
                  <span class="flow-arrow">→</span>
                  <span class="flow-node is-fixed">人工核对</span>
                </span>
              </div>
              <div class="recognition-controls">
                <el-form-item label="允许内网端点" label-width="112px" class="threshold-field">
                  <el-switch
                    v-model="config['invoice-recognition']['allow-private-endpoints']"
                    :disabled="!canManageInvoiceRecognition"
                    inline-prompt
                    active-text="允许"
                    inactive-text="拒绝"
                  />
                </el-form-item>
                <el-form-item label="大模型兜底阈值" label-width="132px" class="threshold-field">
                  <el-input-number
                    v-model="config['invoice-recognition']['fallback-threshold']"
                    :min="0.1"
                    :max="1"
                    :step="0.01"
                    :precision="2"
                    :disabled="!canManageInvoiceRecognition"
                    controls-position="right"
                  />
                </el-form-item>
              </div>
            </div>

            <section class="provider-section">
              <div class="provider-heading">
                <div class="provider-title-line">
                  <h3>百度发票 OCR</h3>
                  <el-tag :type="baiduCredentialsReady ? 'success' : 'info'" effect="plain" size="small">
                    {{ baiduCredentialsReady ? '凭据已配置' : '无凭据' }}
                  </el-tag>
                </div>
                <div class="provider-heading-actions">
                  <el-button
                    v-if="canManageInvoiceRecognition"
                    type="primary"
                    plain
                    size="small"
                    :icon="Connection"
                    :loading="testingProvider === 'baidu'"
                    @click="testProvider('baidu')"
                  >
                    验证凭据
                  </el-button>
                  <el-switch
                    v-model="config['invoice-recognition'].baidu.enabled"
                    :disabled="!canManageInvoiceRecognition"
                    inline-prompt
                    active-text="启用"
                    inactive-text="停用"
                  />
                </div>
              </div>
              <div class="provider-grid">
                <el-form-item label="请求超时" label-width="112px">
                  <el-input-number
                    v-model="config['invoice-recognition'].baidu['timeout-seconds']"
                    :min="1"
                    :max="120"
                    :disabled="!canManageInvoiceRecognition"
                    controls-position="right"
                  />
                  <span class="input-unit">秒</span>
                </el-form-item>
                <el-form-item label="API Key" label-width="112px">
                  <div class="secret-row">
                    <SecretInput
                      v-model.trim="config['invoice-recognition'].baidu['api-key']"
                      secret-path="invoice-recognition.baidu.api-key"
                      :configured="isSecretConfigured('invoice-recognition.baidu.api-key')"
                      :can-reveal="canManageSystemSecrets"
                      :disabled="!canManageInvoiceRecognition || config['invoice-recognition'].baidu['clear-api-key']"
                      placeholder="请输入百度智能云 API Key"
                    />
                    <el-checkbox
                      v-if="config['invoice-recognition'].baidu['api-key-configured']"
                      v-model="config['invoice-recognition'].baidu['clear-api-key']"
                      :disabled="!canManageInvoiceRecognition"
                    >清除</el-checkbox>
                  </div>
                </el-form-item>
                <el-form-item label="Secret Key" label-width="112px">
                  <div class="secret-row">
                    <SecretInput
                      v-model.trim="config['invoice-recognition'].baidu['secret-key']"
                      secret-path="invoice-recognition.baidu.secret-key"
                      :configured="isSecretConfigured('invoice-recognition.baidu.secret-key')"
                      :can-reveal="canManageSystemSecrets"
                      :disabled="!canManageInvoiceRecognition || config['invoice-recognition'].baidu['clear-secret-key']"
                      placeholder="请输入百度智能云 Secret Key"
                    />
                    <el-checkbox
                      v-if="config['invoice-recognition'].baidu['secret-key-configured']"
                      v-model="config['invoice-recognition'].baidu['clear-secret-key']"
                      :disabled="!canManageInvoiceRecognition"
                    >清除</el-checkbox>
                  </div>
                </el-form-item>
              </div>
            </section>

            <section class="provider-section">
              <div class="provider-heading">
                <div class="provider-title-line">
                  <h3>公网 OCR</h3>
                  <el-tag
                    :type="publicOCRKeyReady ? 'success' : 'info'"
                    effect="plain"
                    size="small"
                  >
                    {{ publicOCRKeyReady ? '凭据已配置' : '无凭据' }}
                  </el-tag>
                  <el-tag
                    v-if="publicOCRDetectionLabel"
                    type="info"
                    effect="plain"
                    size="small"
                  >
                    {{ publicOCRDetectionLabel }}
                  </el-tag>
                </div>
                <div class="provider-heading-actions">
                  <el-button
                    v-if="canManageInvoiceRecognition"
                    type="primary"
                    plain
                    size="small"
                    :icon="Connection"
                    :loading="testingProvider === 'public-ocr'"
                    @click="testProvider('public-ocr')"
                  >
                    测试并识别协议
                  </el-button>
                  <el-switch
                    v-model="config['invoice-recognition']['public-ocr'].enabled"
                    :disabled="!canManageInvoiceRecognition"
                    inline-prompt
                    active-text="启用"
                    inactive-text="停用"
                  />
                </div>
              </div>
              <div class="provider-grid">
                <el-form-item label="请求超时" label-width="112px">
                  <el-input-number
                    v-model="config['invoice-recognition']['public-ocr']['timeout-seconds']"
                    :min="1"
                    :max="120"
                    :disabled="!canManageInvoiceRecognition"
                    controls-position="right"
                  />
                  <span class="input-unit">秒</span>
                </el-form-item>
                <el-form-item label="接口地址" label-width="112px" class="grid-full">
                  <el-input
                    v-model.trim="config['invoice-recognition']['public-ocr'].endpoint"
                    :disabled="!canManageInvoiceRecognition"
                    placeholder="https://ocr.example.com/invoice/recognize"
                    @input="resetPublicOCRDetection"
                  />
                </el-form-item>
                <el-form-item label="API Key" label-width="112px" class="grid-full">
                  <div class="secret-row">
                    <SecretInput
                      v-model.trim="config['invoice-recognition']['public-ocr']['api-key']"
                      secret-path="invoice-recognition.public-ocr.api-key"
                      :configured="isSecretConfigured('invoice-recognition.public-ocr.api-key')"
                      :can-reveal="canManageSystemSecrets"
                      :disabled="!canManageInvoiceRecognition || config['invoice-recognition']['public-ocr']['clear-api-key']"
                      placeholder="请输入 API Key（可选）"
                      @input="resetPublicOCRDetection"
                    />
                    <el-checkbox
                      v-if="config['invoice-recognition']['public-ocr']['api-key-configured']"
                      v-model="config['invoice-recognition']['public-ocr']['clear-api-key']"
                      :disabled="!canManageInvoiceRecognition"
                      @change="resetPublicOCRDetection"
                    >
                      清除凭据
                    </el-checkbox>
                  </div>
                </el-form-item>
              </div>
            </section>

            <section class="provider-section">
              <div class="provider-heading">
                <div class="provider-title-line">
                  <h3>权威发票验真</h3>
                  <el-tag :type="verificationCredentialsReady ? 'success' : 'info'" effect="plain" size="small">
                    {{ verificationCredentialsReady ? '凭据已配置' : '无凭据' }}
                  </el-tag>
                  <el-tag
                    v-if="verificationDetectionLabel"
                    type="info"
                    effect="plain"
                    size="small"
                  >
                    {{ verificationDetectionLabel }}
                  </el-tag>
                </div>
                <div class="provider-heading-actions">
                  <el-button
                    v-if="canManageInvoiceRecognition"
                    type="primary"
                    plain
                    size="small"
                    :icon="Connection"
                    :loading="testingProvider === 'verification'"
                    @click="testProvider('verification')"
                  >
                    测试并识别供应商
                  </el-button>
                  <el-switch
                    v-model="config['invoice-recognition'].verification.enabled"
                    :disabled="!canManageInvoiceRecognition"
                    inline-prompt
                    active-text="启用"
                    inactive-text="停用"
                  />
                </div>
              </div>
              <p class="provider-hint">
                全局停用后不调用付费验真接口，发票按人工核对结果确认；启用后必须取得权威查验结果。供应商与协议由服务器自动探测。
              </p>
              <div class="provider-grid">
                <el-form-item label="请求超时" label-width="112px">
                  <el-input-number
                    v-model="config['invoice-recognition'].verification['timeout-seconds']"
                    :min="1"
                    :max="120"
                    :disabled="!canManageInvoiceRecognition"
                    controls-position="right"
                  />
                  <span class="input-unit">秒</span>
                </el-form-item>
                <el-form-item label="接口地址" label-width="112px" class="grid-full">
                  <el-input
                    v-model.trim="config['invoice-recognition'].verification.endpoint"
                    :disabled="!canManageInvoiceRecognition"
                    placeholder="百度请留空；其他厂商填写验真网关地址"
                    @input="resetVerificationDetection"
                  />
                </el-form-item>
                <el-form-item label="API Key" label-width="112px">
                  <div class="secret-row">
                    <SecretInput
                      v-model.trim="config['invoice-recognition'].verification['api-key']"
                      secret-path="invoice-recognition.verification.api-key"
                      :configured="isSecretConfigured('invoice-recognition.verification.api-key')"
                      :can-reveal="canManageSystemSecrets"
                      :disabled="!canManageInvoiceRecognition || config['invoice-recognition'].verification['clear-api-key']"
                      placeholder="请输入验真服务 API Key"
                      @input="resetVerificationDetection"
                    />
                    <el-checkbox
                      v-if="config['invoice-recognition'].verification['api-key-configured']"
                      v-model="config['invoice-recognition'].verification['clear-api-key']"
                      :disabled="!canManageInvoiceRecognition"
                      @change="resetVerificationDetection"
                    >清除</el-checkbox>
                  </div>
                </el-form-item>
                <el-form-item label="Secret Key" label-width="112px">
                  <div class="secret-row">
                    <SecretInput
                      v-model.trim="config['invoice-recognition'].verification['secret-key']"
                      secret-path="invoice-recognition.verification.secret-key"
                      :configured="isSecretConfigured('invoice-recognition.verification.secret-key')"
                      :can-reveal="canManageSystemSecrets"
                      :disabled="!canManageInvoiceRecognition || config['invoice-recognition'].verification['clear-secret-key']"
                      placeholder="请输入 Secret Key（HTTP 网关可选）"
                      @input="resetVerificationDetection"
                    />
                    <el-checkbox
                      v-if="config['invoice-recognition'].verification['secret-key-configured']"
                      v-model="config['invoice-recognition'].verification['clear-secret-key']"
                      :disabled="!canManageInvoiceRecognition"
                      @change="resetVerificationDetection"
                    >清除</el-checkbox>
                  </div>
                </el-form-item>
              </div>
            </section>

            <section class="provider-section">
              <div class="provider-heading">
                <div class="provider-title-line">
                  <h3>多模态大模型</h3>
                  <el-tag
                    :type="multimodalKeyReady ? 'success' : 'info'"
                    effect="plain"
                    size="small"
                  >
                    {{ multimodalKeyReady ? '凭据已配置' : '无凭据' }}
                  </el-tag>
                  <el-tag
                    v-if="multimodalProtocolLabel"
                    type="info"
                    effect="plain"
                    size="small"
                  >
                    {{ multimodalProtocolLabel }}
                  </el-tag>
                </div>
                <div class="provider-heading-actions">
                  <el-button
                    v-if="canManageInvoiceRecognition"
                    type="primary"
                    plain
                    size="small"
                    :icon="Connection"
                    :loading="testingProvider === 'multimodal'"
                    @click="testProvider('multimodal')"
                  >
                    测试连接
                  </el-button>
                  <el-switch
                    v-model="config['invoice-recognition'].multimodal.enabled"
                    :disabled="!canManageInvoiceRecognition"
                    inline-prompt
                    active-text="启用"
                    inactive-text="停用"
                  />
                </div>
              </div>
              <div class="provider-grid">
                <el-form-item label="Base URL" label-width="112px">
                  <el-input
                    v-model.trim="config['invoice-recognition'].multimodal['base-url']"
                    :disabled="!canManageInvoiceRecognition"
                    placeholder="https://api.example.com/v1"
                    @input="resetMultimodalProtocol"
                  />
                </el-form-item>
                <el-form-item label="模型" label-width="112px">
                  <el-input
                    v-model.trim="config['invoice-recognition'].multimodal.model"
                    :disabled="!canManageInvoiceRecognition"
                    placeholder="请输入支持图片的模型名称"
                    @input="resetMultimodalProtocol"
                  />
                </el-form-item>
                <el-form-item label="请求超时" label-width="112px">
                  <el-input-number
                    v-model="config['invoice-recognition'].multimodal['timeout-seconds']"
                    :min="1"
                    :max="120"
                    :disabled="!canManageInvoiceRecognition"
                    controls-position="right"
                  />
                  <span class="input-unit">秒</span>
                </el-form-item>
                <el-form-item label="API Key" label-width="112px">
                  <div class="secret-row">
                    <SecretInput
                      v-model.trim="config['invoice-recognition'].multimodal['api-key']"
                      secret-path="invoice-recognition.multimodal.api-key"
                      :configured="isSecretConfigured('invoice-recognition.multimodal.api-key')"
                      :can-reveal="canManageSystemSecrets"
                      :disabled="!canManageInvoiceRecognition || config['invoice-recognition'].multimodal['clear-api-key']"
                      placeholder="请输入 API Key（可选）"
                      @input="resetMultimodalProtocol"
                    />
                    <el-checkbox
                      v-if="config['invoice-recognition'].multimodal['api-key-configured']"
                      v-model="config['invoice-recognition'].multimodal['clear-api-key']"
                      :disabled="!canManageInvoiceRecognition"
                      @change="resetMultimodalProtocol"
                    >
                      清除凭据
                    </el-checkbox>
                  </div>
                </el-form-item>
              </div>
            </section>
          </div>
        </el-tab-pane>
      </el-tabs>
      </el-form>
      <div v-else-if="configLoadError" class="config-error-state" role="alert">
        <el-icon><WarningFilled /></el-icon>
        <strong>运行配置读取失败</strong>
        <el-button :icon="Refresh" @click="initForm">重新加载</el-button>
      </div>
      <el-skeleton v-else class="config-loading-state" :rows="8" animated />
        </section>
      </div>
    </div>
  </main>
</template>

<script setup>
  import {
    getSystemConfig,
    reloadSystem,
    setSystemConfig,
    testInvoiceRecognitionProvider
  } from '@/api/system'
  import { computed, ref, watch } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import {
    ArrowRight,
    Check,
    CircleCheckFilled,
    Coin,
    Connection,
    DataBoard,
    Document,
    Key,
    Message,
    Minus,
    Plus,
    Refresh,
    Setting,
    Tickets,
    UploadFilled,
    View,
    WarningFilled
  } from '@element-plus/icons-vue'
  import { emailTest } from '@/api/email'
  import { CreateUUID } from '@/utils/format'
  import { useUserStore } from '@/pinia/modules/user'
  import SecretInput from '@/components/secretInput/index.vue'

  defineOptions({
    name: 'Config'
  })

  const activeNames = ref('1')
  const userStore = useUserStore()
  const canManageSystemSecrets = computed(
    () => Number(userStore.userInfo.authorityId) === 888
  )
  const canManageInvoiceRecognition = canManageSystemSecrets
  const testingProvider = ref('')
  const testingEmail = ref(false)
  const reloading = ref(false)
  const saving = ref(false)
  const configLoading = ref(false)
  const configReady = ref(false)
  const configLoadError = ref(false)
  const savedSnapshot = ref('')
  const configuredSecrets = ref({})

  const defaultInvoiceRecognition = () => ({
    'fallback-threshold': 0.82,
    'allow-private-endpoints': false,
    baidu: {
      enabled: false,
      'verification-enabled': false,
      'api-key': '',
      'secret-key': '',
      'api-key-configured': false,
      'secret-key-configured': false,
      'clear-api-key': false,
      'clear-secret-key': false,
      'timeout-seconds': 30
    },
    'public-ocr': {
      enabled: false,
      provider: '',
      protocol: '',
      endpoint: '',
      'api-key': '',
      'api-key-configured': false,
      'clear-api-key': false,
      'timeout-seconds': 30
    },
    verification: {
      enabled: false,
      provider: '',
      protocol: '',
      endpoint: '',
      'api-key': '',
      'secret-key': '',
      'api-key-configured': false,
      'secret-key-configured': false,
      'clear-api-key': false,
      'clear-secret-key': false,
      'timeout-seconds': 30
    },
    multimodal: {
      enabled: false,
      'base-url': '',
      'api-key': '',
      'api-key-configured': false,
      'clear-api-key': false,
      model: '',
      protocol: '',
      'timeout-seconds': 45
    }
  })

  const config = ref({
    system: {
      'iplimit-count': 0,
      'iplimit-time': 0
    },
    jwt: {},
    mysql: {},
    mssql: {},
    sqlite: {},
    pgsql: {},
    oracle: {},
    excel: {},
    autocode: {},
    redis: {},
    mongo: {
      coll: '',
      options: '',
      database: '',
      username: '',
      password: '',
      'min-pool-size': '',
      'max-pool-size': '',
      'socket-timeout-ms': '',
      'connect-timeout-ms': '',
      'is-zap': false,
      hosts: [
        {
          host: '',
          port: ''
        }
      ]
    },
    qiniu: {},
    'tencent-cos': {},
    'aliyun-oss': {},
    'hua-wei-obs': {},
    'cloudflare-r2': {},
    minio: {},
    captcha: {},
    zap: {},
    local: {},
    email: {},
    timer: {
      detail: {}
    },
    'invoice-recognition': defaultInvoiceRecognition()
  })

  const configGroups = [
    {
      key: 'runtime',
      label: '核心运行',
      icon: Setting,
      sections: [
        { name: '1', label: '基础设置', description: '服务端口、存储类型与基础开关', icon: Setting },
        { name: '3', label: '运行日志', description: '日志级别、输出方式与保留策略', icon: Document }
      ]
    },
    {
      key: 'security',
      label: '安全认证',
      icon: Key,
      sections: [
        { name: '2', label: 'JWT 签名', description: '签名密钥、令牌周期与签发身份', icon: Key },
        { name: '7', label: '验证码', description: '验证码尺寸与字符规则', icon: View }
      ]
    },
    {
      key: 'data',
      label: '数据服务',
      icon: DataBoard,
      sections: [
        { name: '9', label: '主数据库', description: '数据库连接、连接池与日志策略', icon: Coin },
        { name: '4', label: 'Redis', description: '缓存连接与数据库选择', icon: Connection, enabled: () => config.value.system['use-redis'] },
        { name: '14', label: 'MongoDB', description: '文档数据库集群与连接参数', icon: DataBoard, enabled: () => config.value.system['use-mongo'] }
      ]
    },
    {
      key: 'integration',
      label: '外部集成',
      icon: Connection,
      sections: [
        { name: '10', label: '文件存储', description: '本地与对象存储服务连接', icon: UploadFilled },
        { name: '5', label: '邮件服务', description: 'SMTP 连接与通知账户', icon: Message }
      ]
    },
    {
      key: 'intelligence',
      label: '智能识别',
      icon: Tickets,
      sections: [
        { name: '11', label: '发票识别', description: 'OCR、权威查验与多模态模型', icon: Tickets }
      ]
    }
  ]

  const visibleConfigGroups = computed(() => configGroups.map((group) => ({
    ...group,
    sections: group.sections.filter((section) => !section.enabled || section.enabled())
  })).filter((group) => group.sections.length))
  const visibleSections = computed(() => visibleConfigGroups.value.flatMap((group) =>
    group.sections.map((section) => ({ ...section, groupLabel: group.label }))
  ))
  const activeSection = computed(() =>
    visibleSections.value.find((section) => section.name === activeNames.value) || visibleSections.value[0]
  )
  const serializeConfig = (value) => JSON.stringify(value)
  const isSecretConfigured = (path) => Boolean(configuredSecrets.value[path])
  const isDirty = computed(() => Boolean(savedSnapshot.value) &&
    serializeConfig(config.value) !== savedSnapshot.value)
  const databaseTypeLabel = computed(() => ({
    mysql: 'MySQL',
    pgsql: 'PostgreSQL',
    mssql: 'SQL Server',
    sqlite: 'SQLite',
    oracle: 'Oracle'
  })[config.value.system['db-type']] || '未设置')
  const storageTypeLabel = computed(() => ({
    local: '本地存储',
    qiniu: '七牛云',
    'tencent-cos': '腾讯云 COS',
    'aliyun-oss': '阿里云 OSS',
    'huawei-obs': '华为云 OBS',
    'cloudflare-r2': 'Cloudflare R2',
    minio: 'MinIO'
  })[config.value.system['oss-type']] || '未设置')
  const dataServiceLabel = computed(() => {
    const services = []
    if (config.value.system['use-redis']) services.push('Redis')
    if (config.value.system['use-mongo']) services.push('MongoDB')
    return services.length ? services.join(' + ') : '未启用'
  })
  const recognitionProviderCount = computed(() => {
    const recognition = config.value['invoice-recognition']
    return [
      recognition.baidu.enabled,
      recognition['public-ocr'].enabled,
      recognition.verification.enabled,
      recognition.multimodal.enabled
    ].filter(Boolean).length
  })

  watch(visibleSections, (sections) => {
    if (!sections.some((section) => section.name === activeNames.value)) {
      activeNames.value = '1'
    }
  })

  const normalizeInvoiceRecognition = () => {
    const defaults = defaultInvoiceRecognition()
    const current = config.value['invoice-recognition'] || {}
    config.value['invoice-recognition'] = {
      ...defaults,
      ...current,
      baidu: { ...defaults.baidu, ...(current.baidu || {}) },
      'public-ocr': { ...defaults['public-ocr'], ...(current['public-ocr'] || {}) },
      verification: { ...defaults.verification, ...(current.verification || {}) },
      multimodal: { ...defaults.multimodal, ...(current.multimodal || {}) }
    }
  }

  const baiduCredentialsReady = computed(() => {
    const provider = config.value['invoice-recognition'].baidu
    const apiKeyReady = !provider['clear-api-key'] && (provider['api-key-configured'] || Boolean(provider['api-key']))
    const secretKeyReady = !provider['clear-secret-key'] && (provider['secret-key-configured'] || Boolean(provider['secret-key']))
    return apiKeyReady && secretKeyReady
  })

  const publicOCRKeyReady = computed(() => {
    const provider = config.value['invoice-recognition']['public-ocr']
    return !provider['clear-api-key'] &&
      (provider['api-key-configured'] || Boolean(provider['api-key']))
  })
  const multimodalKeyReady = computed(() => {
    const provider = config.value['invoice-recognition'].multimodal
    return !provider['clear-api-key'] &&
      (provider['api-key-configured'] || Boolean(provider['api-key']))
  })
  const verificationCredentialsReady = computed(() => {
    const provider = config.value['invoice-recognition'].verification
    const apiKeyReady = !provider['clear-api-key'] &&
      (provider['api-key-configured'] || Boolean(provider['api-key']))
    const secretKeyReady = !provider['clear-secret-key'] &&
      (provider['secret-key-configured'] || Boolean(provider['secret-key']))
    return provider.endpoint ? apiKeyReady || secretKeyReady : apiKeyReady && secretKeyReady
  })
  const multimodalProtocolLabel = computed(() => {
    const protocol = config.value['invoice-recognition'].multimodal.protocol
    if (protocol === 'openai-compatible') return '协议：OpenAI Compatible（自动）'
    if (protocol === 'anthropic') return '协议：Anthropic（自动）'
    return ''
  })
  const publicOCRDetectionLabel = computed(() => {
    const provider = config.value['invoice-recognition']['public-ocr']
    if (provider.provider === 'http-compatible' && provider.protocol === 'multipart-json-v1') {
      return 'HTTP Compatible（自动）'
    }
    return ''
  })
  const verificationDetectionLabel = computed(() => {
    const provider = config.value['invoice-recognition'].verification
    if (provider.provider === 'baidu' && provider.protocol === 'baidu-vat-invoice-v1') {
      return '百度权威验真（自动）'
    }
    if (provider.provider === 'http-compatible' && provider.protocol === 'invoice-verification-json-v1') {
      return 'HTTP 验真网关（自动）'
    }
    return ''
  })

  const resetPublicOCRDetection = () => {
    const provider = config.value['invoice-recognition']['public-ocr']
    provider.provider = ''
    provider.protocol = ''
  }

  const resetVerificationDetection = () => {
    const provider = config.value['invoice-recognition'].verification
    provider.provider = ''
    provider.protocol = ''
  }

  const resetMultimodalProtocol = () => {
    config.value['invoice-recognition'].multimodal.protocol = ''
  }

  const initForm = async () => {
    configLoading.value = true
    const hadConfig = configReady.value
    try {
      const res = await getSystemConfig()
      if (res.code === 0) {
        config.value = res.data.config
        configuredSecrets.value = res.data.configuredSecrets || {}
        normalizeInvoiceRecognition()
        savedSnapshot.value = serializeConfig(config.value)
        configReady.value = true
        configLoadError.value = false
        return true
      }
      if (!hadConfig) configLoadError.value = true
      return false
    } catch {
      if (!hadConfig) configLoadError.value = true
      return false
    } finally {
      configLoading.value = false
    }
  }
  initForm()
  const reload = async () => {
    if (reloading.value) return
    try {
      await ElMessageBox.confirm('确定要重载服务?', '警告', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      })
    } catch (action) {
      if (action === 'cancel' || action === 'close') {
        ElMessage({ type: 'info', message: '已取消重载' })
      }
      return
    }

    reloading.value = true
    try {
      const res = await reloadSystem()
      if (res.code === 0) {
        ElMessage({ type: 'success', message: '系统配置重载完成' })
      }
    } catch {
      // 请求错误由全局拦截器统一展示，避免重复提示。
    } finally {
      reloading.value = false
    }
  }

  const update = async () => {
    if (saving.value || !isDirty.value) return
    saving.value = true
    try {
      const multimodal = config.value['invoice-recognition'].multimodal
      const res = await setSystemConfig({ config: config.value })
      if (res.code === 0) {
        const protocol = res.data?.multimodal?.protocol
        const verificationProvider = res.data?.verification?.provider
        ElMessage({
          type: 'success',
          message: verificationProvider || (multimodal.enabled && protocol)
            ? '配置文件设置成功，连接协议已由服务器自动识别'
            : '配置文件设置成功'
        })
        savedSnapshot.value = serializeConfig(config.value)
        await initForm()
      }
    } finally {
      saving.value = false
    }
  }

  const testProvider = async (target, notify = true) => {
    if (!canManageInvoiceRecognition.value || testingProvider.value) return false
    testingProvider.value = target
    try {
      const res = await testInvoiceRecognitionProvider({
        target,
        config: config.value['invoice-recognition']
      })
      if (res?.code !== 0) return false
      const provider = res.data?.provider
      const protocol = res.data?.protocol
      if (target === 'multimodal') {
        if (!['openai-compatible', 'anthropic'].includes(protocol)) {
          ElMessage.error('连接成功，但未能确定接口协议')
          return false
        }
        config.value['invoice-recognition'].multimodal.protocol = protocol
        if (notify) {
          ElMessage.success(`多模态模型连接正常，已自动识别为 ${protocol === 'anthropic' ? 'Anthropic' : 'OpenAI Compatible'} 协议`)
        }
      } else if (target === 'public-ocr') {
        if (provider !== 'http-compatible' || protocol !== 'multipart-json-v1') {
          ElMessage.error('连接成功，但未能确定公网 OCR 协议')
          return false
        }
        const configuration = config.value['invoice-recognition']['public-ocr']
        configuration.provider = provider
        configuration.protocol = protocol
        if (notify) ElMessage.success('公网 OCR 连接正常，协议已自动识别')
      } else if (target === 'verification') {
        const supported = (provider === 'baidu' && protocol === 'baidu-vat-invoice-v1') ||
          (provider === 'http-compatible' && protocol === 'invoice-verification-json-v1')
        if (!supported) {
          ElMessage.error('连接成功，但未能确定权威验真供应商')
          return false
        }
        const configuration = config.value['invoice-recognition'].verification
        configuration.provider = provider
        configuration.protocol = protocol
        if (notify) {
          ElMessage.success(provider === 'baidu'
            ? '权威验真连接正常，已自动识别为百度协议'
            : '权威验真连接正常，已自动识别为 HTTP 网关协议')
        }
      } else if (target === 'baidu' && notify) {
        ElMessage.success('百度智能云凭据认证成功')
      }
      return true
    } finally {
      testingProvider.value = ''
    }
  }

  const email = async () => {
    if (testingEmail.value || isDirty.value) return
    testingEmail.value = true
    try {
      const res = await emailTest()
      ElMessage({
        type: res.code === 0 ? 'success' : 'error',
        message: res.code === 0 ? '邮件发送成功' : '邮件发送失败'
      })
    } finally {
      testingEmail.value = false
    }
  }

  const getUUID = () => {
    config.value.jwt['signing-key'] = CreateUUID()
  }

  const addNode = () => {
    config.value.mongo.hosts.push({
      host: '',
      port: ''
    })
  }

  const removeNode = (index) => {
    config.value.mongo.hosts.splice(index, 1)
  }
</script>

<style lang="scss" scoped>
  .system-config-page {
    min-width: 0;
    padding: 20px 24px 28px;
    color: var(--na-foreground);
  }

  .system-config-page :deep(.na-page-header) {
    margin-bottom: 12px;
  }

  .system-config-page :deep(.na-page-title) {
    font-size: 22px;
  }

  .header-action-wrap {
    display: inline-flex;
  }

  .form-action-wrap {
    display: inline-flex;
  }

  .runtime-summary {
    display: grid;
    grid-template-columns: minmax(180px, 1.1fr) minmax(0, 3fr) auto;
    align-items: stretch;
    min-width: 0;
    margin-bottom: 16px;
    border-block: 1px solid var(--na-border);
    background: var(--na-card);
  }

  .runtime-summary.unavailable .runtime-summary-lead,
  .runtime-summary.unavailable .runtime-summary-facts {
    opacity: 0.55;
  }

  .runtime-summary-lead {
    display: flex;
    align-items: center;
    gap: 12px;
    min-width: 0;
    padding: 12px 16px;
  }

  .runtime-summary-lead > div {
    display: flex;
    min-width: 0;
    flex-direction: column;
    gap: 2px;
  }

  .runtime-summary-lead span {
    color: var(--na-muted-foreground);
    font-size: 12px;
    line-height: 18px;
  }

  .runtime-summary-lead strong {
    overflow: hidden;
    color: var(--na-foreground);
    font-size: 13px;
    font-weight: 650;
    line-height: 20px;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .signal-dot {
    width: 9px;
    height: 9px;
    flex: 0 0 auto;
    border: 2px solid var(--na-card);
    border-radius: 50%;
    background: var(--na-success);
    box-shadow: 0 0 0 3px var(--na-success-soft);
  }

  .signal-dot.waiting {
    background: var(--na-info);
    box-shadow: 0 0 0 3px var(--na-info-soft);
  }

  .signal-dot.failed {
    background: var(--na-danger);
    box-shadow: 0 0 0 3px var(--na-danger-soft);
  }

  .runtime-summary-facts {
    display: grid;
    grid-template-columns: repeat(4, minmax(120px, 1fr));
    min-width: 0;
    margin: 0;
    border-left: 1px solid var(--na-border);
  }

  .runtime-summary-facts > div {
    display: flex;
    min-width: 0;
    flex-direction: column;
    justify-content: center;
    gap: 2px;
    padding: 10px 14px;
    border-right: 1px solid var(--na-border);
  }

  .runtime-summary-facts dt {
    color: var(--na-muted-foreground);
    font-size: 12px;
    line-height: 18px;
  }

  .runtime-summary-facts dd {
    overflow: hidden;
    margin: 0;
    color: var(--na-foreground);
    font-size: 13px;
    font-weight: 600;
    line-height: 20px;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .config-console {
    display: grid;
    overflow: hidden;
    grid-template-columns: 220px minmax(0, 1fr);
    min-width: 0;
    min-height: 620px;
  }

  .config-sidebar {
    min-width: 0;
    border-right: 1px solid var(--na-border);
    background: color-mix(in srgb, var(--na-muted) 70%, var(--na-card));
  }

  .config-sidebar-header {
    display: flex;
    min-height: 64px;
    flex-direction: column;
    justify-content: center;
    gap: 2px;
    padding: 12px 16px;
    border-bottom: 1px solid var(--na-border);
  }

  .config-sidebar-header span {
    color: var(--na-muted-foreground);
    font-size: 12px;
    line-height: 18px;
  }

  .config-sidebar-header strong {
    color: var(--na-foreground);
    font-size: 14px;
    font-weight: 650;
    line-height: 20px;
  }

  .config-sidebar-nav {
    padding: 12px 8px 20px;
  }

  .config-nav-group + .config-nav-group {
    margin-top: 12px;
  }

  .config-nav-group-title {
    display: grid;
    grid-template-columns: 18px minmax(0, 1fr) auto;
    align-items: center;
    gap: 8px;
    min-height: 28px;
    padding: 4px 8px;
    color: var(--na-muted-foreground);
    font-size: 12px;
    font-weight: 600;
  }

  .config-nav-group-title .el-icon {
    font-size: 14px;
  }

  .config-nav-group-title small {
    font-size: 11px;
    font-variant-numeric: tabular-nums;
    font-weight: 600;
  }

  .config-nav-items {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .config-nav-items button {
    display: grid;
    grid-template-columns: minmax(0, 1fr) 16px;
    align-items: center;
    gap: 8px;
    min-height: 36px;
    padding: 7px 10px 7px 34px;
    border: 0;
    border-radius: 7px;
    color: var(--na-muted-foreground);
    background: transparent;
    cursor: pointer;
    font-size: 13px;
    text-align: left;
    transition: color 180ms cubic-bezier(.22, 1, .36, 1), background-color 180ms cubic-bezier(.22, 1, .36, 1);
  }

  .config-nav-items button:hover {
    color: var(--na-foreground);
    background: var(--na-card);
  }

  .config-nav-items button:focus-visible {
    outline: 2px solid var(--na-primary);
    outline-offset: 2px;
  }

  .config-nav-items button.active {
    color: var(--na-primary);
    background: var(--na-primary-soft);
    font-weight: 650;
  }

  .config-nav-items button:disabled {
    cursor: not-allowed;
    opacity: 0.55;
  }

  .config-nav-arrow {
    opacity: 0;
    transition: opacity 180ms ease, transform 180ms cubic-bezier(.22, 1, .36, 1);
  }

  .config-nav-items button:hover .config-nav-arrow,
  .config-nav-items button.active .config-nav-arrow {
    opacity: 1;
  }

  .config-nav-items button.active .config-nav-arrow {
    transform: translateX(2px);
  }

  .config-main {
    min-width: 0;
    background: var(--na-card);
  }

  .config-mobile-nav {
    display: none;
  }

  .config-workbench {
    overflow: hidden;
    min-width: 0;
    min-height: 620px;
  }

  .config-editor-header {
    display: flex;
    align-items: center;
    gap: 16px;
    min-height: 84px;
    padding: 14px 20px;
    border-bottom: 1px solid var(--na-border);
  }

  .config-editor-heading {
    display: flex;
    align-items: center;
    min-width: 0;
    gap: 12px;
  }

  .config-editor-icon {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 40px;
    height: 40px;
    flex: 0 0 auto;
    border-radius: var(--na-radius-sm);
    color: var(--na-primary);
    background: var(--na-primary-soft);
    font-size: 18px;
  }

  .config-editor-heading > div {
    min-width: 0;
  }

  .config-editor-heading > div > span {
    display: block;
    margin-bottom: 1px;
    color: var(--na-muted-foreground);
    font-size: 12px;
    line-height: 18px;
  }

  .config-editor-heading h2 {
    overflow: hidden;
    margin: 0;
    color: var(--na-foreground);
    font-size: 17px;
    font-weight: 650;
    line-height: 24px;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .config-editor-heading p {
    overflow: hidden;
    margin: 2px 0 0;
    color: var(--na-muted-foreground);
    font-size: 12px;
    line-height: 18px;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .config-sync-state {
    display: inline-flex;
    align-items: center;
    flex: 0 0 auto;
    gap: 8px;
    min-height: 32px;
    align-self: center;
    margin: 0 16px;
    padding: 4px 8px;
    border: 1px solid var(--na-border);
    border-radius: 999px;
    color: var(--na-success);
    background: var(--na-success-soft);
    font-size: 12px;
    font-weight: 600;
  }

  .config-sync-state.dirty {
    color: var(--na-warning);
    background: var(--na-warning-soft);
  }

  .config-sync-state.failed {
    color: var(--na-danger);
    background: var(--na-danger-soft);
  }

  .config-sync-state.waiting {
    color: var(--na-info);
    background: var(--na-info-soft);
  }

  .config-form,
  .config-tabs {
    min-width: 0;
  }

  .config-tabs {
    --config-field-max-width: 300px;
  }

  .config-tabs :deep(.el-tabs__header) {
    display: none;
  }

  .config-tabs :deep(.el-tabs__content) {
    min-width: 0;
    overflow: visible;
  }

  .config-tabs :deep(.el-tab-pane) {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(220px, var(--config-field-max-width)));
    justify-content: start;
    gap: 0 16px;
    width: 100%;
    min-height: 360px;
    margin-top: 0 !important;
    padding: 16px 20px 24px;
  }

  .config-tabs :deep(.el-tab-pane > .el-form-item) {
    width: 100%;
    min-width: 0;
    max-width: var(--config-field-max-width);
    margin-bottom: 12px;
  }

  .config-tabs :deep(.el-tab-pane > .el-form-item:has(h3)),
  .config-tabs :deep(.el-tab-pane > h2),
  .config-tabs :deep(.el-tab-pane > .recognition-settings) {
    grid-column: 1 / -1;
    max-width: none;
  }

  .config-tabs :deep(.el-form-item__label) {
    height: auto;
    padding: 0 0 4px;
    color: var(--na-muted-foreground);
    font-size: 12px;
    font-weight: 550;
    line-height: 18px;
  }

  .config-tabs :deep(.el-form-item__content) {
    min-width: 0;
    max-width: none;
  }

  .config-tabs :deep(.el-input),
  .config-tabs :deep(.el-select),
  .config-tabs :deep(.el-input-number) {
    width: 100%;
  }

  .config-tabs h2,
  .config-tabs :deep(h3) {
    margin: 0 0 12px;
    padding-bottom: 8px;
    border-bottom: 1px solid var(--na-border);
    color: var(--na-foreground);
    font-size: 14px;
    font-weight: 650;
    line-height: 22px;
  }

  .config-error-state {
    display: flex;
    align-items: center;
    justify-content: center;
    flex-direction: column;
    gap: 12px;
    min-height: 360px;
    padding: 32px;
    color: var(--na-danger);
    text-align: center;
  }

  .config-error-state > .el-icon {
    font-size: 28px;
  }

  .config-error-state strong {
    color: var(--na-foreground);
    font-size: 14px;
    font-weight: 600;
  }

  .config-loading-state {
    min-height: 360px;
    padding: 20px;
  }

  .recognition-settings {
    width: 100%;
  }

  .recognition-overview {
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto;
    align-items: center;
    gap: 16px;
    padding: 8px 12px;
    border: 1px solid var(--na-border);
    border-radius: var(--na-radius-sm);
    background: var(--na-muted);
  }

  .recognition-flow {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 8px 12px;
    min-width: 0;
  }

  .flow-step {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    white-space: nowrap;
  }

  .flow-node {
    display: inline-flex;
    align-items: center;
    min-height: 28px;
    padding: 4px 8px;
    border: 1px solid var(--na-ring);
    border-radius: var(--na-radius-sm);
    color: var(--na-primary);
    background: var(--na-primary-soft);
    font-size: 12px;
    line-height: 18px;
    white-space: nowrap;
  }

  .flow-node.is-fixed {
    color: var(--na-muted-foreground);
    border-color: var(--na-border);
    background: var(--na-card);
  }

  .flow-arrow {
    color: var(--na-muted-foreground);
    font-size: 13px;
    line-height: 1;
  }

  .threshold-field {
    flex: 0 0 auto;
    margin-bottom: 0;
  }

  .threshold-field :deep(.el-form-item__label) {
    color: var(--na-muted-foreground);
    font-size: 12px;
  }

  .recognition-controls {
    display: flex;
    align-items: center;
    gap: 20px;
    flex: 0 0 auto;
  }

  .provider-section {
    display: grid;
    grid-template-columns: 220px minmax(0, 1fr);
    gap: 0 24px;
    margin: 0;
    padding: 16px 0 4px;
    border-bottom: 1px solid var(--na-border);
    background: transparent;
  }

  .provider-section:last-child {
    border-bottom: 0;
  }

  .provider-heading,
  .provider-title-line,
  .provider-heading-actions,
  .secret-row {
    display: flex;
    align-items: center;
  }

  .provider-heading {
    grid-column: 1;
    grid-row: span 2;
    align-items: flex-start;
    flex-direction: column;
    justify-content: flex-start;
    gap: 8px;
    margin: 0;
    padding: 0;
  }

  .provider-hint {
    grid-column: 2;
    max-width: 82ch;
    margin: 0 0 12px;
    color: var(--na-muted-foreground);
    font-size: 12px;
    line-height: 19px;
  }

  .provider-title-line {
    flex-wrap: wrap;
    gap: 8px;
    min-width: 0;
  }

  .provider-title-line h3 {
    margin: 0;
    padding: 0;
    border: 0;
    color: var(--na-foreground);
    font-size: 15px;
    font-weight: 600;
    line-height: 24px;
  }

  .provider-title-line :deep(.el-tag) {
    --el-tag-border-radius: 4px;
  }

  .provider-heading-actions {
    flex: 0 0 auto;
    flex-wrap: wrap;
    gap: 8px;
  }

  .provider-grid {
    display: grid;
    grid-column: 2;
    grid-template-columns: repeat(2, minmax(240px, 420px));
    gap: 0 16px;
    max-width: 856px;
  }

  .provider-grid .grid-full {
    grid-column: 1 / -1;
  }

  .provider-grid :deep(.el-form-item) {
    display: block;
    margin-bottom: 12px;
  }

  .provider-grid :deep(.el-form-item__label) {
    width: auto !important;
    height: auto;
    padding: 0 0 4px;
    color: var(--na-muted-foreground);
    font-size: 12px;
    line-height: 20px;
    justify-content: flex-start;
  }

  .provider-grid :deep(.el-form-item__content) {
    min-height: 32px;
    margin-left: 0 !important;
    line-height: 32px;
  }

  .provider-grid :deep(.el-input) {
    width: 100%;
  }

  .provider-grid :deep(.el-input-number) {
    width: 160px;
    max-width: calc(100% - 28px);
  }

  .secret-row {
    flex-wrap: wrap;
    gap: 8px 12px;
    width: 100%;
  }

  .secret-row :deep(.el-input) {
    flex: 1 1 240px;
    min-width: 0;
  }

  .secret-row :deep(.el-checkbox) {
    flex: 0 0 auto;
    margin-right: 0;
  }

  .input-unit {
    margin-left: 8px;
    color: var(--na-muted-foreground);
    font-size: 12px;
  }

  @media (max-width: 1280px) {
    .provider-section {
      grid-template-columns: minmax(0, 1fr);
      gap: 0;
    }

    .provider-heading {
      grid-column: 1;
      grid-row: auto;
      align-items: center;
      flex-direction: row;
      justify-content: space-between;
      margin-bottom: 16px;
    }

    .provider-hint,
    .provider-grid {
      grid-column: 1;
    }
  }

  @media (max-width: 1100px) {
    .runtime-summary {
      grid-template-columns: minmax(0, 1fr) auto;
    }

    .runtime-summary-facts {
      grid-column: 1 / -1;
      grid-row: 2;
      border-top: 1px solid var(--na-border);
      border-left: 0;
    }

    .config-console {
      grid-template-columns: minmax(0, 1fr);
    }

    .config-sidebar {
      display: none;
    }

    .config-mobile-nav {
      display: grid;
      grid-template-columns: 88px minmax(0, 1fr);
      align-items: center;
      gap: 12px;
      padding: 10px 16px;
      border-bottom: 1px solid var(--na-border);
      background: color-mix(in srgb, var(--na-muted) 70%, var(--na-card));
    }

    .config-mobile-nav > span {
      color: var(--na-muted-foreground);
      font-size: 12px;
      font-weight: 600;
    }

    .recognition-overview {
      grid-template-columns: minmax(0, 1fr);
    }

    .recognition-controls {
      justify-content: flex-start;
    }
  }

  @media (max-width: 768px) {
    .runtime-summary-facts {
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }

    .runtime-summary-facts > div:nth-child(2n) {
      border-right: 0;
    }

    .config-tabs :deep(.el-tab-pane) {
      grid-template-columns: minmax(0, 1fr);
      min-height: 360px;
      padding: 16px 12px 20px;
    }

    .config-tabs :deep(.el-tab-pane > .el-form-item) {
      max-width: 420px;
    }

    .config-tabs :deep(.el-tab-pane > .el-form-item:has(h3)),
    .config-tabs :deep(.el-tab-pane > h2),
    .config-tabs :deep(.el-tab-pane > .recognition-settings) {
      grid-column: 1;
    }

    .config-editor-header {
      min-height: 76px;
      padding: 12px 16px;
    }

    .recognition-overview {
      gap: 16px;
      padding: 12px;
    }

    .recognition-controls {
      align-items: flex-start;
      flex-direction: column;
      gap: 12px;
    }

    .threshold-field :deep(.el-form-item__label) {
      justify-content: flex-start;
    }

    .provider-section {
      padding: 20px 0 4px;
    }

    .provider-heading {
      align-items: flex-start;
      flex-direction: column;
      gap: 12px;
    }

    .provider-heading-actions {
      justify-content: space-between;
      width: 100%;
    }

    .provider-grid {
      grid-template-columns: minmax(0, 1fr);
    }

    .provider-grid .grid-full {
      grid-column: auto;
    }
  }

  @media (max-width: 480px) {
    .system-config-page {
      padding-inline: 12px;
    }

    .system-config-page :deep(.na-page-actions) {
      display: grid;
      grid-template-columns: repeat(2, minmax(0, 1fr));
      width: 100%;
    }

    .header-action-wrap,
    .header-action-wrap :deep(.el-button),
    .system-config-page :deep(.na-page-actions > .el-button) {
      width: 100%;
      margin-left: 0;
    }

    .config-tabs :deep(.el-tab-pane > .el-form-item) {
      max-width: none;
    }

    .runtime-summary {
      grid-template-columns: minmax(0, 1fr);
    }

    .runtime-summary-facts {
      grid-row: auto;
    }

    .config-sync-state {
      justify-self: start;
      width: auto;
      margin: 0 12px 12px;
    }

    .config-mobile-nav {
      grid-template-columns: minmax(0, 1fr);
      gap: 8px;
    }

    .flow-step {
      width: 100%;
    }

    .flow-arrow {
      width: 18px;
      text-align: center;
      transform: rotate(90deg);
    }
  }

  @media (prefers-reduced-motion: reduce) {
    .config-nav-items button,
    .config-nav-arrow {
      transition: none;
    }
  }
</style>
