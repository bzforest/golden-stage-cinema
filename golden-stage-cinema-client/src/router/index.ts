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
      path: '/',
      name: 'home',
      component: () => import('@/views/HomeView.vue'),
    },
    {
      path: '/movie/:id',
      name: 'movie-detail',
      component: () => import('@/views/MovieDetailView.vue'),
    },
    {
      path: '/booking/:showtimeId',
      name: 'seat-map',
      component: () => import('@/views/SeatMapView.vue'),
    },
  ],
  scrollBehavior(to, from, savedPosition) {
    // ถ้าย้อนกลับ (Back/Forward) ให้จำตำแหน่งเดิม แต่ถ้ากดไปหน้าใหม่ ให้ขึ้นไปบนสุด
    if (savedPosition) {
      return savedPosition
    } else {
      return { top: 0, behavior: 'smooth' } // เติม smooth ให้มันเลื่อนขึ้นแบบสมูทๆ ได้ด้วยครับ
    }
  }
})

export default router
