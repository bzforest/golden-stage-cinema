<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useMovieStore } from '@/stores/useMovieStore'
import { useAuthStore } from '@/stores/useAuthStore'
import { Button } from '@/components/ui/button'
import Input from '@/components/ui/input/Input.vue'
import { ChevronLeftIcon, TicketIcon, CalendarIcon, ClockIcon, MailIcon, CreditCardIcon } from '@lucide/vue'
import Navbar from '@/components/layout/Navbar.vue'
import Footer from '@/components/layout/Footer.vue'

const route = useRoute()
const router = useRouter()
const movieStore = useMovieStore()
const authStore = useAuthStore()

const showtimeId = route.params.showtimeId as string

// User Info for delivery
const deliveryEmail = ref(authStore.user?.email || '')

// Fetch initial data if accessed directly
onMounted(async () => {
  if (movieStore.seats.length === 0) {
    await movieStore.fetchSeatsByShowtime(showtimeId)
  }
  if (!movieStore.currentShowtime) {
    await movieStore.fetchShowtimeById(showtimeId)
  }
  if (movieStore.currentShowtime && !movieStore.currentMovie) {
    await movieStore.fetchMovieById(movieStore.currentShowtime.movie_id)
  }
  if (movieStore.cinemas.length === 0) {
    await movieStore.fetchCinemas()
  }
})

// Data Computation
const movie = computed(() => movieStore.currentMovie)
const showtime = computed(() => movieStore.currentShowtime)
const cinema = computed(() => {
  if (!showtime.value) return null
  return movieStore.cinemas.find(c => c.id === showtime.value?.cinema_id)
})
const selectedSeats = computed(() => movieStore.seats.filter(s => s.status === 'SELECTED'))

// Date Formatting
const formattedDate = computed(() => {
  if (!showtime.value?.start_time) return ''
  const date = new Date(showtime.value.start_time)
  return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
})

const formattedTime = computed(() => {
  if (!showtime.value?.start_time) return ''
  const date = new Date(showtime.value.start_time)
  return date.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit', hour12: false })
})

// Pricing Calculations
const normalSeats = computed(() => selectedSeats.value.filter(s => s.type !== 'Premium'))
const premiumSeats = computed(() => selectedSeats.value.filter(s => s.type === 'Premium'))

const normalTotal = computed(() => normalSeats.value.reduce((sum, s) => sum + (s.price || 150), 0))
const premiumTotal = computed(() => premiumSeats.value.reduce((sum, s) => sum + (s.price || 300), 0))

const subTotal = computed(() => normalTotal.value + premiumTotal.value)
const taxes = computed(() => subTotal.value * 0.07)
const bookingFee = 20
const totalPayable = computed(() => subTotal.value + taxes.value + bookingFee)

const goBack = () => {
  router.push(`/booking/${showtimeId}`)
}

const proceedToPayment = () => {
  router.push({ name: 'payment-method', params: { showtimeId: showtimeId } })
}
</script>

<template>
  <div class="min-h-screen bg-background flex flex-col font-sans">
    <Navbar />

    <main class="flex-grow pt-10 pb-12 px-4 md:px-8">
      <div class="max-w-6xl mx-auto">
        <!-- Header -->
        <button @click="goBack" class="flex items-center gap-2 text-muted-foreground hover:text-white transition-colors mb-6 group w-max">
          <div class="w-8 h-8 rounded-full bg-muted/50 flex items-center justify-center group-hover:bg-primary/20 transition-colors">
            <ChevronLeftIcon class="w-4 h-4 group-hover:text-primary" />
          </div>
          <span class="text-sm font-semibold">Back to Seats</span>
        </button>

        <h1 class="text-3xl font-black text-white mb-8 border-b border-border/40 pb-4">Booking Summary</h1>

        <div class="flex flex-col lg:flex-row gap-6 lg:gap-8">
          <!-- LEFT COLUMN -->
          <div class="w-full lg:w-[65%] space-y-6">
            
            <!-- Movie & Ticket Info Card -->
            <div class="bg-card/60 border border-border/40 rounded-2xl p-4 md:p-6 backdrop-blur-md">
              <div class="flex flex-col sm:flex-row gap-6">
                <!-- Poster -->
                <div class="shrink-0 mx-auto sm:mx-0 w-32 sm:w-40 aspect-[2/3] rounded-xl overflow-hidden shadow-lg border border-border/20">
                  <img v-if="movie?.posterUrl" :src="movie.posterUrl" :alt="movie.title" class="w-full h-full object-cover" />
                  <div v-else class="w-full h-full bg-muted animate-pulse"></div>
                </div>

                <!-- Info -->
                <div class="flex-grow flex flex-col justify-center">
                  <div class="flex items-center gap-2 mb-2">
                    <TicketIcon class="w-4 h-4 text-primary" />
                    <span class="text-xs font-bold text-primary tracking-wider uppercase">{{ cinema?.name || 'Golden Stage Cinema' }}</span>
                  </div>
                  
                  <h2 class="text-2xl sm:text-3xl font-black text-white leading-tight mb-6">
                    {{ movie?.title || 'Loading Movie...' }}
                  </h2>

                  <div class="grid grid-cols-2 gap-4">
                    <!-- Date & Time -->
                    <div>
                      <p class="text-[10px] font-bold text-muted-foreground uppercase tracking-wider mb-2">Date & Time</p>
                      <div class="flex items-center gap-2 text-sm text-white mb-1">
                        <CalendarIcon class="w-4 h-4 text-muted-foreground" />
                        <span>{{ formattedDate }}</span>
                      </div>
                      <div class="flex items-center gap-2 text-sm text-white">
                        <ClockIcon class="w-4 h-4 text-muted-foreground" />
                        <span>{{ formattedTime }}</span>
                      </div>
                    </div>

                    <!-- Seats -->
                    <div>
                      <p class="text-[10px] font-bold text-muted-foreground uppercase tracking-wider mb-2">Tickets</p>
                      <div class="flex items-center gap-2 text-sm text-white mb-2">
                        <TicketIcon class="w-4 h-4 text-muted-foreground" />
                        <span>{{ selectedSeats.length }} Seats</span>
                      </div>
                      <div class="flex flex-wrap gap-1.5">
                        <span v-for="seat in selectedSeats" :key="seat.id" 
                          class="px-2 py-0.5 rounded text-[11px] font-bold"
                          :class="seat.type === 'Premium' ? 'bg-yellow-500/20 text-yellow-400 border border-yellow-500/30' : 'bg-muted text-muted-foreground border border-border/50'">
                          {{ seat.seat_number }}
                        </span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- e-Ticket Delivery Card -->
            <div class="bg-card/60 border border-border/40 rounded-2xl p-6 backdrop-blur-md">
              <div class="flex items-center gap-2 mb-3">
                <TicketIcon class="w-5 h-5 text-primary" />
                <h3 class="text-lg font-bold text-white">e-Ticket Delivery</h3>
              </div>
              <p class="text-sm text-muted-foreground mb-4">
                Your tickets will be sent to your registered email address and phone number upon successful payment.
              </p>
              
              <div class="max-w-md">
                <div class="relative">
                  <MailIcon class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                  <Input v-model="deliveryEmail" type="email" placeholder="Enter your email" class="pl-9 bg-background/50 border-border/50" />
                </div>
              </div>
            </div>

          </div>

          <!-- RIGHT COLUMN -->
          <div class="w-full lg:w-[35%]">
            <div class="bg-card/60 border border-border/40 rounded-2xl p-6 backdrop-blur-md sticky top-24">
              <h2 class="text-xl font-bold text-white mb-6">Payment Summary</h2>

              <div class="space-y-4 mb-6">
                <!-- Normal Tickets -->
                <div v-if="normalSeats.length > 0" class="flex justify-between items-center text-sm">
                  <div class="flex items-center gap-2 text-muted-foreground">
                    <div class="w-1.5 h-1.5 rounded-full bg-muted-foreground/50"></div>
                    <span>Normal ({{ normalSeats.length }} × ฿{{ (normalSeats[0]?.price || 150).toLocaleString() }})</span>
                  </div>
                  <span class="text-white font-medium">฿{{ normalTotal.toLocaleString(undefined, {minimumFractionDigits: 2}) }}</span>
                </div>

                <!-- Premium Tickets -->
                <div v-if="premiumSeats.length > 0" class="flex justify-between items-center text-sm">
                  <div class="flex items-center gap-2 text-muted-foreground">
                    <div class="w-1.5 h-1.5 rounded-full bg-yellow-500/50"></div>
                    <span>Premium ({{ premiumSeats.length }} × ฿{{ (premiumSeats[0]?.price || 300).toLocaleString() }})</span>
                  </div>
                  <span class="text-white font-medium">฿{{ premiumTotal.toLocaleString(undefined, {minimumFractionDigits: 2}) }}</span>
                </div>

                <!-- Taxes -->
                <div class="flex justify-between items-center text-sm">
                  <div class="flex items-center gap-2 text-muted-foreground">
                    <div class="w-1.5 h-1.5 rounded-full bg-muted-foreground/50"></div>
                    <span>Taxes (7%)</span>
                  </div>
                  <span class="text-white font-medium">฿{{ taxes.toLocaleString(undefined, {minimumFractionDigits: 2, maximumFractionDigits: 2}) }}</span>
                </div>

                <!-- Booking Fee -->
                <div class="flex justify-between items-center text-sm">
                  <div class="flex items-center gap-2 text-muted-foreground">
                    <div class="w-1.5 h-1.5 rounded-full bg-muted-foreground/50"></div>
                    <span>Booking Fee</span>
                  </div>
                  <span class="text-white font-medium">฿{{ bookingFee.toLocaleString(undefined, {minimumFractionDigits: 2}) }}</span>
                </div>
              </div>

              <!-- Total Payable -->
              <div class="border-t border-border/40 pt-4 mb-6 flex justify-between items-end">
                <span class="text-xs font-bold text-muted-foreground uppercase tracking-wider mb-1">Total Payable</span>
                <span class="text-3xl font-black text-white">฿{{ totalPayable.toLocaleString(undefined, {minimumFractionDigits: 2, maximumFractionDigits: 2}) }}</span>
              </div>

              <!-- Proceed Button -->
              <Button @click="proceedToPayment" class="w-full h-14 rounded-xl text-base font-bold bg-primary text-primary-foreground hover:bg-primary/90 shadow-lg shadow-primary/20 mb-4 flex items-center justify-center gap-2 cursor-pointer">
                Proceed to Payment
                <CreditCardIcon class="w-5 h-5" />
              </Button>

              <p class="text-[10px] text-center text-muted-foreground/70 leading-relaxed px-4">
                By proceeding, you agree to our Terms & Conditions.
              </p>
            </div>
          </div>
        </div>

      </div>
    </main>

    <Footer />
  </div>
</template>
