<template>
  <div v-if="thinking" class="thought-box">
    <el-collapse v-model="activeNames">
      <el-collapse-item name="1">
        <template #title>
          <div class="thought-header">
            <el-icon :class="{ 'is-loading': !done }" size="14">
              <Loading />
            </el-icon>
            <span>THINKING</span>
          </div>
        </template>
        <div class="markdown-body" v-html="renderedThinking" />
      </el-collapse-item>
    </el-collapse>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { marked } from 'marked'
import { Loading } from '@element-plus/icons-vue'

const props = defineProps<{
  thinking?: string
  done?: boolean
}>()

const activeNames = ref([])

const renderedThinking = computed(() => {
  if (!props.thinking) return ''
  try {
    marked.setOptions({ gfm: true, breaks: true })
    return marked.parse(props.thinking) as string
  } catch (e) {
    return props.thinking
  }
})
</script>

<style scoped>
.thought-box {
  color: var(--el-text-color-placeholder);
  font-style: italic;
  background: var(--el-fill-color-lighter);
  padding: 4px 10px !important;
  border-radius: 12px;
  margin: 12px 0;
  font-size: 13px;
  border-left: 4px solid var(--el-border-color-lighter);
  line-height: 1.5;
  overflow-wrap: break-word;
}

.thought-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 11px;
  font-weight: 600;
  color: var(--el-text-color-placeholder);
  letter-spacing: 0.5px;
  user-select: none;
}

:deep(.el-collapse) {
  border: none !important;
  background: transparent !important;
}

:deep(.el-collapse-item) {
  border: none !important;
  background: transparent !important;
  margin: 0 !important;
  padding: 0 !important;
}

:deep(.el-collapse-item__header) {
  border: none !important;
  background: transparent !important;
  height: auto !important;
  line-height: normal !important;
  padding: 0 !important;
  margin: 0 !important;
  color: inherit !important;
  font-style: inherit !important;
}

:deep(.el-collapse-item__arrow) {
  margin: 0 0 0 6px !important;
  color: var(--el-text-color-placeholder) !important;
}

:deep(.el-collapse-item__wrap) {
  border: none !important;
  background: transparent !important;
  margin: 0 !important;
  padding: 0 !important;
}

:deep(.el-collapse-item__content) {
  padding-top: 4px !important;
  padding-bottom: 0 !important;
  padding-left: 0 !important;
  padding-right: 0 !important;
  margin: 0 !important;
  color: inherit !important;
  font-style: inherit !important;
}

/* Force headings and text inside thinking box to have same size and color */
:deep(.markdown-body) {
  font-size: 13px !important;
  color: inherit !important;
}

:deep(.markdown-body h1),
:deep(.markdown-body h2),
:deep(.markdown-body h3),
:deep(.markdown-body h4),
:deep(.markdown-body h5),
:deep(.markdown-body h6) {
  font-size: 13px !important;
  color: inherit !important;
  margin-top: 12px;
  margin-bottom: 6px;
}
</style>