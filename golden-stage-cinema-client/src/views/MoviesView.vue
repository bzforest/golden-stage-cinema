<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useMovieStore } from '@/stores/useMovieStore'
import type { Movie } from '@/stores/useMovieStore'
import { Skeleton } from '@/components/ui/skeleton'
import { StarIcon } from '@lucide/vue'
import Navbar from '@/components/layout/Navbar.vue'
import Footer from '@/components/layout/Footer.vue'

const movieStore = useMovieStore()
const router = useRouter()

onMounted(() => {
  if (movieStore.movies.length === 0) {
    movieStore.fetchMovies()
  }
})

const handleBookClick = (movie: Movie) => {
  router.push({ name: 'movie-detail', params: { id: movie.id } })
}
</script>

<template>
  <div class="min-h-screen bg-background text-foreground flex flex-col">
    <Navbar />

    <main class="flex-1 max-w-7xl mx-auto px-6 py-12 w-full">
      <!-- Section Header -->
      <div class="flex items-center gap-3 mb-10">
        <div class="w-1.5 h-7 bg-primary rounded-sm shadow-[0_0_10px_rgba(var(--primary),0.5)]"></div>
        <h1 class="text-3xl font-bold text-heading text-white">All Movies</h1>
      </div>

      <!-- Loading Skeleton Grid -->
      <div v-if="movieStore.isLoading || movieStore.movies.length === 0" class="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-5 gap-6">
        <div v-for="i in 10" :key="i" class="space-y-4">
          <Skeleton class="w-full aspect-[2/3] rounded-2xl" />
          <Skeleton class="h-5 w-full" />
          <Skeleton class="h-4 w-2/3" />
        </div>
      </div>

      <!-- Movie Grid -->
      <div v-else class="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-5 gap-6">
        <div v-for="movie in movieStore.movies" :key="movie.id" class="group cursor-pointer" @click="handleBookClick(movie)">
          <!-- Poster Wrapper -->
          <div class="relative overflow-hidden rounded-2xl mb-4 aspect-[2/3] bg-muted border border-border/40 shadow-lg group-hover:border-primary/50 transition-colors duration-300">
            <img 
              :src="movie.posterUrl" 
              :alt="movie.title"
              class="w-full h-full object-cover transition-transform duration-500 group-hover:scale-110"
            />
            
            <!-- Rating Badge -->
            <div class="absolute top-3 right-3 bg-background/80 backdrop-blur-md px-2 py-1 rounded-md flex items-center gap-1 shadow-sm border border-border/50">
              <StarIcon class="w-3.5 h-3.5 text-primary fill-primary" />
              <span class="text-xs font-bold text-white">{{ (movie.rating || 0).toFixed(1) }}</span>
            </div>
            
            <!-- Hover Overlay (Book Now) -->
            <div class="absolute inset-0 bg-gradient-to-t from-background/90 via-background/20 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300 flex items-end justify-center pb-8">
              <div class="px-6 py-2 bg-primary text-primary-foreground font-semibold rounded-full text-sm transform translate-y-4 group-hover:translate-y-0 transition-transform duration-300 shadow-lg shadow-primary/30">
                Book Now
              </div>
            </div>
          </div>
          
          <!-- Movie Info -->
          <h3 class="font-bold text-heading text-base md:text-lg truncate group-hover:text-primary transition-colors text-white">{{ movie.title }}</h3>
          <div class="flex items-center gap-2 mt-1.5 opacity-70">
            <p class="text-xs truncate">{{ movie.genre }}</p>
            <span class="text-[10px]">•</span>
            <p class="text-xs whitespace-nowrap">{{ movie.duration }}</p>
          </div>
        </div>
      </div>
    </main>

    <Footer />
  </div>
</template>
