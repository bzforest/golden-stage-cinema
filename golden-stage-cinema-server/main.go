package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golden-stage-cinema-server/config"
	"golden-stage-cinema-server/features/bookings"
	"golden-stage-cinema-server/features/cinemas"
	"golden-stage-cinema-server/features/movies"
	"golden-stage-cinema-server/features/realtime"
	"golden-stage-cinema-server/features/showtimes"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// โหลดไฟล์ .env จากระดับโฟลเดอร์นอกสุด
	err := godotenv.Load("../.env")
	if err != nil {
		log.Println("Warning: Error loading .env file from ../.env")
	}

	// ตั้งค่า Firebase Auth
	config.InitFirebase()

	// เชื่อมต่อ MongoDB
	config.ConnectDB()

	// เชื่อมต่อ Redis
	config.ConnectRedis()

	// เชื่อมต่อ RabbitMQ
	config.ConnectRabbitMQ()

	// ดึงค่า PORT จาก Environment Variable ถ้าไม่มีให้ใช้ค่าเริ่มต้นเป็น 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// สร้าง Gin Router
	r := gin.Default()

	// ตั้งค่า CORS Middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// ติดตั้ง Routes ระบบต่างๆ
	movies.MovieRoutes(r)
	cinemas.CinemaRoutes(r)
	bookings.BookingRoutes(r)
	realtime.RealtimeRoutes(r)
	showtimes.ShowtimeRoutes(r)

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

	// เริ่มการทำงานของ Worker
	bookings.StartBookingWorker()
	// เริ่มการทำงานของ Redis Timeout Listener
	bookings.StartRedisTimeoutListener()

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// 2. รัน Server ใน Goroutine เพื่อไม่ให้บล็อกการรอสัญญาณปิด
	go func() {
		log.Printf("Starting Server on Port %s\n", port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// 1. สร้าง Channel รอรับสัญญาณ OS (SIGINT = กด Ctrl+C, SIGTERM = สั่งปิดโปรเซส)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// โค้ดจะหยุดรอที่บรรทัดนี้จนกว่าจะมีสัญญาณเข้ามาใน Channel
	<-quit
	log.Println("Shutting down server...")

	// 3. กำหนดเวลา Timeout ไว้ 5 วินาที เพื่อให้ Request ที่ค้างอยู่ทำงานให้เสร็จ
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}

	// 4. แสดง Log จังหวะที่กำลังจะปิดตัวลง
	log.Println("Server exiting")
}
