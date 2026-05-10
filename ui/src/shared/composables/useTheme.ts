import { computed } from 'vue'
import { useColorMode, usePreferredDark } from '@vueuse/core'
import { StorageKeys } from '../utils/storage'

// Singleton state menggunakan VueUse
const mode = useColorMode({
  selector: 'html',
  attribute: 'class',
  storageKey: StorageKeys.THEME_MODE,
  initialValue: 'auto',
})

const prefersDark = usePreferredDark()

export function useTheme() {
  // isDarkMode: resolve 'auto' ke preferensi sistem, 'dark'/'light' langsung
  const isDarkMode = computed(() => {
    if (mode.value === 'dark') return true
    if (mode.value === 'light') return false
    return prefersDark.value // mode === 'auto'
  })

  return {
    themeMode: mode,
    isDarkMode,
  }
}
