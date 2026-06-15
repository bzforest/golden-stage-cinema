package showtimes

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"golden-stage-cinema-server/config"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetShowtimeByID ดึงข้อมูลรอบฉาย 1 รอบ
func GetShowtimeByID(c *gin.Context) {
	showtimeIDStr := c.Param("showtime_id")
	if showtimeIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "showtime_id is required"})
		return
	}

	showtimeID, err := primitive.ObjectIDFromHex(showtimeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid showtime_id format"})
		return
	}

	collection := config.GetCollection("showtimes")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var showtime Showtime
	err = collection.FindOne(ctx, bson.M{"_id": showtimeID}).Decode(&showtime)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Showtime not found"})
		return
	}

	c.JSON(http.StatusOK, showtime)
}

// GetShowtimesByMovie ดึงข้อมูลรอบฉายทั้งหมดของหนังเรื่องหนึ่ง
func GetShowtimesByMovie(c *gin.Context) {
	movieIDStr := c.Param("movie_id")
	if movieIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "movie_id is required"})
		return
	}

	movieID, err := primitive.ObjectIDFromHex(movieIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid movie_id format"})
		return
	}

	collection := config.GetCollection("showtimes")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// ค้นหารอบฉายที่มี movie_id ตรงกับที่ส่งมา
	cursor, err := collection.Find(ctx, bson.M{"movie_id": movieID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch showtimes"})
		return
	}
	defer cursor.Close(ctx)

	var showtimes []Showtime
	if err = cursor.All(ctx, &showtimes); err != nil {
		log.Printf("Decode showtimes error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode showtimes"})
		return
	}

	// ถ้าหาไม่เจอ ให้ return 404 Not Found
	if len(showtimes) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No showtimes found for this movie"})
		return
	}

	c.JSON(http.StatusOK, showtimes)
}

// GetSeatsByShowtime ดึงข้อมูลผังเก้าอี้และสถานะของแต่ละที่นั่งในรอบฉายหนึ่งๆ
func GetSeatsByShowtime(c *gin.Context) {
	showtimeIDStr := c.Param("showtime_id")
	userID := c.Query("user_id") // รับ user_id จาก query params สำหรับเช็กว่าเราจองไว้เองไหม

	if showtimeIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "showtime_id is required"})
		return
	}

	showtimeID, err := primitive.ObjectIDFromHex(showtimeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid showtime_id format"})
		return
	}

	collection := config.GetCollection("showtime_seats")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. ค้นหาเก้าอี้ทั้งหมดที่เป็นของรอบฉาย (showtime_id) นี้ จาก MongoDB
	cursor, err := collection.Find(ctx, bson.M{"showtime_id": showtimeID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch seats"})
		return
	}
	defer cursor.Close(ctx)

	var seats []ShowtimeSeat
	if err = cursor.All(ctx, &seats); err != nil {
		log.Printf("Decode seats error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode seats"})
		return
	}

	// ถ้าหาไม่เจอ ให้ return 404 Not Found
	if len(seats) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No seats found for this showtime"})
		return
	}

	// 1.5 ดึงประวัติการจองจาก booked_seats ของ showtimes (Single Source of Truth แบบ Array)
	showtimesCollection := config.GetCollection("showtimes")
	var showtime Showtime
	err = showtimesCollection.FindOne(ctx, bson.M{"_id": showtimeID}).Decode(&showtime)
	if err == nil {
		bookedMap := make(map[string]bool)
		for _, seat := range showtime.BookedSeats {
			bookedMap[seat] = true
		}
		for i := range seats {
			// บังคับเปลี่ยนสถานะตาม booked_seats
			if bookedMap[seats[i].SeatNumber] {
				seats[i].Status = "BOOKED"
			} else {
				seats[i].Status = "AVAILABLE"
			}
		}
	}

	// 2. ดึงสถานะการล็อก (LOCKED) จาก Redis เพื่อมา Merge กับข้อมูลใน Database
	keys := make([]string, 0, len(seats))
	for _, seat := range seats {
		keys = append(keys, fmt.Sprintf("lock:seat:%s:%s", showtimeIDStr, seat.SeatNumber))
	}

	// MGet ดึงข้อมูลหลายๆ key พร้อมกันในครั้งเดียว
	vals, err := config.RedisClient.MGet(ctx, keys...).Result()
	if err == nil {
		for i, val := range vals {
			if val != nil { // มีข้อมูลล็อกใน Redis
				lockerIDRaw := val.(string)
				// Clean possible quotation marks / whitespace
				lockerID := strings.TrimSpace(strings.Trim(lockerIDRaw, "\""))
				userIDClean := strings.TrimSpace(strings.Trim(userID, "\""))
				log.Printf("[GetSeats] seat %s lock owner: %s, requester: %s", seats[i].SeatNumber, lockerID, userIDClean)
				// เปลี่ยนสถานะเฉพาะเก้าอี้ที่ยังไม่มีคนซื้อ (AVAILABLE)
				if seats[i].Status == "AVAILABLE" {
					if lockerID == userIDClean && userIDClean != "" {
						seats[i].Status = "SELECTED" // ล็อกโดยตัวเราเอง
					} else {
						seats[i].Status = "LOCKED"   // ล็อกโดยคนอื่น
					}
				}
			}
		}
	} else {
		log.Printf("Redis MGet error: %v", err)
	}

	c.JSON(http.StatusOK, seats)
}
