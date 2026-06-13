<script setup lang="ts">
import { UserIcon } from '@lucide/vue'
import SearchBar from '@/components/SearchBar.vue'
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

      <!-- Right: Profile Icon -->
      <div class="flex items-center justify-end gap-4 md:w-1/3">
        <button class="w-9 h-9 rounded-full flex items-center justify-center hover:bg-secondary/80 transition-colors text-muted-foreground hover:text-foreground">
          <UserIcon class="w-5 h-5 cursor-pointer" />
        </button>
      </div>

    </div>
  </header>
</template>
