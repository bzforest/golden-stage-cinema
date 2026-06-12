package movies

import "github.com/gin-gonic/gin"

// MovieRoutes ทำหน้าที่จัดการผูก Route ที่เกี่ยวกับ Movies เข้ากับ Controller
func MovieRoutes(router *gin.Engine) {
	router.GET("/api/movies", GetMovies)
}
