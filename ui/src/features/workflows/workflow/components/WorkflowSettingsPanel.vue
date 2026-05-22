<template>
  <div class="workflow-settings">
    <!-- Appearance Section -->
    <el-text class="subsection-title">Appearance</el-text>
    <el-form label-position="left" label-width="140px" class="settings-form">
      <el-form-item label="Edge Style" for="">
        <el-select v-model="edgeStyle" style="width: 100%">
          <el-option v-for="opt in edgeOptions" :key="opt.value" :label="opt.label" :value="opt.value" />
        </el-select>
      </el-form-item>

      <el-form-item label="Edge Animation" for="">
        <el-switch v-model="edgeAnimated" />
      </el-form-item>

      <el-form-item>
        <el-button size="small" type="info" dashed @click="resetWorkflowSettings">
          Reset
        </el-button>
      </el-form-item>
    </el-form>

    <el-divider />

    <!-- Danger Zone Section -->
    <div class="danger-zone">
      <el-text class="subsection-title">Danger Zone</el-text>
      <el-card shadow="never" class="danger-card">
        <el-row align="middle" justify="space-between">
          <el-col :span="18">
            <el-text strong>Delete All Workflows</el-text>
            <div style="margin-top: 4px">
              <el-text size="small" type="info">Permanently remove all workflows. This action cannot be
                undone.</el-text>
            </div>
          </el-col>
          <el-col :span="6" style="text-align: right">
            <el-button type="danger" plain size="small" :icon="Delete" @click="handleDeleteAllWorkflows">
              Delete All
            </el-button>
          </el-col>
        </el-row>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Delete } from '@element-plus/icons-vue'
import { useWorkflowStore } from '@workflows/workflow/stores/workflow'
import { useWorkflowEditorStore } from '@workflows/workflow/stores/workflowEditor'
import { DEFAULT_EDGE_STYLE, DEFAULT_EDGE_ANIMATED } from '@workflows/workflow/utils/workflowSettings'

const editor = useWorkflowEditorStore()
const workflows = useWorkflowStore()
const { edgeStyle, edgeAnimated } = storeToRefs(editor)

const edgeOptions = [
  { label: 'Smooth Step', value: 'smoothstep' },
  { label: 'Curved (Bezier)', value: 'default' },
  { label: 'Sharp Step', value: 'step' },
  { label: 'Straight Line', value: 'straight' }
]

function resetWorkflowSettings() {
  edgeStyle.value = DEFAULT_EDGE_STYLE
  edgeAnimated.value = DEFAULT_EDGE_ANIMATED
  ElMessage.success('Settings reset to defaults')
}

async function handleDeleteAllWorkflows() {
  try {
    await ElMessageBox.confirm(
      'This will permanently delete all your workflows. This action cannot be undone. Are you sure?',
      'Delete All Workflows',
      {
        confirmButtonText: 'Delete All',
        cancelButtonText: 'Cancel',
        type: 'warning',
        confirmButtonClass: 'el-button--danger',
        distinguishCancelAndClose: true,
      }
    )
    await workflows.deleteAllWorkflows()
    ElMessage.success('All workflows deleted successfully')
  } catch (err) {
    if (err !== 'cancel' && err !== 'close') {
      console.error(err)
    }
  }
}
</script>

<style scoped>
.workflow-settings {
  width: 100%;
}

.subsection-title {
  font-weight: 600;
  margin-bottom: 16px;
  display: block;
  color: var(--el-text-color-primary);
}

.settings-form {
  max-width: 500px;
  margin-top: 8px;
}

.danger-zone {
  margin-top: 32px;
}

.danger-card {
  border-radius: 12px;
  border: 1px solid var(--el-border-color-lighter);
}
</style>
