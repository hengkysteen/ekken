import { markRaw } from 'vue'
import { Key } from '@element-plus/icons-vue'
import type { EkkenModule } from '../../core/types/module'

const credentialModule: EkkenModule = {
  id: 'credential',
  name: 'Credential',
  
  routes: [
    {
      path: 'credentials',
      name: 'credentials',
      component: () => import('./views/CredentialsView.vue'),
      meta: { breadcrumb: 'Credentials' }
    }
  ],
  
  sidebarItems: [
    { 
      label: 'Credentials', 
      icon: markRaw(Key), 
      path: '/credentials', 
      name: 'credentials',
      order: 110 // In System section
    }
  ]
}

export default credentialModule
