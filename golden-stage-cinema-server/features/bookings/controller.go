package bookings

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"firebase.google.com/go/auth"
	"golden-stage-cinema-server/config"

	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type LockSeatRequest struct {
	ShowtimeID string `json:"showtime_id" binding:"required"`
	SeatNumber string `json:"seat_number" binding:"required"`
	// เอา UserID ออกจาก JSON Request ป้องกันคนส่งมาปลอมแปลง
}

// CachedUser เก็บข้อมูลผู้ใช้งานแบบย่อสำหรับส่งออก
type CachedUser struct {
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
}

// batchFetchUsersCached ดึงข้อมูลจาก Redis ก่อน หากไม่พบจึงจะดึงจาก Firebase
func batchFetchUsersCached(ctx context.Context, uids []string) map[string]CachedUser {
	userMap := make(map[string]CachedUser)
	if len(uids) == 0 {
		return userMap
	}

	var missingIdentifiers []auth.UserIdentifier

	// 1. ตรวจสอบ Redis Cache ทีละรายการ (สามารถประยุกต์ใช้ MGET ได้ในอนาคตเพื่อความเร็วสูงสุด)
	for _, uid := range uids {
		if uid == "" {
			continue
		}
		cacheKey := "firebase_user:" + uid
		val, err := config.RedisClient.Get(ctx, cacheKey).Result()
		if err == nil && val != "" {
			var cachedUser CachedUser
			if err := json.Unmarshal([]byte(val), &cachedUser); err == nil {
				userMap[uid] = cachedUser
				continue
			}
		}
		missingIdentifiers = append(missingIdentifiers, auth.UIDIdentifier{UID: uid})
	}

	// 2. ดึงข้อมูลส่วนที่หายไปจาก Firebase ทีเดียว (Batch)
	if len(missingIdentifiers) > 0 {
		getUsersResult, err := config.FirebaseAuth.GetUsers(ctx, missingIdentifiers)
		if err == nil {
			for _, u := range getUsersResult.Users {
				cu := CachedUser{
					Email:       u.Email,
					DisplayName: u.DisplayName,
				}
				userMap[u.UID] = cu

				// 3. จัดเก็บลง Redis Cache (ตั้งเวลาหมดอายุ 24 ชั่วโมง)
				cacheBytes, _ := json.Marshal(cu)
				config.RedisClient.Set(ctx, "firebase_user:"+u.UID, string(cacheBytes), 24*time.Hour)
			}
		}
	}

	return userMap
}

// LockSeat ฟังก์ชันสำหรับจองและล็อกที่นั่ง
func LockSeat(c *gin.Context) {
	var req LockSeatRequest

	// แกะ JSON ใส่ตัวแปร req
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// ดึงไอดีผู้ใช้ตัวจริงจาก Middleware
	userID := c.MustGet("user_id").(string)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 1. Rate Limiting: Max 20 requests per minute
	rateLimitKey := fmt.Sprintf("rate_limit:lock:%s", userID)
	count, err := config.RedisClient.Incr(ctx, rateLimitKey).Result()
	if err == nil && count == 1 {
		config.RedisClient.Expire(ctx, rateLimitKey, time.Minute)
	}
	if count > 20 {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests. Please slow down."})
		return
	}

	// สร้าง Redis Key
	key := fmt.Sprintf("lock:seat:%s:%s", req.ShowtimeID, req.SeatNumber)

	// ใช้ SETNX (Set if Not eXists) เพื่อล็อกที่นั่งด้วย userID ตั้งเวลา 5 นาที
	success, err := config.RedisClient.SetNX(ctx, key, userID, 5*time.Minute).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process lock request"})
		return
	}

	if success {
		// Publish LOCKED event ไปที่ RabbitMQ เพื่อให้ WebSocket broadcast ไปบอกทุก Client
		config.RabbitChannel.ExchangeDeclare("seat_updates_ex", "fanout", true, false, false, false, nil)
		messageBody, _ := json.Marshal(map[string]string{
			"showtime_id": req.ShowtimeID,
			"seat_number": req.SeatNumber,
			"status":      "LOCKED",
			"user_id":     userID,
		})
		config.RabbitChannel.PublishWithContext(ctx, "seat_updates_ex", "", false, false, amqp.Publishing{
			ContentType: "application/json",
			Body:        messageBody,
		})

		auditCollection := config.GetCollection("audit_logs")
		auditCollection.InsertOne(ctx, AuditLog{
			Action:     "SEAT_LOCKED",
			Details:    fmt.Sprintf("User %s locked seat %s for showtime %s", userID, req.SeatNumber, req.ShowtimeID),
			Timestamp:  time.Now(),
			UID:        userID,
			ShowtimeID: req.ShowtimeID,
			SeatNumber: req.SeatNumber,
		})

		c.JSON(http.StatusOK, gin.H{"message": "Seat locked successfully"})
	} else {
		// Log System Error (Lock fail)
		auditCollection := config.GetCollection("audit_logs")
		auditCollection.InsertOne(ctx, AuditLog{
			Action:     "SYSTEM_ERROR",
			Details:    fmt.Sprintf("Lock failed: Seat %s for showtime %s is already locked or unavailable", req.SeatNumber, req.ShowtimeID),
			Timestamp:  time.Now(),
			UID:        userID,
			ShowtimeID: req.ShowtimeID,
			SeatNumber: req.SeatNumber,
		})
		c.JSON(http.StatusConflict, gin.H{"error": "Seat is already locked by another user"})
	}
}

// UnlockSeat ฟังก์ชันสำหรับปลดล็อกที่นั่งคืนสู่ระบบ
func UnlockSeat(c *gin.Context) {
	showtimeID := c.Param("showtime_id")
	seatNumber := c.Param("seat_number")
	userID := c.MustGet("user_id").(string)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := fmt.Sprintf("lock:seat:%s:%s", showtimeID, seatNumber)
	val, err := config.RedisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		c.JSON(http.StatusOK, gin.H{"message": "Seat is already unlocked"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check seat lock status"})
		return
	}

	// Clean up strings before comparison to prevent formatting issues
	valClean := strings.Trim(val, "\" \n\r\t")
	userIDClean := strings.Trim(userID, "\" \n\r\t")

	log.Printf("[UnlockSeat] Comparing lock. Redis: '%s' (raw: %v), Request: '%s' (raw: %v)", valClean, val, userIDClean, userID)

	if valClean != userIDClean {
		c.JSON(http.StatusForbidden, gin.H{"error": fmt.Sprintf("Cannot unlock a seat locked by someone else. LockedBy: %s, Requester: %s", valClean, userIDClean)})
		return
	}

	_, err = config.RedisClient.Del(ctx, key).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unlock seat"})
		return
	}

	// ประกาศ Exchange และ Publish ให้ RabbitMQ ทราบว่าที่นั่งนี้ AVAILABLE แล้ว
	config.RabbitChannel.ExchangeDeclare("seat_updates_ex", "fanout", true, false, false, false, nil)
	messageBody, _ := json.Marshal(map[string]string{
		"showtime_id": showtimeID,
		"seat_number": seatNumber,
		"status":      "AVAILABLE",
	})
	config.RabbitChannel.PublishWithContext(ctx, "seat_updates_ex", "", false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        messageBody,
	})

	// Log Seat Released
	auditCollection := config.GetCollection("audit_logs")
	auditCollection.InsertOne(ctx, AuditLog{
		Action:     "SEAT_RELEASED",
		Details:    fmt.Sprintf("User %s explicitly released seat %s for showtime %s", userIDClean, seatNumber, showtimeID),
		Timestamp:  time.Now(),
		UID:        userIDClean,
		ShowtimeID: showtimeID,
		SeatNumber: seatNumber,
	})

	c.JSON(http.StatusOK, gin.H{"message": "Seat unlocked successfully"})
}

// ConfirmBooking ฟังก์ชันยืนยันการจอง บันทึกลง Mongo และส่ง Event เข้า RabbitMQ
func ConfirmBooking(c *gin.Context) {
	var req ConfirmBookingRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// ดึงไอดีผู้ใช้ตัวจริงจาก Middleware
	userID := c.MustGet("user_id").(string)
	cleanTokenUserID := strings.TrimSpace(strings.Trim(userID, "\""))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	showtimeID, err := primitive.ObjectIDFromHex(req.ShowtimeID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid showtime_id format"})
		return
	}

	// 1. Pre-validation: ตรวจสอบ Ownership ใน Redis สำหรับทุกที่นั่งก่อนลงมือทำจริง
	var lockKeys []string
	for _, seatNumber := range req.SeatNumbers {
		lockKeys = append(lockKeys, fmt.Sprintf("lock:seat:%s:%s", req.ShowtimeID, seatNumber))
	}

	vals, err := config.RedisClient.MGet(ctx, lockKeys...).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check lock ownership"})
		return
	}

	// เช็กว่าทุกที่นั่งเป็นของเราหรือไม่
	for i, val := range vals {
		if val == nil {
			c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("Seat lock expired or not locked for seat %s", req.SeatNumbers[i])})
			return
		}
		
		redisUserID := val.(string)
		cleanRedisUserID := strings.TrimSpace(strings.Trim(redisUserID, "\""))
		
		if cleanRedisUserID != cleanTokenUserID {
			c.JSON(http.StatusForbidden, gin.H{"error": fmt.Sprintf("You are not the owner of seat lock for seat %s", req.SeatNumbers[i])})
			return
		}
	}

	// 1.5 Atomic Reservation via UpdateOne and $nin
	showtimesCollection := config.GetCollection("showtimes")
	filter := bson.M{
		"_id":          showtimeID,
		"booked_seats": bson.M{"$nin": req.SeatNumbers},
	}
	update := bson.M{
		"$push": bson.M{"booked_seats": bson.M{"$each": req.SeatNumbers}},
	}
	updateResult, err := showtimesCollection.UpdateOne(ctx, filter, update)
	if err != nil || updateResult.ModifiedCount == 0 {
		config.RedisClient.Del(ctx, lockKeys...)
		c.JSON(http.StatusConflict, gin.H{"error": "Sorry, some seats were just taken."})
		return
	}

	// 2. Bulk Insert: บันทึกข้อมูลการจองลง MongoDB
	var bookings []interface{}
	for _, seatNumber := range req.SeatNumbers {
		bookings = append(bookings, Booking{
			ShowtimeID: showtimeID,
			SeatNumber: seatNumber,
			UserID:     userID,
			Status:     "CONFIRMED",
			CreatedAt:  time.Now(),
		})
	}

	collection := config.GetCollection("bookings")
	_, err = collection.InsertMany(ctx, bookings)
	if err != nil {
		config.RedisClient.Del(ctx, lockKeys...)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save bookings to database"})
		return
	}

	// (3. Bulk Update ถูกนำออกแล้ว เนื่องจากใช้ Single Source of Truth จากตาราง bookings)

	// 4. ลบ Lock ออกจาก Redis ทั้งหมด
	config.RedisClient.Del(ctx, lockKeys...)

	// 5. ประกาศ Exchange แบบ Fanout ใน RabbitMQ
	config.RabbitChannel.ExchangeDeclare("seat_updates_ex", "fanout", true, false, false, false, nil)

	// วนลูป Publish แจ้งสถานะของทีละเก้าอี้ เพื่อให้ฝั่ง Frontend (ที่รับทีละตัว) ทำงานได้เหมือนเดิม
	for _, seatNumber := range req.SeatNumbers {
		messageBody, _ := json.Marshal(map[string]string{
			"showtime_id": req.ShowtimeID,
			"seat_number": seatNumber,
			"status":      "BOOKED",
		})
		
		config.RabbitChannel.PublishWithContext(ctx,
			"seat_updates_ex", // exchange
			"",                // routing key
			false,             // mandatory
			false,             // immediate
			amqp.Publishing{
				ContentType: "application/json",
				Body:        messageBody,
			})
	}

	auditCollection := config.GetCollection("audit_logs")
	auditCollection.InsertOne(ctx, AuditLog{
		Action:     "BOOKING_CONFIRMED",
		Details:    fmt.Sprintf("User %s confirmed booking for showtime %s (Seats: %v)", userID, req.ShowtimeID, req.SeatNumbers),
		Timestamp:  time.Now(),
		UID:        userID,
		ShowtimeID: req.ShowtimeID,
		SeatNumber: strings.Join(req.SeatNumbers, ","),
	})

	// ตอบกลับ 200 OK
	c.JSON(http.StatusOK, gin.H{"message": "Bookings confirmed and messages queued"})
}

// GetUserBookings ดึงประวัติการจองของผู้ใช้งาน
func GetUserBookings(c *gin.Context) {
	userID := c.MustGet("user_id").(string)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := config.GetCollection("bookings")

	// ใช้ Aggregate Pipeline เพื่อ Join กับ showtimes และ movies
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.M{"user_id": userID}}},
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "showtimes"},
			{Key: "localField", Value: "showtime_id"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "showtime"},
		}}},
		bson.D{{Key: "$unwind", Value: bson.D{
			{Key: "path", Value: "$showtime"},
			{Key: "preserveNullAndEmptyArrays", Value: true},
		}}},
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "movies"},
			{Key: "localField", Value: "showtime.movie_id"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "movie"},
		}}},
		bson.D{{Key: "$unwind", Value: bson.D{
			{Key: "path", Value: "$movie"},
			{Key: "preserveNullAndEmptyArrays", Value: true},
		}}},
        bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "cinemas"},
			{Key: "localField", Value: "showtime.cinema_id"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "cinema"},
		}}},
		bson.D{{Key: "$unwind", Value: bson.D{
			{Key: "path", Value: "$cinema"},
			{Key: "preserveNullAndEmptyArrays", Value: true},
		}}},
        bson.D{{Key: "$sort", Value: bson.M{"created_at": -1}}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch bookings"})
		return
	}
	defer cursor.Close(ctx)

	var bookings []bson.M
	if err = cursor.All(ctx, &bookings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode bookings"})
		return
	}

	if bookings == nil {
		bookings = []bson.M{}
	}

	c.JSON(http.StatusOK, bookings)
}

// GetAdminBookings ดึงประวัติการจองทั้งหมดสำหรับผู้ดูแลระบบ พร้อมรองรับการ Filter แบบ Server-side
func GetAdminBookings(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	searchStr := c.Query("search")
	dateStr := c.Query("date")
	pageStr := c.Query("page")
	limitStr := c.Query("limit")

	page := 1
	limit := 10
	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		page = p
	}
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		limit = l
	}

	collection := config.GetCollection("bookings")

	// ใช้ Aggregate Pipeline เพื่อ Join กับ showtimes
	pipeline := mongo.Pipeline{}

	// 1. $lookup โยงตาราง showtimes, movies, cinemas
	pipeline = append(pipeline,
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "showtimes"},
			{Key: "localField", Value: "showtime_id"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "showtime"},
		}}},
		bson.D{{Key: "$unwind", Value: bson.D{
			{Key: "path", Value: "$showtime"},
			{Key: "preserveNullAndEmptyArrays", Value: true},
		}}},
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "movies"},
			{Key: "localField", Value: "showtime.movie_id"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "movie"},
		}}},
		bson.D{{Key: "$unwind", Value: bson.D{
			{Key: "path", Value: "$movie"},
			{Key: "preserveNullAndEmptyArrays", Value: true},
		}}},
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "cinemas"},
			{Key: "localField", Value: "showtime.cinema_id"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "cinema"},
		}}},
		bson.D{{Key: "$unwind", Value: bson.D{
			{Key: "path", Value: "$cinema"},
			{Key: "preserveNullAndEmptyArrays", Value: true},
		}}},
	)

	// 2. สร้าง Filter ($match) โดยไม่ต้องพึ่งพา $addFields
	matchFilter := bson.M{}

	// การค้นหาด้วยวันที่แบบเจาะจง
	if dateStr != "" {
		if t, err := time.Parse("2006-01-02", dateStr); err == nil {
			nextDay := t.Add(24 * time.Hour)
			matchFilter["created_at"] = bson.M{"$gte": t, "$lt": nextDay}
		}
	}

	if searchStr != "" {
		cleanSearchStr := strings.TrimSpace(strings.TrimPrefix(searchStr, "#"))
		orConditions := []bson.M{}

		// เงื่อนไข 1: ค้นหาด้วย ObjectID (ถ้า Parse ผ่าน)
		if objID, err := primitive.ObjectIDFromHex(cleanSearchStr); err == nil {
			orConditions = append(orConditions, bson.M{"_id": objID})
		}

		// เงื่อนไข 2: Regex ในฟิลด์ที่เหมาะสม
		orConditions = append(orConditions, bson.M{"movie.title": bson.M{"$regex": cleanSearchStr, "$options": "i"}})

		// หา UIDs จาก Firebase (ถ้าตรงกับ SearchStr) // ยังต้อง Fetch Users อยู่แต่จะดึงแบบจำกัด
		// หมายเหตุ: การ Search จาก Firebase ทุกตัวยังหลีกเลี่ยงไม่ได้ 100% ถ้าไม่ซิงค์ลง MongoDB 
		// แต่อย่างน้อยก็ทำแค่ตอน Search
		var matchingUIDs []string
		iter := config.FirebaseAuth.Users(ctx, "")
		for {
			u, err := iter.Next()
			if err != nil {
				break
			}
			if strings.Contains(strings.ToLower(u.Email), strings.ToLower(cleanSearchStr)) ||
				strings.Contains(strings.ToLower(u.DisplayName), strings.ToLower(cleanSearchStr)) {
				matchingUIDs = append(matchingUIDs, u.UID)
			}
		}

		if len(matchingUIDs) > 0 {
			orConditions = append(orConditions, bson.M{"user_id": bson.M{"$in": matchingUIDs}})
		}

		matchFilter["$or"] = orConditions
	}

	if len(matchFilter) > 0 {
		pipeline = append(pipeline, bson.D{{Key: "$match", Value: matchFilter}})
	}

	// 3. Count Total ก่อน Skip/Limit
	countPipeline := append(pipeline, bson.D{{Key: "$count", Value: "total"}})
	countCursor, err := collection.Aggregate(ctx, countPipeline)
	total := 0
	if err == nil {
		var countResult []bson.M
		if err = countCursor.All(ctx, &countResult); err == nil && len(countResult) > 0 {
			total = int(countResult[0]["total"].(int32))
		}
	}

	// 4. Sort, Skip, Limit
	pipeline = append(pipeline, bson.D{{Key: "$sort", Value: bson.M{"created_at": -1}}})
	pipeline = append(pipeline, bson.D{{Key: "$skip", Value: (page - 1) * limit}})
	pipeline = append(pipeline, bson.D{{Key: "$limit", Value: limit}})

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch admin bookings"})
		return
	}
	defer cursor.Close(ctx)

	var bookings []bson.M
	if err = cursor.All(ctx, &bookings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode admin bookings"})
		return
	}

	if bookings == nil {
		bookings = []bson.M{}
	} else {
		// 5. Batch Fetch Users from Firebase with Redis Cache
		var userIDs []string
		uniqueUIDs := make(map[string]bool)
		for _, b := range bookings {
			if uid, ok := b["user_id"].(string); ok && uid != "" && !uniqueUIDs[uid] {
				uniqueUIDs[uid] = true
				userIDs = append(userIDs, uid)
			}
		}

		userMap := batchFetchUsersCached(ctx, userIDs)

		for _, b := range bookings {
			if uid, ok := b["user_id"].(string); ok && uid != "" {
				if u, exists := userMap[uid]; exists {
					b["user_email"] = u.Email
					b["user_name"] = u.DisplayName
				} else {
					b["user_email"] = "Unknown"
					b["user_name"] = "Unknown"
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  bookings,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// GetAdminLogs ดึงข้อมูล Audit Logs ทั้งหมดพร้อมระบบค้นหาและ Enrich Data
func GetAdminLogs(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pageStr := c.Query("page")
	limitStr := c.Query("limit")
	searchStr := c.Query("search")

	page := 1
	limit := 10
	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		page = p
	}
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		limit = l
	}

	collection := config.GetCollection("audit_logs")

	// 1. สร้าง Filter สำหรับ Search
	matchFilter := bson.M{}
	if searchStr != "" {
		matchFilter["$or"] = []bson.M{
			{"action": bson.M{"$regex": searchStr, "$options": "i"}},
			{"details": bson.M{"$regex": searchStr, "$options": "i"}},
		}
	}

	total, _ := collection.CountDocuments(ctx, matchFilter)

	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "timestamp", Value: -1}})
	findOptions.SetSkip(int64((page - 1) * limit))
	findOptions.SetLimit(int64(limit))

	cursor, err := collection.Find(ctx, matchFilter, findOptions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch logs"})
		return
	}
	defer cursor.Close(ctx)

	var logs []AuditLog
	if err = cursor.All(ctx, &logs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode logs"})
		return
	}
	if logs == nil {
		logs = []AuditLog{}
	}

	// 2. Enrich Data (ดึงชื่อ User, ชื่อภาพยนตร์, เวลาฉาย)
	type EnrichedLog struct {
		ID            string    `json:"id"`
		Action        string    `json:"action"`
		Details       string    `json:"details"`
		Timestamp     time.Time `json:"timestamp"`
		UserEmail     string    `json:"user_email,omitempty"`
		UserName      string    `json:"user_name,omitempty"`
		SeatNumber    string    `json:"seat_number,omitempty"`
		MovieTitle    string    `json:"movie_title,omitempty"`
		ShowtimeDate  string    `json:"showtime_date,omitempty"`
		UID           string    `json:"uid,omitempty"`
		ShowtimeIDStr string    `json:"showtime_id,omitempty"`
		RawJSON       bool      `json:"raw_json"`
	}

	var enrichedLogs []EnrichedLog

	// Regex for parsing Details string
	lockRegex := regexp.MustCompile(`User (.*) locked seat (.*) for showtime (.*)`)
	confirmRegexAlt := regexp.MustCompile(`User (.*) confirmed booking for showtime (.*) \(Seats: (.*)\)`)

	// Sets สำหรับ Batch Processing
	uniqueUIDs := make(map[string]bool)
	uniqueShowtimeIDs := make(map[string]bool)

	// Step 1: Parse Regex และรวบรวม IDs
	for _, log := range logs {
		eLog := EnrichedLog{
			ID:        log.ID.Hex(),
			Action:    log.Action,
			Details:   log.Details,
			Timestamp: log.Timestamp,
			RawJSON:   false,
		}

		// อ่านค่าจาก Struct ทันทีโดยไม่ต้องใช้ Regex (อ้างอิงจาก Struct ใหม่)
		uid := log.UID
		showtimeID := log.ShowtimeID
		seatNum := log.SeatNumber

		// หากค่าว่างแปลว่าเป็น Log เก่า ให้ใช้ Regex งัดข้อมูลเหมือนเดิม
		if uid == "" || showtimeID == "" {
			var jsonDetails map[string]interface{}
			if err := json.Unmarshal([]byte(log.Details), &jsonDetails); err == nil && len(jsonDetails) > 0 {
				eLog.RawJSON = true
				if sn, ok := jsonDetails["seat_number"].(string); ok {
					seatNum = sn
				}
				if stID, ok := jsonDetails["showtime_id"].(string); ok {
					showtimeID = stID
				}
			} else {
				if matches := lockRegex.FindStringSubmatch(log.Details); len(matches) == 4 {
					uid = matches[1]
					seatNum = matches[2]
					showtimeID = matches[3]
				} else if matches := confirmRegexAlt.FindStringSubmatch(log.Details); len(matches) == 4 {
					uid = matches[1]
					showtimeID = matches[2]
					seatNum = matches[3]
				}
			}
		} else {
			// ตรวจสอบว่าเก็บ JSON ไหม (ถ้า Details เป็น JSON จะได้เซต RawJSON = true)
			if strings.HasPrefix(strings.TrimSpace(log.Details), "{") {
				eLog.RawJSON = true
			}
		}

		eLog.UID = uid
		eLog.ShowtimeIDStr = showtimeID
		eLog.SeatNumber = seatNum
		
		if uid != "" {
			uniqueUIDs[uid] = true
		}
		if showtimeID != "" {
			uniqueShowtimeIDs[showtimeID] = true
		}

		enrichedLogs = append(enrichedLogs, eLog)
	}

	// Step 2: Batch Fetch Firebase Users with Redis Cache
	var userIDs []string
	for uid := range uniqueUIDs {
		userIDs = append(userIDs, uid)
	}
	userMap := batchFetchUsersCached(ctx, userIDs)

	// Step 3: Batch Fetch MongoDB Showtimes and Movies
	var stObjIDs []primitive.ObjectID
	for stID := range uniqueShowtimeIDs {
		if objID, err := primitive.ObjectIDFromHex(stID); err == nil {
			stObjIDs = append(stObjIDs, objID)
		}
	}

	showtimeMap := make(map[string]bson.M)
	var movieObjIDs []primitive.ObjectID
	uniqueMovieIDs := make(map[primitive.ObjectID]bool)

	if len(stObjIDs) > 0 {
		stCursor, err := config.GetCollection("showtimes").Find(ctx, bson.M{"_id": bson.M{"$in": stObjIDs}})
		if err == nil {
			var sts []bson.M
			stCursor.All(ctx, &sts)
			for _, st := range sts {
				stIDStr := st["_id"].(primitive.ObjectID).Hex()
				showtimeMap[stIDStr] = st
				if mvID, ok := st["movie_id"].(primitive.ObjectID); ok {
					if !uniqueMovieIDs[mvID] {
						uniqueMovieIDs[mvID] = true
						movieObjIDs = append(movieObjIDs, mvID)
					}
				}
			}
		}
	}

	movieMap := make(map[primitive.ObjectID]string)
	if len(movieObjIDs) > 0 {
		mvCursor, err := config.GetCollection("movies").Find(ctx, bson.M{"_id": bson.M{"$in": movieObjIDs}})
		if err == nil {
			var mvs []bson.M
			mvCursor.All(ctx, &mvs)
			for _, mv := range mvs {
				mvID := mv["_id"].(primitive.ObjectID)
				if title, ok := mv["title"].(string); ok {
					movieMap[mvID] = title
				}
			}
		}
	}

	// Step 4: Map back to EnrichedLogs
	for i, eLog := range enrichedLogs {
		if eLog.UID != "" {
			if u, exists := userMap[eLog.UID]; exists {
				enrichedLogs[i].UserEmail = u.Email
				enrichedLogs[i].UserName = u.DisplayName
			}
		}
		
		if eLog.ShowtimeIDStr != "" {
			if st, exists := showtimeMap[eLog.ShowtimeIDStr]; exists {
				if startTime, ok := st["start_time"].(primitive.DateTime); ok {
					enrichedLogs[i].ShowtimeDate = startTime.Time().Format("2006-01-02 15:04:05")
				}
				if mvID, ok := st["movie_id"].(primitive.ObjectID); ok {
					if title, exists := movieMap[mvID]; exists {
						enrichedLogs[i].MovieTitle = title
					}
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  enrichedLogs,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}
