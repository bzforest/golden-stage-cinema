package bookings

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"golden-stage-cinema-server/config"
)

// StartBookingWorker เริ่มทำงาน Goroutine เพื่อรอรับ Message จาก RabbitMQ
func StartBookingWorker() {
	if config.RabbitChannel == nil {
		log.Println("RabbitMQ is not connected, skipping worker start.")
		return
	}

	config.RabbitChannel.ExchangeDeclare("seat_updates_ex", "fanout", true, false, false, false, nil)
	q, err := config.RabbitChannel.QueueDeclare("booking_worker_queue", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to declare worker queue: %v", err)
	}

	err = config.RabbitChannel.QueueBind(q.Name, "", "seat_updates_ex", false, nil)
	if err != nil {
		log.Fatalf("Failed to bind worker queue: %v", err)
	}

	msgs, err := config.RabbitChannel.Consume(
		q.Name,         // queue name
		"",             // consumer
		true,           // auto-ack
		false,          // exclusive
		false,          // no-local
		false,          // no-wait
		nil,            // args
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %v", err)
	}

	// ใช้ Goroutine เพื่อให้วนลูปรับข้อความแบบ Asynchronous
	go func() {
		for msg := range msgs {
			// แกะ Payload ออกมาเพื่อกรองเอาเฉพาะการจองที่สำเร็จ
			var payload map[string]string
			if err := json.Unmarshal(msg.Body, &payload); err == nil {
				if payload["status"] != "CONFIRMED" {
					continue // กรองเอาเฉพาะข้อมูลการจองสำเร็จเท่านั้น
				}
			}

			log.Printf("[Worker] Received CONFIRMED message from RabbitMQ: %s\n", string(msg.Body))

			// บันทึก Audit Log ลง MongoDB
			auditLog := AuditLog{
				Action:     "BOOKING_CONFIRMED",
				Details:    string(msg.Body),
				Timestamp:  time.Now(),
				UID:        payload["user_id"],
				ShowtimeID: payload["showtime_id"],
				SeatNumber: payload["seat_number"],
			}

			collection := config.GetCollection("audit_logs")
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			_, dbErr := collection.InsertOne(ctx, auditLog)
			cancel() // เรียก cancel ทันทีหลังจากใช้งานเสร็จ

			if dbErr != nil {
				log.Printf("[Worker] Failed to save AuditLog: %v\n", dbErr)
			} else {
				log.Println("[Worker] AuditLog saved successfully.")
			}

			// จำลองการส่งอีเมล
			log.Printf("[Worker] Sending ticket confirmation email for booking data: %s\n", string(msg.Body))
		}
	}()

	log.Println("[Worker] RabbitMQ Consumer is running and waiting for messages...")
}
