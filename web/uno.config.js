import { defineConfig } from '@unocss/vite';
import presetWind3 from '@unocss/preset-wind3';
import transformerDirectives from '@unocss/transformer-directives'

export default defineConfig({
  content: {
    pipeline: {
      include: [
        /[\\/]src[\\/].*\.(?:vue|[jt]sx?|html)(?:\?.*)?$/,
        /[\\/]prototypes[\\/].*\.(?:vue|[jt]sx?|html)(?:\?.*)?$/,
        /[\\/]index\.html(?:\?.*)?$/
      ],
      exclude: [/[\\/]node_modules[\\/]/, /[\\/]dist[\\/]/, /\.css(?:\?.*)?$/]
    }
  },
  theme: {
    colors: {
      // semantic tokens — keep utility classes in sync with the SaaS theme
      primary: 'var(--na-primary)',
      'on-primary': 'var(--na-on-primary)',
      foreground: 'var(--na-foreground)',
      muted: 'var(--na-muted)',
      'muted-foreground': 'var(--na-muted-foreground)',
      card: 'var(--na-card)',
      popover: 'var(--na-popover)',
      background: 'var(--na-background)',
      line: 'var(--na-border)',
      success: 'var(--na-success)',
      warning: 'var(--na-warning)',
      danger: 'var(--na-danger)',
      info: 'var(--na-info)'
    },
    backgroundColor: {
      main: 'var(--na-background)'
    },
    textColor: {
      active: 'var(--el-color-primary)'
    },
    boxShadowColor: {
      active: 'var(--el-color-primary)'
    },
    borderColor: {
      'table-border': 'var(--el-border-color-lighter)'
    }
  },
  shortcuts: {
    'na-shortcut-panel': 'border border-solid border-[var(--na-border)] rounded-[var(--na-radius)] bg-[var(--na-card)] shadow-[var(--na-shadow-sm)]',
    'na-shortcut-toolbar': 'flex items-center flex-wrap gap-2'
  },
  presets: [
    presetWind3({ dark: 'class' })
  ],
  transformers: [
    transformerDirectives(),
  ],
})
