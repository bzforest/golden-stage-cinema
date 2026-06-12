package bookings

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"golden-stage-cinema-server/config"

	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LockSeatRequest struct {
	ShowtimeID string `json:"showtime_id" binding:"required"`
	SeatNumber string `json:"seat_number" binding:"required"`
	// เอา UserID ออกจาก JSON Request ป้องกันคนส่งมาปลอมแปลง
}

// LockSeat ฟังก์ชันสำหรับจองและล็อกที่นั่ง
func LockSeat(c *gin.Context) {
	var req LockSeatRequest

	// แกะ JSON ใส่ตัวแปร req
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// ดึงไอดีผู้ใช้ตัวจริงจาก Middleware
	userID := c.MustGet("user_id").(string)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// สร้าง Redis Key
	key := fmt.Sprintf("lock:seat:%s:%s", req.ShowtimeID, req.SeatNumber)

	// ใช้ SETNX (Set if Not eXists) เพื่อล็อกที่นั่งด้วย userID ตั้งเวลา 5 นาที
	success, err := config.RedisClient.SetNX(ctx, key, userID, 5*time.Minute).Result()
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

type ConfirmBookingRequest struct {
	ShowtimeID string `json:"showtime_id" binding:"required"`
	SeatNumber string `json:"seat_number" binding:"required"`
}

// ConfirmBooking ฟังก์ชันยืนยันการจอง บันทึกลง Mongo และส่ง Event เข้า RabbitMQ
func ConfirmBooking(c *gin.Context) {
	var req ConfirmBookingRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// ดึงไอดีผู้ใช้ตัวจริงจาก Middleware
	userID := c.MustGet("user_id").(string)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	showtimeID, err := primitive.ObjectIDFromHex(req.ShowtimeID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid showtime_id format"})
		return
	}

	// 1. บันทึกข้อมูลการจองลง MongoDB
	collection := config.GetCollection("bookings")
	booking := Booking{
		ShowtimeID: showtimeID,
		SeatNumber: req.SeatNumber,
		UserID:     userID,
		Status:     "CONFIRMED",
		CreatedAt:  time.Now(),
	}

	_, err = collection.InsertOne(ctx, booking)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save booking to database"})
		return
	}

	// 2. ประกาศคิว (Queue) ใน RabbitMQ
	q, err := config.RabbitChannel.QueueDeclare(
		"seat_updates", // name
		true,           // durable
		false,          // delete when unused
		false,          // exclusive
		false,          // no-wait
		nil,            // arguments
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to declare a queue"})
		return
	}

	// สร้าง Payload JSON
	messageBody, _ := json.Marshal(map[string]string{
		"showtime_id": req.ShowtimeID,
		"seat_number": req.SeatNumber,
		"status":      "CONFIRMED",
	})

	// 3. ส่งข้อความ (Publish) เข้า RabbitMQ
	err = config.RabbitChannel.PublishWithContext(ctx,
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        messageBody,
		})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish message"})
		return
	}

	// ตอบกลับ 200 OK
	c.JSON(http.StatusOK, gin.H{"message": "Booking confirmed and message queued"})
}
