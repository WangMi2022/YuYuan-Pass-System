const REPORT_TITLE = '智能日报'

const MIME_TYPES = {
  docx: 'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
  md: 'text/markdown;charset=utf-8',
  xlsx: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet'
}

export const REPORT_EXPORT_FORMATS = [
  { value: 'docx', label: 'Word 文档', extension: '.docx' },
  { value: 'xlsx', label: 'Excel 工作簿', extension: '.xlsx' },
  { value: 'md', label: 'Markdown 文档', extension: '.md' }
]

export function formatMicrosMoney(value) {
  return `¥${(Number(value || 0) / 1000000).toFixed(4)}`
}

const reportDate = (report = {}) => String(report.reportDate || report.generatedAt || '')
  .slice(0, 10)
  .replace(/[^0-9-]/g, '') || new Date().toISOString().slice(0, 10)

const generatedAt = (report = {}) => {
  if (!report.generatedAt) return '未记录'
  const date = new Date(report.generatedAt)
  return Number.isNaN(date.getTime()) ? String(report.generatedAt) : date.toLocaleString('zh-CN', { hour12: false })
}

const normalizeGroups = (groups = []) => groups.map((group) => ({
  label: String(group?.label || '未分类'),
  description: String(group?.description || ''),
  items: Array.isArray(group?.items) ? group.items.map((item) => ({
    label: String(item?.label || '未命名指标'),
    value: String(item?.value ?? '—')
  })) : []
}))

const escapeMarkdownCell = (value) => String(value ?? '')
  .replace(/\|/g, '\\|')
  .replace(/\r?\n/g, '<br>')

const stripInlineMarkdown = (value) => String(value || '')
  .replace(/!\[([^\]]*)\]\([^)]*\)/g, '$1')
  .replace(/\[([^\]]+)\]\([^)]*\)/g, '$1')
  .replace(/`{1,3}([^`]+)`{1,3}/g, '$1')
  .replace(/(\*\*|__)(.*?)\1/g, '$2')
  .replace(/(\*|_)(.*?)\1/g, '$2')
  .replace(/~~(.*?)~~/g, '$1')
  .replace(/<[^>]+>/g, '')
  .trim()

export function buildReportFilename(report, format) {
  const extension = REPORT_EXPORT_FORMATS.find((item) => item.value === format)?.extension
  if (!extension) throw new Error(`不支持的日报格式：${format}`)
  return `${REPORT_TITLE}-${reportDate(report)}${extension}`
}

export function buildReportMarkdown(report = {}, groups = [], options = {}) {
  const generation = options.generationLabel || report.generatedBy || '业务统计'
  const lines = [
    `# ${REPORT_TITLE}`,
    '',
    `- 报告日期：${reportDate(report)}`,
    `- 生成方式：${generation}`,
    `- 生成时间：${generatedAt(report)}`,
    '',
    '## 日报正文',
    '',
    String(report.summary || '日报暂无正文。').trim(),
    '',
    '## 业务指标'
  ]

  normalizeGroups(groups).forEach((group) => {
    lines.push('', `### ${group.label}`)
    if (group.description) lines.push('', group.description)
    lines.push('', '| 指标 | 数值 |', '| --- | ---: |')
    group.items.forEach((item) => lines.push(`| ${escapeMarkdownCell(item.label)} | ${escapeMarkdownCell(item.value)} |`))
  })

  return `${lines.join('\n').trim()}\n`
}

function downloadBlob(blob, filename) {
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  link.style.display = 'none'
  document.body.appendChild(link)
  link.click()
  link.remove()
  window.setTimeout(() => URL.revokeObjectURL(url), 0)
}

async function buildExcelBlob(report, groups, options) {
  const ExcelJSModule = await import('exceljs')
  const ExcelJS = ExcelJSModule.default || ExcelJSModule
  const workbook = new ExcelJS.Workbook()
  workbook.creator = 'mit-assets-admin'
  workbook.created = new Date()
  workbook.modified = new Date()

  const summarySheet = workbook.addWorksheet('日报摘要', {
    views: [{ showGridLines: false }],
    properties: { defaultRowHeight: 22 }
  })
  summarySheet.columns = [{ width: 18 }, { width: 72 }]
  summarySheet.addRows([
    [REPORT_TITLE, reportDate(report)],
    ['生成方式', options.generationLabel || report.generatedBy || '业务统计'],
    ['生成时间', generatedAt(report)],
    ['日报正文', String(report.summary || '日报暂无正文。')]
  ])
  summarySheet.getColumn(1).font = { bold: true, color: { argb: 'FF374151' } }
  summarySheet.getColumn(1).fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFF3F4F6' } }
  summarySheet.getColumn(2).alignment = { vertical: 'top', wrapText: true }
  summarySheet.getRow(1).height = 30
  summarySheet.getRow(4).height = 120
  summarySheet.eachRow((row) => {
    row.eachCell((cell) => {
      cell.border = {
        top: { style: 'thin', color: { argb: 'FFE5E7EB' } },
        left: { style: 'thin', color: { argb: 'FFE5E7EB' } },
        bottom: { style: 'thin', color: { argb: 'FFE5E7EB' } },
        right: { style: 'thin', color: { argb: 'FFE5E7EB' } }
      }
      cell.alignment = { ...cell.alignment, vertical: 'top', wrapText: true }
    })
  })

  const metricsSheet = workbook.addWorksheet('业务指标', {
    views: [{ state: 'frozen', ySplit: 1, showGridLines: false }],
    properties: { defaultRowHeight: 22 }
  })
  metricsSheet.columns = [
    { header: '分类', key: 'group', width: 18 },
    { header: '分类说明', key: 'description', width: 34 },
    { header: '指标', key: 'metric', width: 26 },
    { header: '数值', key: 'value', width: 22 }
  ]
  normalizeGroups(groups).forEach((group) => {
    group.items.forEach((item) => metricsSheet.addRow({
      group: group.label,
      description: group.description,
      metric: item.label,
      value: item.value
    }))
  })
  metricsSheet.getRow(1).font = { bold: true, color: { argb: 'FFFFFFFF' } }
  metricsSheet.getRow(1).fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FF2563EB' } }
  metricsSheet.autoFilter = { from: 'A1', to: 'D1' }
  metricsSheet.eachRow((row, rowNumber) => {
    row.alignment = { vertical: 'center', wrapText: true }
    if (rowNumber > 1 && rowNumber % 2 === 1) {
      row.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFF8FAFC' } }
    }
  })

  const buffer = await workbook.xlsx.writeBuffer()
  return new Blob([buffer], { type: MIME_TYPES.xlsx })
}

const summaryParagraphs = (summary, docx) => {
  const { HeadingLevel, Paragraph, TextRun } = docx
  const lines = String(summary || '日报暂无正文。').split(/\r?\n/)
  return lines.map((source) => {
    const line = source.trim()
    if (!line) return new Paragraph({ spacing: { after: 80 } })
    const heading = line.match(/^(#{1,3})\s+(.+)$/)
    if (heading) {
      const levels = [HeadingLevel.HEADING_1, HeadingLevel.HEADING_2, HeadingLevel.HEADING_3]
      return new Paragraph({
        heading: levels[heading[1].length - 1],
        text: stripInlineMarkdown(heading[2]),
        spacing: { before: 160, after: 80 }
      })
    }
    const bullet = line.match(/^[-*+]\s+(.+)$/)
    if (bullet) {
      return new Paragraph({
        bullet: { level: 0 },
        children: [new TextRun(stripInlineMarkdown(bullet[1]))],
        spacing: { after: 80 }
      })
    }
    return new Paragraph({
      children: [new TextRun(stripInlineMarkdown(line))],
      spacing: { after: 100, line: 360 }
    })
  })
}

async function buildWordBlob(report, groups, options) {
  const docx = await import('docx')
  const {
    AlignmentType,
    BorderStyle,
    Document,
    HeadingLevel,
    Packer,
    Paragraph,
    Table,
    TableCell,
    TableRow,
    TextRun,
    WidthType
  } = docx
  const border = { style: BorderStyle.SINGLE, size: 1, color: 'D1D5DB' }
  const cellBorders = { top: border, right: border, bottom: border, left: border }
  const tableCell = (text, bold = false, shading) => new TableCell({
    borders: cellBorders,
    shading: shading ? { fill: shading } : undefined,
    margins: { top: 100, right: 120, bottom: 100, left: 120 },
    children: [new Paragraph({ children: [new TextRun({ text: String(text), bold })] })]
  })
  const children = [
    new Paragraph({
      heading: HeadingLevel.TITLE,
      alignment: AlignmentType.CENTER,
      children: [new TextRun({ text: REPORT_TITLE, bold: true })],
      spacing: { after: 160 }
    }),
    new Paragraph({
      alignment: AlignmentType.CENTER,
      children: [new TextRun({
        text: `${reportDate(report)}  |  ${options.generationLabel || report.generatedBy || '业务统计'}  |  ${generatedAt(report)}`,
        color: '64748B',
        size: 20
      })],
      spacing: { after: 360 }
    }),
    new Paragraph({ heading: HeadingLevel.HEADING_1, text: '日报正文', spacing: { after: 120 } }),
    ...summaryParagraphs(report.summary, docx),
    new Paragraph({ heading: HeadingLevel.HEADING_1, text: '业务指标', spacing: { before: 240, after: 120 } })
  ]

  normalizeGroups(groups).forEach((group) => {
    children.push(new Paragraph({
      heading: HeadingLevel.HEADING_2,
      children: [new TextRun({ text: group.label, bold: true })],
      spacing: { before: 200, after: 40 }
    }))
    if (group.description) {
      children.push(new Paragraph({
        children: [new TextRun({ text: group.description, color: '64748B', italics: true })],
        spacing: { after: 100 }
      }))
    }
    children.push(new Table({
      width: { size: 100, type: WidthType.PERCENTAGE },
      columnWidths: [5200, 3600],
      rows: [
        new TableRow({ children: [tableCell('指标', true, 'E2E8F0'), tableCell('数值', true, 'E2E8F0')] }),
        ...group.items.map((item) => new TableRow({ children: [tableCell(item.label), tableCell(item.value)] }))
      ]
    }))
  })

  const document = new Document({
    creator: 'mit-assets-admin',
    title: `${REPORT_TITLE}-${reportDate(report)}`,
    description: '智能日报导出文档',
    sections: [{
      properties: {
        page: { margin: { top: 900, right: 900, bottom: 900, left: 900 } }
      },
      children
    }]
  })
  return Packer.toBlob(document)
}

export async function buildReportExportBlob(report, groups, format, options = {}) {
  if (!report) throw new Error('日报数据不存在')
  if (format === 'md') {
    const content = buildReportMarkdown(report, groups, options)
    return new Blob(['\uFEFF', content], { type: MIME_TYPES.md })
  }
  if (format === 'xlsx') return buildExcelBlob(report, groups, options)
  if (format === 'docx') return buildWordBlob(report, groups, options)
  throw new Error(`不支持的日报格式：${format}`)
}

export async function exportSmartReport(report, groups, format, options = {}) {
  const filename = buildReportFilename(report, format)
  const blob = await buildReportExportBlob(report, groups, format, options)
  downloadBlob(blob, filename)
  return filename
}
