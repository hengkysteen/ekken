import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'main',
      component: () => import('./layouts/MainLayout.vue'),
      children: [
        {
          path: '',
          name: 'dashboard',
          component: () => import('./shared/views/AppDashboard.vue'),
          meta: { breadcrumb: 'Dashboard' }
        },
        {
          path: 'settings',
          name: 'settings',
          component: () => import('./shared/views/SettingsView.vue'),
          meta: { breadcrumb: 'Settings' }
        },
        {
          path: 'compo',
          name: 'compo',
          component: () => import('./shared/views/Compo.vue'),
          meta: { breadcrumb: 'Compo' }
        }
      ]
    },
    {
      path: '/:pathMatch(.*)*',
      name: 'not-found',
      component: () => import('./shared/views/NotFoundView.vue')
    }
  ]
})

export default router
