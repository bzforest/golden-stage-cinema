import { defineStore } from 'pinia'
import { ref } from 'vue'
import { auth } from '@/lib/firebase'
import { onAuthStateChanged, signOut, type User } from 'firebase/auth'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const isLoading = ref(true)
  const isAdmin = ref(false)

  // Wait for the first auth state check to complete
  let resolveInit: (value: unknown) => void
  const initPromise = new Promise((resolve) => {
    resolveInit = resolve
  })

  // Listen to Firebase Auth state changes
  onAuthStateChanged(auth, async (currentUser) => {
    if (currentUser) {
      try {
        const idTokenResult = await currentUser.getIdTokenResult(true)
        isAdmin.value = !!idTokenResult.claims.admin || idTokenResult.claims.role === 'admin'
      } catch (err) {
        console.error("Failed to fetch custom claims", err)
        isAdmin.value = false
      }
    } else {
      isAdmin.value = false
    }

    user.value = currentUser

    if (isLoading.value) {
      isLoading.value = false
      resolveInit(true)
    }
  })

  const waitForInit = () => initPromise

  const logout = async () => {
    try {
      await signOut(auth)
    } catch (error) {
      console.error('Logout failed:', error)
      throw error
    }
  }

  return {
    user,
    isLoading,
    isAdmin,
    waitForInit,
    logout
  }
})
