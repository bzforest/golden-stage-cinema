package realtime

import (
	"encoding/json"
	"log"
	"net/http"

	"golden-stage-cinema-server/config"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// สร้างตัวแปร upgrader เพื่อเปลี่ยน HTTP เป็น WebSocket
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // อนุญาตให้เชื่อมต่อจากโดเมนอื่นได้ (สำหรับตอนพัฒนา)
	},
}

// SeatUpdatePayload โครงสร้างข้อมูลที่จะแกะจาก JSON ที่ส่งมาจาก RabbitMQ
type SeatUpdatePayload struct {
	ShowtimeID string `json:"showtime_id"`
	SeatNumber string `json:"seat_number"`
	Status     string `json:"status"`
}

// HandleConnections ฟังก์ชันสำหรับรับ Connection และอัปเดตข้อมูล Real-time ให้ Client
func HandleConnections(c *gin.Context) {
	showtimeID := c.Param("showtime_id")
	if showtimeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "showtime_id is required"})
		return
	}

	// 1. อัปเกรดจาก HTTP เป็น WebSocket
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Error upgrading to websocket:", err)
		return
	}
	defer ws.Close()

	// 2. ดักฟัง (Consume) ข้อความจาก RabbitMQ คิว "seat_updates"
	msgs, err := config.RabbitChannel.Consume(
		"seat_updates", // queue name
		"",             // consumer name (ปล่อยว่างได้)
		true,           // auto-ack (ให้ลบข้อความออกจากคิวอัตโนมัติเมื่ออ่านแล้ว)
		false,          // exclusive
		false,          // no-local
		false,          // no-wait
		nil,            // args
	)
	if err != nil {
		log.Println("Failed to register a consumer:", err)
		return
	}

	// 3. ใช้ Channel เพื่อส่งสัญญาณบอก Goroutine ให้เลิกทำงานเมื่อ Client ปิดเบราว์เซอร์
	done := make(chan struct{})
	defer close(done) // จะถูกเรียกเมื่อหลุดออกจาก for-loop ด้านล่าง (Client หลุด)

	// สร้าง Goroutine เพื่อทำหน้าที่รับข้อมูลจาก RabbitMQ แบบคู่ขนาน
	go func() {
		for {
			select {
			case d, ok := <-msgs:
				// กรณีที่ RabbitMQ ปิดลงหรือมีปัญหา
				if !ok {
					log.Println("RabbitMQ channel closed")
					return 
				}

				// แปลง JSON ให้กลายเป็น Struct
				var update SeatUpdatePayload
				if err := json.Unmarshal(d.Body, &update); err != nil {
					log.Println("Error decoding message:", err)
					continue
				}

				// ถ้าข้อความที่เด้งมา เป็นของรอบฉายเดียวกับที่หน้าเว็บนี้กำลังดูอยู่
				if update.ShowtimeID == showtimeID {
					// 4. ให้ส่งข้อมูลกลับไปที่หน้าเว็บทันที (WriteJSON)
					if err := ws.WriteJSON(update); err != nil {
						log.Println("Error writing message to websocket:", err)
						return // เลิกทำงาน Goroutine ถ้ายิงกลับไปไม่สำเร็จ (เช่น เน็ตลูกค้าหลุด)
					}
				}
			case <-done:
				// ถ้า Client ปิดเบราว์เซอร์ ให้เลิกดักฟัง
				return
			}
		}
	}()

	// ลูปหลักนี้เอาไว้หน่วงฟังก์ชันไม่ให้จบลงทันที 
	// และเอาไว้ดักจับตอนที่ผู้ใช้ปิดหน้าจอหรือหลุด (มันจะเกิด err แล้วเบรกออกจากลูป ไปทำ defer close)
	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			log.Println("Client disconnected:", err)
			break
		}
	}
}
