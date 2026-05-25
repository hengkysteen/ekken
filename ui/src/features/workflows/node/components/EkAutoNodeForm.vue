<template>
  <div style="width: 100%;">
    <template v-if="!nodeDef">
      <EkNodeNotFound :node="props.node" />
    </template>
    <el-form v-else label-position="top" @submit.prevent>


      <el-form-item v-if="nodeDef?.actions?.length" :for="`action-${node?.id}`">
        <template #label>
          <div class="flex items-center gap-2">
            <span>{{ nodeDef?.actions?.length > 1 ? 'Actions' : 'Action' }}</span>
            <el-text v-if="currentActionDescription" type="info" size="small">
              ({{ currentActionDescription }})
            </el-text>
          </div>
        </template>
        <el-select v-if="nodeDef?.actions?.length > 6" 
          v-model="currentActionType" 
          :id="`action-${node?.id}`"
          style="width: 100%;">
          <el-option v-for="action in actionOptions" 
            :key="action.value" 
            :label="action.label" 
            :value="action.value" />
        </el-select>
        <el-segmented v-else 
          :id="`action-${node?.id}`" 
          v-model="currentActionType" 
          :options="actionOptions" />
      </el-form-item>

      <template v-if="localAction">
        <div v-if="isWebhookNode" class="webhook-url-panel">
          <el-form-item label="Local URL">
            <el-input :model-value="webhookLocalUrl" readonly>
              <template #append>
                <el-button :icon="Refresh" @click="regenerateWebhookID" />
                <el-button :icon="CopyDocument" @click="copyWebhookLocalUrl" />
              </template>
            </el-input>
          </el-form-item>
        </div>
        <EkDynamicForm 
          :layout="(actionBlueprint?.auto_layout || []) as any" 
          :fields="hydratedAction.fields"
          :global-fields="localGlobalFields"
          @update:field="handleFieldUpdate"
          @update:global-field="handleGlobalFieldUpdate"
        />
      </template>

    </el-form>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { CopyDocument, Refresh } from '@element-plus/icons-vue'
import { useNodeStore } from '@workflows/node/stores/node'
import {
  buildActionInstance,
  getActionBlueprint,
  hydrateActionForForm,
  hydrateFieldsForForm,
  serializeActionForSave
} from '@workflows/node/utils/node'
import EkDynamicForm from './EkDynamicForm.vue'
import EkNodeNotFound from './EkNodeNotFound.vue'
import type { NodeFormProps, NodeDefinition } from '@workflows/node/types/node'

const props = defineProps<NodeFormProps>()
const emit = defineEmits(['action-change'])
const nodeStore = useNodeStore()

const currentActionType = ref('')
const localAction = ref<any>(null)
const localGlobalFields = ref<any[]>([])

watch(localAction, (newAction) => {
  emit('action-change', newAction)
}, { deep: true, immediate: true })

const catalog = computed(() => nodeStore.catalog)
const actionOptions = computed(() => {
  const options = nodeDef.value?.actions?.map(action => ({
    label: action.label,
    value: action.type
  })) || []
  return options.sort((a, b) => a.label.localeCompare(b.label))
})

const nodeDef = computed(() => {
  const type = props.node?.data?.nodeType || props.node?.type
  return catalog.value?.find((n: NodeDefinition) => n.type === type)
})

const actionBlueprint = computed(() => getActionBlueprint(nodeDef.value, currentActionType.value))
const currentActionDescription = computed(() => actionBlueprint.value?.description || '')
const hydratedAction = computed(() => hydrateActionForForm(localAction.value, nodeDef.value, currentActionType.value))
const isWebhookNode = computed(() => nodeDef.value?.type === 'webhook')
const webhookID = computed(() => getLocalFieldValue('webhook_id'))
const webhookLocalUrl = computed(() => {
  const id = webhookID.value
  if (!id) return ''
  return `${window.location.origin}/api/webhook/${id}`
})

// Handle action switching
watch(currentActionType, (newType) => {
  if (newType && localAction.value && localAction.value.type !== newType) {
    const oldFields = localAction.value.fields || []
    const newAction = buildActionInstance(nodeDef.value, newType)

    // Preserve values for fields with same keys
    newAction.fields = newAction.fields.map((f: any) => {
      const oldField = oldFields.find((of: any) => of.key === f.key)
      return oldField ? { key: f.key, value: oldField.value } : f
    })

    localAction.value = newAction
  }
})

onMounted(() => {
  // Load existing action or build default
  if (props.node?.data?.action) {
    const savedAction = serializeActionForSave(props.node.data.action)
    currentActionType.value = savedAction.type || nodeDef.value?.default_action || nodeDef.value?.actions?.[0]?.type || ''

    // Split fields: separate global fields from action fields
    const globalKeys = new Set(nodeDef.value?.global_fields?.map((f: any) => f.key) || [])
    const actionFields: any[] = []
    const globalFields: any[] = []
    
    savedAction.fields?.forEach((f: any) => {
      if (globalKeys.has(f.key)) {
        globalFields.push(f)
      } else {
        actionFields.push(f)
      }
    })

    localAction.value = { ...savedAction, fields: actionFields }
    localGlobalFields.value = hydrateFieldsForForm(nodeDef.value?.global_fields, globalFields)
  } else if (nodeDef.value) {
    localAction.value = buildActionInstance(nodeDef.value)
    currentActionType.value = localAction.value.type
    localGlobalFields.value = hydrateFieldsForForm(nodeDef.value.global_fields)
  }

  if (isWebhookNode.value && !getLocalFieldValue('webhook_id')) {
    handleFieldUpdate('webhook_id', generateWebhookID())
  }
})

function handleFieldUpdate(key: string, value: any) {
  if (!localAction.value) return
  const field = localAction.value.fields.find((f: any) => f.key === key)
  if (field) {
    field.value = value
  } else {
    localAction.value.fields.push({ key, value })
  }
}

function getLocalFieldValue(key: string): any {
  return localAction.value?.fields?.find((f: any) => f.key === key)?.value
}

function generateWebhookID(): string {
  const bytes = new Uint8Array(8)
  if (window.crypto?.getRandomValues) {
    window.crypto.getRandomValues(bytes)
    return Array.from(bytes, b => b.toString(16).padStart(2, '0')).join('')
  }
  return `${Date.now().toString(36)}${Math.random().toString(36).slice(2, 10)}`
}

function regenerateWebhookID() {
  handleFieldUpdate('webhook_id', generateWebhookID())
}

async function copyWebhookLocalUrl() {
  if (!webhookLocalUrl.value) return
  try {
    await navigator.clipboard.writeText(webhookLocalUrl.value)
    ElMessage.success('Copied')
  } catch {
    ElMessage.error('Copy failed')
  }
}

function handleGlobalFieldUpdate(key: string, value: any) {
  const field = localGlobalFields.value.find((f: any) => f.key === key)
  if (field) {
    field.value = value
  }
}

function getData(): any {
  // Merge global fields into action fields before returning
  const mergedFields = [...(localAction.value?.fields || [])]
  
  localGlobalFields.value.forEach(gf => {
    if (!mergedFields.find(f => f.key === gf.key)) {
      mergedFields.push(gf)
    }
  })
  
  return serializeActionForSave({
    ...localAction.value,
    fields: mergedFields
  })
}
defineExpose({ getData })
</script>

<style scoped>
.advanced-collapse {
  margin-top: 20px;
  border-top: 1px solid var(--el-border-color-lighter);
}

.webhook-url-panel {
  margin-bottom: 16px;
}
</style>
