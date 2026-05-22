import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { profileApi, type Profile, type UpdateProfilePayload } from '@profile/api'
import { Storage, StorageKeys } from '@shared/utils/storage'

const defaultProfile: Profile = {
  name: '',
  pin_enabled: false,
}

function getStoredUnlocked() {
  const value = Storage.get<boolean | string>(StorageKeys.PROFILE_UNLOCKED)
  return value === true || value === 'true'
}

export const useProfileStore = defineStore('profile', () => {
  const profile = ref<Profile>({ ...defaultProfile })
  const loading = ref(false)
  const initialized = ref(false)
  const unlocked = ref(getStoredUnlocked())

  const displayName = computed(() => profile.value.name.trim() || 'John Doe')
  const initials = computed(() => {
    const parts = displayName.value.trim().split(/\s+/).filter(Boolean)
    const chars = parts.length > 1
      ? `${parts[0][0]}${parts[1][0]}`
      : displayName.value.slice(0, 2)
    return chars.toUpperCase()
  })
  const requiresUnlock = computed(() => initialized.value && profile.value.pin_enabled && !unlocked.value)

  async function fetchProfile() {
    loading.value = true
    try {
      profile.value = await profileApi.get()
      if (!profile.value.pin_enabled) {
        unlocked.value = true
        Storage.remove(StorageKeys.PROFILE_UNLOCKED)
      }
    } catch (err) {
      console.error('Failed to load profile:', err)
      profile.value = { ...defaultProfile }
      unlocked.value = true
    } finally {
      loading.value = false
      initialized.value = true
    }
  }

  async function saveProfile(payload: UpdateProfilePayload) {
    profile.value = await profileApi.update(payload)
    unlocked.value = true
    if (profile.value.pin_enabled) {
      Storage.set(StorageKeys.PROFILE_UNLOCKED, 'true')
    } else {
      Storage.remove(StorageKeys.PROFILE_UNLOCKED)
    }
    return profile.value
  }

  async function verifyPin(pin: string) {
    const result = await profileApi.verifyPin(String(pin).trim())
    if (result.valid) {
      unlocked.value = true
      Storage.set(StorageKeys.PROFILE_UNLOCKED, 'true')
    }
    return result.valid
  }

  async function resetPin(answer: string, newPin: string) {
    const result = await profileApi.resetPin({
      answer: String(answer).trim(),
      new_pin: String(newPin).trim(),
    })
    if (result) {
      unlocked.value = true
      Storage.set(StorageKeys.PROFILE_UNLOCKED, 'true')
      await fetchProfile()
    }
    return result
  }

  function lockApp() {
    if (profile.value.pin_enabled) {
      unlocked.value = false
      Storage.remove(StorageKeys.PROFILE_UNLOCKED)
    }
  }

  return {
    profile,
    loading,
    initialized,
    unlocked,
    displayName,
    initials,
    requiresUnlock,
    fetchProfile,
    saveProfile,
    verifyPin,
    resetPin,
    lockApp,
  }
})
