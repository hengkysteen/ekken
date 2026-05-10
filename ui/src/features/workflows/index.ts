import { markRaw, defineAsyncComponent } from 'vue'
import { CopyDocument } from '@element-plus/icons-vue'
import type { EkkenModule } from '../../core/types/module'

const workflowModule: EkkenModule = {
  id: 'workflow',
  name: 'Workflow',

  routes: [
    {
      path: 'workflows',
      name: 'workflows',
      component: () => import('./workflow/views/WorkflowsView.vue'),
      meta: { breadcrumb: 'Workflows' }
    },
    {
      path: 'workflow/:id',
      name: 'workflow',
      component: () => import('./workflow/views/WorkflowEditor.vue'),
      meta: { breadcrumb: ':id', parent: '/workflows' }
    }
  ],

  sidebarItems: [
    {
      label: 'Workflows',
      icon: markRaw(CopyDocument),
      path: '/workflows',
      name: 'workflows',
      order: 20
    }
  ],

  settingsTabs: [
    {
      id: 'workflow',
      label: 'Workflow',
      description: 'Configure workflow editor',
      icon: markRaw(CopyDocument),
      component: markRaw(defineAsyncComponent(() => import('./workflow/components/WorkflowSettingsPanel.vue'))),
      order: 10,
    }
  ]
}

export default workflowModule
