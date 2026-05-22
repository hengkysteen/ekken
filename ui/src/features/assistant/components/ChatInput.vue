<template>
  <el-card shadow="never" class="chat-input-card" :body-style="{ padding: '12px' }">
    <div class="input-container">
      <!-- Seamless Textarea: No borders, autosize for multi-line -->
      <el-input v-model="text" type="textarea" :autosize="{ minRows: 1, maxRows: 12 }"
        :placeholder="`Message (${store.selectedAgent})...`" @keydown.enter.exact.prevent="onSend"
        class="seamless-input" />
      <!-- Toolbar: Bottom row for all menus and actions -->
      <el-row justify="space-between" align="middle" class="input-toolbar">
        <!-- Left: Mode & Thinking Selectors -->
        <el-col :span="10">
          <el-space>
            <el-select v-model="store.selectedAgent" size="small" class="mode-select" popper-class="mode-select-popper" style="width: 100px">
              <el-option v-for="a in store.agents" :key="a.name" :label="a.name.toUpperCase()" :value="a.name" />
            </el-select>
            <!-- <el-select v-model="store.selectedThinking" size="small" class="mode-select" popper-class="mode-select-popper" style="width: 100px">
              <el-option v-for="t in ['high', 'medium', 'low', 'none']" :key="t" :label="t.toUpperCase()" :value="t" />
            </el-select> -->
          </el-space>
        </el-col>
        <!-- Right: Model Selection & Final Actions -->
        <el-col :span="14" style="text-align: right">
          <el-space :size="8">
            <!-- Cascader for Provider & Model -->
            <el-cascader :key="cascaderKey" v-model="cascaderValue" :props="cascaderProps" :show-all-levels="false"
              placeholder="Select Provider / Model" size="small" class="model-cascader"
              popper-class="model-cascader-popper">
              <template #header>
                <div class="cascader-header">
                  <span class="header-item">SELECT MODEL</span>
                </div>
              </template>
              <template #default="{ data }">
                <el-row justify="space-between" align="middle" style="width: 100%; min-width: 150px">
                  <span>{{ data.label }}</span>
                  <el-icon v-if="data.leaf && isFavorite(data.value)" color="orange" style="margin-left: 8px">
                    <StarFilled />
                  </el-icon>
                </el-row>
              </template>
            </el-cascader>
            <!-- Final Platform Actions -->
            <el-button v-if="store.isSending" type="danger" circle :icon="VideoPause" @click="onStop" />
            <el-button v-else type="primary" circle :icon="Promotion" :disabled="!text.trim() || !selectedModel"
              @click="onSend" />
          </el-space>
        </el-col>
      </el-row>
    </div>
  </el-card>
</template>
<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { Promotion, StarFilled, VideoPause } from '@element-plus/icons-vue'
import { useAssistantStore } from '../store/useAssistantStore'
const props = defineProps<{ provider?: string; model?: string }>()
const emit = defineEmits(['send', 'stop', 'update:provider', 'update:model'])
const store = useAssistantStore()
const text = ref('buatkan workflow screenshot web github trending dan path image ke /Users/steen/Desktop/github-trend.png')
const selectedProvider = ref(props.provider || '')
const selectedModel = ref(props.model || '')
const cascaderKey = ref(0)
const favorites = ref<string[]>(JSON.parse(localStorage.getItem('assistant_favorites') || '[]'))
// Cascader logic
const cascaderValue = computed({
  get: () => (selectedProvider.value && selectedModel.value ? [selectedProvider.value, selectedModel.value] : []),
  set: (val: any) => {
    if (val && val.length === 2) {
      selectedProvider.value = val[0]
      selectedModel.value = val[1]
    }
  }
})
const cascaderProps = {
  lazy: true,
  async lazyLoad(node: any, resolve: any) {
    const { level, value } = node
    try {
      if (level === 0) {
        const res = await fetch('/api/assistant/providers')
        const json = await res.json()
        const nodes = json.data.map((p: any) => ({
          value: p.id,
          label: (p.label || p.name).toUpperCase(),
          leaf: false
        }))
        resolve(nodes)
      } else {
        const res = await fetch(`/api/assistant/providers/${value}/models`)
        const json = await res.json()
        const nodes = json.data.map((m: any) => ({
          value: m.model,
          label: (m.name || m.model).toUpperCase(),
          leaf: true
        }))
        resolve(nodes)
      }
    } catch (e) {
      resolve([])
    }
  }
}
watch(favorites, (v) => localStorage.setItem('assistant_favorites', JSON.stringify(v)), { deep: true })
watch(() => props.provider, (v) => { if (v && v !== selectedProvider.value) selectedProvider.value = v }, { immediate: true })
watch(() => props.model, (v) => { if (v && v !== selectedModel.value) selectedModel.value = v }, { immediate: true })
watch(selectedProvider, (v) => emit('update:provider', v))
watch(selectedModel, (v) => emit('update:model', v))
const isFavorite = (id: string) => favorites.value.includes(id)
const onSend = () => {
  if (text.value.trim() && selectedModel.value) {
    emit('send', { content: text.value.trim(), provider: selectedProvider.value, model: selectedModel.value, agent: store.selectedAgent })
    text.value = ''
  }
}
const onStop = () => {
  emit('stop')
}
defineExpose({
  refresh: () => {
    cascaderKey.value++
  }
})
</script>
<style scoped>
.chat-input-card {
  border-radius: 20px;
  border-color: var(--el-border-color-lighter);
}

.chat-input-card:focus-within {
  border-color: var(--el-color-primary-light-3);
  box-shadow: 0 0 0 1px var(--el-color-primary-light-3);
}

.seamless-input :deep(.el-textarea__inner) {
  border: none;
  box-shadow: none !important;
  background: transparent;
  padding: 8px;
  font-size: 15px;
  resize: none;
}

.input-toolbar {
  margin-top: 4px;
}

.toolbar-btn {
  font-size: 18px;
  color: var(--el-text-color-secondary);
}
</style>