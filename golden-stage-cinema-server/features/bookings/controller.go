package bookings

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"golden-stage-cinema-server/config"

	"github.com/gin-gonic/gin"
)

// LockSeatRequest กำหนดรูปแบบของ JSON ที่ส่งมา
type LockSeatRequest struct {
	ShowtimeID string `json:"showtime_id" binding:"required"`
	SeatNumber string `json:"seat_number" binding:"required"`
	UserID     string `json:"user_id" binding:"required"`
}

// LockSeat ฟังก์ชันสำหรับจองและล็อกที่นั่ง
func LockSeat(c *gin.Context) {
	var req LockSeatRequest

	// แกะ JSON ใส่ตัวแปร req
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// สร้าง Redis Key
	key := fmt.Sprintf("lock:seat:%s:%s", req.ShowtimeID, req.SeatNumber)

	// ใช้ SETNX (Set if Not eXists) เพื่อล็อกที่นั่งด้วย user_id ตั้งเวลา 5 นาที
	success, err := config.RedisClient.SetNX(ctx, key, req.UserID, 5*time.Minute).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process lock request"})
		return
	}

	if success {
		c.JSON(http.StatusOK, gin.H{"message": "Seat locked successfully"})
	} else {
		c.JSON(http.StatusConflict, gin.H{"error": "Seat is already locked by another user"})
	}
}
