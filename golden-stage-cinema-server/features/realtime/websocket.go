package realtime

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

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
	UserID     string `json:"user_id,omitempty"`
}

// Hub จัดการ WebSocket Connections ทั้งหมดโดยแยกตาม Showtime
// แทนที่จะให้แต่ละ Client สร้าง Consumer เอง (ซึ่งทำให้ RabbitMQ Round-Robin เลือกส่งให้คนเดียว)
// เราจะมี Consumer ตัวเดียวกลาง แล้วกระจาย (Fan-out) ให้ทุก Client ที่ Subscribe อยู่
type Hub struct {
	mu      sync.RWMutex
	clients map[string]map[*websocket.Conn]bool // map[showtimeID] -> set of connections
}

var hub = &Hub{
	clients: make(map[string]map[*websocket.Conn]bool),
}

// Register เพิ่ม Client เข้ามาในห้องของ showtime นั้น
func (h *Hub) Register(showtimeID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[showtimeID] == nil {
		h.clients[showtimeID] = make(map[*websocket.Conn]bool)
	}
	h.clients[showtimeID][conn] = true
	log.Printf("[Hub] Client registered for showtime %s (total: %d)", showtimeID, len(h.clients[showtimeID]))
}

// Unregister ลบ Client ออกจากห้อง
func (h *Hub) Unregister(showtimeID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if conns, ok := h.clients[showtimeID]; ok {
		delete(conns, conn)
		if len(conns) == 0 {
			delete(h.clients, showtimeID)
		}
		log.Printf("[Hub] Client unregistered from showtime %s (remaining: %d)", showtimeID, len(conns))
	}
}

// Broadcast ส่งข้อความไปยังทุก Client ที่อยู่ในห้อง showtime เดียวกัน
func (h *Hub) Broadcast(showtimeID string, payload SeatUpdatePayload) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	conns, ok := h.clients[showtimeID]
	if !ok {
		return
	}
	for conn := range conns {
		if err := conn.WriteJSON(payload); err != nil {
			log.Println("[Hub] Error writing to client, closing:", err)
			conn.Close()
			// Note: จะถูก Unregister โดย HandleConnections loop เมื่อ ReadMessage fail
		}
	}
}

// StartConsumer เป็น Goroutine กลางที่ดักฟัง RabbitMQ แล้ว Fan-out ให้ทุก Client ผ่าน Hub
// ฟังก์ชันนี้ควรถูกเรียกครั้งเดียวตอน Server เริ่มต้น
func StartConsumer() {
	// ประกาศ Exchange แบบ Fanout
	err := config.RabbitChannel.ExchangeDeclare("seat_updates_ex", "fanout", true, false, false, false, nil)
	if err != nil {
		log.Fatal("Failed to declare fanout exchange:", err)
	}

	// สร้าง Queue แบบ Exclusive (ถูกทำลายเมื่อหลุดการเชื่อมต่อ/ปิดเซิร์ฟเวอร์) 
	q, err := config.RabbitChannel.QueueDeclare("", false, false, true, false, nil)
	if err != nil {
		log.Fatal("Failed to declare exclusive queue:", err)
	}

	// ผูก Queue เข้ากับ Exchange
	err = config.RabbitChannel.QueueBind(q.Name, "", "seat_updates_ex", false, nil)
	if err != nil {
		log.Fatal("Failed to bind queue:", err)
	}

	msgs, err := config.RabbitChannel.Consume(
		q.Name,         // queue name (ใช้คิวผีที่สร้างขึ้นมา)
		"",             // consumer name
		true,           // auto-ack
		false,          // exclusive
		false,          // no-local
		false,          // no-wait
		nil,            // args
	)
	if err != nil {
		log.Fatal("Failed to register seat_updates consumer:", err)
	}

	go func() {
		log.Println("[Consumer] Listening for seat_updates from RabbitMQ...")
		for d := range msgs {
			var update SeatUpdatePayload
			if err := json.Unmarshal(d.Body, &update); err != nil {
				log.Println("[Consumer] Error decoding message:", err)
				continue
			}
			log.Printf("[Consumer] Broadcasting: showtime=%s seat=%s status=%s", update.ShowtimeID, update.SeatNumber, update.Status)
			hub.Broadcast(update.ShowtimeID, update)
		}
	}()
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

	// 2. ลงทะเบียน Client เข้า Hub ของ showtime นี้
	hub.Register(showtimeID, ws)
	defer hub.Unregister(showtimeID, ws)

	// 3. ลูปหลักเพื่อ keep-alive และดักจับเวลา Client ปิดเบราว์เซอร์
	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			log.Println("Client disconnected:", err)
			break
		}
	}
}
