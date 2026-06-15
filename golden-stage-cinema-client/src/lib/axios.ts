import axios from 'axios'
import { auth } from './firebase'

export const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  }
})

api.interceptors.request.use(
  async (config) => {
    const user = auth.currentUser
    if (user) {
      try {
        const token = await user.getIdToken()
        config.headers.Authorization = `Bearer ${token}`
      } catch (error) {
        console.error('Failed to get Firebase token', error)
      }
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)
