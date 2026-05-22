import { defineAsyncComponent, markRaw } from 'vue'
import { User } from '@element-plus/icons-vue'
import type { EkkenModule } from '../../core/types/module'
import { useProfileStore } from '@profile/stores/profile'

const profileModule: EkkenModule = {
  id: 'profile',
  name: 'Profile',

  settingsTabs: [
    {
      id: 'profile',
      label: 'Profile',
      description: 'Local profile and app lock',
      icon: markRaw(User),
      component: markRaw(defineAsyncComponent(() => import('./components/ProfileSettingsPanel.vue'))),
      order: 5,
    },
  ],

  async onStartup() {
    const profileStore = useProfileStore()
    await profileStore.fetchProfile()
  },
}

export default profileModule
