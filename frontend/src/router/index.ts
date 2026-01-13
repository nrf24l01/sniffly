import { useAuthStore } from '@/stores/auth'
import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    { path: '/', name: 'Home', component: () => import('@/views/Home.vue'), meta: { requiresAuth: true } },
    { path: '/captures', name: 'Capturers', component: () => import('@/views/Capturers.vue'), meta: { requiresAuth: true } },
    { path: '/login', name: 'Login', component: () => import('@/views/Login.vue'), meta: { noWhenAuth: true } },
  ],
})

router.beforeEach((to, from, next) => {
  const auth = useAuthStore()

  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    next({ name: 'Login' })
  } else if (to.meta.noWhenAuth && auth.isAuthenticated) {
    next({ name: 'Home' })
  } else {
    next()
  }
})

export default router
