/*
 * Chart theming helpers. ECharts paints to canvas, so CSS variables must be
 * resolved to concrete values at option-build time. Recompute options when
 * appStore.isDark or config.primaryColor changes — both mutate the variables
 * synchronously, so getComputedStyle sees fresh values immediately.
 */

export const cssVar = (name, fallback = '') => {
  if (typeof window === 'undefined') return fallback
  const value = getComputedStyle(document.documentElement).getPropertyValue(name)
  return value ? value.trim() : fallback
}

export const chartPalette = () =>
  [1, 2, 3, 4, 5, 6].map((i) => cssVar(`--na-chart-${i}`, '#6366f1'))

export const chartTheme = () => ({
  label: cssVar('--na-chart-label', '#71717a'),
  grid: cssVar('--na-chart-grid', '#ececee'),
  surface: cssVar('--na-card', '#ffffff'),
  text: cssVar('--na-foreground', '#18181b'),
  muted: cssVar('--na-muted-foreground', '#71717a'),
  primary: cssVar('--na-primary', '#6366f1'),
  success: cssVar('--na-success', '#059669'),
  warning: cssVar('--na-warning', '#d97706'),
  danger: cssVar('--na-danger', '#dc2626'),
  info: cssVar('--na-info', '#0284c7')
})
