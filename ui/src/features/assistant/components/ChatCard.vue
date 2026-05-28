<template>
  <el-row :justify="role === 'user' ? 'end' : 'start'" class="chat-row">
    <el-col :xs="24" :sm="24" :md="24">
      <!-- Layout untuk Assistant -->
      <template v-if="role === 'assistant'">
        <div class="assistant-content-wrapper">
          <!-- Header: Provider & Model Info -->
          <el-space :size="6" alignment="center">
            <el-avatar style="background: transparent;" shape="circle" :size="16" :src="props.providerLogo" fit="fit" />
            <el-text type="info" size="small" strong>
              {{ provider?.toUpperCase() }} > {{ model?.toUpperCase() }}
            </el-text>
          </el-space>
          <!-- Content Body -->
          <div class="content-container">
            <thinking-box :thinking="props.thinking" :done="props.done" />
            <div v-if="state" class="state-text">{{ state }}</div>
            <div v-if="viewMode === 'markdown'" ref="markdownBodyRef" class="markdown-body" v-html="renderedContent" />
            <pre v-else class="raw-content">{{ content }}</pre>
          </div>
          <!-- Actions Footer (Visible on hover when done) -->
          <div v-if="done" class="chat-footer">
            <el-space :size="8">
              <el-tooltip content="Copy Message" placement="top">
                <el-button circle size="small" :icon="DocumentCopy" @click="copyToClipboard" />
              </el-tooltip>
            </el-space>
          </div>
        </div>
      </template>
      <!-- Layout untuk User -->
      <template v-else>
        <div class="user-bubble-wrapper">
          <div v-if="viewMode === 'markdown'" ref="markdownBodyRef" class="markdown-body" v-html="renderedContent" />
          <pre v-else class="raw-content simple">{{ content }}</pre>
        </div>
      </template>
    </el-col>
  </el-row>
</template>
<script setup lang="ts">
import { computed, watch, ref, nextTick } from 'vue'
import { marked } from 'marked'
import hljs from 'highlight.js'
import katex from 'katex'
import 'katex/dist/katex.min.css'
import { DocumentCopy } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import ThinkingBox from './ThinkingBox.vue'
const props = defineProps<{
  role: 'user' | 'assistant'
  content: string
  thinking?: string
  provider?: string
  providerLogo?: string
  model?: string
  done?: boolean
  state?: string
}>()
const markdownBodyRef = ref<HTMLElement | null>(null)
const viewMode = ref<'markdown' | 'raw'>('markdown')
const copyToClipboard = async () => {
  try {
    await navigator.clipboard.writeText(props.content)
    ElMessage.success('Copied to clipboard')
  } catch (err) {
    ElMessage.error('Failed to copy')
  }
}
const renderedContent = computed(() => {
  let text = props.content
  const tokens: { token: string; html: string }[] = []
  const addToken = (raw: string, isBlock: boolean) => {
    const id = tokens.length
    const token = `@@TOKEN_X${id}X@@`
    try {
      const rendered = katex.renderToString(raw, { displayMode: isBlock, throwOnError: false })
      tokens.push({ token, html: isBlock ? `<div class="math-block">${rendered}</div>` : rendered })
      return token
    } catch (e) { return raw }
  }
  // 1. Handle Thought/Think blocks
  text = text.replace(/<(thought|think)>([\s\S]*?)<\/\1>/g, (_, _tag, content) => {
    const token = `@@THOUGHT_TOKEN_${tokens.length}@@`
    tokens.push({ token, html: `<div class="thought-box">${content.trim()}</div>` })
    return token
  })
  // Handle incomplete blocks
  const openTagMatch = text.match(/<(thought|think)>/)
  if (openTagMatch && !text.includes(`</${openTagMatch[1]}>`)) {
    const parts = text.split(openTagMatch[0])
    const thoughtContent = parts.slice(1).join(openTagMatch[0])
    const token = `@@THOUGHT_TOKEN_${tokens.length}@@`
    tokens.push({ token, html: `<div class="thought-box">${thoughtContent.trim()}</div>` })
    text = parts[0] + token
  }
  // 2. LaTeX Environments
  text = text.replace(/\\begin\{([a-z]*\*?)\}([\s\S]+?)\\end\{\1\}/g, (match) => addToken(match, true))
  text = text.replace(/(\$\$|\\\[)([\s\S]+?)(\$\$|\\\])/g, (_, __, math) => addToken(math, true))
  text = text.replace(/((?<!\$)\$|\\\()([^$\n]+?)(\$(?!\$)|\\\))/g, (_, __, math) => addToken(math, false))
  let html = ''
  try {
    marked.setOptions({ gfm: true, breaks: true })
    html = marked.parse(text) as string
  } catch (e) { html = text }
  tokens.forEach(({ token, html: tokenHtml }) => {
    html = html.split(token).join(tokenHtml)
  })
  return html
})
const applyHighlighting = async () => {
  await nextTick()
  if (markdownBodyRef.value && viewMode.value === 'markdown') {
    const blocks = markdownBodyRef.value.querySelectorAll('pre code')
    blocks.forEach((block) => {
      if (block.children.length === 0) {
        block.removeAttribute('data-highlighted')
        hljs.highlightElement(block as HTMLElement)
      }
    })
  }
}
watch(renderedContent, () => { applyHighlighting() }, { immediate: true })
</script>
<style scoped>
.state-text {
  font-size: 13px;
  color: var(--el-text-color-placeholder);
  margin-bottom: 6px;
  animation: pulse 1.2s ease-in-out infinite;
}

@keyframes pulse {

  0%,
  100% {
    opacity: 1;
  }

  50% {
    opacity: 0.4;
  }
}
</style>