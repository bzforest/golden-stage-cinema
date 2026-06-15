package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AdminRequired เป็น Middleware สำหรับตรวจสอบสิทธิ์ผู้ดูแลระบบ
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// ดึงค่า email จาก Context ที่ถูกเซ็ตไว้ใน FirebaseAuthMiddleware
		emailVal, exists := c.Get("user_email")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: email claim is missing"})
			c.Abort()
			return
		}

		email := emailVal.(string)

		// ตรวจสอบว่าเป็น Admin หรือไม่ (Hardcode เพื่อความง่าย หรือเช็กจาก Database ก็ได้)
		if email != "admingoldenstage@gmail.com" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: you do not have administrator privileges"})
			c.Abort()
			return
		}

		// ปล่อยผ่าน
		c.Next()
	}
}
