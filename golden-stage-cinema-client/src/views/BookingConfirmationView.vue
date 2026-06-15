<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useMovieStore, type ShowtimeSeat } from '@/stores/useMovieStore'
import { Button } from '@/components/ui/button'
import Navbar from '@/components/layout/Navbar.vue'
import { CheckCircle2Icon, DownloadIcon, HomeIcon, TicketIcon, MapPinIcon, CalendarIcon, ClockIcon } from '@lucide/vue'

const route = useRoute()
const router = useRouter()
const movieStore = useMovieStore()

const showtimeId = route.params.showtimeId as string

onMounted(async () => {
  if (!movieStore.currentShowtime) {
    await movieStore.fetchShowtimeById(showtimeId)
  }
  if (movieStore.currentShowtime && !movieStore.currentMovie) {
    await movieStore.fetchMovieById(movieStore.currentShowtime.movie_id)
  }
  if (movieStore.cinemas.length === 0) {
    await movieStore.fetchCinemas()
  }
  
  // If no booked seats in store, they might have refreshed. 
  // We can still show the page but it might miss seat data.
})

const movie = computed(() => movieStore.currentMovie)
const showtime = computed(() => movieStore.currentShowtime)
const cinema = computed(() => {
  if (!showtime.value) return null
  return movieStore.cinemas.find(c => c.id === showtime.value?.cinema_id)
})

const bookedSeats = computed<ShowtimeSeat[]>(() => {
  if (movieStore.lastBookedSeats && movieStore.lastBookedSeats.length > 0) {
    return movieStore.lastBookedSeats
  }
  const stored = sessionStorage.getItem('lastBookedSeats')
  if (stored) {
    try {
      const parsed = JSON.parse(stored) as ShowtimeSeat[]
      movieStore.lastBookedSeats = parsed
      return parsed
    } catch(e) {}
  }
  return []
})

// Formatting
const formattedDate = computed(() => {
  if (!showtime.value?.start_time) return 'N/A'
  const date = new Date(showtime.value.start_time)
  return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
})

const formattedTime = computed(() => {
  if (!showtime.value?.start_time) return 'N/A'
  const date = new Date(showtime.value.start_time)
  return date.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit', hour12: false })
})

const seatNumbers = computed(() => {
  if (!bookedSeats.value || bookedSeats.value.length === 0) return 'N/A'
  return bookedSeats.value.map(s => s.seat_number).join(', ')
})

// Order ID Persistence
const orderId = ref(sessionStorage.getItem('lastOrderId') || '#RMC-' + Math.random().toString(36).substring(2, 8).toUpperCase())

const downloadTicket = () => {
  alert('Downloading ticket...')
}

const returnToHome = () => {
  movieStore.lastBookedSeats = [] // clear it out
  sessionStorage.removeItem('lastBookedSeats')
  sessionStorage.removeItem('lastOrderId')
  router.push('/')
}
</script>

<template>
  <div class="min-h-screen bg-background flex flex-col font-sans">
    <Navbar />

    <main class="flex-grow flex items-center justify-center pt-24 pb-12 px-4">
      <div class="w-full max-w-[448px] mx-auto flex flex-col items-center bg-gray-500/20 p-8 rounded-3xl">
        
        <!-- Success Icon & Message -->
        <div class="w-16 h-16 rounded-full bg-primary/20 flex items-center justify-center mb-6">
          <CheckCircle2Icon class="w-8 h-8 text-primary" />
        </div>
        
        <h1 class="text-2xl font-black text-white text-center mb-2">Booking Confirmed!</h1>
        <p class="text-sm text-muted-foreground text-center mb-10">Your ticket has been sent to your email.</p>

        <!-- Ticket Card -->
        <div class="w-full bg-card/80 border border-border/40 rounded-3xl p-6 backdrop-blur-md relative mb-8 shadow-2xl shadow-primary/5">
          <!-- Notches -->
          <div class="absolute left-[-12px] top-[140px] w-6 h-6 rounded-full bg-background border-r border-border/40 z-10"></div>
          <div class="absolute right-[-12px] top-[140px] w-6 h-6 rounded-full bg-background border-l border-border/40 z-10"></div>
          <div class="absolute left-6 right-6 top-[152px] border-t-2 border-dashed border-border/40"></div>

          <!-- Top Section: Movie Info -->
          <div class="flex gap-4 pb-6 mb-6">
            <div class="w-[72px] min-w-[72px] h-[108px] rounded-lg overflow-hidden shadow-md bg-muted">
               <img v-if="movie?.posterUrl" :src="movie.posterUrl" class="w-full h-full object-cover" />
            </div>
            <div class="flex flex-col justify-center overflow-hidden">
              <div class="flex items-center gap-1.5 mb-1">
                <TicketIcon class="w-3.5 h-3.5 text-primary shrink-0" />
                <span class="text-[10px] font-bold text-primary tracking-wider uppercase">E-Ticket</span>
              </div>
              <h2 class="text-lg font-black text-white leading-tight mb-2 truncate">{{ movie?.title || 'Loading...' }}</h2>
              <div class="flex items-center gap-1 text-xs text-muted-foreground">
                <MapPinIcon class="w-3 h-3 shrink-0" />
                <span class="truncate">{{ cinema?.name || 'Golden Stage Cinema' }}</span>
              </div>
            </div>
          </div>

          <!-- Middle Section: Details -->
          <div class="grid grid-cols-2 gap-y-4 gap-x-2 pt-2 mb-8">
            <div>
              <p class="text-[10px] font-bold text-muted-foreground uppercase tracking-wider mb-1">Date</p>
              <div class="flex items-center gap-1.5 text-sm text-white font-medium">
                <CalendarIcon class="w-3.5 h-3.5 text-muted-foreground" />
                {{ formattedDate }}
              </div>
            </div>
            <div>
              <p class="text-[10px] font-bold text-muted-foreground uppercase tracking-wider mb-1">Time</p>
              <div class="flex items-center gap-1.5 text-sm text-white font-medium">
                <ClockIcon class="w-3.5 h-3.5 text-muted-foreground" />
                {{ formattedTime }}
              </div>
            </div>
            <div>
              <p class="text-[10px] font-bold text-muted-foreground uppercase tracking-wider mb-1">Seats</p>
              <div class="flex items-center gap-1.5 text-sm text-white font-medium">
                <TicketIcon class="w-3.5 h-3.5 text-muted-foreground" />
                <span class="truncate">{{ seatNumbers }}</span>
              </div>
            </div>
            <div>
              <p class="text-[10px] font-bold text-muted-foreground uppercase tracking-wider mb-1">Order ID</p>
              <div class="text-sm text-white font-medium">
                {{ orderId }}
              </div>
            </div>
          </div>

          <!-- Bottom Section: Barcode Mock -->
          <div class="w-full h-12 flex justify-between items-center opacity-80">
            <!-- Simple CSS Barcode -->
            <div class="w-1 h-full bg-white mx-[1px]"></div>
            <div class="w-2 h-full bg-white mx-[1px]"></div>
            <div class="w-1 h-full bg-white mx-[1px]"></div>
            <div class="w-[2px] h-full bg-white mx-[1px]"></div>
            <div class="w-3 h-full bg-white mx-[1px]"></div>
            <div class="w-1 h-full bg-white mx-[1px]"></div>
            <div class="w-2 h-full bg-white mx-[1px]"></div>
            <div class="w-[2px] h-full bg-white mx-[1px]"></div>
            <div class="w-1 h-full bg-white mx-[1px]"></div>
            <div class="w-4 h-full bg-white mx-[1px]"></div>
            <div class="w-[2px] h-full bg-white mx-[1px]"></div>
            <div class="w-1 h-full bg-white mx-[1px]"></div>
            <div class="w-2 h-full bg-white mx-[1px]"></div>
            <div class="w-[3px] h-full bg-white mx-[1px]"></div>
            <div class="w-1 h-full bg-white mx-[1px]"></div>
            <div class="w-2 h-full bg-white mx-[1px]"></div>
            <div class="w-1 h-full bg-white mx-[1px]"></div>
            <div class="w-[2px] h-full bg-white mx-[1px]"></div>
            <div class="w-3 h-full bg-white mx-[1px]"></div>
            <div class="w-1 h-full bg-white mx-[1px]"></div>
            <div class="w-[2px] h-full bg-white mx-[1px]"></div>
            <div class="w-2 h-full bg-white mx-[1px]"></div>
            <div class="w-1 h-full bg-white mx-[1px]"></div>
            <div class="w-[2px] h-full bg-white mx-[1px]"></div>
            <div class="w-3 h-full bg-white mx-[1px]"></div>
            <div class="w-1 h-full bg-white mx-[1px]"></div>
          </div>
        </div>

        <!-- Action Buttons -->
        <div class="w-full space-y-3">
          <Button @click="downloadTicket" class="w-full h-12 rounded-xl text-sm font-bold bg-primary text-primary-foreground hover:bg-primary/90 shadow-lg shadow-primary/20 flex items-center justify-center gap-2 cursor-pointer">
            <DownloadIcon class="w-4 h-4" />
            Download Ticket
          </Button>
          <Button @click="returnToHome" variant="outline" class="w-full h-12 rounded-xl text-sm font-bold border-border/40 text-muted-foreground hover:text-white flex items-center justify-center gap-2 cursor-pointer">
            <HomeIcon class="w-4 h-4" />
            Return to Home
          </Button>
        </div>

      </div>
    </main>
  </div>
</template>
