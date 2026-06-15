//go:build ignore

package main

import (
	"context"
	"log"
	"os"

	firebase "firebase.google.com/go"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

func main() {
	// โหลดไฟล์ .env
	err := godotenv.Load(".env")
	if err != nil {
		// ลองโหลดจาก ../.env เผื่อรันจากในโฟลเดอร์ scripts
		godotenv.Load("../.env")
	}

	// 1. ดึง UID จาก Environment Variable หรือ Argument
	targetUID := os.Getenv("ADMIN_UID")
	
	// รองรับการรันแบบส่งค่าผ่าน Argument (เช่น go run scripts/set_admin.go UID123)
	if len(os.Args) > 1 {
		targetUID = os.Args[1]
	}

	if len(targetUID) < 10 {
		log.Fatal("Error: Please provide a valid targetUID in .env (ADMIN_UID=...) or pass it as an argument (go run set_admin.go <UID>).")
	}

	ctx := context.Background()

	// 2. ตั้งค่าการเชื่อมต่อ Firebase Admin SDK
	// (ใช้ Path เดียวกับที่รัน Backend คือโฟลเดอร์ golden-stage-cinema-server)
	opt := option.WithCredentialsFile("firebase-service-account.json")

	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatalf("Error initializing firebase app: %v", err)
	}

	authClient, err := app.Auth(ctx)
	if err != nil {
		log.Fatalf("Error getting firebase auth client: %v", err)
	}

	// 3. กำหนด Custom Claims สำหรับ Admin
	claims := map[string]interface{}{
		"role": "admin",
	}

	// 4. ส่งคำสั่งบันทึก Claims ลงไปที่ผู้ใช้นั้น
	err = authClient.SetCustomUserClaims(ctx, targetUID, claims)
	if err != nil {
		log.Fatalf("Error setting custom claims for user %s: %v", targetUID, err)
	}

	log.Printf("✅ Successfully granted admin privileges to user: %s\n", targetUID)
}
