<template>
  <div class="node-not-found">
    <el-result icon="error" :title="`Node ${nodeType} not found`"
      sub-title="This node type is not registered. It may have been removed or the plugin is not installed." />
  </div>
</template>

<script setup lang="ts">
import type { NodeFormProps } from '@workflows/node/types/node'

const props = defineProps<NodeFormProps>()

const nodeType = props.node?.data?.nodeType || props.node?.type || 'unknown'
</script>