<script setup lang="ts">
import { ref } from 'vue'
import Navbar from '@/components/layout/Navbar.vue'
import Footer from '@/components/layout/Footer.vue'
import SearchBar from '@/components/SearchBar.vue'
import { toast } from 'vue-sonner'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle, CardDescription, CardFooter } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Separator } from '@/components/ui/separator'
import { Skeleton } from '@/components/ui/skeleton'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'

const searchQuery = ref('')
const selectedDate = ref(0)
const selectedTime = ref('')
const sampleName = ref('')

const dates = [
  { day: 13, month: 'JUN', weekday: 'Fri' },
  { day: 14, month: 'JUN', weekday: 'Sat' },
  { day: 15, month: 'JUN', weekday: 'Sun' },
  { day: 16, month: 'JUN', weekday: 'Mon' },
  { day: 17, month: 'JUN', weekday: 'Tue' },
  { day: 18, month: 'JUN', weekday: 'Wed' },
  { day: 19, month: 'JUN', weekday: 'Thu' },
]

const showtimes = ['14:00', '17:30', '20:15', '23:00']
</script>

<template>
  <div class="min-h-screen bg-background text-foreground flex flex-col">
    <!-- Header / Navbar Component -->
    <Navbar />

    <main class="max-w-7xl mx-auto px-6 py-10 space-y-14">

      <!-- ═══════════════════ SECTION 1: Typography ═══════════════════ -->
      <section>
        <h2 class="text-sm font-semibold text-primary uppercase tracking-widest mb-6">Typography</h2>
        <Separator class="mb-8" />
        <div class="space-y-4">
          <h1 class="text-5xl font-extrabold text-heading tracking-tight">
            H1 — Golden Stage Cinema
          </h1>
          <h2 class="text-3xl font-bold text-heading">
            H2 — Now Showing
          </h2>
          <h3 class="text-2xl font-semibold text-heading">
            H3 — Select Your Seats
          </h3>
          <h4 class="text-lg font-medium text-heading">
            H4 — Booking Summary
          </h4>
          <p class="text-base text-muted-foreground max-w-2xl leading-relaxed">
            Paragraph — ยินดีต้อนรับสู่ Golden Stage Cinema ระบบจองตั๋วหนังออนไลน์ที่รองรับการจองพร้อมกันจำนวนมาก ป้องกันปัญหา Double Booking ด้วยเทคโนโลยี Redis Distributed Lock
          </p>
        </div>
      </section>

      <!-- ═══════════════════ SECTION 2: Logo ═══════════════════ -->
      <section>
        <h2 class="text-sm font-semibold text-primary uppercase tracking-widest mb-6">Logo</h2>
        <Separator class="mb-8" />
        <div class="flex items-center gap-10">
          <!-- Logo Large -->
          <div class="flex items-center gap-3">
            <div class="w-12 h-12 bg-primary rounded-xl flex items-center justify-center shadow-lg shadow-primary/25">
              <svg xmlns="http://www.w3.org/2000/svg" class="w-7 h-7 text-primary-foreground" viewBox="0 0 24 24" fill="currentColor"><path d="M18 3v2h-2V3H8v2H6V3H4v18h2v-2h2v2h8v-2h2v2h2V3h-2zM8 17H6v-2h2v2zm0-4H6v-2h2v2zm0-4H6V7h2v2zm10 8h-2v-2h2v2zm0-4h-2v-2h2v2zm0-4h-2V7h2v2z"/></svg>
            </div>
            <span class="text-3xl font-bold text-heading">
              <span class="text-foreground">Golden</span><span class="text-primary">Stage</span>
            </span>
          </div>
          <!-- Logo Small -->
          <div class="flex items-center gap-2">
            <div class="w-8 h-8 bg-primary rounded-lg flex items-center justify-center">
              <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5 text-primary-foreground" viewBox="0 0 24 24" fill="currentColor"><path d="M18 3v2h-2V3H8v2H6V3H4v18h2v-2h2v2h8v-2h2v2h2V3h-2zM8 17H6v-2h2v2zm0-4H6v-2h2v2zm0-4H6V7h2v2zm10 8h-2v-2h2v2zm0-4h-2v-2h2v2zm0-4h-2V7h2v2z"/></svg>
            </div>
            <span class="text-lg font-bold text-heading">
              <span class="text-foreground">Golden</span><span class="text-primary">Stage</span>
            </span>
          </div>
        </div>
      </section>

      <!-- ═══════════════════ SECTION 3: Buttons ═══════════════════ -->
      <section>
        <h2 class="text-sm font-semibold text-primary uppercase tracking-widest mb-6">Buttons</h2>
        <Separator class="mb-8" />
        <div class="flex flex-wrap gap-4">
          <Button class="cursor-pointer">Book Tickets</Button>
          <Button variant="secondary" class="cursor-pointer">Secondary</Button>
          <Button variant="outline" class="cursor-pointer">Outline</Button>
          <Button variant="ghost" class="cursor-pointer">Ghost</Button>
          <Button variant="destructive" class="cursor-pointer">Cancel Booking</Button>
          <Button class="cursor-not-allowed" disabled>Disabled</Button>
        </div>
        <div class="flex flex-wrap gap-4 mt-6">
          <Button size="lg" class="cursor-pointer gap-2 px-8">
            <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polygon points="5 3 19 12 5 21 5 3"/></svg>
            Watch Trailer
          </Button>
          <Button size="lg" variant="outline" class="cursor-pointer gap-2 px-8">
            <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M19 21l-7-5-7 5V5a2 2 0 0 1 2-2h10a2 2 0 0 1 2 2z"/></svg>
            Add to Watchlist
          </Button>
        </div>
      </section>

      <!-- ═══════════════════ SECTION 4: Badges ═══════════════════ -->
      <section>
        <h2 class="text-sm font-semibold text-primary uppercase tracking-widest mb-6">Badges (Genre Tags)</h2>
        <Separator class="mb-8" />
        <div class="flex flex-wrap gap-3">
          <Badge class="bg-primary/15 text-primary border-primary/30 hover:bg-primary/25 cursor-pointer">Action</Badge>
          <Badge class="bg-primary/15 text-primary border-primary/30 hover:bg-primary/25 cursor-pointer">Fantasy</Badge>
          <Badge class="bg-primary/15 text-primary border-primary/30 hover:bg-primary/25 cursor-pointer">Anime</Badge>
          <Badge class="bg-primary/15 text-primary border-primary/30 hover:bg-primary/25 cursor-pointer">Horror</Badge>
          <Badge class="bg-primary/15 text-primary border-primary/30 hover:bg-primary/25 cursor-pointer">Romance</Badge>
          <Badge class="bg-primary/15 text-primary border-primary/30 hover:bg-primary/25 cursor-pointer">Sci-Fi</Badge>
          <Badge variant="outline">IMAX</Badge>
          <Badge variant="secondary">Standard 2D</Badge>
          <Badge class="bg-green-500/15 text-green-400 border-green-500/30">NOW SHOWING</Badge>
        </div>
      </section>

      <!-- ═══════════════════ SECTION 5: Search Bar ═══════════════════ -->
      <section>
        <h2 class="text-sm font-semibold text-primary uppercase tracking-widest mb-6">Search Bar</h2>
        <Separator class="mb-8" />
        <div class="max-w-md">
          <SearchBar />
        </div>
      </section>

      <!-- ═══════════════════ SECTION 6: Select Date & Time ═══════════════════ -->
      <section>
        <h2 class="text-sm font-semibold text-primary uppercase tracking-widest mb-6">Select Date & Time</h2>
        <Separator class="mb-8" />

        <!-- Date Picker -->
        <div class="flex gap-3 mb-8 overflow-x-auto pb-2">
          <button
            v-for="(date, index) in dates"
            :key="index"
            @click="selectedDate = index"
            class="cursor-pointer flex flex-col items-center min-w-[72px] py-3 px-4 rounded-xl border-2 transition-all duration-200"
            :class="selectedDate === index
              ? 'border-primary bg-primary/10 text-primary'
              : 'border-border bg-card text-muted-foreground hover:border-primary/40'"
          >
            <span class="text-xs font-medium uppercase">{{ date.month }}</span>
            <span class="text-2xl font-bold text-heading">{{ date.day }}</span>
            <span class="text-xs">{{ date.weekday }}</span>
          </button>
        </div>

        <!-- Showtime Picker -->
        <Card class="bg-card border-border max-w-2xl">
          <CardHeader class="pb-3">
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-2">
                <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 text-muted-foreground" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M20 10c0 4.993-5.539 10.193-7.399 11.799a1 1 0 0 1-1.202 0C9.539 20.193 4 14.993 4 10a8 8 0 0 1 16 0"/><circle cx="12" cy="10" r="3"/></svg>
                <CardTitle class="text-base">GoldenStage Central Mall</CardTitle>
              </div>
              <Badge variant="secondary">Standard 2D</Badge>
            </div>
          </CardHeader>
          <CardContent>
            <div class="flex flex-wrap gap-3">
              <button
                v-for="time in showtimes"
                :key="time"
                @click="selectedTime = time"
                class="cursor-pointer px-5 py-2.5 rounded-lg border text-sm font-medium transition-all duration-200"
                :class="selectedTime === time
                  ? 'border-primary bg-primary text-primary-foreground'
                  : 'border-border bg-secondary text-foreground hover:border-primary/50'"
              >
                {{ time }}
              </button>
            </div>
          </CardContent>
        </Card>
      </section>

      <!-- ═══════════════════ SECTION 7: Movie Card ═══════════════════ -->
      <section>
        <h2 class="text-sm font-semibold text-primary uppercase tracking-widest mb-6">Movie Card</h2>
        <Separator class="mb-8" />
        <div class="grid grid-cols-2 md:grid-cols-4 gap-6">
          <Card v-for="i in 4" :key="i" class="bg-card border-border overflow-hidden group cursor-pointer hover:border-primary/40 transition-all duration-300">
            <div class="aspect-[2/3] bg-gradient-to-br from-primary/20 to-secondary relative overflow-hidden">
              <div class="absolute inset-0 flex items-center justify-center">
                <svg xmlns="http://www.w3.org/2000/svg" class="w-12 h-12 text-muted-foreground/30" viewBox="0 0 24 24" fill="currentColor"><path d="M18 3v2h-2V3H8v2H6V3H4v18h2v-2h2v2h8v-2h2v2h2V3h-2zM8 17H6v-2h2v2zm0-4H6v-2h2v2zm0-4H6V7h2v2zm10 8h-2v-2h2v2zm0-4h-2v-2h2v2zm0-4h-2V7h2v2z"/></svg>
              </div>
              <div class="absolute top-2 right-2">
                <Badge class="bg-primary/90 text-primary-foreground text-xs gap-1">
                  <svg xmlns="http://www.w3.org/2000/svg" class="w-3 h-3" viewBox="0 0 24 24" fill="currentColor"><polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/></svg>
                  9.0
                </Badge>
              </div>
              <div class="absolute inset-0 bg-gradient-to-t from-black/60 via-transparent to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300 flex items-end p-3">
                <Button size="sm" class="cursor-pointer w-full gap-1.5 text-xs">
                  Book Now
                </Button>
              </div>
            </div>
            <CardContent class="p-3">
              <h3 class="font-semibold text-heading text-sm truncate">Demon Slayer: Infinity Castle</h3>
              <p class="text-xs text-muted-foreground mt-1">Anime / Action / Fantasy</p>
            </CardContent>
          </Card>
        </div>
      </section>

      <!-- ═══════════════════ SECTION 8: My Tickets ═══════════════════ -->
      <section>
        <h2 class="text-sm font-semibold text-primary uppercase tracking-widest mb-6">My Tickets</h2>
        <Separator class="mb-8" />
        <Card class="bg-card border-border max-w-3xl overflow-hidden">
          <div class="flex">
            <div class="w-36 bg-gradient-to-br from-primary/20 to-secondary flex-shrink-0 flex items-center justify-center">
              <svg xmlns="http://www.w3.org/2000/svg" class="w-10 h-10 text-muted-foreground/30" viewBox="0 0 24 24" fill="currentColor"><path d="M18 3v2h-2V3H8v2H6V3H4v18h2v-2h2v2h8v-2h2v2h2V3h-2zM8 17H6v-2h2v2zm0-4H6v-2h2v2zm0-4H6V7h2v2zm10 8h-2v-2h2v2zm0-4h-2v-2h2v2zm0-4h-2V7h2v2z"/></svg>
            </div>
            <div class="flex-1 p-5">
              <div class="flex items-start justify-between">
                <h3 class="text-lg font-bold text-heading">Demon Slayer: Infinity Castle</h3>
                <Badge variant="outline" class="text-xs">IMAX</Badge>
              </div>
              <div class="mt-3 space-y-1.5 text-sm text-muted-foreground">
                <div class="flex items-center gap-2">
                  <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect width="18" height="18" x="3" y="4" rx="2" ry="2"/><line x1="16" x2="16" y1="2" y2="6"/><line x1="8" x2="8" y1="2" y2="6"/><line x1="3" x2="21" y1="10" y2="10"/></svg>
                  Saturday, June 14th, 2026
                </div>
                <div class="flex items-center gap-2">
                  <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
                  8:30 PM (Duration: 2h 15m)
                </div>
                <div class="flex items-center gap-2">
                  <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M20 10c0 4.993-5.539 10.193-7.399 11.799a1 1 0 0 1-1.202 0C9.539 20.193 4 14.993 4 10a8 8 0 0 1 16 0"/><circle cx="12" cy="10" r="3"/></svg>
                  GoldenStage Downtown — Screen 1
                </div>
              </div>
              <Separator class="my-4" />
              <div class="flex items-center gap-6 text-sm">
                <div>
                  <span class="text-muted-foreground">Seats</span>
                  <p class="font-bold text-primary">F12, F13 (VIP)</p>
                </div>
                <div>
                  <span class="text-muted-foreground">Booking ID</span>
                  <p class="font-bold">#GS-84729</p>
                </div>
              </div>
            </div>
          </div>
        </Card>
      </section>

      <!-- ═══════════════════ SECTION 9: Past Movies (History Card) ═══════════════════ -->
      <section>
        <h2 class="text-sm font-semibold text-primary uppercase tracking-widest mb-6">Watch History</h2>
        <Separator class="mb-8" />
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4 max-w-3xl">
          <Card v-for="n in 2" :key="n" class="bg-card border-border overflow-hidden hover:border-primary/30 transition-all">
            <div class="flex">
              <div class="w-20 bg-gradient-to-br from-primary/10 to-secondary flex-shrink-0 flex items-center justify-center">
                <svg xmlns="http://www.w3.org/2000/svg" class="w-8 h-8 text-muted-foreground/30" viewBox="0 0 24 24" fill="currentColor"><path d="M18 3v2h-2V3H8v2H6V3H4v18h2v-2h2v2h8v-2h2v2h2V3h-2zM8 17H6v-2h2v2zm0-4H6v-2h2v2zm0-4H6V7h2v2zm10 8h-2v-2h2v2zm0-4h-2v-2h2v2zm0-4h-2V7h2v2z"/></svg>
              </div>
              <CardContent class="p-4 flex-1">
                <h3 class="font-semibold text-heading text-sm">One Piece Film Gold</h3>
                <p class="text-xs text-muted-foreground mt-1">Sunday, June 8th, 2026 · 4:15 PM</p>
                <p class="text-xs text-muted-foreground mt-0.5">Seats: J4, J5</p>
              </CardContent>
            </div>
          </Card>
        </div>
      </section>

      <!-- ═══════════════════ SECTION 10: Input Text ═══════════════════ -->
      <section>
        <h2 class="text-sm font-semibold text-primary uppercase tracking-widest mb-6">Form Inputs</h2>
        <Separator class="mb-8" />
        <div class="grid grid-cols-1 md:grid-cols-2 gap-6 max-w-2xl">
          <div class="space-y-2">
            <label class="text-sm font-medium text-primary">First Name</label>
            <Input v-model="sampleName" placeholder="Enter your name..." class="bg-card border-border" />
          </div>
          <div class="space-y-2">
            <label class="text-sm font-medium text-primary">Last Name</label>
            <Input placeholder="Enter your last name..." class="bg-card border-border" />
          </div>
          <div class="space-y-2">
            <label class="text-sm font-medium text-primary">Email Address</label>
            <div class="relative">
              <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect width="20" height="16" x="2" y="4" rx="2"/><path d="m22 7-8.97 5.7a1.94 1.94 0 0 1-2.06 0L2 7"/></svg>
              <Input placeholder="you@example.com" class="pl-10 bg-card border-border" />
            </div>
          </div>
          <div class="space-y-2">
            <label class="text-sm font-medium text-primary">Cinema Branch</label>
            <Select>
              <SelectTrigger class="bg-card border-border cursor-pointer">
                <SelectValue placeholder="Select branch" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="central" class="cursor-pointer">GoldenStage Central Mall</SelectItem>
                <SelectItem value="downtown" class="cursor-pointer">GoldenStage Downtown</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>
      </section>

      <!-- ═══════════════════ SECTION 11: Seats ═══════════════════ -->
      <section>
        <h2 class="text-sm font-semibold text-primary uppercase tracking-widest mb-6">Seat Map</h2>
        <Separator class="mb-8" />

        <!-- Legend -->
        <div class="flex items-center gap-8 mb-8">
          <div class="flex items-center gap-2">
            <div class="w-8 h-8 rounded-md bg-secondary border border-border"></div>
            <span class="text-sm text-muted-foreground">Available</span>
          </div>
          <div class="flex items-center gap-2">
            <div class="w-8 h-8 rounded-md bg-primary border border-primary"></div>
            <span class="text-sm text-primary font-medium">Selected</span>
          </div>
          <div class="flex items-center gap-2">
            <div class="w-8 h-8 rounded-md bg-muted-foreground/20 border border-muted-foreground/20"></div>
            <span class="text-sm text-muted-foreground">Booked</span>
          </div>
          <div class="flex items-center gap-2">
            <div class="w-8 h-8 rounded-md bg-red-500/30 border border-red-500/50"></div>
            <span class="text-sm text-red-400">Locked</span>
          </div>
        </div>

        <!-- Seat Grid Preview -->
        <Card class="bg-card border-border max-w-2xl p-8">
          <!-- Screen -->
          <div class="mb-10 relative">
            <div class="h-1.5 bg-gradient-to-r from-transparent via-muted-foreground/40 to-transparent rounded-full mx-10"></div>
            <p class="text-center text-xs text-muted-foreground mt-2 tracking-widest uppercase">Screen</p>
          </div>

          <!-- Seats -->
          <div class="space-y-3">
            <!-- Row A -->
            <div class="flex items-center gap-2 justify-center">
              <span class="w-6 text-xs text-muted-foreground font-medium">A</span>
              <div class="flex gap-2">
                <button class="cursor-pointer w-9 h-9 rounded-md bg-primary border-2 border-primary text-primary-foreground text-xs font-bold flex items-center justify-center hover:scale-105 transition-transform">A1</button>
                <button class="cursor-pointer w-9 h-9 rounded-md bg-secondary border border-border text-xs text-muted-foreground flex items-center justify-center hover:border-primary/50 transition-colors">A2</button>
                <button class="cursor-pointer w-9 h-9 rounded-md bg-secondary border border-border text-xs text-muted-foreground flex items-center justify-center hover:border-primary/50 transition-colors">A3</button>
                <button class="cursor-pointer w-9 h-9 rounded-md bg-primary border-2 border-primary text-primary-foreground text-xs font-bold flex items-center justify-center hover:scale-105 transition-transform">A4</button>
                <div class="w-6"></div>
                <button class="cursor-not-allowed w-9 h-9 rounded-md bg-muted-foreground/20 border border-muted-foreground/20 text-xs text-muted-foreground/40 flex items-center justify-center" disabled>A5</button>
                <button class="cursor-pointer w-9 h-9 rounded-md bg-secondary border border-border text-xs text-muted-foreground flex items-center justify-center hover:border-primary/50 transition-colors">A6</button>
              </div>
              <span class="w-6 text-xs text-muted-foreground font-medium text-right">A</span>
            </div>
            <!-- Row B -->
            <div class="flex items-center gap-2 justify-center">
              <span class="w-6 text-xs text-muted-foreground font-medium">B</span>
              <div class="flex gap-2">
                <button class="cursor-pointer w-9 h-9 rounded-md bg-secondary border border-border text-xs text-muted-foreground flex items-center justify-center hover:border-primary/50 transition-colors">B1</button>
                <button class="cursor-pointer w-9 h-9 rounded-md bg-secondary border border-border text-xs text-muted-foreground flex items-center justify-center hover:border-primary/50 transition-colors">B2</button>
                <button class="cursor-not-allowed w-9 h-9 rounded-md bg-red-500/30 border-2 border-red-500/50 text-xs text-red-400 flex items-center justify-center" disabled>B3</button>
                <button class="cursor-not-allowed w-9 h-9 rounded-md bg-muted-foreground/20 border border-muted-foreground/20 text-xs text-muted-foreground/40 flex items-center justify-center" disabled>B4</button>
                <div class="w-6"></div>
                <button class="cursor-pointer w-9 h-9 rounded-md bg-secondary border border-border text-xs text-muted-foreground flex items-center justify-center hover:border-primary/50 transition-colors">B5</button>
                <button class="cursor-pointer w-9 h-9 rounded-md bg-secondary border border-border text-xs text-muted-foreground flex items-center justify-center hover:border-primary/50 transition-colors">B6</button>
              </div>
              <span class="w-6 text-xs text-muted-foreground font-medium text-right">B</span>
            </div>
            <!-- Row C -->
            <div class="flex items-center gap-2 justify-center">
              <span class="w-6 text-xs text-muted-foreground font-medium">C</span>
              <div class="flex gap-2">
                <button class="cursor-not-allowed w-9 h-9 rounded-md bg-muted-foreground/20 border border-muted-foreground/20 text-xs text-muted-foreground/40 flex items-center justify-center" disabled>C1</button>
                <button class="cursor-not-allowed w-9 h-9 rounded-md bg-muted-foreground/20 border border-muted-foreground/20 text-xs text-muted-foreground/40 flex items-center justify-center" disabled>C2</button>
                <button class="cursor-pointer w-9 h-9 rounded-md bg-secondary border border-border text-xs text-muted-foreground flex items-center justify-center hover:border-primary/50 transition-colors">C3</button>
                <button class="cursor-pointer w-9 h-9 rounded-md bg-secondary border border-border text-xs text-muted-foreground flex items-center justify-center hover:border-primary/50 transition-colors">C4</button>
                <div class="w-6"></div>
                <button class="cursor-pointer w-9 h-9 rounded-md bg-secondary border border-border text-xs text-muted-foreground flex items-center justify-center hover:border-primary/50 transition-colors">C5</button>
                <button class="cursor-not-allowed w-9 h-9 rounded-md bg-red-500/30 border-2 border-red-500/50 text-xs text-red-400 flex items-center justify-center" disabled>C6</button>
              </div>
              <span class="w-6 text-xs text-muted-foreground font-medium text-right">C</span>
            </div>
          </div>
        </Card>
      </section>

      <!-- ═══════════════════ SECTION 12: Tabs ═══════════════════ -->
      <section>
        <h2 class="text-sm font-semibold text-primary uppercase tracking-widest mb-6">Tabs (Profile)</h2>
        <Separator class="mb-8" />
        <Tabs default-value="tickets" class="max-w-2xl">
          <TabsList class="w-full">
            <TabsTrigger value="tickets" class="cursor-pointer flex-1 gap-1.5">
              <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M2 9a3 3 0 0 1 0 6v2a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2v-2a3 3 0 0 1 0-6V7a2 2 0 0 0-2-2H4a2 2 0 0 0-2 2Z"/><path d="M13 5v2"/><path d="M13 17v2"/><path d="M13 11v2"/></svg>
              My Tickets
            </TabsTrigger>
            <TabsTrigger value="history" class="cursor-pointer flex-1 gap-1.5">
              <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
              Watch History
            </TabsTrigger>
            <TabsTrigger value="settings" class="cursor-pointer flex-1 gap-1.5">
              <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z"/><circle cx="12" cy="12" r="3"/></svg>
              Settings
            </TabsTrigger>
          </TabsList>
          <TabsContent value="tickets" class="mt-6">
            <p class="text-muted-foreground text-sm">Your upcoming tickets will be displayed here.</p>
          </TabsContent>
          <TabsContent value="history" class="mt-6">
            <p class="text-muted-foreground text-sm">Your past movie watching history.</p>
          </TabsContent>
          <TabsContent value="settings" class="mt-6">
            <p class="text-muted-foreground text-sm">Account settings and preferences.</p>
          </TabsContent>
        </Tabs>
      </section>

      <!-- ═══════════════════ SECTION 13: Feedback & Loading ═══════════════════ -->
      <section>
        <h2 class="text-sm font-semibold text-primary uppercase tracking-widest mb-6">Feedback & Loading</h2>
        <Separator class="mb-8" />

        <!-- Toast / Sonner -->
        <div class="mb-10">
          <h3 class="text-lg font-semibold text-heading mb-4">Toast Notifications (Sonner)</h3>
          <p class="text-sm text-muted-foreground mb-5 max-w-xl">แจ้งเตือนผู้ใช้งานแบบ Pop-up เมื่อเกิด Event สำคัญ เช่น ล็อกที่นั่งสำเร็จ, หมดเวลา, หรือชำระเงินเรียบร้อย</p>
          <div class="flex flex-wrap gap-3">
            <Button
              class="cursor-pointer gap-2"
              @click="toast.success('ที่นั่ง A1 ถูกล็อกเรียบร้อยแล้ว!', { description: 'กรุณาชำระเงินภายใน 5 นาที' })"
            >
              <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M20 6 9 17l-5-5"/></svg>
              Success Toast
            </Button>
            <Button
              variant="destructive"
              class="cursor-pointer gap-2"
              @click="toast.error('หมดเวลาทำรายการ!', { description: 'ที่นั่งของคุณถูกปลดล็อกอัตโนมัติ' })"
            >
              <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="15" x2="9" y1="9" y2="15"/><line x1="9" x2="15" y1="9" y2="15"/></svg>
              Error Toast
            </Button>
            <Button
              variant="outline"
              class="cursor-pointer gap-2"
              @click="toast.warning('ที่นั่ง B3 กำลังถูกล็อกโดยผู้ใช้อื่น', { description: 'กรุณาเลือกที่นั่งอื่น' })"
            >
              <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3"/><path d="M12 9v4"/><path d="M12 17h.01"/></svg>
              Warning Toast
            </Button>
            <Button
              variant="secondary"
              class="cursor-pointer gap-2"
              @click="toast.info('ชำระเงินสำเร็จ! ตั๋วของคุณถูกส่งไปที่อีเมลแล้ว', { description: 'Booking #GS-84729' })"
            >
              <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><path d="M12 16v-4"/><path d="M12 8h.01"/></svg>
              Info Toast
            </Button>
          </div>
        </div>

        <!-- Skeleton -->
        <div class="mb-10">
          <h3 class="text-lg font-semibold text-heading mb-4">Skeleton Loading</h3>
          <p class="text-sm text-muted-foreground mb-5 max-w-xl">โครงร่างที่แสดงระหว่างรอข้อมูลจาก API เพื่อให้ผู้ใช้รู้สึกว่าแอปโหลดเร็วและลื่นไหล</p>
          <div class="grid grid-cols-2 md:grid-cols-4 gap-6">
            <Card v-for="i in 4" :key="i" class="bg-card border-border overflow-hidden">
              <Skeleton class="aspect-[2/3] w-full rounded-none" />
              <CardContent class="p-3 space-y-2">
                <Skeleton class="h-4 w-3/4" />
                <Skeleton class="h-3 w-1/2" />
              </CardContent>
            </Card>
          </div>
          <div class="mt-6 max-w-3xl">
            <h4 class="text-sm font-medium text-muted-foreground mb-3">Ticket Skeleton</h4>
            <Card class="bg-card border-border overflow-hidden">
              <div class="flex">
                <Skeleton class="w-36 h-40 rounded-none flex-shrink-0" />
                <div class="flex-1 p-5 space-y-3">
                  <Skeleton class="h-5 w-2/3" />
                  <Skeleton class="h-3 w-1/2" />
                  <Skeleton class="h-3 w-2/5" />
                  <Skeleton class="h-3 w-1/3" />
                </div>
              </div>
            </Card>
          </div>
        </div>

        <!-- Tooltip -->
        <div>
          <h3 class="text-lg font-semibold text-heading mb-4">Tooltip</h3>
          <p class="text-sm text-muted-foreground mb-5 max-w-xl">กล่องข้อความอธิบายเมื่อเอาเมาส์ชี้ — เหมาะสำหรับแสดงข้อมูลเพิ่มเติมของที่นั่งในผังที่นั่ง</p>
          <TooltipProvider :delay-duration="200">
            <div class="flex items-center gap-4">
              <!-- Available -->
              <Tooltip>
                <TooltipTrigger as-child>
                  <button class="cursor-pointer w-11 h-11 rounded-md bg-secondary border border-border text-xs text-muted-foreground flex items-center justify-center hover:border-primary/50 transition-colors">A1</button>
                </TooltipTrigger>
                <TooltipContent side="top" class="bg-card border-border">
                  <p class="font-semibold">A1 — ว่าง</p>
                  <p class="text-xs text-muted-foreground">Standard · 250 THB</p>
                </TooltipContent>
              </Tooltip>

              <!-- Selected -->
              <Tooltip>
                <TooltipTrigger as-child>
                  <button class="cursor-pointer w-11 h-11 rounded-md bg-primary border-2 border-primary text-primary-foreground text-xs font-bold flex items-center justify-center">A2</button>
                </TooltipTrigger>
                <TooltipContent side="top" class="bg-card border-primary/50">
                  <p class="font-semibold text-primary">A2 — เลือกแล้ว</p>
                  <p class="text-xs text-muted-foreground">Standard · 250 THB</p>
                </TooltipContent>
              </Tooltip>

              <!-- Booked -->
              <Tooltip>
                <TooltipTrigger as-child>
                  <button class="cursor-not-allowed w-11 h-11 rounded-md bg-muted-foreground/20 border border-muted-foreground/20 text-xs text-muted-foreground/40 flex items-center justify-center" disabled>A3</button>
                </TooltipTrigger>
                <TooltipContent side="top" class="bg-card border-border">
                  <p class="font-semibold text-muted-foreground">A3 — ถูกจองแล้ว</p>
                  <p class="text-xs text-muted-foreground">ไม่สามารถเลือกได้</p>
                </TooltipContent>
              </Tooltip>

              <!-- Locked -->
              <Tooltip>
                <TooltipTrigger as-child>
                  <button class="cursor-not-allowed w-11 h-11 rounded-md bg-red-500/30 border-2 border-red-500/50 text-xs text-red-400 flex items-center justify-center" disabled>A4</button>
                </TooltipTrigger>
                <TooltipContent side="top" class="bg-card border-red-500/30">
                  <p class="font-semibold text-red-400">A4 — กำลังถูกล็อก</p>
                  <p class="text-xs text-muted-foreground">ถูกจองโดยผู้ใช้อื่น (รอชำระเงิน)</p>
                </TooltipContent>
              </Tooltip>
            </div>
          </TooltipProvider>
        </div>
      </section>

    </main>

    <!-- Footer Component -->
    <div class="mt-auto">
      <Footer />
    </div>
  </div>
</template>
