package config

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DB ตัวแปร Global สำหรับเรียกใช้งาน Database
var DB *mongo.Database

// ConnectDB ทำหน้าที่เชื่อมต่อกับ MongoDB
func ConnectDB() {
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		log.Fatal("MONGO_URI environment variable not set")
	}

	dbName := os.Getenv("MONGO_DB_NAME")
	if dbName == "" {
		log.Fatal("MONGO_DB_NAME environment variable not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// ตั้งค่า Option การเชื่อมต่อ
	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB: ", err)
	}

	// ทดสอบ Ping ไปที่ Database ว่าต่อติดจริงๆ
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("Failed to ping MongoDB: ", err)
	}

	DB = client.Database(dbName)
	log.Println("Successfully connected to MongoDB")
}

// GetCollection ดึง Collection ออกมาใช้งาน
func GetCollection(collectionName string) *mongo.Collection {
	return DB.Collection(collectionName)
}
