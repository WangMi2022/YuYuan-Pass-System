import assert from 'node:assert/strict'
import test from 'node:test'
import ExcelJS from 'exceljs'
import mammoth from 'mammoth'
import { buildReportExportBlob, buildReportFilename, buildReportMarkdown, formatMicrosMoney } from './reportExport.js'
import { normalizeReportSummary, reportSummaryBody, reportSummaryHeading } from './reportSummary.js'

const report = {
  reportDate: '2026-08-19T00:00:00+08:00',
  generatedAt: '2026-08-19T09:30:00+08:00',
  generatedBy: 'deterministic+model',
  summary: '## 重点\n\n- 待处理风险 2 项'
}

const groups = [{
  label: '风险处置',
  description: '开放风险及当日处理进度',
  items: [{ label: '开放风险', value: '2' }, { label: '当日处理', value: '1' }]
}]

test('buildReportFilename uses the report date and selected Office extension', () => {
  assert.equal(buildReportFilename(report, 'docx'), '智能日报-2026-08-19.docx')
  assert.equal(buildReportFilename(report, 'xlsx'), '智能日报-2026-08-19.xlsx')
  assert.equal(buildReportFilename(report, 'md'), '智能日报-2026-08-19.md')
  assert.throws(() => buildReportFilename(report, 'pdf'), /不支持的日报格式/)
})

test('formatMicrosMoney converts internal cost micros to a readable yuan amount', () => {
  assert.equal(formatMicrosMoney(25824), '¥0.0258')
  assert.equal(formatMicrosMoney(null), '¥0.0000')
})

test('buildReportMarkdown contains metadata, full summary and grouped metrics', () => {
  const markdown = buildReportMarkdown(report, groups, { generationLabel: '业务统计 + AI 摘要' })
  assert.match(markdown, /^# 智能日报/m)
  assert.match(markdown, /报告日期：2026-08-19/)
  assert.match(markdown, /生成方式：业务统计 \+ AI 摘要/)
  assert.match(markdown, /## 重点/)
  assert.match(markdown, /### 风险处置/)
  assert.match(markdown, /\| 开放风险 \| 2 \|/)
})

test('normalizeReportSummary restores headings and lists from compressed model output', () => {
  const compressed = '# 今日智能日报 ## 一、资产运营 - **今日动态**：新增资产 0 项。 - **存量情况**：在用长期资产 51 项。 - **待办提醒**： - 待入库 9 项 - 维护逾期 10 项 ## 二、风险管控 - **当前开放风险**：1820 条'
  const normalized = normalizeReportSummary(compressed)

  assert.match(normalized, /^# 今日智能日报\n\n## 一、资产运营/m)
  assert.match(normalized, /\n\n- \*\*今日动态\*\*：新增资产 0 项。/)
  assert.match(normalized, /\n- 待入库 9 项/)
  assert.match(normalized, /\n\n## 二、风险管控\n\n- \*\*当前开放风险\*\*：1820 条$/)
  assert.doesNotMatch(normalized, /## .+ ##/)
})

test('normalizeReportSummary preserves already formatted markdown and prose hyphens', () => {
  const formatted = '# 今日智能日报\n\n## 一、资产运营\n\n- **今日动态**：新增资产 0 项。\n- 待入库 9 项'
  assert.equal(normalizeReportSummary(formatted), formatted)
  assert.equal(normalizeReportSummary('方案 A - 方案 B'), '方案 A - 方案 B')
})

test('report summary keeps a visible heading when the formatted body is unavailable', () => {
  const summary = '# 今日智能日报\n\n## 一、资产运营\n\n- **今日新增**：0 项'
  assert.equal(reportSummaryHeading(summary), '今日智能日报')
  assert.equal(reportSummaryBody(summary), '## 一、资产运营\n\n- **今日新增**：0 项')
  assert.equal(reportSummaryHeading('资产运营正常'), '今日业务摘要')
  assert.equal(reportSummaryBody('资产运营正常'), '资产运营正常')
})

test('Office exports create real docx and xlsx zip packages', async () => {
  const [word, excel] = await Promise.all([
    buildReportExportBlob(report, groups, 'docx', { generationLabel: '业务统计 + AI 摘要' }),
    buildReportExportBlob(report, groups, 'xlsx', { generationLabel: '业务统计 + AI 摘要' })
  ])
  assert.equal(word.type, 'application/vnd.openxmlformats-officedocument.wordprocessingml.document')
  assert.equal(excel.type, 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet')
  const wordBuffer = await word.arrayBuffer()
  const excelBuffer = await excel.arrayBuffer()
  assert.deepEqual([...new Uint8Array(wordBuffer).slice(0, 2)], [0x50, 0x4b])
  assert.deepEqual([...new Uint8Array(excelBuffer).slice(0, 2)], [0x50, 0x4b])

  const wordResult = await mammoth.extractRawText({ buffer: Buffer.from(wordBuffer) })
  assert.match(wordResult.value, /智能日报/)
  assert.match(wordResult.value, /开放风险/)

  const workbook = new ExcelJS.Workbook()
  await workbook.xlsx.load(excelBuffer)
  assert.deepEqual(workbook.worksheets.map((sheet) => sheet.name), ['日报摘要', '业务指标'])
  assert.equal(workbook.getWorksheet('业务指标').getCell('C2').value, '开放风险')
})
