package bookings

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"golden-stage-cinema-server/config"

	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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

	// 1. Rate Limiting: Max 20 requests per minute
	rateLimitKey := fmt.Sprintf("rate_limit:lock:%s", userID)
	count, err := config.RedisClient.Incr(ctx, rateLimitKey).Result()
	if err == nil && count == 1 {
		config.RedisClient.Expire(ctx, rateLimitKey, time.Minute)
	}
	if count > 20 {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests. Please slow down."})
		return
	}

	// สร้าง Redis Key
	key := fmt.Sprintf("lock:seat:%s:%s", req.ShowtimeID, req.SeatNumber)

	// ใช้ SETNX (Set if Not eXists) เพื่อล็อกที่นั่งด้วย userID ตั้งเวลา 5 นาที
	success, err := config.RedisClient.SetNX(ctx, key, userID, 5*time.Minute).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process lock request"})
		return
	}

	if success {
		// Publish LOCKED event ไปที่ RabbitMQ เพื่อให้ WebSocket broadcast ไปบอกทุก Client
		config.RabbitChannel.ExchangeDeclare("seat_updates_ex", "fanout", true, false, false, false, nil)
		messageBody, _ := json.Marshal(map[string]string{
			"showtime_id": req.ShowtimeID,
			"seat_number": req.SeatNumber,
			"status":      "LOCKED",
			"user_id":     userID,
		})
		config.RabbitChannel.PublishWithContext(ctx, "seat_updates_ex", "", false, false, amqp.Publishing{
			ContentType: "application/json",
			Body:        messageBody,
		})

		c.JSON(http.StatusOK, gin.H{"message": "Seat locked successfully"})
	} else {
		c.JSON(http.StatusConflict, gin.H{"error": "Seat is already locked by another user"})
	}
}

// UnlockSeat ฟังก์ชันสำหรับปลดล็อกที่นั่งคืนสู่ระบบ
func UnlockSeat(c *gin.Context) {
	showtimeID := c.Param("showtime_id")
	seatNumber := c.Param("seat_number")
	userID := c.MustGet("user_id").(string)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := fmt.Sprintf("lock:seat:%s:%s", showtimeID, seatNumber)
	val, err := config.RedisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		c.JSON(http.StatusOK, gin.H{"message": "Seat is already unlocked"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check seat lock status"})
		return
	}

	// Clean up strings before comparison to prevent formatting issues
	valClean := strings.Trim(val, "\" \n\r\t")
	userIDClean := strings.Trim(userID, "\" \n\r\t")

	log.Printf("[UnlockSeat] Comparing lock. Redis: '%s' (raw: %v), Request: '%s' (raw: %v)", valClean, val, userIDClean, userID)

	if valClean != userIDClean {
		c.JSON(http.StatusForbidden, gin.H{"error": fmt.Sprintf("Cannot unlock a seat locked by someone else. LockedBy: %s, Requester: %s", valClean, userIDClean)})
		return
	}

	_, err = config.RedisClient.Del(ctx, key).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unlock seat"})
		return
	}

	// ประกาศ Exchange และ Publish ให้ RabbitMQ ทราบว่าที่นั่งนี้ AVAILABLE แล้ว
	config.RabbitChannel.ExchangeDeclare("seat_updates_ex", "fanout", true, false, false, false, nil)
	messageBody, _ := json.Marshal(map[string]string{
		"showtime_id": showtimeID,
		"seat_number": seatNumber,
		"status":      "AVAILABLE",
	})
	config.RabbitChannel.PublishWithContext(ctx, "seat_updates_ex", "", false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        messageBody,
	})

	c.JSON(http.StatusOK, gin.H{"message": "Seat unlocked successfully"})
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
	cleanTokenUserID := strings.TrimSpace(strings.Trim(userID, "\""))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	showtimeID, err := primitive.ObjectIDFromHex(req.ShowtimeID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid showtime_id format"})
		return
	}

	// 1. Pre-validation: ตรวจสอบ Ownership ใน Redis สำหรับทุกที่นั่งก่อนลงมือทำจริง
	var lockKeys []string
	for _, seatNumber := range req.SeatNumbers {
		lockKeys = append(lockKeys, fmt.Sprintf("lock:seat:%s:%s", req.ShowtimeID, seatNumber))
	}

	vals, err := config.RedisClient.MGet(ctx, lockKeys...).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check lock ownership"})
		return
	}

	// เช็กว่าทุกที่นั่งเป็นของเราหรือไม่
	for i, val := range vals {
		if val == nil {
			c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("Seat lock expired or not locked for seat %s", req.SeatNumbers[i])})
			return
		}
		
		redisUserID := val.(string)
		cleanRedisUserID := strings.TrimSpace(strings.Trim(redisUserID, "\""))
		
		if cleanRedisUserID != cleanTokenUserID {
			c.JSON(http.StatusForbidden, gin.H{"error": fmt.Sprintf("You are not the owner of seat lock for seat %s", req.SeatNumbers[i])})
			return
		}
	}

	// 1.5 Atomic Reservation via UpdateOne and $nin
	showtimesCollection := config.GetCollection("showtimes")
	filter := bson.M{
		"_id":          showtimeID,
		"booked_seats": bson.M{"$nin": req.SeatNumbers},
	}
	update := bson.M{
		"$push": bson.M{"booked_seats": bson.M{"$each": req.SeatNumbers}},
	}
	updateResult, err := showtimesCollection.UpdateOne(ctx, filter, update)
	if err != nil || updateResult.ModifiedCount == 0 {
		config.RedisClient.Del(ctx, lockKeys...)
		c.JSON(http.StatusConflict, gin.H{"error": "Sorry, some seats were just taken."})
		return
	}

	// 2. Bulk Insert: บันทึกข้อมูลการจองลง MongoDB
	var bookings []interface{}
	for _, seatNumber := range req.SeatNumbers {
		bookings = append(bookings, Booking{
			ShowtimeID: showtimeID,
			SeatNumber: seatNumber,
			UserID:     userID,
			Status:     "CONFIRMED",
			CreatedAt:  time.Now(),
		})
	}

	collection := config.GetCollection("bookings")
	_, err = collection.InsertMany(ctx, bookings)
	if err != nil {
		config.RedisClient.Del(ctx, lockKeys...)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save bookings to database"})
		return
	}

	// (3. Bulk Update ถูกนำออกแล้ว เนื่องจากใช้ Single Source of Truth จากตาราง bookings)

	// 4. ลบ Lock ออกจาก Redis ทั้งหมด
	config.RedisClient.Del(ctx, lockKeys...)

	// 5. ประกาศ Exchange แบบ Fanout ใน RabbitMQ
	config.RabbitChannel.ExchangeDeclare("seat_updates_ex", "fanout", true, false, false, false, nil)

	// วนลูป Publish แจ้งสถานะของทีละเก้าอี้ เพื่อให้ฝั่ง Frontend (ที่รับทีละตัว) ทำงานได้เหมือนเดิม
	for _, seatNumber := range req.SeatNumbers {
		messageBody, _ := json.Marshal(map[string]string{
			"showtime_id": req.ShowtimeID,
			"seat_number": seatNumber,
			"status":      "BOOKED",
		})
		
		config.RabbitChannel.PublishWithContext(ctx,
			"seat_updates_ex", // exchange
			"",                // routing key
			false,             // mandatory
			false,             // immediate
			amqp.Publishing{
				ContentType: "application/json",
				Body:        messageBody,
			})
	}

	// ตอบกลับ 200 OK
	c.JSON(http.StatusOK, gin.H{"message": "Bookings confirmed and messages queued"})
}

// GetUserBookings ดึงประวัติการจองของผู้ใช้งาน
func GetUserBookings(c *gin.Context) {
	userID := c.MustGet("user_id").(string)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := config.GetCollection("bookings")

	cursor, err := collection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch bookings"})
		return
	}
	defer cursor.Close(ctx)

	var bookings []Booking
	if err = cursor.All(ctx, &bookings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode bookings"})
		return
	}

	if bookings == nil {
		bookings = []Booking{}
	}

	c.JSON(http.StatusOK, bookings)
}

// GetAdminBookings ดึงประวัติการจองทั้งหมดสำหรับผู้ดูแลระบบ พร้อมรองรับการ Filter
func GetAdminBookings(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	movieIDStr := c.Query("movie_id")
	dateStr := c.Query("date")

	collection := config.GetCollection("bookings")

	// ใช้ Aggregate Pipeline เพื่อ Join กับ showtimes
	pipeline := mongo.Pipeline{}

	// 1. $lookup โยงตาราง showtimes เพื่อให้รู้ว่า booking นี้เป็นของหนังเรื่องอะไร
	pipeline = append(pipeline, bson.D{{Key: "$lookup", Value: bson.D{
		{Key: "from", Value: "showtimes"},
		{Key: "localField", Value: "showtime_id"},
		{Key: "foreignField", Value: "_id"},
		{Key: "as", Value: "showtime_info"},
	}}})

	// 2. สร้าง Filter ($match)
	matchFilter := bson.M{}

	if movieIDStr != "" {
		if objID, err := primitive.ObjectIDFromHex(movieIDStr); err == nil {
			matchFilter["showtime_info.movie_id"] = objID
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid movie_id format"})
			return
		}
	}

	if dateStr != "" {
		if parsedDate, err := time.Parse("2006-01-02", dateStr); err == nil {
			nextDay := parsedDate.AddDate(0, 0, 1)
			matchFilter["created_at"] = bson.M{
				"$gte": parsedDate,
				"$lt":  nextDay,
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format, expected YYYY-MM-DD"})
			return
		}
	}

	if len(matchFilter) > 0 {
		pipeline = append(pipeline, bson.D{{Key: "$match", Value: matchFilter}})
	}

	// 3. $project เพื่อลบ showtime_info ออกจากผลลัพธ์สุดท้ายให้ตรงกับโครงสร้าง Booking เดิม
	pipeline = append(pipeline, bson.D{{Key: "$project", Value: bson.D{
		{Key: "showtime_info", Value: 0},
	}}})

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch admin bookings"})
		return
	}
	defer cursor.Close(ctx)

	var bookings []Booking
	if err = cursor.All(ctx, &bookings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode admin bookings"})
		return
	}

	if bookings == nil {
		bookings = []Booking{}
	}

	c.JSON(http.StatusOK, bookings)
}
