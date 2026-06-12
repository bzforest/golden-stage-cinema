package routes

import (
	"golden-stage-cinema-server/controllers"

	"github.com/gin-gonic/gin"
)

// MovieRoutes ทำหน้าที่จัดการผูก Route ที่เกี่ยวกับ Movies เข้ากับ Controller
func MovieRoutes(router *gin.Engine) {
	router.GET("/api/movies", controllers.GetMovies)
}
