package bookings

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Booking คือ Struct ที่จับคู่กับ Document ใน Collection 'bookings'
type Booking struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ShowtimeID primitive.ObjectID `bson:"showtime_id" json:"showtime_id"`
	SeatNumber string             `bson:"seat_number" json:"seat_number"`
	UserID     string             `bson:"user_id" json:"user_id"`
	Status     string             `bson:"status" json:"status"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
}

// AuditLog คือ Struct สำหรับเก็บประวัติเหตุการณ์สำคัญในระบบ
type AuditLog struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Action    string             `bson:"action" json:"action"`
	Details   string             `bson:"details" json:"details"`
	Timestamp time.Time          `bson:"timestamp" json:"timestamp"`
}
