package showtimes

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Cinema โมเดลสาขาโรงภาพยนตร์
type Cinema struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name       string             `bson:"name" json:"name"`
	CinemaType string             `bson:"cinema_type" json:"cinema_type"`
}

// Hall โมเดลโรงภาพยนตร์ย่อยในแต่ละสาขา
type Hall struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CinemaID primitive.ObjectID `bson:"cinema_id" json:"cinema_id"`
	Name     string             `bson:"name" json:"name"`
	Capacity int                `bson:"capacity" json:"capacity"`
}

// Showtime โมเดลรอบฉายภาพยนตร์
type Showtime struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	MovieID   primitive.ObjectID `bson:"movie_id" json:"movie_id"`
	CinemaID  primitive.ObjectID `bson:"cinema_id" json:"cinema_id"`
	HallID    primitive.ObjectID `bson:"hall_id" json:"hall_id"`
	StartTime time.Time          `bson:"start_time" json:"start_time"`
	EndTime   time.Time          `bson:"end_time" json:"end_time"`
}

// ShowtimeSeat โมเดลสถานะเก้าอี้ของแต่ละรอบฉาย
type ShowtimeSeat struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ShowtimeID primitive.ObjectID `bson:"showtime_id" json:"showtime_id"`
	SeatNumber string             `bson:"seat_number" json:"seat_number"`
	Type       string             `bson:"type" json:"type"` // เช่น "Premium", "Normal"
	Status     string             `bson:"status" json:"status"` // เช่น "AVAILABLE", "RESERVED", "LOCKED"
	Price      float64            `bson:"price" json:"price"`
}
