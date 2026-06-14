<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useMovieStore } from '@/stores/useMovieStore'
import { useAuthStore } from '@/stores/useAuthStore'
import { api } from '@/lib/axios'
import { Button } from '@/components/ui/button'
import Input from '@/components/ui/input/Input.vue'
import Navbar from '@/components/layout/Navbar.vue'
import Footer from '@/components/layout/Footer.vue'
import { ChevronLeftIcon, CreditCardIcon, SmartphoneIcon, LockIcon, Loader2Icon } from '@lucide/vue'
import { toast } from 'vue-sonner'

const route = useRoute()
const router = useRouter()
const movieStore = useMovieStore()
const authStore = useAuthStore()

const showtimeId = route.params.showtimeId as string

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
  
  // If no selected seats, redirect back to booking
  if (selectedSeats.value.length === 0) {
    router.push(`/booking/${showtimeId}`)
  }
})

// Data Computation
const movie = computed(() => movieStore.currentMovie)
const showtime = computed(() => movieStore.currentShowtime)
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

// Payment State
const isProcessing = ref(false)
const cardNumber = ref('4111 1111 1111 1111')
const expiryDate = ref('12/26')
const cvc = ref('123')
const cardholderName = ref('JOHN DOE')

const goBack = () => {
  router.push(`/booking-summary/${showtimeId}`)
}

const handlePayment = async () => {
  if (isProcessing.value) return
  isProcessing.value = true

  try {
    // Call ConfirmBooking API for all selected seats at once
    const seatsToBook = [...selectedSeats.value]
    const seatNumbers = seatsToBook.map(s => s.seat_number)
    
    await api.post('/bookings/confirm', {
      showtime_id: showtimeId,
      seat_numbers: seatNumbers
    })

    // Generate Order ID
    const orderId = '#RMC-' + Math.random().toString(36).substring(2, 8).toUpperCase()

    // Save booked seats to store for the confirmation page
    movieStore.lastBookedSeats = seatsToBook
    
    // Persist to sessionStorage to survive page refresh
    sessionStorage.setItem('lastBookedSeats', JSON.stringify(seatsToBook))
    sessionStorage.setItem('lastOrderId', orderId)
    
    // Clear 'SELECTED' status by changing them to 'BOOKED' locally immediately for UX
    seatsToBook.forEach(seat => {
      movieStore.updateSeatStatus(seat.seat_number, 'BOOKED')
    })

    // Wait a brief moment to simulate processing
    await new Promise(resolve => setTimeout(resolve, 800))

    // Redirect to confirmation page
    router.push({ name: 'booking-confirmation', params: { showtimeId: showtimeId } })

  } catch (error: any) {
    console.error("Payment failed:", error)
    if (error.response && error.response.status === 409) {
      toast.error("ที่นั่งบางส่วนถูกจองไปแล้ว กรุณาทำรายการใหม่ครับ")
      router.push(`/seat-map/${showtimeId}`)
    } else {
      toast.error("เกิดข้อผิดพลาดในการชำระเงิน กรุณาลองใหม่อีกครั้ง")
    }
    isProcessing.value = false
  }
}
</script>

<template>
  <div class="min-h-screen bg-background flex flex-col font-sans">
    <Navbar />

    <main class="flex-grow pt-24 pb-12 px-4 md:px-8">
      <div class="max-w-5xl mx-auto">
        <!-- Header -->
        <button @click="goBack" class="flex items-center gap-2 text-muted-foreground hover:text-white transition-colors mb-8 group w-max">
          <div class="w-8 h-8 rounded-full bg-muted/50 flex items-center justify-center group-hover:bg-primary/20 transition-colors">
            <ChevronLeftIcon class="w-4 h-4 group-hover:text-primary" />
          </div>
          <span class="text-sm font-semibold">Back to Summary</span>
        </button>

        <div class="flex items-center justify-between mb-8 pb-4 border-b border-border/40">
          <h1 class="text-3xl font-black text-white">Payment Method</h1>
          <div class="hidden sm:flex items-center gap-2 px-3 py-1.5 rounded-full bg-green-500/10 text-green-500 border border-green-500/20">
            <LockIcon class="w-4 h-4" />
            <span class="text-xs font-bold uppercase tracking-wide">Secure Checkout</span>
          </div>
        </div>

        <div class="flex flex-col lg:flex-row gap-8">
          <!-- LEFT COLUMN: Payment Form -->
          <div class="w-full lg:w-[60%]">
            
            <!-- Payment Methods Tabs -->
            <div class="grid grid-cols-3 gap-3 mb-8">
              <button class="flex flex-col items-center justify-center gap-2 p-4 rounded-xl border-2 border-primary bg-primary/10 text-primary transition-all">
                <CreditCardIcon class="w-6 h-6" />
                <span class="text-xs font-bold">Credit Card</span>
              </button>
              <button class="flex flex-col items-center justify-center gap-2 p-4 rounded-xl border-2 border-border/40 bg-muted/20 text-muted-foreground hover:border-border transition-all">
                <SmartphoneIcon class="w-6 h-6" />
                <span class="text-xs font-bold">Apple Pay</span>
              </button>
              <button class="flex flex-col items-center justify-center gap-2 p-4 rounded-xl border-2 border-border/40 bg-muted/20 text-muted-foreground hover:border-border transition-all">
                <SmartphoneIcon class="w-6 h-6" />
                <span class="text-xs font-bold">Google Pay</span>
              </button>
            </div>

            <!-- Card Form -->
            <div class="bg-card/60 border border-border/40 rounded-2xl p-6 backdrop-blur-md">
              <div class="space-y-6">
                <!-- Card Number -->
                <div class="space-y-2">
                  <label class="text-[10px] font-bold text-muted-foreground uppercase tracking-wider">Card Number</label>
                  <div class="relative">
                    <CreditCardIcon class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                    <Input v-model="cardNumber" placeholder="0000 0000 0000 0000" class="pl-10 font-mono tracking-widest bg-background/50" />
                  </div>
                </div>

                <div class="grid grid-cols-2 gap-4">
                  <!-- Expiry Date -->
                  <div class="space-y-2">
                    <label class="text-[10px] font-bold text-muted-foreground uppercase tracking-wider">Expiry Date</label>
                    <Input v-model="expiryDate" placeholder="MM/YY" class="bg-background/50" />
                  </div>
                  <!-- CVC -->
                  <div class="space-y-2">
                    <label class="flex items-center gap-1 text-[10px] font-bold text-muted-foreground uppercase tracking-wider">
                      CVC <LockIcon class="w-3 h-3 opacity-70" />
                    </label>
                    <Input v-model="cvc" type="password" placeholder="123" class="bg-background/50" />
                  </div>
                </div>

                <!-- Cardholder Name -->
                <div class="space-y-2">
                  <label class="text-[10px] font-bold text-muted-foreground uppercase tracking-wider">Cardholder Name</label>
                  <Input v-model="cardholderName" placeholder="Name on card" class="bg-background/50 uppercase" />
                </div>

                <!-- Pay Button -->
                <div class="pt-4">
                  <Button @click="handlePayment" :disabled="isProcessing" class="w-full h-14 rounded-xl text-base font-bold transition-all shadow-lg shadow-primary/20 bg-primary text-primary-foreground hover:bg-primary/90 cursor-pointer">
                    <template v-if="isProcessing">
                      <Loader2Icon class="w-5 h-5 animate-spin mr-2" />
                      Processing Payment...
                    </template>
                    <template v-else>
                      Pay ฿{{ totalPayable.toLocaleString(undefined, {minimumFractionDigits: 2}) }}
                    </template>
                  </Button>
                </div>
              </div>
            </div>

          </div>

          <!-- RIGHT COLUMN: Order Summary -->
          <div class="w-full lg:w-[40%]">
            <div class="bg-card/60 border border-border/40 rounded-2xl p-6 backdrop-blur-md sticky top-24">
              <h2 class="text-lg font-bold text-white mb-6">Order Summary</h2>

              <div class="flex gap-4 mb-6 pb-6 border-b border-border/40">
                <div class="w-20 aspect-[2/3] rounded-lg overflow-hidden shrink-0">
                  <img v-if="movie?.posterUrl" :src="movie.posterUrl" class="w-full h-full object-cover" />
                  <div v-else class="w-full h-full bg-muted"></div>
                </div>
                <div class="flex flex-col justify-center">
                  <h3 class="text-base font-bold text-white leading-snug mb-1">{{ movie?.title || 'Loading...' }}</h3>
                  <p class="text-xs text-primary font-medium mb-1">{{ formattedDate }}</p>
                  <p class="text-xs text-muted-foreground">{{ formattedTime }}</p>
                  <p class="text-xs text-muted-foreground mt-1">{{ selectedSeats.length }} Tickets</p>
                </div>
              </div>

              <div class="flex items-center justify-between">
                <span class="text-xl font-black text-white">Total</span>
                <span class="text-2xl font-black text-primary">฿{{ totalPayable.toLocaleString(undefined, {minimumFractionDigits: 2}) }}</span>
              </div>
            </div>
          </div>

        </div>
      </div>
    </main>

    <Footer />
  </div>
</template>
