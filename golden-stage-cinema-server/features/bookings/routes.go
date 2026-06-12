package bookings

import "github.com/gin-gonic/gin"

// BookingRoutes ทำหน้าที่จัดการ Route ที่เกี่ยวกับการจองที่นั่ง
func BookingRoutes(router *gin.Engine) {
	router.POST("/api/bookings/lock", LockSeat)
	router.POST("/api/bookings/confirm", ConfirmBooking)
}
