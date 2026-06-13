package bookings

import (
	"context"
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

	msgs, err := config.RabbitChannel.Consume(
		"seat_updates", // queue name
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
			log.Printf("[Worker] Received message from RabbitMQ: %s\n", string(msg.Body))

			// บันทึก Audit Log ลง MongoDB
			auditLog := AuditLog{
				Action:    "BOOKING_CONFIRMED",
				Details:   string(msg.Body),
				Timestamp: time.Now(),
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
