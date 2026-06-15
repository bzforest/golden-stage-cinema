package bookings

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"golden-stage-cinema-server/config"

	amqp "github.com/rabbitmq/amqp091-go"
)

// StartRedisTimeoutListener ทำหน้าที่คอยดักฟัง Event การหมดอายุ (Expired) ของ Redis Key
// เพื่อปลดล็อกที่นั่ง (AVAILABLE) และแจงสถานะผ่าน WebSocket อัตโนมัติ
func StartRedisTimeoutListener() {
	if config.RedisClient == nil {
		log.Println("Redis is not connected, skipping timeout listener start.")
		return
	}

	ctx := context.Background()

	// Subscribe ดักฟัง Event Key Expired ใน Redis DB 0
	pubsub := config.RedisClient.Subscribe(ctx, "__keyevent@0__:expired")
	
	// ใช้ Goroutine แยกการทำงานแบบ Asynchronous
	go func() {
		defer pubsub.Close()
		ch := pubsub.Channel()

		for msg := range ch {
			key := msg.Payload // รูปแบบเช่น lock:seat:ST123:A1

			// เช็กว่าใช่คีย์ของการจองที่นั่งที่หมดอายุหรือไม่
			if strings.HasPrefix(key, "lock:seat:") {
				log.Printf("⏳ [Timeout Listener] Seat lock expired: %s\n", key)

				// สกัดเอาค่า showtime_id และ seat_number
				parts := strings.Split(key, ":")
				if len(parts) != 4 {
					continue
				}
				showtimeID := parts[2]
				seatNumber := parts[3]

				// 1. บันทึก Audit Log ลง MongoDB แบบ Asynchronous
				go func(sID, seat string) {
					collection := config.GetCollection("audit_logs")
					logCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()

					details := map[string]string{
						"showtime_id": sID,
						"seat_number": seat,
						"reason":      "Redis Lock Timeout (5 mins)",
					}
					detailsBytes, _ := json.Marshal(details)

					auditLog := AuditLog{
						Action:     "BOOKING_TIMEOUT",
						Details:    string(detailsBytes),
						Timestamp:  time.Now(),
						ShowtimeID: sID,
						SeatNumber: seat,
					}
					
					_, err := collection.InsertOne(logCtx, auditLog)
					if err != nil {
						log.Printf("❌ [Timeout Listener] Failed to save AuditLog: %v\n", err)
					}
				}(showtimeID, seatNumber)

				// 2. Broadcast สถานะกลับไปหา Frontend ผ่าน WebSocket
				if config.RabbitChannel != nil {
					config.RabbitChannel.ExchangeDeclare("seat_updates_ex", "fanout", true, false, false, false, nil)
					messageBody, _ := json.Marshal(map[string]string{
						"showtime_id": showtimeID,
						"seat_number": seatNumber,
						"status":      "AVAILABLE",
					})

					pubCtx, cancelPub := context.WithTimeout(context.Background(), 5*time.Second)
					config.RabbitChannel.PublishWithContext(pubCtx, "seat_updates_ex", "", false, false, amqp.Publishing{
						ContentType: "application/json",
						Body:        messageBody,
					})
					cancelPub()
				}
			}
		}
	}()

	log.Println("🔍 [Timeout Listener] Redis Keyspace Notification subscriber is running...")
}
