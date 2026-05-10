<template>
  <div style="width: 100%;">
    <el-form label-position="top" @submit.prevent>


      <el-form-item v-if="nodeDef?.actions?.length" label="Action" :for="`action-${node?.id}`">

        <el-segmented :id="`action-${node?.id}`" v-model="currentAction" :options="actionOptions" />


      </el-form-item>

      <template v-if="nodeDef?.actions?.length">
        <div v-for="action in nodeDef.actions" :key="action.key">
          <EkDynamicForm v-if="currentAction === action.key" :form="(action.form || []) as any" :fields="action.fields"
            v-model="currentConfig" />
        </div>
      </template>

      <el-collapse v-if="nodeDef?.global_fields" class="advanced-collapse">
        <el-collapse-item title="Advanced Settings" name="advanced">
          <EkDynamicForm :form="nodeDef.global_form || []" :fields="nodeDef.global_fields" v-model="globalConfig" />
        </el-collapse-item>
      </el-collapse>
    </el-form>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useNodeStore } from '@workflows/node/stores/node'
import EkDynamicForm from './EkDynamicForm.vue'
import type { NodeFormProps, NodeDefinition } from '@workflows/node/types/node'

const props = defineProps<NodeFormProps>()
const nodeStore = useNodeStore()

const currentAction = ref('')
const actionConfigs = ref<Record<string, Record<string, any>>>({})

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

const currentConfig = computed({
  get: () => actionConfigs.value[currentAction.value] || {},
  set: (val) => {
    if (currentAction.value) {
      actionConfigs.value[currentAction.value] = val
    }
  }
})

const globalConfig = computed({
  get: () => actionConfigs.value['global'] || {},
  set: (val) => {
    actionConfigs.value['global'] = val
  }
})

onMounted(() => {
  // Load existing config
  const existingConfig = props.node?.data?.config || {}
  currentAction.value = existingConfig.action || nodeDef.value?.default_action || ''

  // 1. Initialize Action Configs
  if (nodeDef.value?.actions?.length) {
    nodeDef.value.actions.forEach(action => {
      const actionData: Record<string, any> = {}
      action.fields?.forEach(field => {
        if (field.key in existingConfig) {
          actionData[field.key] = existingConfig[field.key]
        } else if (field.default !== undefined) {
          actionData[field.key] = field.default
        }
      })
      actionConfigs.value[action.key] = actionData
    })
  }

  // 2. Initialize Global Config
  if (nodeDef.value?.global_fields) {
    const globalData: Record<string, any> = {}
    nodeDef.value.global_fields.forEach(field => {
      if (field.key in existingConfig) {
        globalData[field.key] = existingConfig[field.key]
      } else if (field.default !== undefined) {
        globalData[field.key] = field.default
      }
    })
    actionConfigs.value['global'] = globalData
  }
})

function getData(): Record<string, any> {
  return {
    ...globalConfig.value,
    ...currentConfig.value,
    action: currentAction.value,
  }
}
defineExpose({ getData })
</script>

<style scoped>
.advanced-collapse {
  margin-top: 20px;
  border-top: 1px solid var(--el-border-color-lighter);
}
</style>
