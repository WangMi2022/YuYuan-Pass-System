import { readFileSync, readdirSync } from 'fs'

const svgTitle = /<svg([^>+].*?)>/
const clearHeightWidth = /(width|height)="([^>+].*?)"/g
const hasViewBox = /(viewBox="[^>+].*?")/g
const clearReturn = /(\r)|(\n)/g

function findSvgFile(dirs) {
  const svgRes = []
  for (const dir of dirs) {
    const dirents = readdirSync(dir, { withFileTypes: true })
    for (const dirent of dirents) {
      let pluginName = ''
      if (dir.startsWith('./src/plugin')) {
        pluginName = `${dir.split('/')[3]}-`
      }
      if (dirent.isDirectory()) {
        svgRes.push(...findSvgFile([dir + dirent.name + '/']))
        continue
      }
      if (!dirent.name.endsWith('.svg')) continue
      const svg = readFileSync(dir + dirent.name)
        .toString()
        .replace(clearReturn, '')
        .replace(svgTitle, ($1, $2) => {
          let width = 0
          let height = 0
          let content = $2.replace(clearHeightWidth, (s1, s2, s3) => {
            if (s2 === 'width') width = s3
            if (s2 === 'height') height = s3
            return ''
          })
          if (!hasViewBox.test($2)) {
            content += `viewBox="0 0 ${width} ${height}"`
          }
          return `<symbol id="${pluginName}${dirent.name.replace('.svg', '')}" ${content}>`
        })
        .replace('</svg>', '</symbol>')
      svgRes.push(svg)
    }
  }
  return svgRes
}

export function svgBuilder(paths) {
  if (!paths) return undefined
  if (typeof paths === 'string') paths = [paths]
  return {
    name: 'svg-transform',
    transformIndexHtml(html) {
      const res = findSvgFile(paths)
      return html.replace('<body>', `
<body>
  <svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" style="position: absolute; width: 0; height: 0">
    ${res.join('')}
  </svg>
`)
    }
  }
}
