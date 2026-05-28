import { markRaw, defineAsyncComponent } from 'vue'
import { MagicStick } from '@element-plus/icons-vue'
import type { EkkenModule } from '../../core/types/module'

const assistantModule: EkkenModule = {
  id: 'assistant',
  name: 'Assistant',

  routes: [
    {
      path: 'assistant',
      name: 'assistant',
      component: () => import('./views/AssistantView.vue'),
      meta: { breadcrumb: 'Assistant' }
    },
    {
      path: 'assistant/:id',
      name: 'assistant-chat',
      component: () => import('./views/AssistantView.vue'),
      meta: { breadcrumb: ':id', parent: '/assistant' }
    }
  ],

  sidebarItems: [
    {
      label: 'Assistant',
      icon: markRaw(MagicStick),
      path: '/assistant',
      name: 'assistant',
      order: 30
    }
  ],

  settingsTabs: [
    {
      id: 'assistant',
      label: 'Assistant',
      icon: markRaw(MagicStick),
      component: markRaw(defineAsyncComponent(() => import('./components/AssistantSettingsPanel.vue'))),
      order: 30
    }
  ]
}

export default assistantModule
