<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { SearchIcon } from '@lucide/vue'
import { useRouter } from 'vue-router'
import { useMovieStore } from '@/stores/useMovieStore'
import { refDebounced } from '@vueuse/core'

const emit = defineEmits<{
  (e: 'search', query: string): void
}>()

const router = useRouter()
const movieStore = useMovieStore()

const query = ref('')
const debouncedQuery = refDebounced(query, 500)
const isFocused = ref(false)
const searchContainer = ref<HTMLElement | null>(null)

// Ensure movies are loaded for local search
onMounted(async () => {
  if (movieStore.movies.length === 0) {
    await movieStore.fetchMovies()
  }
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})

const handleClickOutside = (event: MouseEvent) => {
  if (searchContainer.value && !searchContainer.value.contains(event.target as Node)) {
    isFocused.value = false
  }
}

const handleSearch = () => {
  emit('search', query.value)
  if (searchResults.value.length > 0) {
  }
}

const searchResults = computed(() => {
  if (!debouncedQuery.value.trim()) return []
  const lowerQuery = debouncedQuery.value.toLowerCase()
  return movieStore.movies.filter(movie => 
    movie.title.toLowerCase().includes(lowerQuery) || 
    movie.genre.toLowerCase().includes(lowerQuery)
  ).slice(0, 5)
})

const navigateToMovie = (id: string) => {
  query.value = ''
  isFocused.value = false
  router.push(`/movie/${id}`)
}
</script>

<template>
  <div class="relative w-full" ref="searchContainer">
    <SearchIcon class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground z-10" />
    <input 
      v-model="query"
      @focus="isFocused = true"
      @keyup.enter="handleSearch"
      type="text" 
      placeholder="Search movies..." 
      class="relative z-10 w-full bg-secondary/40 border border-border/50 rounded-full pl-11 pr-4 py-2 text-sm text-foreground focus:outline-none focus:ring-1 focus:ring-primary/50 focus:bg-secondary/60 transition-all placeholder:text-muted-foreground/70" 
    />

    <!-- Dropdown Results -->
    <div 
      v-if="isFocused && debouncedQuery.length > 0" 
      class="absolute top-full left-0 right-0 mt-2 bg-card border border-border/50 rounded-xl shadow-xl overflow-hidden z-50 flex flex-col max-h-[400px] overflow-y-auto hide-scrollbar"
    >
      <div v-if="searchResults.length === 0" class="p-4 text-sm text-muted-foreground text-center">
        No movies found
      </div>
      <template v-else>
        <button
          v-for="movie in searchResults"
          :key="movie.id"
          @click="navigateToMovie(movie.id)"
          class="flex items-center gap-4 p-3 hover:bg-secondary/50 transition-colors text-left border-b border-border/50 last:border-0 cursor-pointer"
        >
          <img :src="movie.posterUrl" :alt="movie.title" class="w-12 h-16 object-cover rounded-md" />
          <div class="flex flex-col flex-1 min-w-0">
            <h4 class="font-bold text-sm text-foreground truncate">{{ movie.title }}</h4>
            <p class="text-xs text-muted-foreground truncate">{{ movie.genre }} • {{ movie.duration }}</p>
          </div>
        </button>
      </template>
    </div>
  </div>
</template>

<style scoped>
.hide-scrollbar {
  -ms-overflow-style: none;  /* IE and Edge */
  scrollbar-width: none;  /* Firefox */
}
.hide-scrollbar::-webkit-scrollbar {
  display: none; /* Chrome, Safari and Opera */
}
</style>
