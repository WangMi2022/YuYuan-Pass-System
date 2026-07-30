<template>
  <div class="system">
    <el-form ref="form" :model="config" label-width="240px">
      <!--  System start  -->
      <el-tabs v-model="activeNames">
        <el-tab-pane label="系统配置" name="1" class="mt-3.5">
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
          <el-form-item label="Oss类型">
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
          <el-form-item label="开启redis">
            <el-switch v-model="config.system['use-redis']" />
          </el-form-item>
          <el-form-item label="开启Mongo">
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
        <el-tab-pane label="jwt签名" name="2" class="mt-3.5">
          <el-form-item label="jwt签名">
            <el-input
              v-model.trim="config.jwt['signing-key']"
              placeholder="请输入jwt签名"
            >
              <template #append>
                <el-button @click="getUUID">生成</el-button>
              </template>
            </el-input>
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
        <el-tab-pane label="Zap日志配置" name="3" class="mt-3.5">
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
          class="mt-3.5"
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
        <el-tab-pane label="邮箱配置" name="5" class="mt-3.5">
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
            <el-input
              v-model.trim="config.email.secret"
              placeholder="请输入secret"
            />
          </el-form-item>
          <el-form-item label="测试邮件">
            <el-button @click="email">测试邮件</el-button>
          </el-form-item>
        </el-tab-pane>
        <el-tab-pane
          label="Mongo 数据库配置"
          name="14"
          class="mt-3.5"
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
        <el-tab-pane label="验证码配置" name="7" class="mt-3.5">
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
        <el-tab-pane label="数据库配置" name="9" class="mt-3.5">
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
        <el-tab-pane label="oss配置" name="10" class="mt-3.5">
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
              <el-input
                v-model.trim="config.qiniu['access-key']"
                placeholder="请输入accessKey"
              />
            </el-form-item>
            <el-form-item label="secretKey">
              <el-input
                v-model.trim="config.qiniu['secret-key']"
                placeholder="请输入secretKey"
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
              <el-input
                v-model.trim="config['tencent-cos']['secret-id']"
                placeholder="请输入secretID"
              />
            </el-form-item>
            <el-form-item label="secretKey">
              <el-input
                v-model.trim="config['tencent-cos']['secret-key']"
                placeholder="请输入secretKey"
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
              <el-input
                v-model.trim="config['aliyun-oss']['access-key-id']"
                placeholder="请输入accessKeyId"
              />
            </el-form-item>
            <el-form-item label="accessKeySecret">
              <el-input
                v-model.trim="config['aliyun-oss']['access-key-secret']"
                placeholder="请输入accessKeySecret"
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
              <el-input
                v-model.trim="config['hua-wei-obs']['access-key']"
                placeholder="请输入accessKey"
              />
            </el-form-item>
            <el-form-item label="secretKey">
              <el-input
                v-model.trim="config['hua-wei-obs']['secret-key']"
                placeholder="请输入secretKey"
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
              <el-input
                v-model.trim="config['cloudflare-r2']['access-key-id']"
                placeholder="请输入secretKey"
              />
            </el-form-item>
            <el-form-item label="Secret Access Key">
              <el-input
                v-model.trim="config['cloudflare-r2']['secret-access-key']"
                placeholder="请输入secretKey"
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
              <el-input
                v-model.trim="config.minio['access-key-id']"
                placeholder="请输入Access Key ID"
              />
            </el-form-item>
            <el-form-item label="Access Key Secret">
              <el-input
                v-model.trim="config.minio['access-key-secret']"
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
        <el-tab-pane label="发票智能识别" name="11" class="mt-3.5">
          <div class="recognition-settings">
            <div class="recognition-overview">
              <div class="recognition-flow" aria-label="发票识别顺序">
                <span class="flow-node is-fixed">二维码</span>
                <span class="flow-arrow">→</span>
                <span class="flow-node">百度 / 公网 OCR</span>
                <span class="flow-arrow">→</span>
                <span class="flow-node">多模态模型</span>
                <span class="flow-arrow">→</span>
                <span class="flow-node is-fixed">人工核对</span>
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
                <el-switch
                  v-model="config['invoice-recognition'].baidu.enabled"
                  :disabled="!canManageInvoiceRecognition"
                  inline-prompt
                  active-text="启用"
                  inactive-text="停用"
                />
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
                    <el-input
                      v-model.trim="config['invoice-recognition'].baidu['api-key']"
                      type="password"
                      show-password
                      autocomplete="new-password"
                      :disabled="!canManageInvoiceRecognition || config['invoice-recognition'].baidu['clear-api-key']"
                      :placeholder="baiduAPIKeyPlaceholder"
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
                    <el-input
                      v-model.trim="config['invoice-recognition'].baidu['secret-key']"
                      type="password"
                      show-password
                      autocomplete="new-password"
                      :disabled="!canManageInvoiceRecognition || config['invoice-recognition'].baidu['clear-secret-key']"
                      :placeholder="baiduSecretKeyPlaceholder"
                    />
                    <el-checkbox
                      v-if="config['invoice-recognition'].baidu['secret-key-configured']"
                      v-model="config['invoice-recognition'].baidu['clear-secret-key']"
                      :disabled="!canManageInvoiceRecognition"
                    >清除</el-checkbox>
                  </div>
                </el-form-item>
              </div>
              <div v-if="canManageInvoiceRecognition" class="provider-actions">
                <el-button
                  type="primary"
                  plain
                  :icon="Connection"
                  :loading="testingProvider === 'baidu'"
                  @click="testProvider('baidu')"
                >
                  验证凭据
                </el-button>
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
                <el-switch
                  v-model="config['invoice-recognition']['public-ocr'].enabled"
                  :disabled="!canManageInvoiceRecognition"
                  inline-prompt
                  active-text="启用"
                  inactive-text="停用"
                />
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
                    <el-input
                      v-model.trim="config['invoice-recognition']['public-ocr']['api-key']"
                      type="password"
                      show-password
                      autocomplete="new-password"
                      :disabled="!canManageInvoiceRecognition || config['invoice-recognition']['public-ocr']['clear-api-key']"
                      :placeholder="publicOCRKeyPlaceholder"
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
              <div v-if="canManageInvoiceRecognition" class="provider-actions">
                <el-button
                  type="primary"
                  plain
                  :icon="Connection"
                  :loading="testingProvider === 'public-ocr'"
                  @click="testProvider('public-ocr')"
                >
                  测试并识别协议
                </el-button>
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
                <el-switch
                  v-model="config['invoice-recognition'].verification.enabled"
                  :disabled="!canManageInvoiceRecognition"
                  inline-prompt
                  active-text="启用"
                  inactive-text="停用"
                />
              </div>
              <p class="provider-hint">
                留空接口地址时自动探测百度权威验真；填写网关地址时自动探测系统统一的 HTTP 验真协议。供应商与协议由服务器保存，无需手动选择。
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
                    <el-input
                      v-model.trim="config['invoice-recognition'].verification['api-key']"
                      type="password"
                      show-password
                      autocomplete="new-password"
                      :disabled="!canManageInvoiceRecognition || config['invoice-recognition'].verification['clear-api-key']"
                      :placeholder="verificationAPIKeyPlaceholder"
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
                    <el-input
                      v-model.trim="config['invoice-recognition'].verification['secret-key']"
                      type="password"
                      show-password
                      autocomplete="new-password"
                      :disabled="!canManageInvoiceRecognition || config['invoice-recognition'].verification['clear-secret-key']"
                      :placeholder="verificationSecretKeyPlaceholder"
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
              <div v-if="canManageInvoiceRecognition" class="provider-actions">
                <el-button
                  type="primary"
                  plain
                  :icon="Connection"
                  :loading="testingProvider === 'verification'"
                  @click="testProvider('verification')"
                >
                  测试并识别供应商
                </el-button>
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
                <el-switch
                  v-model="config['invoice-recognition'].multimodal.enabled"
                  :disabled="!canManageInvoiceRecognition"
                  inline-prompt
                  active-text="启用"
                  inactive-text="停用"
                />
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
                    <el-input
                      v-model.trim="config['invoice-recognition'].multimodal['api-key']"
                      type="password"
                      show-password
                      autocomplete="new-password"
                      :disabled="!canManageInvoiceRecognition || config['invoice-recognition'].multimodal['clear-api-key']"
                      :placeholder="multimodalKeyPlaceholder"
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
              <div v-if="canManageInvoiceRecognition" class="provider-actions">
                <el-button
                  type="primary"
                  plain
                  :icon="Connection"
                  :loading="testingProvider === 'multimodal'"
                  @click="testProvider('multimodal')"
                >
                  测试连接
                </el-button>
              </div>
            </section>
          </div>
        </el-tab-pane>
        <el-tab-pane label="自动化代码配置" name="12" class="mt-3.5">
          <el-form-item label="是否自动重启(linux)">
            <el-switch v-model="config.autocode['transfer-restart']" />
          </el-form-item>
          <el-form-item label="root(项目根路径)">
            <el-input v-model="config.autocode.root" disabled />
          </el-form-item>
          <el-form-item label="Server(后端代码地址)">
            <el-input
              v-model.trim="config.autocode['server']"
              placeholder="请输入后端代码地址"
            />
          </el-form-item>
          <el-form-item label="SApi(后端api文件夹地址)">
            <el-input
              v-model.trim="config.autocode['server-api']"
              placeholder="请输入后端api文件夹地址"
            />
          </el-form-item>
          <el-form-item label="SInitialize(后端Initialize文件夹)">
            <el-input
              v-model.trim="config.autocode['server-initialize']"
              placeholder="请输入后端Initialize文件夹"
            />
          </el-form-item>
          <el-form-item label="SModel(后端Model文件地址)">
            <el-input
              v-model.trim="config.autocode['server-model']"
              placeholder="请输入后端Model文件地址"
            />
          </el-form-item>
          <el-form-item label="SRequest(后端Request文件夹地址)">
            <el-input
              v-model.trim="config.autocode['server-request']"
              placeholder="请输入后端Request文件夹地址"
            />
          </el-form-item>
          <el-form-item label="SRouter(后端Router文件夹地址)">
            <el-input
              v-model.trim="config.autocode['server-router']"
              placeholder="请输入后端Router文件夹地址"
            />
          </el-form-item>
          <el-form-item label="SService(后端Service文件夹地址)">
            <el-input
              v-model.trim="config.autocode['server-service']"
              placeholder="请输入后端Service文件夹地址"
            />
          </el-form-item>
          <el-form-item label="Web(前端文件夹地址)">
            <el-input
              v-model.trim="config.autocode.web"
              placeholder="请输入前端文件夹地址"
            />
          </el-form-item>
          <el-form-item label="WApi(后端WApi文件夹地址)">
            <el-input
              v-model.trim="config.autocode['web-api']"
              placeholder="请输入后端WApi文件夹地址"
            />
          </el-form-item>
          <el-form-item label="WForm(后端WForm文件夹地址)">
            <el-input
              v-model.trim="config.autocode['web-form']"
              placeholder="请输入后端WForm文件夹地址"
            />
          </el-form-item>
          <el-form-item label="WTable(后端WTable文件夹地址)">
            <el-input
              v-model.trim="config.autocode['web-table']"
              placeholder="请输入后端WTable文件夹地址"
            />
          </el-form-item>
        </el-tab-pane>
      </el-tabs>
    </el-form>
    <div class="mt-4">
      <el-button type="primary" :disabled="reloading" @click="update">立即更新 </el-button>
      <el-button type="primary" :loading="reloading" :disabled="reloading" @click="reload">重载服务 </el-button>
    </div>
  </div>
</template>

<script setup>
  import {
    getSystemConfig,
    reloadSystem,
    setSystemConfig,
    testInvoiceRecognitionProvider
  } from '@/api/system'
  import { computed, ref } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { Connection, Minus, Plus } from '@element-plus/icons-vue'
  import { emailTest } from '@/api/email'
  import { CreateUUID } from '@/utils/format'
  import { useUserStore } from '@/pinia/modules/user'

  defineOptions({
    name: 'Config'
  })

  const activeNames = ref('1')
  const userStore = useUserStore()
  const canManageInvoiceRecognition = computed(
    () => Number(userStore.userInfo.authorityId) === 888
  )
  const testingProvider = ref('')
  const reloading = ref(false)

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
  const publicOCRKeyPlaceholder = computed(() =>
    config.value['invoice-recognition']['public-ocr']['api-key-configured']
      ? '已配置，留空保持不变'
      : '请输入 API Key（可选）'
  )
  const baiduAPIKeyPlaceholder = computed(() =>
    config.value['invoice-recognition'].baidu['api-key-configured']
      ? '已配置，留空保持不变'
      : '请输入百度智能云 API Key'
  )
  const baiduSecretKeyPlaceholder = computed(() =>
    config.value['invoice-recognition'].baidu['secret-key-configured']
      ? '已配置，留空保持不变'
      : '请输入百度智能云 Secret Key'
  )
  const verificationAPIKeyPlaceholder = computed(() =>
    config.value['invoice-recognition'].verification['api-key-configured']
      ? '已配置，留空保持不变'
      : '请输入验真服务 API Key'
  )
  const verificationSecretKeyPlaceholder = computed(() =>
    config.value['invoice-recognition'].verification['secret-key-configured']
      ? '已配置，留空保持不变'
      : '请输入 Secret Key（HTTP 网关可选）'
  )
  const multimodalKeyPlaceholder = computed(() =>
    config.value['invoice-recognition'].multimodal['api-key-configured']
      ? '已配置，留空保持不变'
      : '请输入 API Key（可选）'
  )
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
    const res = await getSystemConfig()
    if (res.code === 0) {
      config.value = res.data.config
      normalizeInvoiceRecognition()
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
    const multimodal = config.value['invoice-recognition'].multimodal
    const res = await setSystemConfig({ config: config.value })
    if (res.code === 0) {
      const protocol = res.data?.multimodal?.protocol
      const verificationProvider = res.data?.verification?.provider
      ElMessage({
        type: 'success',
        message: verificationProvider || (multimodal.enabled && protocol)
          ? `配置文件设置成功，连接协议已由服务器自动识别`
          : '配置文件设置成功'
      })
      await initForm()
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
    const res = await emailTest()
    if (res.code === 0) {
      ElMessage({
        type: 'success',
        message: '邮件发送成功'
      })
      await initForm()
    } else {
      ElMessage({
        type: 'error',
        message: '邮件发送失败'
      })
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
  .system {
    @apply bg-white p-9 rounded dark:bg-slate-900;
    h2 {
      @apply p-2.5 my-2.5 text-lg shadow;
    }
  }

  .recognition-settings {
    max-width: 1120px;
  }

  .recognition-overview {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 24px;
    padding: 4px 0 20px;
    border-bottom: 1px solid var(--el-border-color-lighter);
  }

  .recognition-flow {
    display: flex;
    align-items: center;
    gap: 8px;
    min-width: 0;
  }

  .flow-node {
    padding: 5px 9px;
    border: 1px solid var(--el-color-primary-light-5);
    border-radius: 4px;
    color: var(--el-color-primary);
    background: var(--el-color-primary-light-9);
    font-size: 12px;
    line-height: 18px;
    white-space: nowrap;
  }

  .flow-node.is-fixed {
    color: var(--el-text-color-regular);
    border-color: var(--el-border-color);
    background: var(--el-fill-color-light);
  }

  .flow-arrow {
    color: var(--el-text-color-placeholder);
    font-size: 13px;
  }

  .threshold-field {
    flex: 0 0 auto;
    margin-bottom: 0;
  }

  .recognition-controls {
    display: flex;
    align-items: center;
    gap: 16px;
    flex: 0 0 auto;
  }

  .provider-section {
    padding: 22px 0;
    border-bottom: 1px solid var(--el-border-color-lighter);
  }

  .provider-heading,
  .provider-title-line,
  .provider-actions,
  .secret-row {
    display: flex;
    align-items: center;
  }

  .provider-heading {
    justify-content: space-between;
    margin-bottom: 18px;
  }

  .provider-hint {
    margin: -8px 0 18px 112px;
    color: var(--el-text-color-secondary);
    font-size: 12px;
    line-height: 20px;
  }

  .provider-title-line {
    flex-wrap: wrap;
    gap: 10px;
  }

  .provider-title-line h3 {
    margin: 0;
    color: var(--el-text-color-primary);
    font-size: 15px;
    font-weight: 600;
    line-height: 24px;
  }

  .provider-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    column-gap: 24px;
  }

  .provider-grid .grid-full {
    grid-column: 1 / -1;
  }

  .provider-grid :deep(.el-input),
  .provider-grid :deep(.el-input-number) {
    width: 100%;
  }

  .secret-row {
    gap: 14px;
    width: 100%;
  }

  .secret-row :deep(.el-checkbox) {
    flex: 0 0 auto;
    margin-right: 0;
  }

  .input-unit {
    margin-left: 8px;
    color: var(--el-text-color-secondary);
    font-size: 12px;
  }

  .provider-actions {
    justify-content: flex-end;
  }

  @media (max-width: 860px) {
    .recognition-overview {
      align-items: flex-start;
      flex-direction: column;
    }

    .recognition-flow {
      flex-wrap: wrap;
    }

    .recognition-controls {
      align-items: flex-start;
      flex-direction: column;
    }

    .provider-grid {
      grid-template-columns: minmax(0, 1fr);
    }

    .provider-grid .grid-full {
      grid-column: auto;
    }

    .provider-hint {
      margin-left: 0;
    }
  }

  @media (max-width: 560px) {
    .secret-row {
      align-items: flex-start;
      flex-direction: column;
    }
  }
</style>
