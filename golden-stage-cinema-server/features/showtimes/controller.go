package showtimes

import (
	"context"
	"log"
	"net/http"
	"time"

	"golden-stage-cinema-server/config"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

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

	// ค้นหาเก้าอี้ทั้งหมดที่เป็นของรอบฉาย (showtime_id) นี้
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

	c.JSON(http.StatusOK, seats)
}
