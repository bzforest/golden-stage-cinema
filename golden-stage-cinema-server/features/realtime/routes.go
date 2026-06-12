package realtime

import "github.com/gin-gonic/gin"

// RealtimeRoutes ผูก Route ของระบบ WebSocket
func RealtimeRoutes(router *gin.Engine) {
	// ใช้ GET เพราะ WebSocket อาศัย HTTP GET ในการอัปเกรดตอนเริ่มแรก
	router.GET("/ws/seats/:showtime_id", HandleConnections)
}
