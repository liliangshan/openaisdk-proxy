import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

// 布局组件
import Layout from '@/views/layout.vue'
// 页面组件
import Login from '@/views/login.vue'
import Dashboard from '@/views/dashboard.vue'
import Providers from '@/views/providers.vue'
import Models from '@/views/models.vue'
import APIKeys from '@/views/api-keys.vue'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: Login,
    meta: { requiresAuth: false }
  },
  {
    path: '/',
    component: Layout,
    meta: { requiresAuth: true },
    children: [
      {
        path: '',
        name: 'Dashboard',
        component: Dashboard
      },
      {
        path: 'providers',
        name: 'Providers',
        component: Providers
      },
      {
        path: 'models',
        name: 'Models',
        component: Models
      },
      {
        path: 'api-keys',
        name: 'APIKeys',
        component: APIKeys
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory('/user/'),
  routes
})

// 路由守卫
router.beforeEach((to, _from, next) => {
  const authStore = useAuthStore()
  
  // 检查是否需要登录
  if (to.meta.requiresAuth && !authStore.isLoggedIn) {
    // 如果没有登录，跳转到登录页
    next({
      name: 'Login',
      query: { redirect: to.fullPath }
    })
  } else if (to.name === 'Login' && authStore.isLoggedIn) {
    // 如果已登录，访问登录页则跳转到首页
    next({ name: 'Dashboard' })
  } else {
    next()
  }
})

export default router
