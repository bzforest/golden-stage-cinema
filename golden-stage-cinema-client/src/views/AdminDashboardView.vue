<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '@/lib/axios'
import Navbar from '@/components/layout/Navbar.vue'
import Footer from '@/components/layout/Footer.vue'
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs'
import { Input } from '@/components/ui/input'
import CustomDatepicker from '@/components/ui/CustomDatepicker.vue'
import { Skeleton } from '@/components/ui/skeleton'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from '@/components/ui/dialog'
import { SearchIcon, ChevronLeftIcon, ChevronRightIcon, MenuIcon } from '@lucide/vue'

const activeTab = ref('bookings')

// Bookings State
const bookings = ref<any[]>([])
const bookingsTotal = ref(0)
const bookingsPage = ref(1)
const bookingsLimit = ref(10)
const searchQuery = ref('')
const searchDate = ref('')
const isLoadingBookings = ref(false)

// Logs State
const logs = ref<any[]>([])
const logsTotal = ref(0)
const logsPage = ref(1)
const logsLimit = ref(10)
const logsSearchQuery = ref('')
const isLoadingLogs = ref(false)

// Modal State
const selectedLog = ref<any>(null)
const isLogModalOpen = ref(false)

const openLogModal = (log: any) => {
  selectedLog.value = log
  isLogModalOpen.value = true
}

// Fetch Functions
const fetchBookings = async () => {
  try {
    isLoadingBookings.value = true
    
    const params = {
      search: searchQuery.value.trim(),
      date: searchDate.value,
      page: bookingsPage.value,
      limit: bookingsLimit.value
    }

    const res = await api.get(`/admin/bookings`, { params })
    bookings.value = res.data?.data || []
    bookingsTotal.value = res.data?.total || 0
  } catch (err) {
    console.error('Failed to fetch admin bookings', err)
  } finally {
    isLoadingBookings.value = false
  }
}

const fetchLogs = async () => {
  isLoadingLogs.value = true
  try {
    const res = await api.get(`/admin/logs`, {
      params: {
        search: logsSearchQuery.value,
        page: logsPage.value,
        limit: logsLimit.value
      }
    })
    logs.value = res.data?.data || []
    logsTotal.value = res.data?.total || 0
  } catch (err) {
    console.error('Failed to fetch admin logs', err)
  } finally {
    isLoadingLogs.value = false
  }
}

const onSearchClick = () => {
  bookingsPage.value = 1
  fetchBookings()
}

const onLogsSearchClick = () => {
  logsPage.value = 1
  fetchLogs()
}

// Pagination Handlers
const changeBookingsPage = (delta: number) => {
  bookingsPage.value += delta
  fetchBookings()
}

const changeLogsPage = (delta: number) => {
  logsPage.value += delta
  fetchLogs()
}

// Initial Fetch
onMounted(() => {
  fetchBookings()
  fetchLogs()
})

const formatDate = (dateStr: string) => {
  if (!dateStr) return ''
  return new Intl.DateTimeFormat('en-US', { 
    dateStyle: 'medium', 
    timeStyle: 'short' 
  }).format(new Date(dateStr))
}

const isJson = (str: string) => {
  if (!str) return false
  try {
    const parsed = JSON.parse(str)
    return typeof parsed === 'object' && parsed !== null
  } catch (e) {
    return false
  }
}

const parseJson = (str: string) => {
  try {
    return JSON.parse(str)
  } catch (e) {
    return {}
  }
}

const formatLogText = (text: string) => {
  // Highlight UIDs or specific keywords if needed, for now just return text
  return text
}
</script>

<template>
  <div class="min-h-screen bg-background flex flex-col font-sans text-foreground">
    <Navbar />

    <main class="flex-1 py-12 px-6">
      <div class="max-w-7xl mx-auto space-y-8">
        <h1 class="text-3xl font-bold text-heading">Admin Dashboard</h1>

        <Tabs v-model="activeTab" class="w-full">
          <TabsList class="mb-8 bg-gray-500/50 border border-border p-1">
            <TabsTrigger value="bookings" class="px-8 py-2 cursor-pointer">All Bookings</TabsTrigger>
            <TabsTrigger value="logs" class="px-8 py-2 cursor-pointer">Audit Logs</TabsTrigger>
          </TabsList>

          <TabsContent value="bookings" class="space-y-6">
            <!-- Filter Bar -->
            <div class="flex items-center gap-4 bg-card border border-border rounded-xl p-4">
              <div class="relative flex-1 max-w-md">
                <SearchIcon class="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-muted-foreground" />
                <Input 
                  v-model="searchQuery" 
                  @keyup.enter="onSearchClick"
                  placeholder="Search by ID, Email, Name, Movie..." 
                  class="pl-10 w-full"
                />
              </div>
              <CustomDatepicker 
                v-model="searchDate" 
                class="w-40" 
              />
              <Button @click="onSearchClick" variant="default" class="cursor-pointer">Search</Button>
            </div>

            <!-- Bookings Table -->
            <div class="bg-card border border-border rounded-xl overflow-hidden">
              <div class="overflow-x-auto">
                <table class="w-full text-left border-collapse">
                  <thead>
                    <tr class="bg-secondary/50 text-muted-foreground text-sm uppercase tracking-wider">
                      <th class="px-6 py-4 font-medium border-b border-border">Booking ID</th>
                      <th class="px-6 py-4 font-medium border-b border-border">User Info</th>
                      <th class="px-6 py-4 font-medium border-b border-border">Movie</th>
                      <th class="px-6 py-4 font-medium border-b border-border">Seat</th>
                      <th class="px-6 py-4 font-medium border-b border-border">Status</th>
                      <th class="px-6 py-4 font-medium border-b border-border">Date</th>
                    </tr>
                  </thead>
                  <tbody class="divide-y divide-border">
                    <template v-if="isLoadingBookings">
                      <tr v-for="i in 5" :key="i" class="hover:bg-secondary/20 transition-colors">
                        <td class="px-6 py-4"><Skeleton class="h-4 w-20" /></td>
                        <td class="px-6 py-4">
                          <Skeleton class="h-4 w-32 mb-2" />
                          <Skeleton class="h-3 w-40" />
                        </td>
                        <td class="px-6 py-4">
                          <Skeleton class="h-4 w-48 mb-2" />
                          <Skeleton class="h-3 w-32" />
                        </td>
                        <td class="px-6 py-4"><Skeleton class="h-4 w-8" /></td>
                        <td class="px-6 py-4"><Skeleton class="h-6 w-24 rounded-full" /></td>
                        <td class="px-6 py-4"><Skeleton class="h-4 w-24" /></td>
                      </tr>
                    </template>
                    <tr v-else-if="bookings.length === 0">
                      <td colspan="6" class="px-6 py-12 text-center text-muted-foreground">No bookings found.</td>
                    </tr>
                    <tr v-for="b in bookings" :key="b._id || b.id" class="hover:bg-secondary/20 transition-colors">
                      <td class="px-6 py-4 text-sm font-mono text-muted-foreground">#{{ (b._id || b.id)?.substring(0, 8).toUpperCase() || '-' }}</td>
                      <td class="px-6 py-4">
                        <div class="font-medium text-heading">{{ b.user_name }}</div>
                        <div class="text-xs text-muted-foreground">{{ b.user_email }}</div>
                      </td>
                      <td class="px-6 py-4">
                        <div class="font-medium text-heading">{{ b.movie?.title }}</div>
                        <div class="text-xs text-muted-foreground">{{ formatDate(b.showtime?.start_time) }}</div>
                      </td>
                      <td class="px-6 py-4 font-medium text-primary">{{ b.seat_number }}</td>
                      <td class="px-6 py-4">
                        <span class="px-2 py-1 rounded-full text-xs font-medium bg-green-500/10 text-green-500">
                          {{ b.status }}
                        </span>
                      </td>
                      <td class="px-6 py-4 text-sm text-muted-foreground">{{ formatDate(b.created_at) }}</td>
                    </tr>
                  </tbody>
                </table>
              </div>
              
              <!-- Pagination -->
              <div class="flex items-center justify-between px-6 py-4 border-t border-border bg-secondary/20">
                <div class="text-sm text-muted-foreground">
                  Showing {{ bookingsTotal > 0 ? (bookingsPage - 1) * bookingsLimit + 1 : 0 }} to 
                  {{ Math.min(bookingsPage * bookingsLimit, bookingsTotal) }} of {{ bookingsTotal }}
                </div>
                <div class="flex gap-2">
                  <Button class="cursor-pointer" variant="outline" size="sm" :disabled="bookingsPage === 1 || isLoadingBookings" @click="changeBookingsPage(-1)">
                    <ChevronLeftIcon class="w-4 h-4 mr-1" /> Prev
                  </Button>
                  <Button class="cursor-pointer" variant="outline" size="sm" :disabled="bookingsPage * bookingsLimit >= bookingsTotal || isLoadingBookings" @click="changeBookingsPage(1)">
                    Next <ChevronRightIcon class="w-4 h-4 ml-1" />
                  </Button>
                </div>
              </div>
            </div>
          </TabsContent>

          <TabsContent value="logs" class="space-y-6">
            <!-- Filter Bar for Logs -->
            <div class="flex items-center gap-4 bg-card border border-border rounded-xl p-4">
              <div class="relative flex-1 max-w-md">
                <SearchIcon class="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-muted-foreground" />
                <Input 
                  v-model="logsSearchQuery" 
                  @keyup.enter="onLogsSearchClick"
                  placeholder="Search logs by Event Type or Details..." 
                  class="pl-10 w-full"
                />
              </div>
              <Button @click="onLogsSearchClick" variant="default" class="cursor-pointer">Search Logs</Button>
            </div>

            <!-- Logs Table -->
            <div class="bg-card border border-border rounded-xl overflow-hidden">
              <div class="overflow-x-auto">
                <table class="w-full text-left border-collapse">
                  <thead>
                    <tr class="bg-secondary/50 text-muted-foreground text-sm uppercase tracking-wider">
                      <th class="px-6 py-4 font-medium border-b border-border w-48">Timestamp</th>
                      <th class="px-6 py-4 font-medium border-b border-border w-48">Event Type</th>
                      <th class="px-6 py-4 font-medium border-b border-border">Details</th>
                      <th class="px-6 py-4 font-medium border-b border-border w-16"></th>
                    </tr>
                  </thead>
                  <tbody class="divide-y divide-border">
                    <tr v-if="isLoadingLogs">
                      <td colspan="3" class="px-6 py-12 text-center text-muted-foreground">Loading logs...</td>
                    </tr>
                    <tr v-else-if="logs.length === 0">
                      <td colspan="3" class="px-6 py-12 text-center text-muted-foreground">No logs found.</td>
                    </tr>
                    <tr v-for="log in logs" :key="log.id" class="hover:bg-secondary/20 transition-colors">
                      <td class="px-6 py-4 text-sm text-muted-foreground">{{ formatDate(log.timestamp) }}</td>
                      <td class="px-6 py-4">
                        <span class="px-2 py-1 rounded-md text-xs font-bold"
                              :class="{
                                'bg-blue-500/10 text-blue-500': log.action === 'SEAT_LOCKED',
                                'bg-green-500/10 text-green-500': log.action === 'BOOKING_CONFIRMED',
                                'bg-primary/10 text-primary': !['SEAT_LOCKED', 'BOOKING_CONFIRMED'].includes(log.action)
                              }">
                          {{ log.action }}
                        </span>
                      </td>
                      <td class="px-6 py-4 text-sm">
                        <!-- Enriched Display for Non-JSON -->
                        <div v-if="!log.raw_json && (log.user_name || log.seat_number)" class="flex flex-wrap items-center gap-2">
                          <span v-if="log.user_name" class="px-2 py-1 bg-primary/20 text-primary rounded-md text-xs font-medium border border-primary/20">
                            User: {{ log.user_name }}
                          </span>
                          <span v-if="log.seat_number" class="px-2 py-1 bg-secondary/50 text-foreground rounded-md text-xs font-medium border border-border">
                            Seat: {{ log.seat_number }}
                          </span>
                          <span v-if="log.showtime_date" class="px-2 py-1 bg-secondary/50 text-foreground rounded-md text-xs font-medium border border-border">
                            Showtime: {{ formatDate(log.showtime_date) }}
                          </span>
                        </div>
                        <!-- Enriched Display for JSON (Timeout) -->
                        <div v-else-if="log.raw_json && log.seat_number" class="flex flex-wrap items-center gap-2">
                          <span class="px-2 py-1 bg-red-500/20 text-red-500 rounded-md text-xs font-medium border border-red-500/20">
                            Reason: Lock Timeout
                          </span>
                          <span class="px-2 py-1 bg-secondary/50 text-foreground rounded-md text-xs font-medium border border-border">
                            Seat: {{ log.seat_number }}
                          </span>
                        </div>
                        <!-- Fallback to original layout if somehow enrichment failed -->
                        <div v-else-if="isJson(log.details)" class="space-y-1.5 p-3 bg-secondary/10 rounded-lg border border-border/50">
                          <div v-for="(value, key) in parseJson(log.details)" :key="key" class="flex flex-wrap gap-x-2 gap-y-1 items-start text-xs">
                            <span class="font-medium text-muted-foreground uppercase tracking-wider">{{ String(key).replace(/_/g, ' ') }}:</span>
                            <span class="font-mono text-primary break-all">{{ value }}</span>
                          </div>
                        </div>
                        <div v-else class="text-muted-foreground leading-relaxed">
                          {{ formatLogText(log.details) }}
                        </div>
                      </td>
                      <td class="px-6 py-4 text-right">
                        <Button variant="ghost" size="icon" class="h-8 w-8 text-muted-foreground hover:text-foreground cursor-pointer" @click="openLogModal(log)">
                          <MenuIcon class="w-4 h-4" />
                        </Button>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
              
              <!-- Pagination -->
              <div class="flex items-center justify-between px-6 py-4 border-t border-border bg-secondary/20">
                <div class="text-sm text-muted-foreground">
                  Showing {{ logsTotal > 0 ? (logsPage - 1) * logsLimit + 1 : 0 }} to 
                  {{ Math.min(logsPage * logsLimit, logsTotal) }} of {{ logsTotal }}
                </div>
                <div class="flex gap-2">
                  <Button variant="outline" size="sm" :disabled="logsPage === 1 || isLoadingLogs" @click="changeLogsPage(-1)">
                    <ChevronLeftIcon class="w-4 h-4 mr-1" /> Prev
                  </Button>
                  <Button variant="outline" size="sm" :disabled="logsPage * logsLimit >= logsTotal || isLoadingLogs" @click="changeLogsPage(1)">
                    Next <ChevronRightIcon class="w-4 h-4 ml-1" />
                  </Button>
                </div>
              </div>
            </div>
          </TabsContent>
        </Tabs>
      </div>
    </main>

    <Footer />

    <!-- Audit Log Details Modal -->
    <Dialog :open="isLogModalOpen" @update:open="isLogModalOpen = $event">
      <DialogContent class="max-w-2xl bg-card border-border text-foreground">
        <DialogHeader>
          <DialogTitle class="text-xl">Audit Log Details</DialogTitle>
          <DialogDescription>
            Comprehensive event information captured by the system.
          </DialogDescription>
        </DialogHeader>
        
        <div v-if="selectedLog" class="grid grid-cols-1 md:grid-cols-2 gap-6 mt-4">
          <div class="space-y-4">
            <div>
              <div class="text-xs font-semibold text-muted-foreground uppercase mb-1">User Information</div>
              <div v-if="selectedLog.uid">
                <div class="font-medium">{{ selectedLog.user_name || 'N/A' }}</div>
                <div class="text-sm text-muted-foreground">{{ selectedLog.user_email || selectedLog.uid }}</div>
              </div>
              <div v-else class="text-sm text-muted-foreground italic">System / No User</div>
            </div>
            <div>
              <div class="text-xs font-semibold text-muted-foreground uppercase mb-1">Showtime & Movie</div>
              <div v-if="selectedLog.showtime_id">
                <div class="font-medium text-primary">{{ selectedLog.movie_title || 'Unknown Movie' }}</div>
                <div class="text-sm text-muted-foreground">{{ formatDate(selectedLog.showtime_date) || 'Unknown Date' }}</div>
                <div class="text-xs font-mono text-muted-foreground mt-1">ID: {{ selectedLog.showtime_id }}</div>
              </div>
              <div v-else class="text-sm text-muted-foreground italic">N/A</div>
            </div>
          </div>
          
          <div class="space-y-4">
            <div>
              <div class="text-xs font-semibold text-muted-foreground uppercase mb-1">Event Action</div>
              <span class="px-2 py-1 rounded-md text-xs font-bold"
                    :class="{
                      'bg-blue-500/10 text-blue-500': selectedLog.action === 'SEAT_LOCKED',
                      'bg-green-500/10 text-green-500': selectedLog.action === 'BOOKING_CONFIRMED',
                      'bg-primary/10 text-primary': !['SEAT_LOCKED', 'BOOKING_CONFIRMED'].includes(selectedLog.action)
                    }">
                {{ selectedLog.action }}
              </span>
            </div>
            <div>
              <div class="text-xs font-semibold text-muted-foreground uppercase mb-1">Target Seat</div>
              <div class="font-medium">{{ selectedLog.seat_number || 'N/A' }}</div>
            </div>
            <div>
              <div class="text-xs font-semibold text-muted-foreground uppercase mb-1">Timestamp</div>
              <div class="text-sm text-muted-foreground">{{ formatDate(selectedLog.timestamp) }}</div>
            </div>
          </div>
          
          <div class="md:col-span-2">
            <div class="text-xs font-semibold text-muted-foreground uppercase mb-2">Raw JSON / Original Log Text</div>
            <div class="p-3 bg-secondary/30 rounded-lg border border-border/50 text-sm font-mono text-muted-foreground break-all whitespace-pre-wrap">
              {{ isJson(selectedLog.details) ? JSON.stringify(parseJson(selectedLog.details), null, 2) : selectedLog.details }}
            </div>
          </div>
        </div>
      </DialogContent>
    </Dialog>

  </div>
</template>
