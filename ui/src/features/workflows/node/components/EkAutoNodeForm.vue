<template>
  <div style="width: 100%;">
    <el-form label-position="top" @submit.prevent>


      <el-form-item v-if="nodeDef?.actions?.length" label="Action" :for="`action-${node?.id}`">
        <el-segmented :id="`action-${node?.id}`" v-model="currentActionKey" :options="actionOptions" />
      </el-form-item>

      <template v-if="localAction">
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
import { useNodeStore } from '@workflows/node/stores/node'
import {
  buildActionInstance,
  getActionBlueprint,
  hydrateActionForForm,
  hydrateFieldsForForm,
  serializeActionForSave
} from '@workflows/node/utils/node'
import EkDynamicForm from './EkDynamicForm.vue'
import type { NodeFormProps, NodeDefinition } from '@workflows/node/types/node'

const props = defineProps<NodeFormProps>()
const nodeStore = useNodeStore()

const currentActionKey = ref('')
const localAction = ref<any>(null)
const localGlobalFields = ref<any[]>([])

const catalog = computed(() => nodeStore.catalog)
const actionOptions = computed(() => {
  return nodeDef.value?.actions?.map(action => ({
    label: action.label,
    value: action.key
  })) || []
})

const nodeDef = computed(() => {
  const type = props.node?.data?.nodeType || props.node?.type
  return catalog.value?.find((n: NodeDefinition) => n.type === type)
})

const actionBlueprint = computed(() => getActionBlueprint(nodeDef.value, currentActionKey.value))
const hydratedAction = computed(() => hydrateActionForForm(localAction.value, nodeDef.value, currentActionKey.value))

// Handle action switching
watch(currentActionKey, (newKey) => {
  if (newKey && localAction.value && localAction.value.key !== newKey) {
    const oldFields = localAction.value.fields || []
    const newAction = buildActionInstance(nodeDef.value, newKey)

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
    currentActionKey.value = savedAction.key || nodeDef.value?.default_action || nodeDef.value?.actions?.[0]?.key || ''

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
    currentActionKey.value = localAction.value.key
    localGlobalFields.value = hydrateFieldsForForm(nodeDef.value.global_fields)
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
</style>
