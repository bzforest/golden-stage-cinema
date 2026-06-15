package cinemas

import (
	"github.com/gin-gonic/gin"
)

// CinemaRoutes ทำหน้าที่จัดการ Route ที่เกี่ยวกับ Cinemas
func CinemaRoutes(r *gin.Engine) {
	cinemaGroup := r.Group("/api/cinemas")
	{
		cinemaGroup.GET("", GetAllCinemas)
		cinemaGroup.GET("/:cinema_id/halls", GetHallsByCinema)
	}
}
