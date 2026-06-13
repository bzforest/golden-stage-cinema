<script setup lang="ts">
import { ref } from 'vue'
import { SearchIcon } from '@lucide/vue'

const emit = defineEmits<{
  (e: 'search', query: string): void
}>()

const query = ref('')

const handleSearch = () => {
  emit('search', query.value)
}
</script>

<template>
  <div class="relative w-full">
    <SearchIcon class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
    <input 
      v-model="query"
      @keyup.enter="handleSearch"
      type="text" 
      placeholder="Search movies..." 
      class="w-full bg-secondary/40 border border-border/50 rounded-full pl-11 pr-4 py-2 text-sm text-foreground focus:outline-none focus:ring-1 focus:ring-primary/50 focus:bg-secondary/60 transition-all placeholder:text-muted-foreground/70" 
    />
  </div>
</template>
