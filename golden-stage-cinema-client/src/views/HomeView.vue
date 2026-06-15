<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useMovieStore } from '@/stores/useMovieStore'
import type { Movie } from '@/stores/useMovieStore'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { StarIcon, ClockIcon, PlayIcon } from '@lucide/vue'
import Navbar from '@/components/layout/Navbar.vue'
import Footer from '@/components/layout/Footer.vue'
import { toast } from 'vue-sonner'

const movieStore = useMovieStore()
const router = useRouter()

const heroMovie = ref<Movie | null>(null)
const trendingMovies = ref<Movie[]>([])

onMounted(() => {
  movieStore.fetchMovies()
})

watch(() => movieStore.movies, (newMovies) => {
  if (newMovies && newMovies.length > 0) {
    // 1. สุ่มหนังสำหรับ Hero Section (1 เรื่อง)
    const randomHeroIndex = Math.floor(Math.random() * newMovies.length)
    heroMovie.value = newMovies[randomHeroIndex] || null

    // 2. สุ่มหนังสำหรับ Trending Movies (5 เรื่อง)
    const shuffled = [...newMovies].sort(() => 0.5 - Math.random())
    trendingMovies.value = shuffled.slice(0, 5)
  }
}, { immediate: true })

const handleBookClick = (movie: Movie) => {
  toast.success(`กำลังนำคุณเข้าสู่หน้าจอรายละเอียดภาพยนตร์: ${movie.title}`)
  router.push({ name: 'movie-detail', params: { id: movie.id } })
}
</script>

<template>
  <div class="min-h-screen bg-background text-foreground flex flex-col">
    <!-- Navbar -->
    <Navbar />

    <main class="flex-1">
      <!-- Error Notification Banner (if API fails) -->
      <div v-if="movieStore.error" class="bg-yellow-500/10 border-b border-yellow-500/20 px-6 py-2 text-center text-sm text-yellow-500/90 backdrop-blur-md sticky top-[73px] z-40">
        {{ movieStore.error }}
      </div>

      <!-- Hero Section -->
      <section class="relative w-full h-[80vh] min-h-[600px] flex items-center">
        <!-- Loading State for Hero -->
        <div v-if="movieStore.isLoading || !heroMovie" class="absolute inset-0 z-0 bg-muted animate-pulse"></div>

        <!-- Hero Background -->
        <div v-else-if="heroMovie" class="absolute inset-0 z-0 overflow-hidden">
          <img 
            :src="heroMovie.backdropUrl || heroMovie.posterUrl" 
            alt="Hero Backdrop"
            class="w-full h-full object-cover mix-blend-screen"
          />
          <div class="absolute inset-0 bg-linear-to-t from-background via-background/30 to-transparent"></div>
          <div class="absolute inset-0 bg-linear-to-r from-background via-background/30 to-transparent"></div>
        </div>

        <!-- Hero Content -->
        <div class="relative z-10 max-w-7xl mx-auto px-6 w-full">
          <div v-if="movieStore.isLoading || !heroMovie" class="max-w-2xl space-y-5">
            <Skeleton class="h-8 w-36 rounded-full" />
            <Skeleton class="h-16 w-full" />
            <Skeleton class="h-16 w-3/4" />
            <Skeleton class="h-24 w-full" />
            <div class="flex gap-4 pt-4">
              <Skeleton class="h-12 w-40" />
              <Skeleton class="h-12 w-40" />
            </div>
          </div>

          <div v-else-if="heroMovie" class="max-w-2xl">
            <!-- Badge -->
            <div class="inline-flex items-center rounded-full border border-primary/30 bg-primary/10 px-3 py-1 text-xs font-semibold text-primary mb-6 backdrop-blur-sm">
              NOW SHOWING
            </div>
            
            <!-- Title -->
            <h1 class="text-5xl md:text-6xl font-extrabold text-heading tracking-tight mb-5 leading-tight text-white drop-shadow-md">
              {{ heroMovie.title }}
            </h1>
            
            <!-- Description -->
            <p class="text-muted-foreground text-sm md:text-base leading-relaxed mb-8 max-w-xl line-clamp-3">
              {{ heroMovie.description }}
            </p>

            <!-- Meta Info -->
            <div class="flex items-center gap-6 mb-10 text-sm font-medium">
              <div class="flex items-center gap-1.5 text-primary">
                <StarIcon class="w-4 h-4 fill-primary" />
                <span class="text-white drop-shadow">{{ (heroMovie.rating || 0).toFixed(1) }}</span>
              </div>
              <div class="flex items-center gap-1.5 text-muted-foreground">
                <ClockIcon class="w-4 h-4" />
                <span>{{ heroMovie.duration }}</span>
              </div>
              <div class="text-muted-foreground">
                {{ heroMovie.genre }}
              </div>
            </div>

            <!-- Buttons -->
            <div class="flex items-center gap-4">
              <Button size="lg" class="px-8 h-12 rounded-full font-semibold text-primary-foreground bg-primary hover:bg-primary/90 shadow-lg shadow-primary/20 hover:scale-105 transition-transform cursor-pointer" @click="handleBookClick(heroMovie)">
                Book Tickets
              </Button>
              <Button size="lg" variant="secondary" class="px-8 h-12 rounded-full font-semibold bg-secondary/80 text-foreground backdrop-blur-md hover:bg-secondary border border-border/50 hover:scale-105 transition-transform cursor-pointer">
                <PlayIcon class="w-5 h-5 mr-2" />
                Watch Trailer
              </Button>
            </div>
          </div>
        </div>
      </section>

      <!-- Trending Movies Section -->
      <section class="max-w-7xl mx-auto px-6 py-16 w-full -mt-10 relative z-20">
        <!-- Section Header -->
        <div class="flex items-end justify-between mb-8">
          <div class="flex items-center gap-3">
            <div class="w-1.5 h-7 bg-primary rounded-sm shadow-[0_0_10px_rgba(var(--primary),0.5)]"></div>
            <h2 class="text-2xl md:text-3xl font-bold text-heading text-white">Trending Movies</h2>
          </div>
          <router-link :to="{ name: 'movies' }" class="text-sm font-semibold text-primary hover:text-primary/80 transition-colors">View All</router-link>
        </div>

        <!-- Loading Skeleton Grid -->
        <div v-if="movieStore.isLoading || trendingMovies.length === 0" class="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-5 gap-6">
          <div v-for="i in 5" :key="i" class="space-y-4 last:hidden lg:last:block">
            <Skeleton class="w-full aspect-[2/3] rounded-2xl" />
            <Skeleton class="h-5 w-full" />
            <Skeleton class="h-4 w-2/3" />
          </div>
        </div>

        <!-- Movie Grid -->
        <div v-else class="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-5 gap-6">
          <div v-for="movie in trendingMovies" :key="movie.id" class="group cursor-pointer last:hidden lg:last:block" @click="handleBookClick(movie)">
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
      </section>
    </main>

    <!-- Footer -->
    <Footer />
  </div>
</template>
