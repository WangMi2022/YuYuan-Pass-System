const HEADING_BOUNDARY = /[ \t]+(?=#{1,6}[ \t]+)/g
const LIST_BOUNDARY = /[ \t]+(?=(?:[-*+]|\d{1,2}[.)])[ \t]+(?:\*\*|__|[\u3400-\u9fff]|\d))/g
const HEADING_LINE = /^#{1,6}\s+/
const LIST_LINE = /^\s*(?:[-*+]|\d{1,2}[.)])\s+/

export function normalizeReportSummary(value) {
  const source = String(value ?? '')
    .replace(/\r\n?/g, '\n')
    .replace(/[ \t]+\n/g, '\n')
    .trim()

  if (!source) return ''

  const hasMarkdownStructure = /(^|\n)#{1,6}\s+/.test(source) || /\*\*[^*]+\*\*/.test(source)
  let expanded = source.replace(HEADING_BOUNDARY, '\n\n')
  if (hasMarkdownStructure) expanded = expanded.replace(LIST_BOUNDARY, '\n')

  const lines = []
  let previous = ''
  for (const rawLine of expanded.split('\n')) {
    const line = rawLine.trimEnd()
    const trimmed = line.trim()

    if (!trimmed) {
      if (lines.length && lines.at(-1) !== '') lines.push('')
      continue
    }

    const startsBlock = HEADING_LINE.test(trimmed) || (LIST_LINE.test(line) && !LIST_LINE.test(previous))
    if (startsBlock && lines.length && lines.at(-1) !== '') lines.push('')
    lines.push(line)
    previous = line
  }

  return lines.join('\n').replace(/\n{3,}/g, '\n\n').trim()
}
