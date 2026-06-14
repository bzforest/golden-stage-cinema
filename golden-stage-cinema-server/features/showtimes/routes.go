package showtimes

import "github.com/gin-gonic/gin"

// ShowtimeRoutes ผูก Route ที่เกี่ยวกับรอบฉายและเก้าอี้
func ShowtimeRoutes(router *gin.Engine) {
	// กลุ่มนี้เป็น Public API ให้ทุกคนเข้ามาดูรอบฉายและผังที่นั่งได้ โดยไม่ต้องล็อกอิน
	publicGroup := router.Group("/api")
	{
		publicGroup.GET("/showtimes/:showtime_id", GetShowtimeByID)
		publicGroup.GET("/movies/:movie_id/showtimes", GetShowtimesByMovie)
		publicGroup.GET("/showtimes/:showtime_id/seats", GetSeatsByShowtime)
	}
}
