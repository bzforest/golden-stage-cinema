package main

import (
	"log"
	"net/http"
	"os"

	"golden-stage-cinema-server/config"
	"golden-stage-cinema-server/features/bookings"
	"golden-stage-cinema-server/features/movies"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// โหลดไฟล์ .env จากระดับโฟลเดอร์นอกสุด
	err := godotenv.Load("../.env")
	if err != nil {
		log.Println("Warning: Error loading .env file from ../.env")
	}

	// เชื่อมต่อ MongoDB
	config.ConnectDB()

	// เชื่อมต่อ Redis
	config.ConnectRedis()

	// ดึงค่า PORT จาก Environment Variable ถ้าไม่มีให้ใช้ค่าเริ่มต้นเป็น 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// สร้าง Gin Router
	r := gin.Default()

	// ติดตั้ง Routes ระบบต่างๆ
	movies.MovieRoutes(r)
	bookings.BookingRoutes(r)

	// สร้าง Group สำหรับ API route
	api := r.Group("/api")
	{
		// สร้าง Health Check Route ทดสอบระบบ
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status": "ok",
			})
		})
	}

	log.Printf("Starting Server on Port %s", port)
	// รัน Server
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}
