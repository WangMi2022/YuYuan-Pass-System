/*
 * THROWAWAY PROTOTYPE — 10 admin-shell directions, switchable with ?variant=01..10.
 * No API calls and no persistent state. Delete after a design direction is selected.
 */

const variants = [
  { key: '01', name: '秩序蓝图', note: '理性、精密、技术图纸式' },
  { key: '02', name: '资产指挥舱', note: '实时态势与风险优先' },
  { key: '03', name: '财务台账', note: '高密度、可核对、重数据' },
  { key: '04', name: '生命周期轨道', note: '围绕资产状态流转组织' },
  { key: '05', name: '仓库地图', note: '按空间、库区和楼层管理' },
  { key: '06', name: '今日行动队列', note: '以任务而不是指标驱动' },
  { key: '07', name: '数据分析画布', note: '图表优先、面向分析决策' },
  { key: '08', name: '管理层简报', note: '克制、叙事化、结论先行' },
  { key: '09', name: '模块矩阵', note: '高效扫描、功能即信息' },
  { key: '10', name: '命令中心', note: '搜索优先、键盘驱动' }
]

const navItems = [
  ['home', '首页驾驶舱'], ['box', '资产管理'], ['move', '协同办公'],
  ['screen', '监控状态'], ['gear', '系统管理']
]

const assets = [
  ['ZC-2026-0182', 'ThinkPad X1 Carbon', '信息技术部', '在用', '¥12,499'],
  ['ZC-2026-0179', 'Dell PowerEdge R760', '数据中心', '维保', '¥86,200'],
  ['ZC-2026-0168', 'Canon imageRUNNER', '行政中心', '在用', '¥24,680'],
  ['ZC-2026-0154', '海康威视门禁终端', '安全管理部', '调拨中', '¥6,980'],
  ['ZC-2026-0141', 'MacBook Pro 14', '设计中心', '在用', '¥18,999']
]

const tasks = [
  ['09:30', '二季度固定资产盘点', '信息技术部 · 86 项', '进行中'],
  ['11:00', '服务器延保审批', '数据中心 · ¥32,800', '待审批'],
  ['14:20', '新员工设备领用', '人力资源部 · 12 项', '待办理'],
  ['16:00', '办公区资产调拨', 'A 座 → C 座 · 24 项', '待确认']
]

const iconPaths = {
  home: '<path d="M3 11.5 12 4l9 7.5"/><path d="M5.5 10v10h13V10M9 20v-6h6v6"/>',
  box: '<path d="m4 7 8-4 8 4-8 4-8-4Z"/><path d="M4 7v10l8 4 8-4V7M12 11v10"/>',
  move: '<path d="M5 9h11M12 5l4 4-4 4M19 15H8M12 11l-4 4 4 4"/>',
  screen: '<rect x="3" y="4" width="18" height="13" rx="2"/><path d="M8 21h8M12 17v4M7 12l2-2 2 2 4-4 2 2"/>',
  gear: '<circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.7 1.7 0 0 0 .34 1.88l.06.06-2.83 2.83-.06-.06A1.7 1.7 0 0 0 15 19.4a1.7 1.7 0 0 0-1 .6 1.7 1.7 0 0 0-.4 1.1V21H9.6v-.1A1.7 1.7 0 0 0 8.5 19.4a1.7 1.7 0 0 0-1.88.34l-.06.06-2.83-2.83.06-.06A1.7 1.7 0 0 0 4.6 15a1.7 1.7 0 0 0-1.6-1H3v-4h.1A1.7 1.7 0 0 0 4.6 8.5a1.7 1.7 0 0 0-.34-1.88l-.06-.06 2.83-2.83.06.06A1.7 1.7 0 0 0 9 4.6a1.7 1.7 0 0 0 1-1.5V3h4v.1A1.7 1.7 0 0 0 15.5 4.6a1.7 1.7 0 0 0 1.88-.34l.06-.06 2.83 2.83-.06.06A1.7 1.7 0 0 0 19.4 9c.13.4.6 1 1.5 1h.1v4h-.1a1.7 1.7 0 0 0-1.5 1Z"/>',
  search: '<circle cx="11" cy="11" r="7"/><path d="m20 20-4-4"/>',
  bell: '<path d="M18 8a6 6 0 0 0-12 0c0 7-3 7-3 9h18c0-2-3-2-3-9M10 21h4"/>',
  plus: '<path d="M12 5v14M5 12h14"/>',
  arrow: '<path d="M5 12h14M13 6l6 6-6 6"/>',
  check: '<path d="m5 12 4 4L19 6"/>',
  alert: '<path d="M12 3 2.5 20h19L12 3Z"/><path d="M12 9v4M12 17h.01"/>',
  scan: '<path d="M4 8V4h4M16 4h4v4M20 16v4h-4M8 20H4v-4M8 12h8"/>',
  filter: '<path d="M4 5h16M7 12h10M10 19h4"/>',
  chevron: '<path d="m9 18 6-6-6-6"/>',
  more: '<circle cx="5" cy="12" r="1"/><circle cx="12" cy="12" r="1"/><circle cx="19" cy="12" r="1"/>',
  user: '<circle cx="12" cy="8" r="4"/><path d="M4 21a8 8 0 0 1 16 0"/>',
  layers: '<path d="m12 2 9 5-9 5-9-5 9-5Z"/><path d="m3 12 9 5 9-5M3 17l9 5 9-5"/>',
  calendar: '<rect x="3" y="5" width="18" height="16" rx="2"/><path d="M16 3v4M8 3v4M3 10h18"/>'
}

const icon = (name, className = '') => `
  <svg class="ui-icon ${className}" viewBox="0 0 24 24" aria-hidden="true" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round">
    ${iconPaths[name] || iconPaths.box}
  </svg>`

const brand = (compact = false, inverse = false) => `
  <div class="brand ${compact ? 'is-compact' : ''} ${inverse ? 'is-inverse' : ''}">
    <span class="brand-mark"><i></i></span>
    ${compact ? '' : '<span><strong>资产管理中心</strong><small>ASSET CONTROL</small></span>'}
  </div>`

const avatar = (name = '王') => `<span class="avatar" aria-label="当前用户：王管理员">${name}</span>`

const sideNav = (active = 0, compact = false) => `
  <nav class="side-nav" aria-label="主菜单">
    ${navItems.map(([i, label], index) => `
      <button class="side-nav-item ${index === active ? 'is-active' : ''}" type="button" data-action="${label}" title="${label}">
        ${icon(i)}<span>${label}</span>
      </button>`).join('')}
  </nav>
  ${compact ? '' : '<div class="side-foot"><span class="status-dot is-ok"></span><span>服务运行正常</span><small>v2.9.2</small></div>'}`

const status = (text) => {
  const type = text.includes('在用') || text.includes('完成') ? 'ok' : text.includes('维保') || text.includes('审批') ? 'warn' : text.includes('调拨') || text.includes('进行') ? 'info' : 'neutral'
  return `<span class="status status-${type}"><i></i>${text}</span>`
}

const table = (className = '') => `
  <div class="asset-table-wrap ${className}">
    <table class="asset-table">
      <thead><tr><th>资产编号</th><th>资产名称</th><th>使用部门</th><th>状态</th><th>入账原值</th><th></th></tr></thead>
      <tbody>${assets.map(row => `<tr>
        <td><code>${row[0]}</code></td><td><strong>${row[1]}</strong></td><td>${row[2]}</td><td>${status(row[3])}</td><td class="num">${row[4]}</td><td><button class="row-more" type="button" aria-label="查看 ${row[1]}">${icon('more')}</button></td>
      </tr>`).join('')}</tbody>
    </table>
  </div>`

const bars = (values, labels = []) => `<div class="mini-bars" aria-label="资产趋势柱状图">
  ${values.map((v, i) => `<span style="--bar:${v}%"><i></i>${labels[i] ? `<small>${labels[i]}</small>` : ''}</span>`).join('')}
</div>`

const spark = (points = '2,35 18,28 34,31 50,16 66,22 82,8 98,14') => `
  <svg class="spark" viewBox="0 0 100 42" preserveAspectRatio="none" aria-hidden="true">
    <polyline points="${points}" fill="none" stroke="currentColor" stroke-width="2" vector-effect="non-scaling-stroke" />
  </svg>`

const topActions = (dark = false) => `
  <div class="top-actions ${dark ? 'is-dark' : ''}">
    <button type="button" class="icon-button" aria-label="搜索">${icon('search')}</button>
    <button type="button" class="icon-button has-dot" aria-label="通知">${icon('bell')}</button>
    ${avatar()}
  </div>`

function variant01() {
  return `<div class="prototype-canvas v01">
    <aside class="v01-side">
      ${brand()}
      <p class="nav-eyebrow">主控制区 / 01</p>
      ${sideNav(0)}
    </aside>
    <section class="v01-shell">
      <header class="v01-top">
        <div><span>资产中心</span><b>/</b><strong>首页驾驶舱</strong></div>
        <label class="compact-search">${icon('search')}<input aria-label="搜索资产" placeholder="搜索资产、人员或单据" /><kbd>⌘ K</kbd></label>
        ${topActions()}
      </header>
      <main id="prototype-main" class="v01-main">
        <section class="blueprint-hero">
          <div class="blueprint-copy"><span class="coordinate">N 31°14′ / E 121°29′</span><h1>早上好，王管理员</h1><p>今日 23 项业务待处理，资产整体健康度保持稳定。</p><div class="hero-buttons"><button class="primary-button" data-action="登记资产">${icon('plus')}登记资产</button><button class="ghost-button" data-action="资产大屏">查看资产大屏${icon('arrow')}</button></div></div>
          <div class="health-blueprint">
            <div class="crosshair x1"></div><div class="crosshair x2"></div>
            <span class="plate-label">ASSET HEALTH / LIVE</span>
            <strong>96.8<small>%</small></strong>
            <div class="health-line"><i style="width:96.8%"></i></div>
            <p><span class="status-dot is-ok"></span>12,680 件资产在线受控</p>
          </div>
        </section>
        <section class="v01-kpis">
          ${[['资产实物总量','12,680','+2.4%'],['资产档案','4,286','+18'],['资产原值','¥3,286万','+1.8%'],['当前估值','¥2,741万','83.4%']].map((k,i)=>`<article><span>0${i+1} / ${k[0]}</span><div><strong>${k[1]}</strong><em>${k[2]}</em></div>${spark()}</article>`).join('')}
        </section>
        <section class="v01-lower">
          <article class="sheet-panel"><header><div><span>REGISTER / 资产台账</span><h2>最近登记</h2></div><button class="text-button">查看全部 ${icon('arrow')}</button></header>${table()}</article>
          <aside class="sheet-panel task-panel"><header><div><span>QUEUE / TODAY</span><h2>今日待办</h2></div><b>23</b></header>${tasks.slice(0,3).map(t=>`<button class="task-line"><time>${t[0]}</time><span><strong>${t[1]}</strong><small>${t[2]}</small></span>${icon('chevron')}</button>`).join('')}</aside>
        </section>
      </main>
    </section>
  </div>`
}

function variant02() {
  return `<div class="prototype-canvas v02">
    <aside class="v02-rail">
      ${brand(true, true)}
      ${sideNav(3, true)}
      <button class="rail-avatar">WY</button>
    </aside>
    <section class="v02-shell">
      <header class="v02-ticker">
        <strong><span></span>系统态势正常</strong>
        <div><span>在线资产 <b>12,438</b></span><span>告警 <b class="amber">08</b></span><span>同步延迟 <b>12s</b></span><span>最后更新 09:41:28</span></div>
        <button>${icon('bell')}<i></i></button>
      </header>
      <main id="prototype-main" class="v02-main">
        <header class="command-head"><div><p>MONDAY · 20 JUL 2026</p><h1>资产运行指挥舱</h1></div><div class="command-actions"><button>${icon('scan')}扫码识别</button><button class="accent">${icon('plus')}创建任务</button></div></header>
        <section class="command-grid">
          <article class="radar-panel">
            <header><span>全域资产态势</span><b>LIVE</b></header>
            <div class="radar-body">
              <div class="radar"><i class="sweep"></i><span class="blip b1"></span><span class="blip b2"></span><span class="blip b3 warn"></span><span class="blip b4"></span><strong>96.8<small>%</small></strong><em>受控率</em></div>
              <div class="radar-stats"><div><span>正常运行</span><strong>11,982</strong><i style="--w:94%"></i></div><div><span>维保处理中</span><strong>186</strong><i style="--w:28%"></i></div><div><span>调拨途中</span><strong>214</strong><i style="--w:33%"></i></div><div class="is-warn"><span>异常告警</span><strong>08</strong><i style="--w:8%"></i></div></div>
            </div>
          </article>
          <article class="alert-panel"><header><span>风险队列</span><b>8 ALERTS</b></header>
            ${[['严重','机房 UPS 电池组巡检逾期','数据中心 · 逾期 2 天'],['警告','12 台设备维保即将到期','本周内需处理'],['注意','固定资产盘点差异待复核','差异 4 项']].map((a,i)=>`<button class="alert-row severity-${i}"><i></i><span><em>${a[0]}</em><strong>${a[1]}</strong><small>${a[2]}</small></span><time>0${i+1}</time></button>`).join('')}
          </article>
          <article class="flow-panel"><header><span>今日业务流量</span><small>过去 12 小时</small></header>${bars([25,40,34,52,48,72,60,84,68,92,78,66], ['06','','','','10','','','','14','','','18'])}<footer><span>入库 <b>32</b></span><span>领用 <b>48</b></span><span>调拨 <b>26</b></span><span>归还 <b>19</b></span></footer></article>
          <article class="mission-panel"><header><span>当前任务</span><b>23 OPEN</b></header>${tasks.map((t,i)=>`<div class="mission-row"><time>${t[0]}</time><i class="mission-node ${i===0?'active':''}"></i><span><strong>${t[1]}</strong><small>${t[2]}</small></span>${status(t[3])}</div>`).join('')}</article>
        </section>
      </main>
    </section>
  </div>`
}

function variant03() {
  return `<div class="prototype-canvas v03">
    <header class="ledger-top">
      ${brand()}
      <nav><button class="active">总账概览</button><button>资产明细</button><button>折旧管理</button><button>盘点对账</button><button>报表中心</button></nav>
      ${topActions()}
    </header>
    <div class="ledger-sub"><span>账套：华东总部 / 2026 年度</span><span>会计期间：07 月</span><span>数据日期：2026-07-20</span><button>切换账套⌄</button></div>
    <main id="prototype-main" class="v03-main">
      <header class="ledger-title"><div><span>FIXED ASSET LEDGER</span><h1>固定资产总账</h1><p>总账、实物账与责任人台账已于今日 08:40 完成同步。</p></div><div><button>导出月报</button><button class="ledger-primary">+ 新增凭证</button></div></header>
      <section class="ledger-summary">
        <article class="summary-main"><span>资产原值合计</span><strong><small>¥</small>32,864,920.00</strong><p>较上月增加 <b>¥584,320.00</b></p></article>
        <article><span>累计折旧</span><strong>¥5,451,880.00</strong><p>本月计提 ¥184,620.00</p></article>
        <article><span>资产净值</span><strong>¥27,413,040.00</strong><p>净值率 83.41%</p></article>
        <article><span>待对账差异</span><strong class="ledger-red">4 项</strong><p>涉及金额 ¥28,600.00</p></article>
      </section>
      <section class="ledger-grid">
        <article class="ledger-book"><header><div><h2>资产分类账</h2><span>单位：人民币元</span></div><button>查看明细 →</button></header>
          <table><thead><tr><th>资产类别</th><th>数量</th><th>期初原值</th><th>本期增加</th><th>本期减少</th><th>期末原值</th><th>净值率</th></tr></thead><tbody>
          ${[['电子设备','2,486','12,418,200','426,800','18,200','12,826,800','78.2%'],['办公设备','4,210','8,426,120','84,600','32,000','8,478,720','81.6%'],['运输设备','86','6,892,000','0','0','6,892,000','72.4%'],['机械设备','1,248','4,286,400','72,920','0','4,359,320','91.8%'],['其他资产','364','308,080','0','0','308,080','88.5%']].map(r=>`<tr>${r.map((c,i)=>`<td class="${i>0?'num':''}">${c}</td>`).join('')}</tr>`).join('')}
          </tbody><tfoot><tr><th>合计</th><th class="num">8,394</th><th class="num">32,330,800</th><th class="num">584,320</th><th class="num">50,200</th><th class="num">32,864,920</th><th class="num">83.4%</th></tr></tfoot></table>
        </article>
        <aside class="reconcile-note"><span class="paperclip">⌁</span><p>对账批注</p><h3>4 项实物状态与财务账不一致</h3><ul><li><b>2 项</b> 已报废未销账</li><li><b>1 项</b> 调拨部门未更新</li><li><b>1 项</b> 原值录入差异</li></ul><button>开始核对 ${icon('arrow')}</button><small>财务部 · 陈敏<br>今天 09:12 更新</small></aside>
      </section>
    </main>
  </div>`
}

function variant04() {
  const stages = [['登记入库','126','+18'],['领用在途','42','+6'],['正常在用','11,982','94.5%'],['维修保养','186','+12'],['调拨归还','214','-8'],['报废处置','130','+4']]
  return `<div class="prototype-canvas v04">
    <header class="orbit-top">${brand()}<nav><button class="active">工作台</button><button>资产</button><button>流程</button><button>分析</button></nav><label>${icon('search')}<input placeholder="查找资产" /></label>${topActions()}</header>
    <main id="prototype-main" class="v04-main">
      <header class="orbit-welcome"><div><span>2026 年 7 月 20 日 · 星期一</span><h1>让每一件资产，沿正确的轨道流转。</h1></div><button>${icon('plus')}登记新资产</button></header>
      <section class="orbit-layout">
        <article class="orbit-card">
          <header><div><span>资产生命周期</span><strong>12,680 件</strong></div><div class="orbit-legend"><i></i>实时状态</div></header>
          <div class="orbit-stage">
            <div class="orbit-ring ring-one"></div><div class="orbit-ring ring-two"></div><div class="orbit-ring ring-three"></div>
            <div class="orbit-core"><span>全生命周期</span><strong>96.8%</strong><small>流程合规率</small></div>
            ${stages.map((s,i)=>`<button class="orbit-node node-${i+1}"><i>${String(i+1).padStart(2,'0')}</i><span><strong>${s[0]}</strong><b>${s[1]}</b><small>${s[2]}</small></span></button>`).join('')}
          </div>
        </article>
        <aside class="orbit-side">
          <article><header><span>需要你的关注</span><b>8</b></header>${[['维修单超时','3 项 · 最长 2 天'],['盘点差异','4 项 · 等待复核'],['调拨待签收','1 项 · 已到达']].map((x,i)=>`<button><i class="attention-${i}"></i><span><strong>${x[0]}</strong><small>${x[1]}</small></span>${icon('chevron')}</button>`).join('')}</article>
          <article class="quick-create"><span>快速发起</span><div><button>${icon('box')}资产入库</button><button>${icon('move')}资产调拨</button><button>${icon('gear')}维修申请</button><button>${icon('calendar')}盘点计划</button></div></article>
        </aside>
      </section>
      <section class="orbit-foot">${[['本月新增','168','+12.4%'],['流转完成','326','+8.2%'],['平均办理时长','1.8 天','-0.4 天'],['按时完成率','97.2%','+1.6%']].map(x=>`<article><span>${x[0]}</span><strong>${x[1]}</strong><small>${x[2]}</small></article>`).join('')}</section>
    </main>
  </div>`
}

function variant05() {
  return `<div class="prototype-canvas v05">
    <header class="map-top">${brand()}<div class="facility-switch"><span>当前园区</span><strong>上海总部园区⌄</strong></div><label>${icon('search')}<input placeholder="搜索位置、资产或责任人" /></label>${topActions()}</header>
    <aside class="map-side"><span class="side-title">空间目录</span><nav>
      <button class="active"><span><i>A</i><b>A 座 · 行政办公</b></span><em>3,842</em></button>
      <button><span><i>B</i><b>B 座 · 研发中心</b></span><em>4,126</em></button>
      <button><span><i>C</i><b>C 座 · 数据中心</b></span><em>1,864</em></button>
      <button><span><i>W</i><b>中央仓库</b></span><em>2,436</em></button>
      <button><span><i>O</i><b>外部存放点</b></span><em>412</em></button>
    </nav><div class="map-side-summary"><span>园区资产总数</span><strong>12,680</strong><small><i></i> 今日已同步</small></div></aside>
    <main id="prototype-main" class="v05-main">
      <header class="map-head"><div><p>空间资产 / A 座</p><h1>楼层资产分布</h1></div><div><button>列表视图</button><button class="active">地图视图</button><button class="map-add">${icon('plus')}资产上图</button></div></header>
      <section class="map-layout">
        <article class="floor-map">
          <header><div><button>←</button><strong>A 座 · 6F</strong><button>→</button></div><span><i class="dot-blue"></i>在用 <i class="dot-yellow"></i>待处理 <i class="dot-red"></i>异常</span></header>
          <div class="floor-canvas">
            <div class="room r601"><span>601<br><small>会议室</small></span><b>32</b></div><div class="room r602"><span>602<br><small>财务部</small></span><b>86</b></div><div class="room r603"><span>603<br><small>行政部</small></span><b>124</b></div><div class="room r604"><span>604<br><small>人力资源</small></span><b>78</b></div><div class="room r605"><span>605<br><small>共享工位</small></span><b>96</b></div><div class="room r606"><span>606<br><small>档案室</small></span><b>214</b></div>
            <div class="corridor">6F · CENTRAL CORRIDOR</div><span class="asset-pin p1">12</span><span class="asset-pin p2 warn">4</span><span class="asset-pin p3">8</span><span class="asset-pin p4 danger">1</span><div class="north">N ↑</div>
          </div>
          <footer><strong>本层合计 630 件</strong><span>在用 608</span><span>待处理 17</span><span>异常 5</span><button>查看楼层清单 →</button></footer>
        </article>
        <aside class="work-orders"><header><div><span>现场作业</span><strong>今天 · 7 项</strong></div><button>${icon('filter')}</button></header>
          ${[['10:00','602 财务部','打印机更换碳粉','陈伟'],['11:30','606 档案室','盘点标签复核','李倩'],['14:00','603 行政部','办公设备调拨','周航'],['16:30','601 会议室','投影设备巡检','孙超']].map((x,i)=>`<article><time>${x[0]}</time><div><span>${x[1]}</span><strong>${x[2]}</strong><small><i style="--avatar:${i}"></i>${x[3]} · ${i?'待开始':'进行中'}</small></div><button>···</button></article>`).join('')}
          <button class="all-orders">查看全部现场作业 ${icon('arrow')}</button>
        </aside>
      </section>
    </main>
  </div>`
}

function variant06() {
  return `<div class="prototype-canvas v06">
    <aside class="action-rail">${brand(true)}<nav>${navItems.map(([i,l],n)=>`<button class="${n===0?'active':''}" title="${l}">${icon(i)}</button>`).join('')}</nav><button class="action-avatar">王</button></aside>
    <header class="action-top"><div><span>今日</span><strong>7 月 20 日，星期一</strong></div><label>${icon('search')}<input placeholder="搜索" /></label><button>${icon('bell')}<i></i></button></header>
    <main id="prototype-main" class="v06-main">
      <section class="action-intro"><div><span>GOOD MORNING</span><h1>先把重要的事做完。</h1><p>今天有 4 个时间节点，23 项待办，其中 3 项会影响其他部门。</p></div><div class="action-progress"><strong>7<small>/ 23</small></strong><span>今日已完成</span><i><b style="width:30%"></b></i></div></section>
      <section class="action-layout">
        <div class="day-timeline">
          <div class="time-now"><time>09:41</time><span>现在</span></div>
          ${tasks.map((t,i)=>`<article class="timeline-item ${i===0?'is-current':''}"><time>${t[0]}</time><div class="timeline-dot"><i></i></div><div class="timeline-card"><header><span>${i===0?'正在处理':i===1?'需你审批':'待开始'}</span>${status(t[3])}</header><h2>${t[1]}</h2><p>${t[2]}</p><div class="timeline-meta"><span>${icon(i===0?'calendar':i===1?'user':'box')}${i===0?'截止 11:30':i===1?'申请人：张强':'关联 12 件资产'}</span><span class="people"><i>陈</i><i>李</i><i>+2</i></span></div><footer><button class="outline">查看详情</button><button class="action-primary">${i===1?'去审批':'开始处理'} ${icon('arrow')}</button></footer></div></article>`).join('')}
        </div>
        <aside class="impact-side">
          <section><span>工作影响</span><strong>今天完成后</strong><div class="impact-score"><b>+18%</b><small>流程准时率</small>${spark('2,34 16,32 30,36 44,22 58,26 72,15 86,18 98,6')}</div><ul><li><i></i>释放 12 台待领用设备</li><li><i></i>关闭 3 个维保告警</li><li><i></i>完成 IT 部门季度盘点</li></ul></section>
          <section class="action-shortcuts"><span>随手处理</span>${[['check','确认 6 条签收','6'],['move','审核 3 条调拨','3'],['alert','复核 4 项差异','4']].map(x=>`<button>${icon(x[0])}<strong>${x[1]}</strong><b>${x[2]}</b></button>`).join('')}</section>
        </aside>
      </section>
    </main>
  </div>`
}

function variant07() {
  return `<div class="prototype-canvas v07">
    <aside class="analytics-side">${brand()}<div class="analytics-workspace"><span>工作区</span><button><i>HQ</i><strong>总部资产中心</strong>⌄</button></div>${sideNav(3)}<button class="collapse-side">« 收起导航</button></aside>
    <section class="v07-shell"><header class="analytics-top"><div><h1>数据分析画布</h1><span>数据更新于 09:40</span></div><div class="period-tabs"><button>今日</button><button>本月</button><button class="active">本季度</button><button>本年度</button></div><button class="export">导出报告</button>${topActions(true)}</header>
    <main id="prototype-main" class="v07-main">
      <section class="analytics-kpis">${[['资产总量','12,680','+2.4%','72'],['资产原值','¥3,286万','+1.8%','58'],['资产净值率','83.4%','-0.6%','44'],['闲置资产','126','-12.5%','30']].map((k,i)=>`<article><header><span>${k[0]}</span><i class="kpi-color c${i}"></i></header><strong>${k[1]}</strong><footer><b class="${i===2?'down':''}">${k[2]}</b><span>较上季度</span></footer>${spark(i===2?'2,8 18,12 34,10 50,18 66,20 82,28 98,30':undefined)}</article>`).join('')}</section>
      <section class="analytics-grid">
        <article class="value-chart"><header><div><span>资产价值趋势</span><small>原值 / 净值 · 单位万元</small></div><button>按月⌄</button></header><div class="line-chart"><div class="axis-y"><span>4000</span><span>3000</span><span>2000</span><span>1000</span><span>0</span></div><div class="grid-lines"><i></i><i></i><i></i><i></i><i></i><svg viewBox="0 0 600 220" preserveAspectRatio="none"><defs><linearGradient id="fill07" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stop-color="#00a8e8" stop-opacity=".28"/><stop offset="1" stop-color="#00a8e8" stop-opacity="0"/></linearGradient></defs><path class="area" d="M0 170 C70 150 100 158 150 122 S245 110 300 96 S400 85 450 52 S540 56 600 26 L600 220 L0 220Z"/><path class="line-a" d="M0 170 C70 150 100 158 150 122 S245 110 300 96 S400 85 450 52 S540 56 600 26"/><path class="line-b" d="M0 188 C70 178 100 180 150 154 S245 145 300 130 S400 120 450 90 S540 96 600 72"/></svg><div class="axis-x"><span>Q3 24</span><span>Q4 24</span><span>Q1 25</span><span>Q2 25</span><span>Q3 25</span><span>Q4 25</span><span>Q1 26</span><span>Q2 26</span></div></div></div><footer><span><i class="cyan"></i>资产原值 <b>¥3,286万</b></span><span><i class="purple"></i>资产净值 <b>¥2,741万</b></span></footer></article>
        <article class="category-chart"><header><span>分类构成</span><button>···</button></header><div class="donut-07"><div><strong>12,680</strong><span>资产总量</span></div></div><ul>${[['电子设备','38.6%','c1'],['办公设备','31.2%','c2'],['机械设备','17.8%','c3'],['运输设备','8.4%','c4'],['其他','4.0%','c5']].map(x=>`<li><i class="${x[2]}"></i><span>${x[0]}</span><b>${x[1]}</b></li>`).join('')}</ul></article>
        <article class="department-bars"><header><span>部门资产排行</span><small>按资产原值</small></header>${[['信息技术部','824万',88],['研发中心','686万',72],['生产运营部','542万',58],['行政中心','328万',35],['市场中心','246万',26]].map((x,i)=>`<div><span>${i+1}</span><strong>${x[0]}</strong><i><b style="width:${x[2]}%"></b></i><em>${x[1]}</em></div>`).join('')}</article>
        <article class="state-matrix"><header><span>状态矩阵</span><small>数量 / 环比</small></header><div>${[['正常在用','11,982','+1.2%'],['维修保养','186','+8.4%'],['闲置封存','126','-12.5%'],['调拨途中','214','+3.1%'],['报废处置','130','-2.8%'],['异常告警','8','+2']].map((x,i)=>`<button class="s${i}"><span>${x[0]}</span><strong>${x[1]}</strong><small>${x[2]}</small></button>`).join('')}</div></article>
      </section>
    </main></section>
  </div>`
}

function variant08() {
  return `<div class="prototype-canvas v08">
    <header class="brief-top">${brand()}<nav><button class="active">晨间简报</button><button>经营视图</button><button>资产专题</button><button>管理报告</button></nav><div><span>2026 / 07 / 20</span>${avatar()}</div></header>
    <main id="prototype-main" class="v08-main">
      <header class="brief-mast"><span>ASSET MANAGEMENT BRIEF · 第 201 期</span><h1>资产管理晨间简报</h1><p>供管理层快速掌握资产规模、经营效率与今日风险</p></header>
      <section class="brief-lead">
        <div class="brief-number"><span>当前资产净值</span><strong><small>¥</small>27,413,040</strong><p>净值率 83.41%，较年初下降 2.6 个百分点，处于预算控制范围内。</p></div>
        <div class="brief-verdict"><span>今日结论</span><h2>整体运行稳健，<br>三项风险需要关注。</h2><p>二季度盘点完成率已达 92%，但数据中心维保逾期、4 项账实差异及 12 台设备续保将在本周形成风险窗口。</p><button>阅读完整分析 ${icon('arrow')}</button></div>
      </section>
      <section class="brief-rule"><span>关键指标</span><i></i><small>截至今日 09:40</small></section>
      <section class="brief-metrics">${[['01','资产总量','12,680','较上月 +2.4%'],['02','本月新增','168','原值 ¥584.3万'],['03','闲置资产','126','环比下降 12.5%'],['04','盘点完成率','92%','剩余 684 项']].map(x=>`<article><span>${x[0]}</span><p>${x[1]}</p><strong>${x[2]}</strong><small>${x[3]}</small></article>`).join('')}</section>
      <section class="brief-columns">
        <article class="brief-story"><span class="story-label">运营观察</span><h2>设备流转效率提升，但跨部门签收仍是主要等待点</h2><p class="dropcap">本</p><p>月累计完成资产流转 326 次，平均办理时长由 2.2 天缩短至 1.8 天。流程节点数据显示，审批阶段已明显提速，等待主要集中在跨部门签收。</p><div class="story-chart">${bars([32,42,38,52,58,66,61,74,82,78,88,92],['8日','','10日','','12日','','14日','','16日','','18日','20日'])}</div><footer><span>流转按时完成率</span><strong>97.2%</strong><small>+1.6%</small></footer></article>
        <article class="brief-risks"><span class="story-label">风险备忘</span>${[['01','机房 UPS 电池组巡检逾期','建议今日 12:00 前确认服务商到场时间。'],['02','固定资产账实差异 4 项','涉及金额 ¥28,600，需财务与行政联合复核。'],['03','12 台核心设备维保即将到期','续保预算已提交，等待财务负责人审批。']].map(x=>`<div><b>${x[0]}</b><span><strong>${x[1]}</strong><p>${x[2]}</p></span></div>`).join('')}<button>查看全部风险事项 →</button></article>
        <aside class="brief-agenda"><span class="story-label">今日议程</span>${tasks.slice(0,3).map(x=>`<div><time>${x[0]}</time><span><strong>${x[1]}</strong><small>${x[2]}</small></span></div>`).join('')}<blockquote>“让资产数据成为经营决策的可靠底稿。”</blockquote></aside>
      </section>
    </main>
  </div>`
}

function variant09() {
  return `<div class="prototype-canvas v09">
    <aside class="matrix-dock">${brand(true,true)}<nav>${navItems.map(([i,l],n)=>`<button class="${n===0?'active':''}" title="${l}">${icon(i)}<span>${l.slice(0,2)}</span></button>`).join('')}</nav><button>${icon('gear')}<span>设置</span></button></aside>
    <section class="v09-shell"><header class="matrix-top"><div><p>资产管理中心</p><h1>工作矩阵</h1></div><label>${icon('search')}<input placeholder="搜索全部模块" /></label><div class="matrix-top-actions"><button>${icon('bell')}<b>8</b></button>${avatar()}</div></header>
    <main id="prototype-main" class="v09-main">
      <section class="matrix-greeting"><div><span>MON / 20 JUL</span><strong>上午好，王管理员</strong><p>把工作拆成模块，一眼看到哪里需要你。</p></div><button>${icon('plus')}新建业务</button></section>
      <section class="module-matrix">
        <button class="matrix-card card-total"><span>资产总量</span><strong>12,680</strong><small>4,286 份资产档案</small><div>${spark()}</div><i>+2.4%</i></button>
        <button class="matrix-card card-value"><span>资产净值</span><strong>¥2,741<small>万</small></strong><p>原值 ¥3,286 万</p><div class="value-meter"><i style="width:83%"></i></div><em>净值率 83.4%</em></button>
        <button class="matrix-card card-tasks"><header><span>待办任务</span><b>23</b></header>${tasks.slice(0,3).map(x=>`<div><time>${x[0]}</time><span>${x[1]}</span><i></i></div>`).join('')}<footer>打开任务中心 →</footer></button>
        <button class="matrix-card card-inventory"><span>${icon('box')}</span><strong>资产档案</strong><small>登记、领用与维护</small><b>4,286</b></button>
        <button class="matrix-card card-transfer"><span>${icon('move')}</span><strong>资产调拨</strong><small>214 件正在流转</small><b>26 待签收</b></button>
        <button class="matrix-card card-maintain"><span>${icon('gear')}</span><strong>维修维保</strong><small>186 件处理中</small><b>3 项超时</b></button>
        <button class="matrix-card card-screen"><span>${icon('screen')}</span><strong>资产大屏</strong><small>全域结构与价值</small><div class="screen-bars"><i></i><i></i><i></i><i></i><i></i></div></button>
        <button class="matrix-card card-alert"><span>${icon('alert')}</span><strong>异常告警</strong><small>需要立即处理</small><b>08</b></button>
        <button class="matrix-card card-quick"><header><span>快捷操作</span><i>⌘ K</i></header><div><span>${icon('scan')}扫码识别</span><span>${icon('plus')}登记资产</span><span>${icon('calendar')}创建盘点</span><span>${icon('user')}人员交接</span></div></button>
        <button class="matrix-card card-notice"><header><span>最新公告</span><b>2 未读</b></header><strong>关于开展 2026 年第二季度固定资产盘点工作的通知</strong><small>资产管理部 · 30 分钟前</small><footer>下一条：数据中心维保窗口调整</footer></button>
      </section>
    </main></section>
  </div>`
}

function variant10() {
  return `<div class="prototype-canvas v10">
    <header class="cmd-top">${brand()}<div class="cmd-context"><span>工作空间</span><b>总部资产中心⌄</b></div><div class="cmd-top-right"><button>${icon('bell')}<i></i></button>${avatar()}</div></header>
    <main id="prototype-main" class="v10-main">
      <section class="cmd-intro"><p>早上好，王管理员</p><h1>今天想处理什么？</h1><label class="command-search">${icon('search')}<input autofocus placeholder="搜索资产，或输入一个操作…" /><kbd>⌘ K</kbd></label><div class="command-hints"><span>试试：</span><button>查看待审批的调拨</button><button>登记一台新设备</button><button>导出本月资产报表</button></div></section>
      <section class="cmd-layout">
        <article class="recent-work"><header><div><span>继续处理</span><small>最近打开</small></div><button>查看全部</button></header>
          ${[['clipboard','二季度固定资产盘点','盘点计划 · 完成 92%','18 分钟前'],['move','A 座至 C 座设备调拨','调拨单 DB-202607-018','昨天 16:42'],['box','ThinkPad X1 Carbon','资产 ZC-2026-0182','昨天 14:10']].map((x,i)=>`<button class="recent-row"><span class="recent-icon i${i}">${icon(x[0] === 'clipboard'?'calendar':x[0])}</span><span><strong>${x[1]}</strong><small>${x[2]}</small></span><time>${x[3]}</time>${icon('chevron')}</button>`).join('')}
        </article>
        <aside class="cmd-today"><header><span>今日概览</span><b>7 / 23 已完成</b></header><div class="today-ring"><strong>30%</strong></div><ul><li><i class="red"></i><span>需立即处理</span><b>3</b></li><li><i class="blue"></i><span>待我审批</span><b>6</b></li><li><i class="gray"></i><span>其他待办</span><b>14</b></li></ul><button>打开我的待办 ${icon('arrow')}</button></aside>
      </section>
      <section class="cmd-suggestions"><header><span>常用操作</span><small>输入命令可直接执行</small></header><div>
        ${[['plus','登记资产','创建新的资产档案','A'],['scan','扫码查资产','使用编号快速定位','S'],['move','发起调拨','在部门间流转资产','T'],['calendar','创建盘点','按部门或区域盘点','P'],['gear','提交维修','创建维修维保工单','M'],['screen','打开大屏','查看资产结构与价值','D']].map(x=>`<button><span>${icon(x[0])}</span><div><strong>${x[1]}</strong><small>${x[2]}</small></div><kbd>${x[3]}</kbd></button>`).join('')}
      </div></section>
      <footer class="cmd-footer"><span><i class="status-dot is-ok"></i>所有系统运行正常</span><span>12,680 件资产 · 数据更新于 09:40</span><button>查看系统状态</button></footer>
    </main>
  </div>`
}

const renderers = {
  '01': variant01, '02': variant02, '03': variant03, '04': variant04, '05': variant05,
  '06': variant06, '07': variant07, '08': variant08, '09': variant09, '10': variant10
}

function getVariant() {
  const raw = new URLSearchParams(window.location.search).get('variant') || '01'
  return variants.some(item => item.key === raw) ? raw : '01'
}

function switcher(current) {
  const info = variants.find(item => item.key === current)
  return `<aside class="prototype-switcher" aria-label="设计方案切换器">
    <button class="switch-arrow" type="button" data-switch="prev" aria-label="上一套方案">←</button>
    <div class="switch-current" aria-live="polite"><span>后台重设计提案</span><strong>${current} — ${info.name}</strong><small>${info.note}</small></div>
    <div class="switch-dots" role="tablist" aria-label="10 套设计方案">
      ${variants.map(v => `<button type="button" role="tab" aria-selected="${v.key === current}" class="${v.key === current ? 'active' : ''}" data-variant="${v.key}" title="${v.key} ${v.name}">${v.key}</button>`).join('')}
    </div>
    <button class="switch-arrow" type="button" data-switch="next" aria-label="下一套方案">→</button>
  </aside>`
}

function render() {
  const current = getVariant()
  const info = variants.find(item => item.key === current)
  document.documentElement.dataset.prototype = current
  document.title = `${current} ${info.name} · 后台重设计提案`
  document.querySelector('#app').innerHTML = `${renderers[current]()}${switcher(current)}<div class="prototype-toast" role="status" aria-live="polite"></div>`
}

function setVariant(key) {
  const url = new URL(window.location.href)
  url.searchParams.set('variant', key)
  history.replaceState({}, '', url)
  render()
  window.scrollTo({ top: 0, behavior: 'instant' })
}

function cycle(direction) {
  const current = variants.findIndex(item => item.key === getVariant())
  const next = (current + direction + variants.length) % variants.length
  setVariant(variants[next].key)
}

document.addEventListener('click', (event) => {
  const direct = event.target.closest('[data-variant]')
  if (direct) return setVariant(direct.dataset.variant)
  const arrow = event.target.closest('[data-switch]')
  if (arrow) return cycle(arrow.dataset.switch === 'next' ? 1 : -1)
  const action = event.target.closest('[data-action]')
  if (action) {
    const toast = document.querySelector('.prototype-toast')
    toast.textContent = `原型交互：${action.dataset.action}`
    toast.classList.add('show')
    window.clearTimeout(window.prototypeToastTimer)
    window.prototypeToastTimer = window.setTimeout(() => toast.classList.remove('show'), 1600)
  }
})

document.addEventListener('keydown', (event) => {
  const tag = event.target.tagName
  if (['INPUT', 'TEXTAREA'].includes(tag) || event.target.isContentEditable) return
  if (event.key === 'ArrowLeft') cycle(-1)
  if (event.key === 'ArrowRight') cycle(1)
})

window.addEventListener('popstate', render)
render()
