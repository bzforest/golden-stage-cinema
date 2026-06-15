<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/useAuthStore'
import { api } from '@/lib/axios'
import Navbar from '@/components/layout/Navbar.vue'
import Footer from '@/components/layout/Footer.vue'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Card, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { UserIcon, LogOutIcon, TicketIcon, HistoryIcon, SettingsIcon, ShieldIcon } from '@lucide/vue'

const router = useRouter()
const authStore = useAuthStore()

const user = computed(() => authStore.user)
const isAdmin = computed(() => authStore.isAdmin)

const handleLogout = async () => {
  try {
    await authStore.logout()
    router.push('/login')
  } catch (error) {
    console.error(error)
  }
}

// Data fetching
const rawBookings = ref<any[]>([])
const isLoading = ref(true)

onMounted(async () => {
  try {
    const res = await api.get('/bookings/me')
    rawBookings.value = res.data || []
  } catch (err) {
    console.error("Failed to fetch bookings", err)
  } finally {
    isLoading.value = false
  }
})

// Group bookings by showtime_id to combine seats for the same showtime
const groupedBookings = computed(() => {
  const groups: Record<string, any> = {}
  rawBookings.value.forEach(b => {
    // skip invalid data
    if (!b.showtime || !b.movie) return

    const key = b.showtime_id
    if (!groups[key]) {
      groups[key] = {
        showtime: b.showtime,
        movie: b.movie,
        cinema: b.cinema,
        seats: [],
        booking_id: b._id || b.id, // using the first booking ID to represent the group
        created_at: new Date(b.created_at)
      }
    }
    groups[key].seats.push(b.seat_number)
  })
  return Object.values(groups).sort((a, b) => b.created_at.getTime() - a.created_at.getTime())
})

const upcomingTickets = computed(() => {
  const now = new Date()
  return groupedBookings.value.filter(g => new Date(g.showtime.start_time) >= now)
})

const pastHistory = computed(() => {
  const now = new Date()
  return groupedBookings.value.filter(g => new Date(g.showtime.start_time) < now)
})

// Formatting helpers
const formatDate = (dateString: string) => {
  const d = new Date(dateString)
  return d.toLocaleDateString('en-US', { weekday: 'long', month: 'long', day: 'numeric', year: 'numeric' })
}

const formatTime = (dateString: string) => {
  const d = new Date(dateString)
  return d.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' })
}
</script>

<template>
  <div class="min-h-screen bg-background flex flex-col font-sans">
    <Navbar />

    <main class="flex-1 py-12 px-6">
      <div class="max-w-4xl mx-auto space-y-12">
        
        <!-- Profile Header -->
        <div class="bg-card border border-border rounded-2xl p-6 md:p-8 flex items-center justify-between shadow-sm">
          <div class="flex items-center gap-6">
            <div class="w-20 h-20 md:w-24 md:h-24 rounded-full bg-secondary border-2 border-border flex items-center justify-center flex-shrink-0 text-muted-foreground overflow-hidden">
              <UserIcon class="w-10 h-10" />
            </div>
            <div>
              <h1 class="text-2xl md:text-3xl font-bold text-heading">
                {{ user?.displayName || 'Cinephile User' }}
              </h1>
              <p class="text-muted-foreground mt-1 text-sm">{{ user?.email }}</p>
              <div class="flex flex-wrap gap-2 mt-3">
                <Badge variant="outline" class="text-xs bg-primary/10 text-primary border-primary/20">
                  🍿 Movie Fan
                </Badge>
                <Badge variant="secondary" class="text-xs text-muted-foreground">
                  ⭐ 1,250 Points
                </Badge>
              </div>
            </div>
          </div>
          <Button variant="outline" @click="handleLogout" class="hidden md:flex gap-2 text-muted-foreground hover:text-red-500 hover:border-red-500 hover:bg-red-500/10 transition-colors">
            <LogOutIcon class="w-4 h-4" />
            Sign Out
          </Button>
        </div>

        <!-- Tabs Section -->
        <Tabs default-value="tickets" class="w-full">
          <TabsList class="w-full max-w-2xl bg-transparent border-b border-border/50 rounded-none h-14 p-0">
            <TabsTrigger value="tickets" class="cursor-pointer flex-1 gap-2 text-muted-foreground data-[state=active]:text-primary data-[state=active]:border-b-2 data-[state=active]:border-primary rounded-none h-full bg-transparent">
              <TicketIcon class="w-4 h-4" />
              My Tickets
            </TabsTrigger>
            <TabsTrigger value="history" class="cursor-pointer flex-1 gap-2 text-muted-foreground data-[state=active]:text-primary data-[state=active]:border-b-2 data-[state=active]:border-primary rounded-none h-full bg-transparent">
              <HistoryIcon class="w-4 h-4" />
              Watch History
            </TabsTrigger>
            <TabsTrigger value="settings" class="cursor-pointer flex-1 gap-2 text-muted-foreground data-[state=active]:text-primary data-[state=active]:border-b-2 data-[state=active]:border-primary rounded-none h-full bg-transparent">
              <SettingsIcon class="w-4 h-4" />
              Settings
            </TabsTrigger>
            <TabsTrigger v-if="isAdmin" value="admin" @click="router.push('/admin')" class="cursor-pointer flex-1 gap-2 text-blue-400 data-[state=active]:text-blue-500 data-[state=active]:border-b-2 data-[state=active]:border-blue-500 rounded-none h-full bg-transparent">
              <ShieldIcon class="w-4 h-4" />
              Admin
            </TabsTrigger>
          </TabsList>

          <!-- Tickets Tab -->
          <TabsContent value="tickets" class="mt-8">
            <h2 class="text-xl font-bold text-heading mb-6">Upcoming Screenings</h2>
            <div v-if="isLoading" class="text-muted-foreground">Loading tickets...</div>
            <div v-else-if="upcomingTickets.length === 0" class="text-muted-foreground text-sm border border-border/50 bg-card p-10 rounded-xl text-center">
              No upcoming tickets. Let's go watch a movie!
            </div>
            <div v-else class="space-y-6">
              <Card v-for="ticket in upcomingTickets" :key="ticket.booking_id" class="bg-card border-border max-w-3xl overflow-hidden">
                <div class="flex flex-col md:flex-row">
                  <div class="w-full md:w-48 aspect-[2/3] md:aspect-auto bg-gradient-to-br from-primary/20 to-secondary flex-shrink-0 flex items-center justify-center relative overflow-hidden">
                    <img v-if="ticket.movie?.poster_url" :src="ticket.movie.poster_url" class="absolute inset-0 w-full h-full object-cover" />
                    <svg v-else xmlns="http://www.w3.org/2000/svg" class="w-10 h-10 text-muted-foreground/30 relative z-10" viewBox="0 0 24 24" fill="currentColor"><path d="M18 3v2h-2V3H8v2H6V3H4v18h2v-2h2v2h8v-2h2v2h2V3h-2zM8 17H6v-2h2v2zm0-4H6v-2h2v2zm0-4H6V7h2v2zm10 8h-2v-2h2v2zm0-4h-2v-2h2v2zm0-4h-2V7h2v2z"/></svg>
                  </div>
                  <div class="flex-1 p-5 md:p-6 flex flex-col">
                    <div class="flex items-start justify-between">
                      <h3 class="text-xl font-bold text-heading">{{ ticket.movie?.title }}</h3>
                      <Badge variant="outline" class="text-xs">{{ ticket.movie?.format || 'Standard' }}</Badge>
                    </div>
                    <div class="mt-4 space-y-2 text-sm text-muted-foreground">
                      <div class="flex items-center gap-2">
                        <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect width="18" height="18" x="3" y="4" rx="2" ry="2"/><line x1="16" x2="16" y1="2" y2="6"/><line x1="8" x2="8" y1="2" y2="6"/><line x1="3" x2="21" y1="10" y2="10"/></svg>
                        {{ formatDate(ticket.showtime?.start_time) }}
                      </div>
                      <div class="flex items-center gap-2">
                        <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
                        {{ formatTime(ticket.showtime?.start_time) }} (Duration: {{ ticket.movie?.duration_mins }}m)
                      </div>
                      <div class="flex items-center gap-2">
                        <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M20 10c0 4.993-5.539 10.193-7.399 11.799a1 1 0 0 1-1.202 0C9.539 20.193 4 14.993 4 10a8 8 0 0 1 16 0"/><circle cx="12" cy="10" r="3"/></svg>
                        {{ ticket.cinema?.name || 'GoldenStage Cinema' }}
                      </div>
                    </div>
                    <div class="mt-auto pt-6">
                      <Separator class="mb-4" />
                      <div class="flex items-center justify-between text-sm">
                        <div>
                          <span class="text-muted-foreground">Seats</span>
                          <p class="font-bold text-primary">{{ ticket.seats.join(', ') }}</p>
                        </div>
                        <div class="text-right">
                          <span class="text-muted-foreground">Booking ID</span>
                          <p class="font-bold font-mono">#{{ String(ticket.booking_id).substring(0, 8).toUpperCase() }}</p>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </Card>
            </div>
          </TabsContent>

          <!-- History Tab -->
          <TabsContent value="history" class="mt-8">
            <h2 class="text-xl font-bold text-heading mb-6">Past Movies</h2>
            <div v-if="isLoading" class="text-muted-foreground">Loading history...</div>
            <div v-else-if="pastHistory.length === 0" class="text-muted-foreground text-sm border border-border/50 bg-card p-10 rounded-xl text-center">
              No watch history found.
            </div>
            <div v-else class="grid grid-cols-1 md:grid-cols-2 gap-4 max-w-4xl">
              <Card v-for="ticket in pastHistory" :key="ticket.booking_id" class="bg-card border-border overflow-hidden hover:border-primary/30 transition-all">
                <div class="flex">
                  <div class="w-24 bg-gradient-to-br from-primary/10 to-secondary flex-shrink-0 flex items-center justify-center relative overflow-hidden">
                    <img v-if="ticket.movie?.poster_url" :src="ticket.movie.poster_url" class="absolute inset-0 w-full h-full object-cover" />
                    <svg v-else xmlns="http://www.w3.org/2000/svg" class="w-8 h-8 text-muted-foreground/30 relative z-10" viewBox="0 0 24 24" fill="currentColor"><path d="M18 3v2h-2V3H8v2H6V3H4v18h2v-2h2v2h8v-2h2v2h2V3h-2zM8 17H6v-2h2v2zm0-4H6v-2h2v2zm0-4H6V7h2v2zm10 8h-2v-2h2v2zm0-4h-2v-2h2v2zm0-4h-2V7h2v2z"/></svg>
                  </div>
                  <CardContent class="p-4 flex-1">
                    <h3 class="font-bold text-heading text-base line-clamp-1">{{ ticket.movie?.title }}</h3>
                    <p class="text-xs text-muted-foreground mt-1">{{ formatDate(ticket.showtime?.start_time) }} · {{ formatTime(ticket.showtime?.start_time) }}</p>
                    <p class="text-xs text-muted-foreground mt-1">Seats: {{ ticket.seats.join(', ') }}</p>
                  </CardContent>
                </div>
              </Card>
            </div>
          </TabsContent>

          <!-- Settings Tab -->
          <TabsContent value="settings" class="mt-8">
            <Card class="bg-card border-border max-w-2xl">
              <CardContent class="p-6 md:p-8">
                <h2 class="text-xl font-bold text-heading mb-6">Account Settings</h2>
                <div class="space-y-6">
                  <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                    <div class="space-y-2">
                      <label class="text-sm font-medium text-muted-foreground">Display Name</label>
                      <Input :value="user?.displayName" placeholder="Enter your name" class="bg-background border-border" readonly />
                    </div>
                  </div>
                  <div class="space-y-2">
                    <label class="text-sm font-medium text-muted-foreground">Email Address</label>
                    <div class="relative">
                      <Input :value="user?.email" class="bg-background border-border" readonly />
                    </div>
                  </div>
                  <div class="space-y-2">
                    <label class="text-sm font-medium text-muted-foreground">Member ID</label>
                    <Input :value="user?.uid" class="bg-background border-border font-mono text-xs" readonly />
                  </div>
                  <div class="pt-4 border-t border-border mt-6">
                    <p class="text-xs text-muted-foreground mb-4">Contact support to update your email or member verification.</p>
                  </div>
                </div>
              </CardContent>
            </Card>
            <div class="mt-6 md:hidden">
              <Button variant="destructive" class="w-full" @click="handleLogout">
                Sign Out
              </Button>
            </div>
          </TabsContent>
        </Tabs>
      </div>
    </main>

    <Footer />
  </div>
</template>
