import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/design',
      name: 'design-system',
      component: () => import('@/views/DesignSystemView.vue'),
    },
    {
      // Redirect หน้าแรกไป Design System ไว้ก่อน (จะเปลี่ยนทีหลังเป็น Landing Page)
      path: '/',
      redirect: '/design',
    },
  ],
})

export default router
