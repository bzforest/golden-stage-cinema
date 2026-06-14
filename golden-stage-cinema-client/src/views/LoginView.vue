<script setup lang="ts">
import { ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { auth } from '@/lib/firebase'
import { signInWithEmailAndPassword, createUserWithEmailAndPassword, updateProfile } from 'firebase/auth'
import Navbar from '@/components/layout/Navbar.vue'
import Footer from '@/components/layout/Footer.vue'
import { Button } from '@/components/ui/button'

const router = useRouter()
const route = useRoute()

const mode = ref<'login' | 'register'>('login')
const displayName = ref('')
const email = ref('')
const password = ref('')
const confirmPassword = ref('')
const isLoading = ref(false)
const errorMessage = ref('')

const handleSubmit = async () => {
  if (!email.value || !password.value) {
    errorMessage.value = 'Please enter both email and password.'
    return
  }
  
  if (mode.value === 'register' && password.value !== confirmPassword.value) {
    errorMessage.value = 'Passwords do not match.'
    return
  }

  isLoading.value = true
  errorMessage.value = ''

  try {
    if (mode.value === 'login') {
      await signInWithEmailAndPassword(auth, email.value, password.value)
    } else {
      const userCred = await createUserWithEmailAndPassword(auth, email.value, password.value)
      if (displayName.value) {
        await updateProfile(userCred.user, { displayName: displayName.value })
      }
    }
    // Smart Redirect
    const redirectPath = route.query.redirect?.toString() || '/'
    router.push(redirectPath)
  } catch (error: any) {
    console.error('Authentication error:', error)
    if (error.code === 'auth/invalid-credential' || error.code === 'auth/wrong-password') {
      errorMessage.value = 'Invalid email or password.'
    } else if (error.code === 'auth/email-already-in-use') {
      errorMessage.value = 'This email is already registered.'
    } else if (error.code === 'auth/weak-password') {
      errorMessage.value = 'Password should be at least 6 characters.'
    } else {
      errorMessage.value = error.message || 'An error occurred during authentication.'
    }
  } finally {
    isLoading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen bg-background text-foreground flex flex-col">
    <Navbar />
    
    <main class="flex-1 flex items-center justify-center p-4 relative overflow-hidden">
      <!-- Decorative Background Elements -->
      <div class="absolute top-1/4 left-1/4 w-96 h-96 bg-primary/10 rounded-full blur-3xl -z-10 animate-pulse"></div>
      <div class="absolute bottom-1/4 right-1/4 w-96 h-96 bg-blue-500/10 rounded-full blur-3xl -z-10 animate-pulse" style="animation-delay: 2s;"></div>

      <div class="w-full max-w-md bg-card/80 backdrop-blur-md border border-border/50 rounded-2xl shadow-2xl overflow-hidden animate-in fade-in zoom-in-95 duration-500">
        
        <!-- Header & Tabs -->
        <div class="p-6 pb-0">
          <div class="text-center mb-8">
            <h1 class="text-3xl font-bold font-heading text-primary bg-clip-text text-transparent bg-gradient-to-r from-primary to-yellow-300">Golden Stage</h1>
            <p class="text-muted-foreground text-sm mt-2">Welcome to the ultimate cinema experience</p>
          </div>
          
          <div class="flex p-1.5 bg-muted/50 rounded-xl mb-6">
            <button 
              @click="mode = 'login'; errorMessage = ''" 
              class="flex-1 py-2 text-sm font-semibold rounded-lg transition-all duration-300"
              :class="mode === 'login' ? 'bg-background text-primary shadow-sm' : 'text-muted-foreground hover:text-foreground'"
            >
              Sign In
            </button>
            <button 
              @click="mode = 'register'; errorMessage = ''" 
              class="flex-1 py-2 text-sm font-semibold rounded-lg transition-all duration-300"
              :class="mode === 'register' ? 'bg-background text-primary shadow-sm' : 'text-muted-foreground hover:text-foreground'"
            >
              Register
            </button>
          </div>
        </div>
        
        <!-- Form -->
        <div class="p-6 pt-0">
          <form @submit.prevent="handleSubmit" class="space-y-5">
            
            <div v-if="errorMessage" class="p-3 bg-red-900/20 border border-red-900/50 rounded-lg text-red-500 text-sm text-center flex items-center justify-center gap-2">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 shrink-0" viewBox="0 0 20 20" fill="currentColor">
                <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clip-rule="evenodd" />
              </svg>
              <span>{{ errorMessage }}</span>
            </div>
            
            <div v-if="mode === 'register'" class="space-y-2">
              <label for="displayName" class="text-sm font-medium text-foreground ml-1">Display Name</label>
              <input 
                id="displayName" 
                v-model="displayName" 
                type="text" 
                placeholder="John Doe"
                :required="mode === 'register'"
                class="w-full px-4 py-3 bg-background/50 border border-input rounded-xl text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent transition-all"
              />
            </div>
            
            <div class="space-y-2">
              <label for="email" class="text-sm font-medium text-foreground ml-1">Email Address</label>
              <input 
                id="email" 
                v-model="email" 
                type="email" 
                placeholder="mockuser01@gmail.com"
                required
                class="w-full px-4 py-3 bg-background/50 border border-input rounded-xl text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent transition-all"
              />
            </div>
            
            <div class="space-y-2">
              <label for="password" class="text-sm font-medium text-foreground ml-1">Password</label>
              <input 
                id="password" 
                v-model="password" 
                type="password" 
                placeholder="••••••••"
                required
                class="w-full px-4 py-3 bg-background/50 border border-input rounded-xl text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent transition-all"
              />
            </div>

            <div v-if="mode === 'register'" class="space-y-2">
              <label for="confirmPassword" class="text-sm font-medium text-foreground ml-1">Confirm Password</label>
              <input 
                id="confirmPassword" 
                v-model="confirmPassword" 
                type="password" 
                placeholder="••••••••"
                :required="mode === 'register'"
                class="w-full px-4 py-3 bg-background/50 border border-input rounded-xl text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent transition-all"
              />
            </div>
            
            <Button type="submit" class="w-full mt-8 h-12 text-base rounded-xl font-bold tracking-wide" :disabled="isLoading">
              <span v-if="isLoading" class="flex items-center gap-2">
                <div class="w-5 h-5 border-2 border-primary-foreground border-t-transparent rounded-full animate-spin"></div>
                Please wait...
              </span>
              <span v-else>{{ mode === 'login' ? 'Sign In to Account' : 'Create New Account' }}</span>
            </Button>
          </form>
        </div>
        
      </div>
    </main>

    <Footer />
  </div>
</template>
