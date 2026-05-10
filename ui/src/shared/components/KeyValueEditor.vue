<template>
  <div class="kv-editor">
    <div v-for="(row, i) in rows" :key="i" class="kv-row">
      <el-input v-model="row.key" :placeholder="placeholderKey" size="small" class="kv-key" />
      <el-input v-model="row.value" :placeholder="placeholderValue" size="small" class="kv-value" />
      <el-button size="small" circle @click="removeRow(i)" class="kv-remove">
        <el-icon><Close /></el-icon>
      </el-button>
    </div>
    <div class="kv-footer">
      <el-button link type="primary" size="small" @click="addRow" class="kv-add">
        + {{ addLabel }}
      </el-button>
      <el-button v-if="rows.length > 1 || rows[0].key || rows[0].value" link type="danger" size="small" @click="clearAll" class="kv-clear">
        Clear All
      </el-button>
    </div>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import { Close } from '@element-plus/icons-vue'

defineOptions({ inheritAttrs: false })

const props = defineProps({
  modelValue: { type: Array, default: () => [] },
  placeholderKey: { type: String, default: 'Key' },
  placeholderValue: { type: String, default: 'Value' },
  addLabel: { type: String, default: 'Add' }
})

const emit = defineEmits(['update:modelValue'])

const rows = ref(props.modelValue.length ? [...props.modelValue] : [{ key: '', value: '' }])
const isSyncing = ref(false)

watch(rows, (val) => {
  if (isSyncing.value) return
  if (JSON.stringify(val) !== JSON.stringify(props.modelValue)) {
    emit('update:modelValue', [...val])
  }
}, { deep: true })

watch(() => props.modelValue, (val) => {
  if (JSON.stringify(val) !== JSON.stringify(rows.value)) {
    isSyncing.value = true
    rows.value = val && val.length ? [...val] : [{ key: '', value: '' }]
    import('vue').then(({ nextTick }) => {
      nextTick(() => { isSyncing.value = false })
    })
  }
}, { deep: true })

const addRow = () => {
  rows.value.push({ key: '', value: '' })
}

const removeRow = (i) => {
  if (rows.value.length === 1) {
    rows.value[0] = { key: '', value: '' }
  } else {
    rows.value.splice(i, 1)
  }
}

const clearAll = () => {
  rows.value = [{ key: '', value: '' }]
}
</script>

<style scoped>
.kv-editor {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.kv-row {
  display: flex;
  gap: 6px;
  align-items: center;
}

.kv-key {
  flex: 0 0 40%;
}

.kv-value {
  flex: 1;
}

.kv-key :deep(.el-input__inner),
.kv-value :deep(.el-input__inner) {
  font-family: var(--el-font-family-mono);
  font-size: 12px;
}

.kv-remove {
  flex-shrink: 0;
}

.kv-add {
  font-size: 12px;
  padding: 4px 0;
  height: auto;
}

.kv-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 4px;
}

.kv-clear {
  font-size: 11px;
}
</style>
