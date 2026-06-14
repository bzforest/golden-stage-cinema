<script setup lang="ts">
import { UserIcon, LogOutIcon } from '@lucide/vue'
import SearchBar from '@/components/SearchBar.vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/useAuthStore'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const handleLogout = async () => {
  try {
    await authStore.logout()
    router.push('/login')
  } catch (error) {
    console.error(error)
  }
}
</script>

<template>
  <header class="border-b border-border/50 bg-background/80 backdrop-blur-md sticky top-0 z-50">
    <div class="max-w-7xl mx-auto px-6 py-4 flex items-center justify-between">
      
      <!-- Left: Logo -->
      <div class="flex items-center gap-2 md:w-1/3 cursor-pointer" @click="$router.push('/')">
        <div class="w-8 h-8 flex items-center justify-center">
          <!-- Using a clapperboard or ticket icon like the reference if possible, but the current logo works -->
          <svg xmlns="http://www.w3.org/2000/svg" class="w-6 h-6 text-primary" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M19.82 2H4.18C2.97 2 2 2.97 2 4.18v15.64C2 21.03 2.97 22 4.18 22h15.64c1.21 0 2.18-.97 2.18-2.18V4.18C22 2.97 21.03 2 19.82 2Z"/>
            <path d="M7 2v20"/>
            <path d="M17 2v20"/>
            <path d="M2 12h20"/>
            <path d="M2 7h5"/>
            <path d="M2 17h5"/>
            <path d="M17 17h5"/>
            <path d="M17 7h5"/>
          </svg>
        </div>
        <span class="text-xl font-bold text-heading tracking-wide hidden sm:inline">
          <span class="text-foreground">Golden</span><span class="text-primary">Stage</span>
        </span>
      </div>

      <!-- Center: Search Bar -->
      <div class="flex-1 flex justify-center px-4 md:px-0 md:w-1/3">
        <div class="w-full max-w-md">
          <SearchBar />
        </div>
      </div>

      <!-- Right: Profile Icon & Auth -->
      <div class="flex items-center justify-end gap-3 md:w-1/3">
        <template v-if="!authStore.isLoading">
          <template v-if="authStore.user">
            <span class="text-xs text-muted-foreground hidden sm:inline-block border border-border/50 px-2 py-1 rounded-full bg-muted/20">
              {{ authStore.user.displayName || authStore.user.email }}
            </span>
            <button 
              @click="handleLogout"
              title="Sign Out"
              class="w-9 h-9 rounded-full flex items-center justify-center hover:bg-red-900/20 hover:text-red-500 transition-colors text-muted-foreground"
            >
              <LogOutIcon class="w-4 h-4 cursor-pointer" />
            </button>
          </template>
          <template v-else>
            <button 
              @click="router.push({ path: '/login', query: { redirect: route.fullPath } })"
              class="px-5 py-2 text-sm font-bold rounded-lg bg-primary text-primary-foreground hover:bg-primary/90 transition-all shadow-md shadow-primary/20 hover:scale-105"
            >
              Sign In
            </button>
          </template>
        </template>
      </div>

    </div>
  </header>
</template>
