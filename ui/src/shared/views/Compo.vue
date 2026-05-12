<template>
    <AppPage scrollable title="EkDynamicForm Example">
        <div style="padding: 20px; border: 1px solid #ddd; border-radius: 8px;">
            <h3 style="margin-top: 0;">Dynamic Form (Auto Layout)</h3>
            <el-divider></el-divider>
            <EkDynamicForm v-model="dynamicFormData" :layout="autoLayout" :fields="fields" />
            <pre style="margin-top: 20px; padding: 10px; background: #f5f5f5; border-radius: 4px;">{{ dynamicFormData }}</pre>
        </div>
    </AppPage>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import AppPage from '../components/AppPage.vue';
import EkDynamicForm from '@workflows/node/components/EkDynamicForm.vue';

// Dynamic form data
const dynamicFormData = ref({
  username: '',
  email: '',
  api_key: '',
  port: 8080,
  timeout: 30,
  region: '',
  file_path: '',
  description: '',
  status: 'active',
  volume: 50,
  enabled: true,
  config: '{\n  "key": "value"\n}',
  color: '#409EFF',
  date: '',
  time: ''
})

// Auto layout: array of rows, each row has items with flex
const autoLayout = [
  [{ key: 'info_header', component: 'text', flex: 1, options: { text: 'User Information', size: '18px', bold: true } }],
  [{ key: 'username', flex: 1, component: 'input' }, { key: 'email', flex: 1, component: 'input' }],
  [{ key: 'api_key', flex: 1, component: 'input' }],
  [{ key: 'info_network', component: 'text', flex: 1, options: { text: 'Network Settings', size: '16px', bold: true, color: '#409EFF' } }],
  [{ key: 'port', flex: 1, component: 'number-s1', options: { min: 100 } }, { key: 'timeout', flex: 1, component: 'number-s2' }],
  [{ key: 'region', flex: 1, component: 'select' }],
  [{ key: 'description', flex: 1, component: 'textarea' }],
  [{ key: 'status', flex: 1, component: 'radio' }, { key: 'enabled', flex: 1, component: 'switch' }],
  [{ key: 'volume', flex: 1, component: 'slider' }],
  [{ key: 'color', flex: 1, component: 'colorPicker' }, { key: 'date', flex: 1, component: 'datePicker' }, { key: 'time', flex: 1, component: 'timePicker' }],
  [{ key: 'config', flex: 1, component: 'jsonEditor' }],
  [{ key: 'file_path', flex: 1, component: 'input' }]
]

// Field definitions
const fields = [
  { key: 'username', type: 'input', label: 'Username', helper: 'Enter your username', placeholder: 'Type username...' },
  { key: 'email', type: 'input', label: 'Email', helper: 'Format: user@example.com', placeholder: 'email@example.com' },
  { key: 'api_key', type: 'input', label: 'API Key', helper: 'Click icon to select credential', placeholder: 'Enter API key...', credential_picker: true },
  { key: 'port', type: 'number', label: 'Port (Style 1)', helper: 'Controls on right', min: 1, max: 65535 },
  { key: 'timeout', type: 'number', label: 'Timeout (Style 2)', helper: 'Controls both sides', min: 0, max: 300 },
  { key: 'region', type: 'select', label: 'Region', helper: 'Select server region', placeholder: 'Select region', clearable: true, options: ['Asia Pacific', 'Europe', 'US East', 'US West'] },
  { key: 'description', type: 'textarea', label: 'Description', helper: 'Max 200 characters', placeholder: 'Enter description...', rows: 4, maxlength: 200 },
  { key: 'status', type: 'radio', label: 'Status', helper: 'Select status', options: ['active', 'inactive', 'pending'] },
  { key: 'enabled', type: 'switch', label: 'Enabled', helper: 'Toggle to enable/disable' },
  { key: 'volume', type: 'slider', label: 'Volume', helper: 'Adjust volume level', min: 0, max: 100, show_input: true },
  { key: 'color', type: 'colorPicker', label: 'Color', helper: 'Pick a color' },
  { key: 'date', type: 'datePicker', label: 'Date', helper: 'Select a date', placeholder: 'Pick a date' },
  { key: 'time', type: 'timePicker', label: 'Time', helper: 'Select a time', placeholder: 'Pick a time' },
  { key: 'config', type: 'jsonEditor', label: 'Config (JSON)', helper: 'Enter valid JSON', height: '250px' },
  { key: 'file_path', type: 'input', label: 'File Path', helper: 'Click icon to select file', placeholder: 'Select file...', native_file_picker: true }
]
</script>