package cinemas

import (
	"context"
	"log"
	"net/http"
	"time"

	"golden-stage-cinema-server/config"
	"golden-stage-cinema-server/features/showtimes"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetAllCinemas ดึงข้อมูลสาขาทั้งหมด
func GetAllCinemas(c *gin.Context) {
	collection := config.GetCollection("cinemas")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch cinemas"})
		return
	}
	defer cursor.Close(ctx)

	var cinemas []showtimes.Cinema
	if err = cursor.All(ctx, &cinemas); err != nil {
		log.Printf("Decode cinemas error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode cinemas"})
		return
	}

	c.JSON(http.StatusOK, cinemas)
}

// GetHallsByCinema ดึงข้อมูลโรงหนังของสาขาที่ระบุ
func GetHallsByCinema(c *gin.Context) {
	cinemaIDStr := c.Param("cinema_id")
	if cinemaIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cinema_id is required"})
		return
	}

	cinemaID, err := primitive.ObjectIDFromHex(cinemaIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cinema_id format"})
		return
	}

	collection := config.GetCollection("halls")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{"cinema_id": cinemaID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch halls"})
		return
	}
	defer cursor.Close(ctx)

	var halls []showtimes.Hall
	if err = cursor.All(ctx, &halls); err != nil {
		log.Printf("Decode halls error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode halls"})
		return
	}

	c.JSON(http.StatusOK, halls)
}
