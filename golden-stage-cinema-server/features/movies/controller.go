package movies

import (
	"context"
	"net/http"
	"time"

	"golden-stage-cinema-server/config"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

// GetMovies รับ Request เพื่อดึงข้อมูลหนังทั้งหมดจาก MongoDB
func GetMovies(c *gin.Context) {
	// กำหนด Timeout เพื่อป้องกัน Request ค้าง
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// เรียกใช้งาน Collection 'movies'
	collection := config.GetCollection("movies")

	// สั่งค้นหาหนังทั้งหมด (เงื่อนไขว่างเปล่า bson.M{})
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch movies from database"})
		return
	}
	defer cursor.Close(ctx)

	var moviesList []Movie
	// วนลูปอ่านข้อมูลมาแปลงใส่ใน Slice ของ moviesList
	if err = cursor.All(ctx, &moviesList); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode movies data"})
		return
	}

	// ถ้าไม่มีหนังเลย ให้ส่งกลับเป็น array เปล่าแทนที่จะเป็น null
	if moviesList == nil {
		moviesList = []Movie{}
	}

	// ส่งข้อมูลกลับไปเป็น JSON
	c.JSON(http.StatusOK, moviesList)
}
