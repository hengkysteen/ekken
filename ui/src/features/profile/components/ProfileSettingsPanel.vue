<template>
  <div>
    <el-text strong style="display: block; margin-bottom: 16px; font-size: 16px">Profile</el-text>
    <el-card shadow="never" style="max-width: 600px">
      <el-form label-position="left" label-width="140px" @submit.prevent="save">
        <el-form-item label="Name">
          <el-input v-model="form.name" placeholder="Display name" clearable />
        </el-form-item>

        <el-form-item label="App Lock" for="">
          <el-switch v-model="form.pin_enabled" />
        </el-form-item>

        <el-form-item>
          <el-text type="info" size="small">
            After you set a PIN and save, a lock shortcut will appear in the app bar.
          </el-text>
        </el-form-item>

        <template v-if="form.pin_enabled">
          <el-form-item :label="hasExistingPin ? 'New PIN' : 'PIN'" for="">
            <el-input-otp
              v-model="pin"
              :length="4"
              inputmode="numeric"
              mask
              :validator="isDigit"
            />
          </el-form-item>

          <el-form-item v-if="pin || !hasExistingPin" label="Confirm PIN" for="">
            <el-input-otp
              v-model="confirmPin"
              :length="4"
              inputmode="numeric"
              mask
              :validator="isDigit"
            />
          </el-form-item>

          <!-- Security Question Section -->
          <el-form-item label="Security Question">
            <el-select
              v-model="selectedQuestion"
              :placeholder="hasExistingQuestion ? `Keep existing: ${profileStore.profile.security_question}` : 'Select a security question'"
              clearable
              style="width: 100%"
            >
              <el-option v-for="q in predefinedQuestions" :key="q" :label="q" :value="q" />
              <el-option label="Write my own question..." value="custom" />
            </el-select>
          </el-form-item>

          <el-form-item v-if="selectedQuestion === 'custom'" label="Custom Question">
            <el-input v-model="customQuestion" placeholder="Enter your custom security question" clearable />
          </el-form-item>

          <el-form-item label="Security Answer">
            <el-input
              v-model="securityAnswer"
              :placeholder="hasExistingQuestion ? 'Keep existing answer' : 'Answer to the security question'"
              clearable
              show-password
            />
          </el-form-item>

          <el-form-item v-if="profileStore.profile.pin_updated_at">
            <el-text type="info" size="small">
              PIN last updated: {{ formatDate(profileStore.profile.pin_updated_at) }}
            </el-text>
          </el-form-item>
        </template>

        <el-form-item>
          <el-button type="primary" :loading="saving" @click="save">Save</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { useProfileStore } from '@profile/stores/profile'

const profileStore = useProfileStore()
const saving = ref(false)
const pin = ref('')
const confirmPin = ref('')
const selectedQuestion = ref('')
const customQuestion = ref('')
const securityAnswer = ref('')

const form = reactive({
  name: '',
  pin_enabled: false,
})

const predefinedQuestions = [
  "What was the name of your first pet?",
  "In what city were you born?",
  "What is your mother's maiden name?",
  "What was the name of your first school?",
  "What was the model of your first car?"
]

const hasExistingPin = computed(() => Boolean(profileStore.profile.pin_updated_at))
const hasExistingQuestion = computed(() => Boolean(profileStore.profile.security_question))

watch(
  () => profileStore.profile,
  (profile) => {
    form.name = profile.name
    form.pin_enabled = profile.pin_enabled
    pin.value = ''
    confirmPin.value = ''
    selectedQuestion.value = ''
    customQuestion.value = ''
    securityAnswer.value = ''
  },
  { immediate: true }
)

async function save() {
  if (form.pin_enabled) {
    if (!hasExistingPin.value && pin.value.length !== 4) {
      ElMessage.warning('PIN is required to enable App Lock')
      return
    }
    if (pin.value && pin.value.length !== 4) {
      ElMessage.warning('PIN must be 4 digits')
      return
    }
    if (pin.value && pin.value !== confirmPin.value) {
      ElMessage.warning('PIN confirmation does not match')
      return
    }

    // Security Question validation
    const question = selectedQuestion.value === 'custom' ? customQuestion.value.trim() : selectedQuestion.value
    const answer = securityAnswer.value.trim()

    if (!hasExistingQuestion.value) {
      if (!question || !answer) {
        ElMessage.warning('Security Question and Answer are required to enable App Lock')
        return
      }
    } else {
      if ((question && !answer) || (!question && answer)) {
        ElMessage.warning('Both Security Question and Answer must be filled to update them')
        return
      }
    }
  }

  try {
    saving.value = true
    const question = selectedQuestion.value === 'custom' ? customQuestion.value.trim() : selectedQuestion.value
    const answer = securityAnswer.value.trim()

    await profileStore.saveProfile({
      name: form.name,
      pin_enabled: form.pin_enabled,
      pin: pin.value.trim() || undefined,
      security_question: question || undefined,
      security_answer: answer || undefined,
    })
    pin.value = ''
    confirmPin.value = ''
    selectedQuestion.value = ''
    customQuestion.value = ''
    securityAnswer.value = ''
    ElMessage.success('Profile saved')
  } catch (err: any) {
    ElMessage.error(err.message || 'Failed to save profile')
  } finally {
    saving.value = false
  }
}

function isDigit(char: string) {
  return /^[0-9]$/.test(char)
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat(undefined, {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(new Date(value))
}
</script>
