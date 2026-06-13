import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api } from '@/lib/axios'

export interface Movie {
  id: string
  title: string
  description: string
  posterUrl: string
  backdropUrl: string
  genre: string
  duration: string
  rating: number
}

export const useMovieStore = defineStore('movie', () => {
  const movies = ref<Movie[]>([])
  const isLoading = ref(false)
  const error = ref<string | null>(null)

  const fetchMovies = async () => {
    isLoading.value = true
    error.value = null
    try {
      // In a real scenario, this will call the backend API
      const response = await api.get('/movies')
      
      // แปลงข้อมูลจาก API ให้ตรงกับ Interface (Mapping)
      const rawMovies = response.data || []
      movies.value = rawMovies.map((movie: any) => ({
        id: movie.id || movie._id || String(Math.random()),
        title: movie.title || 'Unknown Title',
        description: movie.synopsis || movie.description || '',
        posterUrl: movie.poster_url || movie.posterUrl || '',
        backdropUrl: movie.backdrop_url || movie.backdropUrl || '',
        genre: movie.genre || 'Unknown Genre',
        duration: movie.duration_mins ? `${movie.duration_mins}m` : (movie.duration || '0m'),
        rating: movie.rating || 0
      }))
    } catch (err: any) {
      console.error('Failed to fetch movies:', err)
      error.value = 'ไม่สามารถเชื่อมต่อเซิร์ฟเวอร์ได้ (กำลังแสดงข้อมูลจำลอง)'
      
      // Fallback mock data matching the reference image for demonstration
      movies.value = [
        {
          id: '1',
          title: 'Demon Slayer Infinite Castle : Part 1',
          description: 'a massive, disorienting pocket dimension serving as the primary lair for Muzan Kibutsuji and the Twelve Kizuki, characterized by an ever-shifting, gravity-defying structure of Japanese wooden rooms and hanging staircases',
          posterUrl: 'https://images.unsplash.com/photo-1578681994506-b8f463449011?q=80&w=400&auto=format&fit=crop',
          backdropUrl: 'https://images.unsplash.com/photo-1578681994506-b8f463449011?q=80&w=2070&auto=format&fit=crop',
          genre: 'Anime / Action / Fantasy',
          duration: '2h 15m',
          rating: 9.0
        },
        {
          id: '2',
          title: 'One Piece Film Gold',
          description: 'The Straw Hat Pirates take on Gild Tesoro, one of the richest and most ambitious men in the world.',
          posterUrl: 'https://images.unsplash.com/photo-1612438214708-f428a707dd4e?q=80&w=400&auto=format&fit=crop',
          backdropUrl: '',
          genre: 'Anime / Action / Adventure',
          duration: '2h 0m',
          rating: 8.2
        },
        {
          id: '3',
          title: 'JoJo\'s Bizarre Adventure: Steel Ball Run',
          description: 'Set in 1890, the story follows Johnny Joestar and Gyro Zeppeli as they enter a cross-country horse race across the United States.',
          posterUrl: 'https://images.unsplash.com/photo-1560930950-5c20d80c318a?q=80&w=400&auto=format&fit=crop',
          backdropUrl: '',
          genre: 'Anime / Action / Adventure',
          duration: '2h 20m',
          rating: 9.8
        },
        {
          id: '4',
          title: 'Chainsaw man the movie Reze Arc',
          description: 'Denji meets Reze, a girl who shows interest in him. However, things take a dark turn as her true identity is revealed.',
          posterUrl: 'https://images.unsplash.com/photo-1606606764516-09252c5025d5?q=80&w=400&auto=format&fit=crop',
          backdropUrl: '',
          genre: 'Anime / Action / Dark Fantasy',
          duration: '1h 45m',
          rating: 9.6
        },
        {
          id: '5',
          title: 'Dragon Ball Super: Broly',
          description: 'Goku and Vegeta encounter Broly, a Saiyan warrior unlike any fighter they\'ve faced before.',
          posterUrl: 'https://images.unsplash.com/photo-1535666669445-e8c15cd2e7d9?q=80&w=400&auto=format&fit=crop',
          backdropUrl: '',
          genre: 'Anime / Action / Adventure',
          duration: '1h 40m',
          rating: 9.3
        }
      ]
    } finally {
      isLoading.value = false
    }
  }

  return {
    movies,
    isLoading,
    error,
    fetchMovies
  }
})
