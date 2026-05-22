<template>
  <div class="Container">


    <div class="content">
      <div class="logo-wrapper">
        <AppLogo style="width:40px; height: 40px;" />
      </div>
      <h2 class="lock-title">Ekken is Locked</h2>
      <p class="lock-subtitle">Hi , {{ profileStore.displayName }}</p>
      <p class="otp-hint">Enter your 4 digit PIN</p>
      <el-input-otp v-model="otp" :length="4" inputmode="numeric" size="large" mask :validator="isDigit"
        @finish="unlock" />
      <div class="status-message">
        <span v-if="loading" class="loading-dots">Verifying PIN</span>
      </div>
      <el-button v-if="profileStore.profile.security_question" link class="forgot-btn" @click="showForgotDialog = true">
        Forgot PIN?
      </el-button>
    </div>



    <!-- Forgot PIN / Reset Dialog -->
    <el-dialog v-model="showForgotDialog" title="Reset PIN" width="90%" style="max-width: 440px" destroy-on-close
      :close-on-click-modal="false" @close="closeForgotDialog">
      <el-form label-position="top" style="margin-top: 8px">
        <el-form-item :label="`Security Question: ${profileStore.profile.security_question}`">
          <el-input v-model="resetAnswer" placeholder="Your Answer" clearable />
        </el-form-item>
        <el-form-item label="New 4-Digit PIN">
          <el-input-otp v-model="resetNewPin" :length="4" inputmode="numeric" mask :validator="isDigit" />
        </el-form-item>
      </el-form>
      <template #footer>
        <div style="display: flex; justify-content: flex-end; gap: 12px">
          <el-button @click="closeForgotDialog">Cancel</el-button>
          <el-button type="primary" :loading="resetLoading" @click="handleReset">Reset</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>



<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useProfileStore } from '@profile/stores/profile'
import AppLogo from '@shared/components/AppLogo.vue'
const profileStore = useProfileStore()
const otp = ref('')
const loading = ref(false)
const isShaking = ref(false)
const showForgotDialog = ref(false)
const resetAnswer = ref('')
const resetNewPin = ref('')
const resetLoading = ref(false)
function isDigit(char: string) {
  return /^[0-9]$/.test(char)
}
async function unlock() {
  const pin = String(otp.value).trim()
  if (pin.length !== 4) return
  try {
    loading.value = true
    const valid = await profileStore.verifyPin(pin)
    if (!valid) {
      otp.value = ''
      triggerShake()
      ElMessage.error('Invalid PIN')
      return
    }
  } catch (err: any) {
    otp.value = ''
    triggerShake()
    if (err.message === 'invalid pin') {
      ElMessage.error('Invalid PIN')
      return
    }
    ElMessage.error(err.message || 'Failed to verify PIN')
  } finally {
    loading.value = false
  }
}
function triggerShake() {
  isShaking.value = true
  setTimeout(() => {
    isShaking.value = false
  }, 500)
}
function closeForgotDialog() {
  showForgotDialog.value = false
  resetAnswer.value = ''
  resetNewPin.value = ''
}
async function handleReset() {
  const answer = resetAnswer.value.trim()
  const pin = resetNewPin.value.trim()
  if (!answer) {
    ElMessage.warning('Answer is required')
    return
  }
  if (pin.length !== 4) {
    ElMessage.warning('New PIN must be 4 digits')
    return
  }
  try {
    resetLoading.value = true
    const success = await profileStore.resetPin(answer, pin)
    if (success) {
      ElMessage.success('PIN reset successfully')
      closeForgotDialog()
    } else {
      triggerShake()
      ElMessage.error('Incorrect answer')
    }
  } catch (err: any) {
    ElMessage.error(err.message || 'Failed to reset PIN')
  } finally {
    resetLoading.value = false
  }
}
</script>

<style scoped>
.Container {
  width: 100%;
  height: 100vh;
  display: flex;
  justify-content: center;
  align-items: center;
  background-color: var(--el-bg-color);
}

.content {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  gap: 16px;
}
</style>