package middlewares

import (
	"context"
	"net/http"
	"strings"

	"golden-stage-cinema-server/config"

	"github.com/gin-gonic/gin"
)

// FirebaseAuthMiddleware ดักจับ Request เพื่อตรวจสอบ Token ของ Firebase ก่อนอนุญาตให้เข้าถึง API
func FirebaseAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. ดึงค่าจาก Header ที่ชื่อว่า "Authorization"
		authHeader := c.GetHeader("Authorization")

		// 2. ตรวจสอบว่ามีการส่งมา และมีคำว่า "Bearer " นำหน้าหรือไม่
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: missing or invalid token"})
			c.Abort() // สั่งหยุด ไม่ให้ทำ Controller ถัดไป
			return
		}

		// 3. ตัดคำว่า "Bearer " ออก (7 ตัวอักษร) เพื่อเอาเฉพาะก้อน Token จริงๆ
		tokenString := authHeader[7:]

		// 4. ให้ Firebase Auth เช็กว่า Token ถูกต้องและยังไม่หมดอายุใช่มั้ย
		token, err := config.FirebaseAuth.VerifyIDToken(context.Background(), tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: invalid or expired token"})
			c.Abort()
			return
		}

		// 5. ถ้าถูกต้อง ดึงค่า UID ของผู้ใช้มาฝังไว้ใน Context 
		// เพื่อให้ API ฝั่ง Booking เอาไปใช้ต่อได้ทันที
		c.Set("user_id", token.UID)

		// ปล่อยผ่านให้ไปทำ API ถัดไปได้
		c.Next()
	}
}
