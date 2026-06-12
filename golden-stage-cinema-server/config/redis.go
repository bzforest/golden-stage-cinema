package config

import (
	"context"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

// RedisClient ตัวแปร Global สำหรับเรียกใช้งาน Redis
var RedisClient *redis.Client

// ConnectRedis ทำหน้าที่เชื่อมต่อกับ Redis
func ConnectRedis() {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379" // ค่าเริ่มต้นเผื่อไม่ได้ตั้งไว้
	}

	// สร้าง Client โดยอ่านค่า Address และกำหนดรหัสผ่านเป็นค่าว่าง
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "", // ตามที่ระบุไว้ให้ตั้งเป็นค่าว่างไปก่อน
		DB:       0,  // ใช้ default DB
	})

	// ทดสอบ Ping เชื่อมต่อ
	ctx := context.Background()
	_, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatal("Failed to connect to Redis: ", err)
	}

	log.Println("Successfully connected to Redis")
}
