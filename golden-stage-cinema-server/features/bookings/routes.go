package bookings

import (
	"github.com/gin-gonic/gin"
	"golden-stage-cinema-server/middlewares"
)

// BookingRoutes ทำหน้าที่จัดการ Route ที่เกี่ยวกับการจองที่นั่ง
func BookingRoutes(router *gin.Engine) {
	// จัดกลุ่ม Route แล้วเอา Middleware ขวางไว้ก่อน
	bookingGroup := router.Group("/api/bookings")
	bookingGroup.Use(middlewares.FirebaseAuthMiddleware())
	{
		bookingGroup.POST("/lock", LockSeat)
		bookingGroup.POST("/confirm", ConfirmBooking)
		bookingGroup.GET("/me", GetUserBookings)
	}
}
