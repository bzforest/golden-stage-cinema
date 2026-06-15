# 💻 Code Walkthrough: เจาะลึกหัวใจสำคัญของระบบ Booking

เอกสารฉบับนี้จัดทำขึ้นเพื่อพาคุณเจาะลึกดู Code Snippets ของฟีเจอร์สำคัญในโปรเจกต์นี้ โดยเฉพาะเรื่องที่เกี่ยวข้องกับ Concurrency, Real-time Updates, และ Data Integrity ครับ

---

## 1. 🔒 ระบบ Distributed Lock ด้วย Redis SETNX
**ไฟล์:** `features/bookings/controller.go` (ฟังก์ชัน `LockSeat`)

ปัญหาที่ระบบจองตั๋วทุกระบบต้องเจอคือ "ทำยังไงไม่ให้คน 2 คนกดเก้าอี้เดียวกันในเวลาเดียวกัน" เราใช้ Redis SETNX เข้ามาแก้ปัญหานี้ครับ:
```go
// SETNX = Set if Not eXists (จะสร้างคีย์สำเร็จก็ต่อเมื่อคีย์นั้นยังไม่มีอยู่เท่านั้น)
success, err := config.RedisClient.SetNX(ctx, key, userID, 5*time.Minute).Result()

if !success {
    // ถ้า false แปลว่ามีคนกดตัดหน้าไปแล้วระดับเสี้ยววินาที ระบบจะตีกลับ 409
    c.JSON(http.StatusConflict, gin.H{"error": "Seat is already locked by someone else"})
    return
}
```
*💡 สังเกตว่าเราแนบ `userID` เข้าไปใน Value และให้เวลาตาย (TTL) ที่ 5 นาที เพื่อให้เกิดกลไก Auto-unlock หาก User หายตัวไป*

---

## 2. 🧱 ระบบ Bulk Confirmation (All-or-Nothing)
**ไฟล์:** `features/bookings/controller.go` (ฟังก์ชัน `ConfirmBooking`)

เมื่อผู้ใช้กดชำระเงินตั๋วหลายใบ ระบบจะต้องมั่นใจว่า "ทุกใบ" ยังเป็นของเขาอยู่ เพื่อป้องกัน Partial Failures (ตัดเงินเต็ม แต่ได้ตั๋วไม่ครบ):
```go
// 1. Pre-validation: เช็กกุญแจทีเดียวทั้งพวงด้วย MGet
vals, _ := config.RedisClient.MGet(ctx, lockKeys...).Result()

for i, val := range vals {
    // ถ้าที่นั่งไหนหลุด Lock หรือไม่ใช่ของตัวเอง ให้ Abort All ทันที!
    if val == nil || strings.Trim(val.(string), "\"") != cleanTokenUserID {
        c.JSON(http.StatusForbidden, gin.H{"error": "Lock invalid"})
        return
    }
}

// 2. Bulk Insert & Update: ทำทีเดียวเพื่อลดโอกาสพังกลางคัน
collection.InsertMany(ctx, bookings)
seatsCollection.UpdateMany(
    ctx,
    bson.M{"showtime_id": showtimeID, "seat_number": bson.M{"$in": req.SeatNumbers}},
    bson.M{"$set": bson.M{"status": "BOOKED"}},
)
```
*💡 การทำ Pre-validation ช่วยให้เราไม่ต้องกังวลเรื่อง Rollback Database มากนัก เพราะเราตรวจเช็กสิทธิ์ก่อนลงมือทำจริง*

---

## 3. 📡 การกระจายข่าวด้วย RabbitMQ
**ไฟล์:** `features/bookings/controller.go` (ฟังก์ชัน `LockSeat` / `ConfirmBooking`)

เราไม่ให้ Client โหลด API ซ้ำๆ เพื่อดูสถานะเก้าอี้ แต่เราใช้ RabbitMQ ยิง Event แทน:
```go
// ประกาศว่าห้องส่งนี้ชื่อ seat_updates_ex และเป็นแบบ Fanout (กระจายให้ทุกคนที่ต่อสาย)
config.RabbitChannel.ExchangeDeclare("seat_updates_ex", "fanout", true, false, false, false, nil)

// ยิงข้อความใส่ห้องส่งทันที
messageBody, _ := json.Marshal(map[string]string{
    "showtime_id": req.ShowtimeID,
    "seat_number": req.SeatNumber,
    "status":      "LOCKED", // หรือ BOOKED, AVAILABLE
})
config.RabbitChannel.PublishWithContext(ctx, "seat_updates_ex", "", false, false, amqp.Publishing{...})
```
*💡 การใช้ Fanout Exchange หมายความว่า ถ้ามี WebSocket Hub 3 ตัว (รันแอป 3 Instance) ทุก Hub จะได้รับข้อความนี้เหมือนกันหมด ทำให้สเกลระบบได้ง่าย*

---

## 4. 🌐 ฝั่ง Frontend: อัปเดต UI แบบ Real-time
**ไฟล์:** `src/views/SeatMapView.vue` (ส่วน WebSocket)

เมื่อมีข้อความจาก RabbitMQ ถูกส่งต่อมายัง WebSocket ฝั่ง Frontend จะตอบสนองดังนี้:
```javascript
ws.onmessage = (event) => {
  const data = JSON.parse(event.data)
  // หากเป็นเรื่องของรอบฉายอื่น ให้เมิน
  if (data.showtime_id !== showtimeId) return

  const seatIndex = seats.value.findIndex(s => s.seat_number === data.seat_number)
  if (seatIndex !== -1) {
    if (data.status === 'LOCKED' && data.user_id !== authStore.user?.uid) {
        // หากคนอื่นล็อค ให้ทาเก้าอี้เป็นสีแดง
        seats.value[seatIndex].status = 'LOCKED'
    } else if (data.status === 'LOCKED' && data.user_id === authStore.user?.uid) {
        // หากเป็นเราเองที่ล็อค (Boomerang effect) เปลี่ยนเป็นสีเหลือง (SELECTED)
        seats.value[seatIndex].status = 'SELECTED'
    } else {
        // อัปเดตสถานะ AVAILABLE หรือ BOOKED ตามปกติ
        seats.value[seatIndex].status = data.status
    }
  }
}
```
*💡 ลอจิกการเช็ก `user_id` ตรงนี้สำคัญมาก เพราะมันป้องกันไม่ให้เก้าอี้ที่เราเพิ่งกดเลือกเอง กลายเป็นสีแดง!*

---
## 5. ⚡ Frontend: การคืนเก้าอี้อัตโนมัติ (Cleanup)
**ไฟล์:** `src/views/SeatMapView.vue`

ถ้าผู้ใช้กด Back หรือสลับไปหน้าอื่นโดยไม่ได้ตั้งใจ เก้าอี้ควรถูกปลดล็อกทันที:
```javascript
let isProceeding = false // ตัวแปรกันไม่ให้ปลดล็อกหากกำลังเดินหน้าไปจ่ายเงิน

onBeforeUnmount(() => {
  // หากไม่ได้กำลังไปจ่ายเงิน...
  if (!isProceeding) {
    // ให้วนลูปคืนเก้าอี้ (ยิง DELETE /api/bookings/lock)
    releaseSelectedSeats()
  }
})
```
*💡 วิธีนี้ช่วยป้องกันเก้าอี้ว่างเปล่า (Ghost Seats) ที่ค้างอยู่ในระบบนาน 5 นาทีโดยไม่จำเป็น เพิ่มโอกาสในการขายตั๋ว*

---
## 6. 🗄️ Single Source of Truth
**ไฟล์:** `features/showtimes/controller.go` (ฟังก์ชัน `GetSeatsByShowtime`)

แทนที่จะต้องคอยซิงค์ข้อมูลระหว่าง 2 Tables (เช่น อัปเดตผังที่นั่งพร้อมกับอัปเดตใบเสร็จ) เราเปลี่ยนมาใช้การ **คำนวณสด** จากตารางใบเสร็จ (`bookings`) แทน:
```go
// ดึงประวัติการจองจาก bookings ทั้งหมด
bookingsCollection.Find(ctx, bson.M{"showtime_id": showtimeID})

// นำมาเช็กกับผังที่นั่งดั้งเดิม
for i := range seats {
    if bookedMap[seats[i].SeatNumber] {
        seats[i].Status = "BOOKED"
    } else {
        seats[i].Status = "AVAILABLE"
    }
}
```
*💡 ลอจิกนี้เป็นหัวใจสำคัญของการทำ Single Source of Truth ช่วยขจัดบั๊กข้อมูลไม่ตรงกัน (Data Inconsistency) ในระดับโครงสร้าง*

---
## 7. 🔌 ระบบซ่อมภาพ (WebSocket Reconnection Sync)
**ไฟล์:** `src/views/SeatMapView.vue` (ส่วนฟังก์ชัน `connectWebSocket`)

ป้องกันปัญหาเน็ตหลุดชั่วคราวแล้วพลาด Message ที่คนอื่นส่งมา:
```javascript
let hasConnectedBefore = false

ws.onopen = () => {
  if (hasConnectedBefore) {
    // หากเป็นการต่อใหม่ (Reconnect) สั่งโหลดผังที่นั่งจาก Backend มาทับทันที!
    movieStore.fetchSeatsByShowtime(showtimeId)
  }
  hasConnectedBefore = true
}
```
*💡 เป็นกลไก Fallback ที่ง่ายแต่ทรงพลัง รับประกันได้ว่าทันทีที่เน็ตกลับมาต่อติด ผังที่นั่งบนจอจะต้องตรงกับ Database 100% เสมอ*
