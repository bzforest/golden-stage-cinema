//go:build ignore

package main

import (
	"context"
	"log"

	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

func main() {
	// 1. ระบุ UID ของผู้ใช้ที่ต้องการให้สิทธิ์ Admin
	// สามารถไปก๊อปปี้ UID ได้จากหน้า Authentication ใน Firebase Console
	targetUID := "OQ2vt82VeNfhhQ5MImjkqyBcCeM2"

	if len(targetUID) < 10 {
		log.Fatal("Error: Please provide a valid targetUID at the top of the script.")
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
