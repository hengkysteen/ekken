<template>
  <div class="assistant-settings">
    <!-- Data Management Section -->
    <div class="danger-zone">
      <el-text class="subsection-title">Data Management</el-text>
      <el-card shadow="never" class="danger-card">
        <el-row align="middle" justify="space-between">
          <el-col :span="18">
            <el-text strong>Delete All Conversations</el-text>
            <div style="margin-top: 4px">
              <el-text size="small" type="info">Permanently remove all chat history from your local workspace. This action cannot be undone.</el-text>
            </div>
          </el-col>
          <el-col :span="6" style="text-align: right">
            <el-button type="danger" plain size="small" :icon="Delete" @click="handleDeleteAllConversations">
              Delete All
            </el-button>
          </el-col>
        </el-row>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ElMessage, ElMessageBox } from 'element-plus'
import { Delete } from '@element-plus/icons-vue'
import { deleteAllConversations } from '../api'

async function handleDeleteAllConversations() {
  try {
    await ElMessageBox.confirm(
      'This will permanently delete all your chat history. This action cannot be undone. Are you sure?',
      'Delete All Conversations',
      {
        confirmButtonText: 'Delete All',
        cancelButtonText: 'Cancel',
        type: 'warning',
        confirmButtonClass: 'el-button--danger',
        distinguishCancelAndClose: true,
      }
    )
    
    await deleteAllConversations()
    ElMessage.success('All conversations deleted successfully')
  } catch (err) {
    if (err !== 'cancel' && err !== 'close') {
      console.error(err)
    }
  }
}
</script>

<style scoped>
.assistant-settings {
  width: 100%;
}

.subsection-title {
  font-weight: 600;
  margin-bottom: 16px;
  display: block;
  color: var(--el-text-color-primary);
}

.danger-zone {
  margin-top: 8px;
}

.danger-card {
  border-radius: 12px;
  border: 1px solid var(--el-border-color-lighter);
}
</style>
