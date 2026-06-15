<template>
  <div class="relative" ref="containerRef">
    <div 
      @click="isOpen = !isOpen" 
      class="flex items-center justify-between h-9 px-3 rounded-md border border-input bg-transparent text-sm shadow-sm cursor-pointer hover:border-primary/50 transition-colors"
      :class="!modelValue ? 'text-muted-foreground' : 'text-foreground'"
    >
      <span>{{ formattedDate }}</span>
      <CalendarIcon class="w-4 h-4 opacity-50" />
    </div>

    <div v-if="isOpen" class="absolute z-50 top-full mt-1 left-0 w-64 bg-card border border-border rounded-lg shadow-lg p-3">
      <div class="flex items-center justify-between mb-3">
        <Button variant="ghost" size="icon" class="h-7 w-7" @click.stop="prevMonth">
          <ChevronLeftIcon class="h-4 w-4" />
        </Button>
        <span class="font-medium text-sm">{{ currentMonthName }} {{ currentYear }}</span>
        <Button variant="ghost" size="icon" class="h-7 w-7" @click.stop="nextMonth">
          <ChevronRightIcon class="h-4 w-4" />
        </Button>
      </div>
      
      <div class="grid grid-cols-7 gap-1 text-center text-xs text-muted-foreground mb-1">
        <div v-for="day in ['Su', 'Mo', 'Tu', 'We', 'Th', 'Fr', 'Sa']" :key="day">{{ day }}</div>
      </div>
      
      <div class="grid grid-cols-7 gap-1 text-sm">
        <div v-for="blank in blankDays" :key="'blank-'+blank" class="h-8 w-8"></div>
        <div 
          v-for="day in daysInMonth" 
          :key="day"
          @click.stop="selectDate(day)"
          class="h-8 w-8 flex items-center justify-center rounded-md cursor-pointer transition-colors"
          :class="isSameDate(day) ? 'bg-primary text-primary-foreground font-bold' : 'hover:bg-secondary text-foreground'"
        >
          {{ day }}
        </div>
      </div>

      <div class="mt-3 text-center border-t border-border pt-2 flex items-center gap-2">
        <Button variant="ghost" size="sm" class="text-xs text-muted-foreground flex-1 h-7 hover:text-foreground" @click.stop="clearDate">Clear</Button>
        <Button variant="outline" size="sm" class="text-xs flex-1 h-7" @click.stop="setToday">Today</Button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { CalendarIcon, ChevronLeftIcon, ChevronRightIcon } from '@lucide/vue'
import { Button } from '@/components/ui/button'

const props = defineProps<{
  modelValue: string // YYYY-MM-DD
}>()

const emit = defineEmits(['update:modelValue'])

const isOpen = ref(false)
const containerRef = ref<HTMLElement | null>(null)

// Current viewing month
const currentDate = ref(props.modelValue ? new Date(props.modelValue) : new Date())
const currentMonth = ref(currentDate.value.getMonth())
const currentYear = ref(currentDate.value.getFullYear())

// Sync viewing month when modelValue changes
watch(() => props.modelValue, (newVal) => {
  if (newVal) {
    const d = new Date(newVal)
    currentMonth.value = d.getMonth()
    currentYear.value = d.getFullYear()
  }
})

// Close on outside click
const handleClickOutside = (event: MouseEvent) => {
  if (containerRef.value && !containerRef.value.contains(event.target as Node)) {
    isOpen.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})

const monthNames = ['January', 'February', 'March', 'April', 'May', 'June', 'July', 'August', 'September', 'October', 'November', 'December']
const currentMonthName = computed(() => monthNames[currentMonth.value])

const daysInMonth = computed(() => new Date(currentYear.value, currentMonth.value + 1, 0).getDate())
const blankDays = computed(() => new Date(currentYear.value, currentMonth.value, 1).getDay())

const formattedDate = computed(() => {
  if (!props.modelValue) return 'Select Date'
  const d = new Date(props.modelValue)
  if (isNaN(d.getTime())) return 'Invalid Date'
  return `${d.getDate().toString().padStart(2, '0')}/${(d.getMonth() + 1).toString().padStart(2, '0')}/${d.getFullYear()}`
})

const isSameDate = (day: number) => {
  if (!props.modelValue) return false
  const d = new Date(props.modelValue)
  return d.getDate() === day && d.getMonth() === currentMonth.value && d.getFullYear() === currentYear.value
}

const selectDate = (day: number) => {
  const m = (currentMonth.value + 1).toString().padStart(2, '0')
  const d = day.toString().padStart(2, '0')
  const val = `${currentYear.value}-${m}-${d}`
  emit('update:modelValue', val)
  isOpen.value = false
}

const clearDate = () => {
  emit('update:modelValue', '')
  isOpen.value = false
}

const setToday = () => {
  const d = new Date()
  const m = (d.getMonth() + 1).toString().padStart(2, '0')
  const day = d.getDate().toString().padStart(2, '0')
  const val = `${d.getFullYear()}-${m}-${day}`
  emit('update:modelValue', val)
  currentMonth.value = d.getMonth()
  currentYear.value = d.getFullYear()
  isOpen.value = false
}

const prevMonth = () => {
  if (currentMonth.value === 0) {
    currentMonth.value = 11
    currentYear.value--
  } else {
    currentMonth.value--
  }
}

const nextMonth = () => {
  if (currentMonth.value === 11) {
    currentMonth.value = 0
    currentYear.value++
  } else {
    currentMonth.value++
  }
}
</script>
