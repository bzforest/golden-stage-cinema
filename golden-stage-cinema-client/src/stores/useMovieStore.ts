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

export interface Cinema {
  id: string
  name: string
  cinema_type?: string
}

export interface Hall {
  id: string
  cinema_id: string
  name: string
  capacity: number
}

export interface Showtime {
  id: string
  movie_id: string
  cinema_id: string
  hall_id: string
  start_time: string
  end_time: string
}

export const useMovieStore = defineStore('movie', () => {
  const movies = ref<Movie[]>([])
  const currentMovie = ref<Movie | null>(null)
  const showtimes = ref<Showtime[]>([])
  const cinemas = ref<Cinema[]>([])
  const halls = ref<Hall[]>([])
  
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

  const fetchMovieById = async (id: string) => {
    isLoading.value = true
    error.value = null
    try {
      const response = await api.get(`/movies/${id}`)
      const movie = response.data
      currentMovie.value = {
        id: movie.id || movie._id,
        title: movie.title || 'Unknown Title',
        description: movie.synopsis || movie.description || '',
        posterUrl: movie.poster_url || movie.posterUrl || '',
        backdropUrl: movie.backdrop_url || movie.backdropUrl || '',
        genre: movie.genre || 'Unknown Genre',
        duration: movie.duration_mins ? `${movie.duration_mins}m` : (movie.duration || '0m'),
        rating: movie.rating || 0
      }
    } catch (err: any) {
      console.error('Failed to fetch movie details:', err)
      // Fallback: try to find in existing movies list
      const existing = movies.value.find(m => m.id === id)
      if (existing) {
        currentMovie.value = existing
      } else {
        error.value = 'ไม่พบข้อมูลภาพยนตร์'
      }
    } finally {
      isLoading.value = false
    }
  }

  const fetchCinemas = async () => {
    try {
      const response = await api.get('/cinemas')
      cinemas.value = response.data
    } catch (err: any) {
      console.error('Failed to fetch cinemas:', err)
    }
  }

  const fetchHallsByCinema = async (cinemaId: string) => {
    try {
      const response = await api.get(`/cinemas/${cinemaId}/halls`)
      // Merge with existing halls array uniquely
      const newHalls = response.data
      const existingIds = new Set(halls.value.map(h => h.id))
      newHalls.forEach((h: Hall) => {
        if (!existingIds.has(h.id)) {
          halls.value.push(h)
        }
      })
    } catch (err: any) {
      console.error(`Failed to fetch halls for cinema ${cinemaId}:`, err)
    }
  }

  const fetchShowtimes = async (movieId: string) => {
    isLoading.value = true
    error.value = null
    showtimes.value = []
    
    try {
      const response = await api.get(`/movies/${movieId}/showtimes`)
      showtimes.value = response.data
      
      // After fetching showtimes, fetch cinemas if we haven't
      if (cinemas.value.length === 0) {
        await fetchCinemas()
      }
      
      // For each unique cinema in the showtimes, fetch its halls
      const uniqueCinemaIds = [...new Set(showtimes.value.map(s => s.cinema_id))]
      await Promise.all(uniqueCinemaIds.map(cid => fetchHallsByCinema(cid)))
      
    } catch (err: any) {
      console.error('Failed to fetch showtimes:', err)
      // We don't set error.value here because movie details might have loaded successfully
      // Just leave showtimes empty
    } finally {
      isLoading.value = false
    }
  }

  return {
    movies,
    currentMovie,
    showtimes,
    cinemas,
    halls,
    isLoading,
    error,
    fetchMovies,
    fetchMovieById,
    fetchShowtimes
  }
})
