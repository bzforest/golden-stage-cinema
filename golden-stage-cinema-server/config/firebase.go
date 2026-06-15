package config

import (
	"context"
	"log"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"google.golang.org/api/option"
)

// FirebaseAuth ตัวแปร Global สำหรับเรียกใช้งานฟังก์ชันของ Firebase Auth
var FirebaseAuth *auth.Client

// InitFirebase ทำหน้าที่เชื่อมต่อและตั้งค่า Firebase ด้วย Service Account Key
func InitFirebase() {
	// ใช้ context.Background() สำหรับการตั้งค่าเริ่มต้นแบบไม่มี Timeout
	ctx := context.Background()

	// ชี้ไปยังไฟล์ Service Account Key ที่เราดาวน์โหลดมาจาก Firebase Console
	// (ต้องมีไฟล์นี้อยู่ในระดับเดียวกับ main.go หรือแก้ Path ให้ตรงกัน)
	opt := option.WithCredentialsFile("firebase-service-account.json")

	// สั่ง Initialize Firebase App
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatalf("Error initializing firebase app: %v", err)
	}

	// สร้าง Client สำหรับเข้าถึงระบบ Auth ของ Firebase
	client, err := app.Auth(ctx)
	if err != nil {
		log.Fatalf("Error getting firebase auth client: %v", err)
	}

	FirebaseAuth = client
	log.Println("Successfully connected to Firebase Auth")
}
