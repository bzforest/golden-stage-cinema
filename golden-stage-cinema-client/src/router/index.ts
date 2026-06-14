import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/useAuthStore'

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
      meta: { requiresAuth: true }
    },
    {
      path: '/booking-summary/:showtimeId',
      name: 'booking-summary',
      component: () => import('@/views/BookingSummaryView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/payment/:showtimeId',
      name: 'payment-method',
      component: () => import('@/views/PaymentMethodView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/booking-confirmation/:showtimeId',
      name: 'booking-confirmation',
      component: () => import('@/views/BookingConfirmationView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/LoginView.vue'),
      meta: { requiresGuest: true }
    },
  ],
  scrollBehavior(to, from, savedPosition) {
    if (savedPosition) {
      return savedPosition
    } else {
      return { top: 0, behavior: 'smooth' }
    }
  }
})

router.beforeEach(async (to, from, next) => {
  const authStore = useAuthStore()
  await authStore.waitForInit()

  if (to.meta.requiresAuth && !authStore.user) {
    next({ path: '/login', query: { redirect: to.fullPath } })
  } else if (to.meta.requiresGuest && authStore.user) {
    next({ path: '/' })
  } else {
    next()
  }
})

export default router
