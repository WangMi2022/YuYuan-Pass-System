<template>
  <main class="na-page na-page--list system-config-page">
    <div class="config-console na-panel">
      <div class="config-main">
        <section class="config-workbench" :aria-labelledby="`config-section-${activeNames}`">
          <header class="config-editor-header">
            <div class="config-editor-heading">
              <span class="config-editor-icon" aria-hidden="true">
                <el-icon><component :is="activeSection.icon" /></el-icon>
              </span>
              <div>
                <h2 :id="`config-section-${activeNames}`">{{ activeSection.label }}</h2>
                <p>{{ activeSection.description }}</p>
              </div>
            </div>
            <div class="config-editor-actions" aria-label="基础设置操作">
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
          <el-form-item label="发件人名称">
            <el-input
              v-model.trim="config.email.nickname"
              placeholder="例如：资产管理平台"
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
        <el-tab-pane label="验证服务" name="7" lazy>
          <div class="verification-config-heading">
            <div>
              <strong>登录图形验证码</strong>
              <p>控制登录安全校验图片的尺寸与字符长度</p>
            </div>
          </div>
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

          <div class="verification-config-heading">
            <div>
              <strong>联系方式验证总开关</strong>
              <p>关闭时不发送验证码，也不允许用户修改手机号或邮箱；可先配置渠道，接入完成后再开启。</p>
            </div>
            <el-tag :type="contactVerificationReady ? 'success' : 'info'" effect="plain">
              {{ contactVerificationReady ? '可开启' : '等待渠道配置' }}
            </el-tag>
          </div>
          <el-form-item label="启用联系方式验证">
            <div class="verification-toggle-field">
              <el-switch
                v-model="config['contact-verification'].enabled"
                :disabled="!contactVerificationReady && !config['contact-verification'].enabled"
              />
              <span>{{ contactVerificationReady ? '开启后，用户修改手机或邮箱必须完成对应渠道验证码' : '请先完成并开启邮件或接码中心渠道' }}</span>
            </div>
          </el-form-item>

          <div class="verification-config-heading">
            <div>
              <strong>邮件验证码</strong>
              <p>复用“邮件服务”中的 SMTP 账户，配置完整后才可开启</p>
            </div>
            <el-tag :type="emailVerificationReady ? 'success' : 'info'" effect="plain">
              {{ emailVerificationReady ? '配置完整' : '待配置' }}
            </el-tag>
          </div>
          <el-form-item label="开启邮件验证码">
            <div class="verification-toggle-field">
              <el-switch
                v-model="config['contact-verification'].email.enabled"
                :disabled="!emailVerificationReady && !config['contact-verification'].email.enabled"
              />
              <span>{{ emailVerificationReady ? '开启后，修改邮箱必须完成验证码校验' : '请先补齐 SMTP 地址、端口、发件人和密钥' }}</span>
            </div>
          </el-form-item>
          <el-form-item label="验证码邮件主题">
            <el-input
              v-model.trim="config['contact-verification'].email.subject"
              placeholder="账号安全验证码"
            />
          </el-form-item>

          <div class="verification-config-heading">
            <div>
              <strong>接码中心</strong>
              <p>通过统一 Webhook 接入接码中心，后续可替换为正式短信供应商</p>
            </div>
            <el-tag :type="smsVerificationReady ? 'success' : 'info'" effect="plain">
              {{ smsVerificationReady ? '配置完整' : '待配置' }}
            </el-tag>
          </div>
          <el-form-item label="接入方式">
            <el-select v-model="config['contact-verification'].sms.provider">
              <el-option label="HTTP Webhook" value="webhook" />
            </el-select>
          </el-form-item>
          <el-form-item label="接码中心地址">
            <el-input
              v-model.trim="config['contact-verification'].sms.endpoint"
              placeholder="https://sms.example.com/send"
            />
          </el-form-item>
          <el-form-item label="接码中心签名">
            <el-input
              v-model.trim="config['contact-verification'].sms['sign-name']"
              placeholder="请输入短信签名"
            />
          </el-form-item>
          <el-form-item label="接码中心模板 ID">
            <el-input
              v-model.trim="config['contact-verification'].sms['template-id']"
              placeholder="请输入验证码模板 ID"
            />
          </el-form-item>
          <el-form-item label="接码中心访问令牌">
            <SecretInput
              v-model.trim="config['contact-verification'].sms['access-token']"
              secret-path="contact-verification.sms.access-token"
              :configured="isSecretConfigured('contact-verification.sms.access-token')"
              :can-reveal="canManageSystemSecrets"
              placeholder="请输入 Webhook Bearer Token"
            />
          </el-form-item>
          <el-form-item label="开启接码中心验证码">
            <div class="verification-toggle-field">
              <el-switch
                v-model="config['contact-verification'].sms.enabled"
                :disabled="!smsVerificationReady && !config['contact-verification'].sms.enabled"
              />
              <span>{{ smsVerificationReady ? '开启后，修改手机号必须完成验证码校验' : '请先补齐地址、签名、模板和访问令牌' }}</span>
            </div>
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
      </el-tabs>
      </el-form>
      <div v-else-if="configLoadError" class="config-error-state" role="alert">
        <el-icon><WarningFilled /></el-icon>
        <strong>基础设置读取失败</strong>
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
    setSystemConfig
  } from '@/api/system'
  import { computed, ref, watch } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import {
    Check,
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
  const testingEmail = ref(false)
  const reloading = ref(false)
  const saving = ref(false)
  const configLoading = ref(false)
  const configReady = ref(false)
  const configLoadError = ref(false)
  const savedSnapshot = ref('')
  const configuredSecrets = ref({})


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
    'contact-verification': {
      enabled: false,
      sms: {
        enabled: false,
        provider: 'webhook',
        endpoint: '',
        'access-token': '',
        'sign-name': '',
        'template-id': ''
      },
      email: {
        enabled: false,
        subject: '账号安全验证码'
      }
    },
    timer: {
      detail: {}
    }
  })

  const configSections = [
    { name: '1', label: '基础设置', description: '服务端口、存储类型与基础开关', icon: Setting },
    { name: '2', label: 'JWT 签名', description: '签名密钥、令牌周期与签发身份', icon: Key },
    { name: '3', label: '运行日志', description: '日志级别、输出方式与保留策略', icon: Document },
    { name: '4', label: 'Redis', description: '缓存连接与数据库选择', icon: Connection, enabled: () => config.value.system['use-redis'] },
    { name: '5', label: '邮件服务', description: 'SMTP 连接与通知账户', icon: Message },
    { name: '7', label: '验证服务', description: '登录图形验证码与联系方式验证渠道', icon: View },
    { name: '9', label: '主数据库', description: '数据库连接、连接池与日志策略', icon: Coin },
    { name: '10', label: '文件存储', description: '本地与对象存储服务连接', icon: UploadFilled },
    { name: '14', label: 'MongoDB', description: '文档数据库集群与连接参数', icon: DataBoard, enabled: () => config.value.system['use-mongo'] }
  ]

  const visibleSections = computed(() => configSections.filter((section) => !section.enabled || section.enabled()))
  const activeSection = computed(() =>
    visibleSections.value.find((section) => section.name === activeNames.value) || visibleSections.value[0]
  )
  const serializeConfig = (value) => JSON.stringify(value)
  const isSecretConfigured = (path) => Boolean(configuredSecrets.value[path])
  const hasConfiguredSecret = (path, value) => Boolean(
    String(value || '').trim() || isSecretConfigured(path)
  )
  const emailVerificationReady = computed(() => {
    const email = config.value.email || {}
    return Boolean(
      String(email.from || '').trim() &&
      String(email.host || '').trim() &&
      Number(email.port) > 0 &&
      Number(email.port) <= 65535 &&
      hasConfiguredSecret('email.secret', email.secret)
    )
  })
  const smsVerificationReady = computed(() => {
    const sms = config.value['contact-verification']?.sms || {}
    return Boolean(
      sms.provider === 'webhook' &&
      /^https?:\/\/\S+$/i.test(String(sms.endpoint || '').trim()) &&
      String(sms['sign-name'] || '').trim() &&
      String(sms['template-id'] || '').trim() &&
      hasConfiguredSecret('contact-verification.sms.access-token', sms['access-token'])
    )
  })
  const contactVerificationReady = computed(() => {
    const verification = config.value['contact-verification'] || {}
    return Boolean(
      (verification.email?.enabled && emailVerificationReady.value) ||
      (verification.sms?.enabled && smsVerificationReady.value)
    )
  })
  const isDirty = computed(() => Boolean(savedSnapshot.value) &&
    serializeConfig(config.value) !== savedSnapshot.value)
  watch(visibleSections, (sections) => {
    if (!sections.some((section) => section.name === activeNames.value)) {
      activeNames.value = '1'
    }
  })

  const withVerificationDefaults = (value) => {
    const verification = value?.['contact-verification'] || {}
    value['contact-verification'] = {
      ...verification,
      enabled: Boolean(verification.enabled),
      sms: {
        enabled: false,
        provider: 'webhook',
        endpoint: '',
        'access-token': '',
        'sign-name': '',
        'template-id': '',
        ...(verification.sms || {})
      },
      email: {
        enabled: false,
        subject: '账号安全验证码',
        ...(verification.email || {})
      }
    }
    return value
  }

  const initForm = async () => {
    configLoading.value = true
    const hadConfig = configReady.value
    try {
      const res = await getSystemConfig()
      if (res.code === 0) {
        config.value = withVerificationDefaults(res.data.config)
        configuredSecrets.value = res.data.configuredSecrets || {}
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
      const res = await setSystemConfig({ config: config.value })
      if (res.code === 0) {
        ElMessage({ type: 'success', message: '配置文件设置成功' })
        savedSnapshot.value = serializeConfig(config.value)
        await initForm()
      }
    } finally {
      saving.value = false
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
    padding: 16px 18px 24px;
    color: var(--na-foreground);
  }

  .header-action-wrap {
    display: inline-flex;
  }

  .form-action-wrap {
    display: inline-flex;
  }

  .config-console {
    display: flex;
    flex-direction: column;
    overflow: hidden;
    min-width: 0;
    border-radius: 12px;
  }

  .config-main {
    display: flex;
    flex: 1 1 auto;
    flex-direction: column;
    min-width: 0;
    min-height: 0;
    background: var(--na-card);
  }

  .config-workbench {
    display: flex;
    flex: 1 1 auto;
    flex-direction: column;
    overflow: hidden;
    min-width: 0;
  }

  .config-editor-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    min-height: 74px;
    padding: 14px 16px;
    border-bottom: 1px solid var(--na-border);
  }

  .config-editor-heading {
    display: flex;
    align-items: center;
    min-width: 0;
    gap: 12px;
  }

  .config-editor-actions {
    display: flex;
    align-items: center;
    flex: 0 0 auto;
    gap: 8px;
  }

  .config-editor-icon {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 38px;
    height: 38px;
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

  .config-form,
  .config-tabs {
    min-width: 0;
  }

  .config-form {
    flex: 1 1 auto;
    min-height: 0;
    overflow: auto;
    background: color-mix(in srgb, var(--na-muted) 30%, var(--na-card));
  }

  .config-tabs {
    --config-field-max-width: 300px;
  }

  .config-tabs :deep(.el-tabs__header) {
    display: block;
    margin: 0;
    padding: 0 16px;
    border-bottom: 1px solid var(--na-border);
    background: var(--na-card);
  }

  .config-tabs :deep(.el-tabs__nav-wrap::after) {
    display: none;
  }

  .config-tabs :deep(.el-tabs__nav) {
    gap: 18px;
  }

  .config-tabs :deep(.el-tabs__item) {
    height: 44px;
    padding: 0 2px;
    color: var(--na-muted-foreground);
    font-size: 13px;
    font-weight: 560;
  }

  .config-tabs :deep(.el-tabs__item:hover),
  .config-tabs :deep(.el-tabs__item.is-active) {
    color: var(--na-primary);
  }

  .config-tabs :deep(.el-tabs__active-bar) {
    height: 2px;
    border-radius: 2px 2px 0 0;
    background: var(--na-primary);
  }

  .config-tabs :deep(.el-tabs__content) {
    min-width: 0;
    overflow: visible;
  }

  .config-tabs :deep(.el-tab-pane) {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(220px, var(--config-field-max-width)));
    justify-content: start;
    gap: 2px 16px;
    width: 100%;
    margin-top: 0 !important;
    padding: 16px 16px 24px;
  }

  .config-tabs :deep(.el-tab-pane > .el-form-item) {
    width: 100%;
    min-width: 0;
    max-width: var(--config-field-max-width);
    margin-bottom: 12px;
  }

  .config-tabs :deep(.el-tab-pane > .el-form-item:has(h3)),
  .config-tabs :deep(.el-tab-pane > h2) {
    grid-column: 1 / -1;
    max-width: none;
  }

  .config-tabs :deep(.el-form-item__label) {
    height: auto;
    padding: 0 0 4px;
    color: color-mix(in srgb, var(--na-muted-foreground) 82%, var(--na-foreground));
    font-size: 12px;
    font-weight: 580;
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

  .verification-config-heading {
    grid-column: 1 / -1;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    min-width: 0;
    margin: 6px 0 10px;
    padding-bottom: 10px;
    border-bottom: 1px solid var(--na-border);
  }

  .verification-config-heading:not(:first-child) {
    margin-top: 16px;
  }

  .verification-config-heading strong {
    display: block;
    color: var(--na-foreground);
    font-size: 14px;
    font-weight: 650;
    line-height: 22px;
  }

  .verification-config-heading p {
    margin: 2px 0 0;
    color: var(--na-muted-foreground);
    font-size: 12px;
    line-height: 18px;
  }

  .verification-toggle-field {
    display: flex;
    align-items: flex-start;
    gap: 10px;
    min-height: 32px;
  }

  .verification-toggle-field :deep(.el-switch) {
    flex: 0 0 auto;
  }

  .verification-toggle-field span {
    color: var(--na-muted-foreground);
    font-size: 12px;
    line-height: 18px;
  }

  .config-error-state {
    display: flex;
    align-items: center;
    justify-content: center;
    flex-direction: column;
    gap: 12px;
    min-height: 220px;
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
    min-height: 220px;
    padding: 20px;
  }



  @media (max-width: 768px) {
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
    .verification-config-heading {
      grid-column: 1;
    }

    .config-editor-header {
      min-height: 76px;
      padding: 12px 16px;
    }

  }

  @media (max-width: 480px) {
    .system-config-page {
      padding-inline: 12px;
    }

    .config-editor-header {
      align-items: flex-start;
      flex-wrap: wrap;
    }

    .config-editor-actions {
      display: grid;
      grid-template-columns: repeat(2, minmax(0, 1fr));
      width: 100%;
    }

    .header-action-wrap,
    .header-action-wrap :deep(.el-button),
    .config-editor-actions > :deep(.el-button) {
      width: 100%;
      margin-left: 0;
    }

    .config-tabs :deep(.el-tab-pane > .el-form-item) {
      max-width: none;
    }

  }

  @media (prefers-reduced-motion: reduce) {
    .config-tabs :deep(.el-tabs__item),
    .config-tabs :deep(.el-tabs__active-bar) {
      transition: none;
    }
  }
</style>
