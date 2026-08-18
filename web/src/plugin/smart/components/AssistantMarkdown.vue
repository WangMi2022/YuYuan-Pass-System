<template>
  <div class="assistant-markdown" v-html="html" />
</template>

<script setup>
import { computed } from 'vue'
import { marked } from 'marked'

const props = defineProps({
  source: {
    type: String,
    default: ''
  }
})

const escapeHTML = (value = '') => String(value)
  .replace(/&/g, '&amp;')
  .replace(/</g, '&lt;')
  .replace(/>/g, '&gt;')
  .replace(/"/g, '&quot;')
  .replace(/'/g, '&#39;')

const safeHref = (value = '') => {
  const href = String(value).trim()
  return /^(https?:|mailto:|\/|#)/i.test(href) ? href : ''
}

const html = computed(() => marked.parse(props.source || '', {
  async: false,
  breaks: true,
  gfm: true,
  renderer: {
    html: ({ text }) => escapeHTML(text),
    image: ({ text }) => escapeHTML(text),
    link({ href, tokens }) {
      const text = this.parser.parseInline(tokens)
      const target = safeHref(href)
      if (!target) return text
      return `<a href="${escapeHTML(target)}" target="_blank" rel="noopener noreferrer">${text}</a>`
    }
  }
}))
</script>

<style scoped lang="scss">
.assistant-markdown {
  min-width: 0;
  color: var(--na-foreground);
  font-size: .875rem;
  line-height: 1.72;
  overflow-wrap: anywhere;
}

.assistant-markdown :deep(p),
.assistant-markdown :deep(ul),
.assistant-markdown :deep(ol),
.assistant-markdown :deep(blockquote),
.assistant-markdown :deep(pre),
.assistant-markdown :deep(table) { margin: 0 0 var(--na-space-sm); }

.assistant-markdown :deep(p:last-child),
.assistant-markdown :deep(ul:last-child),
.assistant-markdown :deep(ol:last-child),
.assistant-markdown :deep(blockquote:last-child),
.assistant-markdown :deep(pre:last-child),
.assistant-markdown :deep(table:last-child) { margin-bottom: 0; }

.assistant-markdown :deep(h1),
.assistant-markdown :deep(h2),
.assistant-markdown :deep(h3),
.assistant-markdown :deep(h4) {
  margin: var(--na-space-md) 0 var(--na-space-xs);
  color: var(--na-foreground);
  font-size: .9375rem;
  font-weight: 700;
  line-height: 1.45;
}

.assistant-markdown :deep(ul),
.assistant-markdown :deep(ol) { padding-left: 1.35rem; }
.assistant-markdown :deep(li + li) { margin-top: var(--na-space-2xs); }

.assistant-markdown :deep(a) {
  color: var(--na-primary);
  text-decoration: underline;
  text-underline-offset: 2px;
}

.assistant-markdown :deep(code) {
  padding: 1px 4px;
  border-radius: 4px;
  background: var(--na-muted);
  color: var(--na-foreground);
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: .82em;
}

.assistant-markdown :deep(pre) {
  overflow: auto;
  padding: var(--na-space-sm);
  border: 1px solid var(--na-border);
  border-radius: var(--na-radius-sm);
  background: var(--na-muted);
}

.assistant-markdown :deep(pre code) { padding: 0; background: transparent; }

.assistant-markdown :deep(blockquote) {
  padding: var(--na-space-xs) var(--na-space-sm);
  border: 1px solid var(--na-border);
  border-radius: var(--na-radius-sm);
  background: var(--na-muted);
  color: var(--na-muted-foreground);
}

.assistant-markdown :deep(table) {
  display: block;
  max-width: 100%;
  overflow: auto;
  border-collapse: collapse;
}

.assistant-markdown :deep(th),
.assistant-markdown :deep(td) {
  min-width: 96px;
  padding: 6px 8px;
  border: 1px solid var(--na-border);
  text-align: left;
  vertical-align: top;
}

.assistant-markdown :deep(th) { background: var(--na-muted); font-weight: 650; }
</style>
