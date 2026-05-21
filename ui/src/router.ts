import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'main',
      component: () => import('./layouts/MainLayout.vue'),
      children: []
    },
    {
      path: '/:pathMatch(.*)*',
      name: 'not-found',
      component: () => import('./shared/views/NotFoundView.vue')
    }
  ]
})

export default router
