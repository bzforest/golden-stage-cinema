package config

import (
	"log"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitChannel ตัวแปร Global สำหรับเรียกใช้งาน RabbitMQ Channel
var RabbitChannel *amqp.Channel

// ConnectRabbitMQ ทำหน้าที่เชื่อมต่อกับ RabbitMQ
func ConnectRabbitMQ() {
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@localhost:5672/" // ค่าเริ่มต้น
	}

	// เชื่อมต่อไปที่ RabbitMQ Server
	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ: ", err)
	}

	// สร้าง Channel (ช่องทางการสื่อสารส่งข้อมูล)
	ch, err := conn.Channel()
	if err != nil {
		log.Fatal("Failed to open a channel: ", err)
	}

	RabbitChannel = ch
	log.Println("Successfully connected to RabbitMQ")
}
