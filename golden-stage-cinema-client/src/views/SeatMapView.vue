<script setup lang="ts">
import { onMounted, onBeforeUnmount, ref, computed, nextTick, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useMovieStore, type ShowtimeSeat } from '@/stores/useMovieStore'
import { useAuthStore } from '@/stores/useAuthStore'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { ArrowLeftIcon, MonitorIcon } from '@lucide/vue'
import Navbar from '@/components/layout/Navbar.vue'
import Footer from '@/components/layout/Footer.vue'
import { api } from '@/lib/axios'

const route = useRoute()
const router = useRouter()
const movieStore = useMovieStore()
const authStore = useAuthStore()

const showtimeId = route.params.showtimeId as string
const selectedSeatIds = ref<Set<string>>(new Set())
const seatMapContainer = ref<HTMLElement | null>(null)

// Seat rows configuration
const allRows = ['H', 'G', 'F', 'E', 'D', 'C', 'B', 'A'] // H is closest to screen, A is farthest
const premiumRows = ['A', 'B']
const normalRows = ['C', 'D', 'E', 'F', 'G', 'H']

onMounted(async () => {
  await movieStore.fetchSeatsByShowtime(showtimeId)

  // Try to find the showtime info from the store
  // If currentShowtime is not set, try to extract movie info from the first seat
  if (!movieStore.currentShowtime) {
    // Fetch showtime details by looking through our data
    await movieStore.fetchShowtimeById(showtimeId)
  }

  // If we still don't have movie info, try to load it from the showtime
  if (movieStore.currentShowtime && !movieStore.currentMovie) {
    await movieStore.fetchMovieById(movieStore.currentShowtime.movie_id)
  }

  // Populate selected seats from the fetched data
  const initialSelected = new Set<string>()
  movieStore.seats.forEach(s => {
    if (s.status === 'SELECTED') {
      initialSelected.add(s.id)
    }
  })
  selectedSeatIds.value = initialSelected
})

// WebSocket Connection
let ws: WebSocket | null = null
let reconnectTimeout: ReturnType<typeof setTimeout> | null = null

const connectWebSocket = () => {
  if (ws) return
  
  const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const host = import.meta.env.VITE_API_BASE_URL 
    ? new URL(import.meta.env.VITE_API_BASE_URL).host 
    : 'localhost:8080'
    
  ws = new WebSocket(`${wsProtocol}//${host}/ws/seats/${showtimeId}`)
  
  ws.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data)
      if (data.seat_number && data.status) {
        let finalStatus = data.status
        
        // Handle WebSocket Boomerang (our own lock comes back as LOCKED)
        if (data.status === 'LOCKED') {
          if (data.user_id && authStore.user && data.user_id === authStore.user.uid) {
            finalStatus = 'SELECTED'
          }
        }
        
        movieStore.updateSeatStatus(data.seat_number, finalStatus as any)
        
        // Note: If seat gets booked by someone else, we might want to unselect it
        if (data.status === 'BOOKED' && selectedSeatIds.value.has(data.seat_id)) {
           const newSet = new Set(selectedSeatIds.value)
           newSet.delete(data.seat_id)
           selectedSeatIds.value = newSet
        }
      }
    } catch (e) {
      console.error('Failed to parse WS message:', e)
    }
  }
  
  ws.onclose = () => {
    ws = null
    reconnectTimeout = setTimeout(connectWebSocket, 3000)
  }
  
  ws.onerror = (error) => {
    console.error('WebSocket error:', error)
    ws?.close()
  }
}

// Auto-center the seat map on mobile when data finishes loading
watch(() => movieStore.isLoading, (isLoading) => {
  if (!isLoading && movieStore.seats.length > 0) {
    nextTick(() => {
      centerSeatMap()
    })
    connectWebSocket()
  }
})

// Prepare seat release for unmounting or navigating away
const releaseSelectedSeats = () => {
  if (selectedSeatIds.value.size > 0) {
    const seatsToRelease = movieStore.seats.filter(s => selectedSeatIds.value.has(s.id))
    seatsToRelease.forEach(seat => {
      api.delete(`/bookings/lock/${showtimeId}/${seat.seat_number}`).catch(e => console.error('Unlock failed', e))
    })
  }
}

onBeforeUnmount(() => {
  if (reconnectTimeout) clearTimeout(reconnectTimeout)
  if (ws) {
    ws.onclose = null // prevent reconnect
    ws.close()
    ws = null
  }
  releaseSelectedSeats()
  selectedSeatIds.value.clear()
})

const centerSeatMap = () => {
  if (seatMapContainer.value) {
    const el = seatMapContainer.value
    el.scrollLeft = (el.scrollWidth - el.clientWidth) / 2
  }
}

// Group seats by row
const seatsByRow = computed(() => {
  const map = new Map<string, ShowtimeSeat[]>()

  movieStore.seats.forEach(seat => {
    const row = seat.seat_number.charAt(0) // e.g., "A" from "A5"
    if (!map.has(row)) {
      map.set(row, [])
    }
    map.get(row)!.push(seat)
  })

  // Sort seats within each row by number
  map.forEach((seats, _row) => {
    seats.sort((a, b) => {
      const numA = parseInt(a.seat_number.slice(1))
      const numB = parseInt(b.seat_number.slice(1))
      return numA - numB
    })
  })

  return map
})

// Split row seats into left (1-5) and right (6-10)
const getSeatGroups = (row: string) => {
  const seats = seatsByRow.value.get(row) || []
  const left = seats.filter(s => {
    const num = parseInt(s.seat_number.slice(1))
    return num <= 5
  })
  const right = seats.filter(s => {
    const num = parseInt(s.seat_number.slice(1))
    return num > 5
  })
  return { left, right }
}

// Seat interaction
const toggleSeat = async (seat: ShowtimeSeat) => {
  if (seat.status !== 'AVAILABLE' && !isSeatSelected(seat.id)) return

  const newSet = new Set(selectedSeatIds.value)
  if (newSet.has(seat.id)) {
    // 1. OPTIMISTIC UPDATE FIRST (เปลี่ยนสถานะทันทีไม่ต้องรอ API ตอบ)
    const prevStatus = seat.status
    newSet.delete(seat.id)
    selectedSeatIds.value = newSet
    movieStore.updateSeatStatus(seat.seat_number, 'AVAILABLE')

    try {
      // 2. Local deselect & API unlock
      await api.delete(`/bookings/lock/${showtimeId}/${seat.seat_number}`)
    } catch (error) {
      console.error('Failed to unlock seat:', error)
      // REVERT if failed
      newSet.add(seat.id)
      selectedSeatIds.value = newSet
      movieStore.updateSeatStatus(seat.seat_number, prevStatus as any)
      alert('มีปัญหาในการปลดล็อกที่นั่ง')
    }
  } else {
    try {
      // 1. Send lock request to API
      // Note: This will return 401 if user is not logged in
      await api.post('/bookings/lock', {
        showtime_id: showtimeId,
        seat_number: seat.seat_number
      })
      
      // 2. Lock success -> update state
      newSet.add(seat.id)
      selectedSeatIds.value = newSet
      
    } catch (error: any) {
      console.error('Failed to lock seat:', error)
      if (error.response?.status === 401) {
        alert('กรุณาเข้าสู่ระบบก่อนจองที่นั่งครับ (Please login first)')
      } else {
        alert('ที่นั่งนี้ถูกจองไปแล้วโดยผู้ใช้อื่น หรือมีปัญหาในการจอง')
        // Refresh seats to get actual state
        movieStore.fetchSeatsByShowtime(showtimeId)
      }
    }
  }
}

const isSeatSelected = (seatId: string) => selectedSeatIds.value.has(seatId)

// Get selected seat objects
const selectedSeats = computed(() => {
  return movieStore.seats.filter(s => selectedSeatIds.value.has(s.id))
    .sort((a, b) => {
      // Sort by row then number
      if (a.seat_number.charAt(0) !== b.seat_number.charAt(0)) {
        return a.seat_number.charAt(0).localeCompare(b.seat_number.charAt(0))
      }
      return parseInt(a.seat_number.slice(1)) - parseInt(b.seat_number.slice(1))
    })
})

// Total price
const totalPrice = computed(() => {
  return selectedSeats.value.reduce((sum, seat) => sum + seat.price, 0)
})

// Showtime info
const showtimeInfo = computed(() => {
  const st = movieStore.currentShowtime
  if (!st) return { date: '', time: '' }
  const d = new Date(st.start_time)
  return {
    date: d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' }),
    time: d.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit', hour12: false })
  }
})

const handleContinue = () => {
  // TODO: Implement payment/lock flow
  console.log('Selected seats:', selectedSeats.value.map(s => s.seat_number))
  console.log('Total price:', totalPrice.value)
}
</script>

<template>
  <div class="min-h-screen bg-background text-foreground flex flex-col">
    <Navbar />

    <main class="flex-1 py-6 px-4 md:px-6">
      <div class="max-w-7xl mx-auto">

        <!-- Back Button -->
        <button @click="router.back()"
          class="inline-flex items-center gap-2 px-4 py-2 rounded-full border border-border/50 bg-card/40 backdrop-blur-sm hover:bg-card hover:border-primary/50 transition-colors mb-6 text-sm font-medium cursor-pointer">
          <ArrowLeftIcon class="w-4 h-4" /> Back
        </button>

        <!-- Main Content: 2 Column Layout -->
        <div class="flex flex-col lg:flex-row gap-8">

          <!-- LEFT: Seat Map (70%) -->
          <div class="w-full lg:w-[70%]">

            <!-- Movie Info Header -->
            <div class="text-center mb-8 h-16">
              <div v-if="movieStore.isLoading || !movieStore.currentMovie">
                <h1 class="text-2xl md:text-3xl font-extrabold text-white mb-1 animate-pulse">
                  Loading...
                </h1>
                <div class="h-5 w-32 bg-muted/30 mx-auto rounded animate-pulse mt-2"></div>
              </div>
              <div v-else>
                <h1 class="text-2xl md:text-3xl font-extrabold text-white mb-1">
                  {{ movieStore.currentMovie?.title || 'Unknown Title' }}
                </h1>
                <p class="text-primary font-semibold text-sm md:text-base">
                  {{ showtimeInfo.date }} · {{ showtimeInfo.time }}
                </p>
              </div>
            </div>
            <!-- Seat Grid Area -->
            <div v-if="movieStore.isLoading || movieStore.seats.length === 0" class="flex flex-col items-center justify-center space-y-4 py-12">
              <div class="w-10 h-10 border-4 border-primary border-t-transparent rounded-full animate-spin"></div>
              <p class="text-muted-foreground font-medium">Loading seat map...</p>
            </div>
            <div v-else class="w-full overflow-x-auto hide-scrollbar pb-6" ref="seatMapContainer">
              
              <!-- Wrapper to force minimum width for mobile -->
              <div class="min-w-[600px] md:min-w-fit flex flex-col items-center mx-auto relative px-4">
                
                <!-- Screen Graphic (Aligned to grid width) -->
                <div class="relative mb-12 w-full max-w-[500px] mx-auto text-center">
                  <div class="w-full px-8 md:px-12">
                    <!-- Screen curve -->
                    <div class="h-3 bg-gradient-to-r from-transparent via-muted-foreground/50 to-transparent rounded-b-[100%]"></div>
                    <div class="h-1 bg-gradient-to-r from-transparent via-muted-foreground/20 to-transparent rounded-b-[100%] mt-1 mx-6"></div>
                  </div>
                  <div class="flex items-center justify-center gap-2 mt-3 text-muted-foreground text-xs font-medium">
                    <MonitorIcon class="w-4 h-4" />
                    <span>SCREEN</span>
                  </div>
                </div>

                <!-- Normal Rows (C-H, closest to screen) -->
                <div class="space-y-2 mb-8">
                  <div v-for="row in normalRows.slice().reverse()" :key="row" class="flex items-center justify-center gap-2 md:gap-3 w-max mx-auto">
                    <!-- Left Row Label -->
                    <div class="sticky left-0 z-10 bg-background flex items-center justify-center w-6 md:w-8 h-8 -ml-2 pl-2">
                      <span class="text-xs font-bold text-muted-foreground">{{ row }}</span>
                    </div>

                    <!-- Left Group (1-5) -->
                    <div class="flex gap-1.5 md:gap-2">
                      <button v-for="seat in getSeatGroups(row).left" :key="seat.id" @click="toggleSeat(seat)" :disabled="seat.status !== 'AVAILABLE' && seat.status !== 'SELECTED'"
                        :title="`${seat.seat_number} - ฿${seat.price}`"
                        class="w-8 h-8 md:w-9 md:h-9 rounded-t-lg text-xs font-bold transition-all duration-200 flex items-center justify-center cursor-pointer shrink-0"
                        :class="{
                          'bg-muted/60 border border-border/40 hover:bg-muted hover:border-primary/50 text-muted-foreground hover:text-white': seat.status === 'AVAILABLE' && !isSeatSelected(seat.id),
                          'bg-yellow-500 text-yellow-950 shadow-md shadow-yellow-500/30 scale-105 border-2 border-yellow-500': isSeatSelected(seat.id),
                          'bg-red-500 border border-red-600 text-white cursor-not-allowed': seat.status === 'LOCKED',
                          'bg-muted opacity-50 border border-border/30 text-muted-foreground cursor-not-allowed': seat.status === 'RESERVED' || seat.status === 'BOOKED',
                        }">
                        {{ seat.seat_number.slice(1) }}
                      </button>
                    </div>

                    <!-- Center Aisle -->
                    <div class="w-8 md:w-12 shrink-0"></div>

                    <!-- Right Group (6-10) -->
                    <div class="flex gap-1.5 md:gap-2">
                      <button v-for="seat in getSeatGroups(row).right" :key="seat.id" @click="toggleSeat(seat)" :disabled="seat.status !== 'AVAILABLE' && seat.status !== 'SELECTED'"
                        :title="`${seat.seat_number} - ฿${seat.price}`"
                        class="w-8 h-8 md:w-9 md:h-9 rounded-t-lg text-xs font-bold transition-all duration-200 flex items-center justify-center cursor-pointer shrink-0"
                        :class="{
                          'bg-muted/60 border border-border/40 hover:bg-muted hover:border-primary/50 text-muted-foreground hover:text-white': seat.status === 'AVAILABLE' && !isSeatSelected(seat.id),
                          'bg-yellow-500 text-yellow-950 shadow-md shadow-yellow-500/30 scale-105 border-2 border-yellow-500': isSeatSelected(seat.id),
                          'bg-red-500 border border-red-600 text-white cursor-not-allowed': seat.status === 'LOCKED',
                          'bg-muted opacity-50 border border-border/30 text-muted-foreground cursor-not-allowed': seat.status === 'RESERVED' || seat.status === 'BOOKED',
                        }">
                        {{ seat.seat_number.slice(1) }}
                      </button>
                    </div>

                    <!-- Right Row Label -->
                    <div class="sticky right-0 z-10 bg-background flex items-center justify-center w-6 md:w-8 h-8 -mr-2 pr-2">
                      <span class="text-xs font-bold text-muted-foreground">{{ row }}</span>
                    </div>
                  </div>
                </div>

                <!-- Zone Separator (Aligned to grid width) -->
                <div class="flex items-center gap-4 my-6 w-full max-w-[500px] mx-auto px-10">
                  <div class="flex-1 border-t border-dashed border-border/40"></div>
                  <span class="text-[10px] font-bold text-primary/80 tracking-widest uppercase whitespace-nowrap">Premium Zone</span>
                  <div class="flex-1 border-t border-dashed border-border/40"></div>
                </div>

                <!-- Premium Rows (A-B, farthest from screen) -->
                <div class="space-y-2">
                  <div v-for="row in premiumRows.slice().reverse()" :key="row" class="flex items-center justify-center gap-2 md:gap-3 w-max mx-auto">
                    <!-- Left Row Label -->
                    <div class="sticky left-0 z-10 bg-background flex items-center justify-center w-6 md:w-8 h-8 -ml-2 pl-2">
                      <span class="text-xs font-bold text-primary">{{ row }}</span>
                    </div>

                    <!-- Left Group (1-5) -->
                    <div class="flex gap-1.5 md:gap-2">
                      <button v-for="seat in getSeatGroups(row).left" :key="seat.id" @click="toggleSeat(seat)" :disabled="seat.status !== 'AVAILABLE' && seat.status !== 'SELECTED'"
                        :title="`${seat.seat_number} - ฿${seat.price} (Premium)`"
                        class="w-8 h-8 md:w-9 md:h-9 rounded-t-lg text-xs font-bold transition-all duration-200 flex items-center justify-center relative cursor-pointer shrink-0"
                        :class="{
                          'bg-yellow-900/20 border-2 border-yellow-500/50 hover:bg-yellow-900/30 hover:border-yellow-400 text-yellow-500/70 hover:text-yellow-400': seat.status === 'AVAILABLE' && !isSeatSelected(seat.id),
                          'bg-yellow-500 text-yellow-950 shadow-md shadow-yellow-500/30 scale-105 border-2 border-yellow-500': isSeatSelected(seat.id),
                          'bg-red-500 border border-red-600 text-white cursor-not-allowed': seat.status === 'LOCKED',
                          'bg-muted opacity-50 border border-border/30 text-muted-foreground cursor-not-allowed': seat.status === 'RESERVED' || seat.status === 'BOOKED',
                        }">
                        {{ seat.seat_number.slice(1) }}
                      </button>
                    </div>

                    <!-- Center Aisle -->
                    <div class="w-8 md:w-12 shrink-0"></div>

                    <!-- Right Group (6-10) -->
                    <div class="flex gap-1.5 md:gap-2">
                      <button v-for="seat in getSeatGroups(row).right" :key="seat.id" @click="toggleSeat(seat)" :disabled="seat.status !== 'AVAILABLE' && seat.status !== 'SELECTED'"
                        :title="`${seat.seat_number} - ฿${seat.price} (Premium)`"
                        class="w-8 h-8 md:w-9 md:h-9 rounded-t-lg text-xs font-bold transition-all duration-200 flex items-center justify-center relative cursor-pointer shrink-0"
                        :class="{
                          'bg-yellow-900/20 border-2 border-yellow-500/50 hover:bg-yellow-900/30 hover:border-yellow-400 text-yellow-500/70 hover:text-yellow-400': seat.status === 'AVAILABLE' && !isSeatSelected(seat.id),
                          'bg-yellow-500 text-yellow-950 shadow-md shadow-yellow-500/30 scale-105 border-2 border-yellow-500': isSeatSelected(seat.id),
                          'bg-red-500 border border-red-600 text-white cursor-not-allowed': seat.status === 'LOCKED',
                          'bg-muted opacity-50 border border-border/30 text-muted-foreground cursor-not-allowed': seat.status === 'RESERVED' || seat.status === 'BOOKED',
                        }">
                        {{ seat.seat_number.slice(1) }}
                      </button>
                    </div>

                    <!-- Right Row Label -->
                    <div class="sticky right-0 z-10 bg-background flex items-center justify-center w-6 md:w-8 h-8 -mr-2 pr-2">
                      <span class="text-xs font-bold text-primary">{{ row }}</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- Legend -->
            <div class="flex items-center justify-center gap-4 md:gap-6 mt-10 pb-4 flex-wrap">
              <div class="flex items-center gap-2">
                <div class="w-5 h-5 rounded-sm bg-muted/60 border border-border/40"></div>
                <span class="text-xs text-muted-foreground">Available</span>
              </div>
              <div class="flex items-center gap-2">
                <div class="w-5 h-5 rounded-sm bg-yellow-500 border-2 border-yellow-500"></div>
                <span class="text-xs text-muted-foreground">Selected</span>
              </div>
              <div class="flex items-center gap-2">
                <div class="w-5 h-5 rounded-sm bg-red-500 border border-red-600"></div>
                <span class="text-xs text-muted-foreground">Locked</span>
              </div>
              <div class="flex items-center gap-2">
                <div class="w-5 h-5 rounded-sm bg-muted opacity-50 border border-border/30"></div>
                <span class="text-xs text-muted-foreground">Booked</span>
              </div>
              <div class="flex items-center gap-2">
                <div class="w-5 h-5 rounded-sm bg-yellow-900/20 border-2 border-yellow-500/50"></div>
                <span class="text-xs text-muted-foreground">Premium</span>
              </div>
            </div>
          </div>

          <!-- RIGHT: Booking Summary (30%) -->
          <div class="w-full lg:w-[30%]">
            <div class="bg-card/60 border border-border/40 rounded-2xl p-6 backdrop-blur-md sticky top-24">
              <h2 class="text-xl font-bold text-white mb-1">Booking Summary</h2>
              <p class="text-sm text-muted-foreground mb-6">Select your seats to proceed.</p>

              <!-- Ticket Count & Price Overview -->
              <div class="bg-background/50 border border-border/30 rounded-xl p-4 mb-6">
                <div class="flex items-center justify-between">
                  <div>
                    <p class="text-[10px] font-bold text-muted-foreground uppercase tracking-wider">Tickets</p>
                    <p class="text-lg font-black text-white">{{ selectedSeats.length }} Selected</p>
                  </div>
                  <div class="text-right">
                    <p class="text-[10px] font-bold text-muted-foreground uppercase tracking-wider">Total Price</p>
                    <p class="text-lg font-black text-primary">฿{{ totalPrice.toLocaleString() }}</p>
                  </div>
                </div>
              </div>

              <!-- Selected Seats List -->
              <div v-if="selectedSeats.length > 0" class="mb-6">
                <p class="text-sm font-semibold text-muted-foreground mb-3">Selected Seats</p>
                <div class="flex flex-wrap gap-2">
                  <span v-for="seat in selectedSeats" :key="seat.id"
                    class="inline-flex items-center gap-1 px-3 py-1.5 rounded-lg text-xs font-bold cursor-pointer transition-all hover:scale-105"
                    :class="seat.type === 'Premium' ? 'bg-yellow-500/20 text-yellow-400 border border-yellow-500/30' : 'bg-primary/20 text-primary border border-primary/30'"
                    @click="toggleSeat(seat)">
                    {{ seat.seat_number }}
                    <span class="text-[10px] opacity-60">×</span>
                  </span>
                </div>
              </div>

              <!-- Price Breakdown -->
              <div v-if="selectedSeats.length > 0" class="space-y-2 mb-6 pt-4 border-t border-border/30">
                <div v-if="selectedSeats.filter(s => s.type === 'Normal').length > 0" class="flex justify-between text-sm">
                  <span class="text-muted-foreground">Normal × {{ selectedSeats.filter(s => s.type === 'Normal').length }}</span>
                  <span class="font-semibold text-white">฿{{ (selectedSeats.filter(s => s.type === 'Normal').length * 150).toLocaleString() }}</span>
                </div>
                <div v-if="selectedSeats.filter(s => s.type === 'Premium').length > 0" class="flex justify-between text-sm">
                  <span class="text-muted-foreground">Premium × {{ selectedSeats.filter(s => s.type === 'Premium').length }}</span>
                  <span class="font-semibold text-yellow-400">฿{{ (selectedSeats.filter(s => s.type === 'Premium').length * 300).toLocaleString() }}</span>
                </div>
                <div class="flex justify-between text-sm pt-2 border-t border-border/20">
                  <span class="font-bold text-white">Total</span>
                  <span class="font-black text-primary text-base">฿{{ totalPrice.toLocaleString() }}</span>
                </div>
              </div>

              <!-- Continue Button -->
              <Button @click="handleContinue" :disabled="selectedSeats.length === 0"
                class="w-full h-12 rounded-xl text-base font-bold transition-all duration-300 cursor-pointer"
                :class="selectedSeats.length > 0 ? 'bg-primary text-primary-foreground hover:bg-primary/90 shadow-lg shadow-primary/20' : 'bg-muted text-muted-foreground cursor-not-allowed'">
                Continue to Payment
              </Button>
            </div>
          </div>
        </div>
      </div>
    </main>

    <Footer />
  </div>
</template>
<style scoped>
/* Hide scrollbar for Chrome, Safari and Opera */
.hide-scrollbar::-webkit-scrollbar {
  display: none;
}

/* Hide scrollbar for IE, Edge and Firefox */
.hide-scrollbar {
  -ms-overflow-style: none;  /* IE and Edge */
  scrollbar-width: none;  /* Firefox */
}
</style>
