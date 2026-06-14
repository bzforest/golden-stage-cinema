<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useMovieStore } from '@/stores/useMovieStore'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { StarIcon, ClockIcon, PlayIcon, CalendarIcon, MapPinIcon, ArrowLeftIcon } from '@lucide/vue'
import Navbar from '@/components/layout/Navbar.vue'
import Footer from '@/components/layout/Footer.vue'

const route = useRoute()
const router = useRouter()
const movieStore = useMovieStore()

const movieId = route.params.id as string
const selectedDate = ref<string>('')

onMounted(async () => {
  await movieStore.fetchMovieById(movieId)
  await movieStore.fetchShowtimes(movieId)
  
  if (availableDates.value.length > 0) {
    selectedDate.value = availableDates.value[0]?.raw || ''
  } else {
    // Fallback Mock Dates for UI demonstration if backend showtimes are empty
    selectedDate.value = new Date().toISOString().split('T')[0] || ''
  }
})

// Process dates from showtimes
const availableDates = computed(() => {
  if (!movieStore.showtimes || movieStore.showtimes.length === 0) {
    // Generate some mock dates for the UI to look good if no showtimes exist
    const mock = []
    const today = new Date()
    for (let i=0; i<5; i++) {
      const d = new Date(today)
      d.setDate(today.getDate() + i)
      mock.push({
        month: d.toLocaleDateString('en-US', { month: 'short' }),
        day: d.toLocaleDateString('en-US', { day: '2-digit' }),
        weekday: d.toLocaleDateString('en-US', { weekday: 'short' }),
        raw: d.toISOString().split('T')[0]
      })
    }
    return mock
  }
  
  const datesMap = new Map<string, { month: string, day: string, weekday: string, raw: string }>()
  
  movieStore.showtimes.forEach(st => {
    const d = new Date(st.start_time)
    // Use local timezone to group dates correctly and prevent UTC mismatch duplicates
    const dateKey = `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
    
    if (dateKey && !datesMap.has(dateKey)) {
      datesMap.set(dateKey, {
        month: d.toLocaleDateString('en-US', { month: 'short' }).toUpperCase(),
        day: d.toLocaleDateString('en-US', { day: '2-digit' }),
        weekday: d.toLocaleDateString('en-US', { weekday: 'short' }),
        raw: dateKey
      })
    }
  })
  
  // Sort dates
  return Array.from(datesMap.values()).sort((a, b) => a.raw.localeCompare(b.raw))
})

const formatTime = (timeStr: string) => {
  const d = new Date(timeStr)
  return d.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit', hour12: false })
}

// Group showtimes by cinema and hall for the selected date
const filteredShowtimes = computed(() => {
  if (!selectedDate.value) return []
  
  // 1. Filter by selected date
  const dayShowtimes = movieStore.showtimes.filter(st => {
    const d = new Date(st.start_time)
    const stDateKey = `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
    return stDateKey === selectedDate.value
  })
  
  if (dayShowtimes.length === 0) {
    // Return mock showtime structure if no real ones exist for visual demonstration
    return [
      {
        cinema: { id: 'mock1', name: 'RomeoCinematic Central Mall' },
        halls: [
          {
            hall: { id: 'h1', name: 'Hall 1' },
            showtimes: [
              { id: 's1', start_time: `${selectedDate.value}T12:30:00Z` },
              { id: 's2', start_time: `${selectedDate.value}T15:45:00Z` },
              { id: 's3', start_time: `${selectedDate.value}T18:15:00Z` },
              { id: 's4', start_time: `${selectedDate.value}T20:30:00Z` },
            ]
          }
        ]
      }
    ]
  }
  
  // 2. Group by Cinema -> Hall
  const grouped: any[] = []
  
  // Find unique cinemas in today's showtimes
  const cinemaIds = [...new Set(dayShowtimes.map(st => st.cinema_id))]
  
  cinemaIds.forEach(cid => {
    const cinema = movieStore.cinemas.find(c => c.id === cid) || { id: cid, name: 'Unknown Cinema' }
    const cinemaShowtimes = dayShowtimes.filter(st => st.cinema_id === cid)
    
    // Find unique halls in this cinema
    const hallIds = [...new Set(cinemaShowtimes.map(st => st.hall_id))]
    const hallsGroup = hallIds.map(hid => {
      const hall = movieStore.halls.find(h => h.id === hid) || { id: hid, name: 'Unknown Hall' }
      const hallShowtimes = cinemaShowtimes.filter(st => st.hall_id === hid)
        .sort((a, b) => new Date(a.start_time).getTime() - new Date(b.start_time).getTime())
        
      return { hall, showtimes: hallShowtimes }
    })
    
    grouped.push({
      cinema,
      halls: hallsGroup
    })
  })
  
  return grouped
})

const handleShowtimeClick = (showtimeId: string) => {
  // Find the showtime and set it in the store before navigating
  const showtime = movieStore.showtimes.find(s => s.id === showtimeId)
  if (showtime) {
    movieStore.currentShowtime = showtime
  }
  router.push({ name: 'seat-map', params: { showtimeId } })
}
</script>

<template>
  <div class="min-h-screen bg-background text-foreground flex flex-col relative">
    <Navbar />

    <!-- Backdrop Image (Fixed) -->
    <div class="fixed inset-0 z-0 pointer-events-none" v-if="movieStore.currentMovie">
      <img :src="movieStore.currentMovie.backdropUrl || movieStore.currentMovie.posterUrl" class="w-full h-full object-cover opacity-80 mix-blend-screen" />
      <div class="absolute inset-0 bg-gradient-to-t from-background via-background/80 to-transparent"></div>
      <div class="absolute inset-0 bg-gradient-to-r from-background via-background/60 to-transparent"></div>
    </div>

    <main class="flex-1 relative z-10 py-10 px-6">
      <div class="max-w-6xl mx-auto">
        <!-- Back Button -->
        <button @click="router.back()" class="inline-flex items-center gap-2 px-4 py-2 rounded-full border border-border/50 bg-card/40 backdrop-blur-sm hover:bg-card hover:border-primary/50 transition-colors mb-8 text-sm font-medium">
          <ArrowLeftIcon class="w-4 h-4" /> Back
        </button>

        <div v-if="movieStore.isLoading && !movieStore.currentMovie" class="grid grid-cols-1 md:grid-cols-3 gap-10">
          <div class="col-span-1">
            <Skeleton class="w-full aspect-[2/3] rounded-2xl" />
          </div>
          <div class="col-span-1 md:col-span-2 space-y-4">
            <Skeleton class="h-8 w-40 rounded-full" />
            <Skeleton class="h-16 w-3/4" />
            <Skeleton class="h-32 w-full" />
          </div>
        </div>

        <div v-else-if="movieStore.currentMovie" class="flex flex-col md:flex-row gap-10 lg:gap-16">
          
          <!-- Left Column (Poster) -->
          <div class="w-full md:w-1/3 lg:w-1/4 flex flex-col gap-6 shrink-0 items-center">
            <div class="w-[270px] h-[406px] rounded-2xl overflow-hidden shadow-[0_0_30px_rgba(var(--primary),0.15)] border border-border/30 bg-muted">
              <img :src="movieStore.currentMovie.posterUrl" class="w-full h-full object-cover" />
            </div>
            <Button variant="secondary" class="w-72 md:w-full h-12 rounded-xl text-base font-semibold bg-secondary/80 text-foreground border border-border/50 backdrop-blur-md hover:bg-secondary cursor-pointer">
              <PlayIcon class="w-5 h-5 mr-2" /> Watch Trailer
            </Button>
          </div>

          <!-- Right Column (Details & Showtimes) -->
          <div class="w-full md:w-2/3 lg:w-3/4 flex-1">
            
            <!-- Badges -->
            <div class="flex flex-wrap items-center gap-3 mb-5">
              <span class="px-3 py-1.5 rounded-full border border-primary/30 bg-primary/10 text-primary text-xs font-bold">{{ movieStore.currentMovie.genre }}</span>
              <span class="px-3 py-1.5 rounded-full border border-border/50 bg-card/50 text-white text-xs font-bold flex items-center gap-1.5">
                <StarIcon class="w-3.5 h-3.5 fill-primary text-primary" /> {{ (movieStore.currentMovie.rating || 0).toFixed(1) }}
              </span>
              <span class="px-3 py-1.5 rounded-full border border-border/50 bg-card/50 text-muted-foreground text-xs font-semibold flex items-center gap-1.5">
                <ClockIcon class="w-3.5 h-3.5" /> {{ movieStore.currentMovie.duration }}
              </span>
            </div>
            
            <!-- Title -->
            <h1 class="text-4xl md:text-5xl lg:text-6xl font-extrabold text-heading mb-6 leading-tight drop-shadow-sm text-white">
              {{ movieStore.currentMovie.title }}
            </h1>
            
            <!-- Synopsis -->
            <p class="text-muted-foreground/90 text-sm md:text-base leading-relaxed mb-12 max-w-3xl">
              {{ movieStore.currentMovie.description }}
            </p>
            
            <!-- Showtimes Section -->
            <div>
              <div class="flex items-center gap-2.5 mb-6">
                <CalendarIcon class="w-5 h-5 text-primary" />
                <h2 class="text-xl md:text-2xl font-bold text-heading text-white">Select Date & Time</h2>
              </div>
              
              <!-- Date Selector -->
              <div class="flex gap-3 overflow-x-auto pb-4 mb-8 scrollbar-hide">
                <button v-for="date in availableDates" :key="date.raw" 
                  @click="selectedDate = date.raw || ''"
                  class="flex flex-col items-center justify-center min-w-[70px] h-[84px] rounded-xl border transition-all duration-300 cursor-pointer"
                  :class="selectedDate === date.raw ? 'bg-primary border-primary text-primary-foreground shadow-lg shadow-primary/20 scale-105' : 'bg-card border-border/50 hover:border-primary/50 text-muted-foreground'">
                  <span class="text-[10px] font-bold uppercase tracking-widest mt-1">{{ date.month }}</span>
                  <span class="text-2xl font-black leading-none my-1" :class="{'text-white': selectedDate !== date.raw}">{{ date.day }}</span>
                  <span class="text-[10px] font-medium mb-1">{{ date.weekday }}</span>
                </button>
              </div>
              
              <!-- Cinema List -->
              <div class="space-y-6">
                <div v-for="group in filteredShowtimes" :key="group.cinema.id" class="bg-card/40 rounded-2xl border border-border/40 p-5 md:p-6 backdrop-blur-md">
                  <div class="flex flex-col sm:flex-row sm:items-center justify-between mb-5 pb-5 border-b border-border/30 gap-2">
                    <div class="flex items-center gap-2.5">
                      <MapPinIcon class="w-5 h-5 text-muted-foreground" />
                      <h3 class="font-bold text-lg text-white">{{ group.cinema.name }}</h3>
                    </div>
                    <span class="text-xs font-semibold text-muted-foreground">{{ group.cinema.cinema_type || 'Standard 2D' }}</span>
                  </div>
                  
                  <div v-for="hallGroup in group.halls" :key="hallGroup.hall.id" class="mb-6 last:mb-0">
                    <div class="flex items-center gap-2 mb-3">
                      <div class="w-1 h-4 bg-primary rounded-full"></div>
                      <h4 class="text-sm font-bold text-white">{{ hallGroup.hall.name }}</h4>
                    </div>
                    <div class="flex flex-wrap gap-3">
                      <button v-for="showtime in hallGroup.showtimes" :key="showtime.id"
                        @click="handleShowtimeClick(showtime.id)"
                        class="px-5 py-2.5 rounded-lg border border-border/60 hover:border-primary hover:text-primary transition-colors text-sm font-medium bg-background/50 hover:bg-primary/5 cursor-pointer">
                        {{ formatTime(showtime.start_time) }}
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            </div>

          </div>
        </div>

        <div v-else class="text-center py-20 text-muted-foreground">
          Movie not found.
        </div>
      </div>
    </main>

    <Footer />
  </div>
</template>

<style scoped>
/* Hide scrollbar for Chrome, Safari and Opera */
.scrollbar-hide::-webkit-scrollbar {
  display: none;
}
/* Hide scrollbar for IE, Edge and Firefox */
.scrollbar-hide {
  -ms-overflow-style: none;  /* IE and Edge */
  scrollbar-width: none;  /* Firefox */
}
</style>
