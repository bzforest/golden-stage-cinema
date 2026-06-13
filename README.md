# 🎬 Cinema Ticket Booking System

## 1. System Architecture Diagram

```text
[ ผู้ใช้งาน (User) ]
       |
       | 1. Login
       v
[ Firebase Auth ] ---> (คืนค่า idToken ให้ Vue นำไปใช้)
       |
       | 2. เปิดหน้าเว็บ / กดจองที่นั่ง (REST API)
       | 3. รับสถานะที่นั่ง Real-time (WebSocket)
       v
[ Frontend (Vue 3) ] <====================> [ Backend (Go / Gin) ]
                                                    |
                                                    |--- 4. ล็อกที่นั่ง (SETNX)
                                                    |---> [ Redis ]
                                                    |
                                                    |--- 5. ดึงข้อมูลหนัง / บันทึกการจอง
                                                    |---> [ MongoDB ]
                                                    |
                                                    |--- 6. โยน Event เมื่อจ่ายเงินสำเร็จ
                                                    |---> [ RabbitMQ ]
                                                            |
                                                            v
                                                  [ Worker (Go Routine) ]
                                            (คอยรับ Event ไปบันทึก Log หรือจำลองส่ง Email)
```

## 2. Tech Stack Overview

* **Backend: Go (Gin Framework)**
  เหตุผล: Go โดดเด่นเรื่องการจัดการ Concurrency ผ่าน Goroutines ซึ่งตอบโจทย์ระบบที่ต้องรับ Request พร้อมกันจำนวนมาก ผนวกกับ Gin Framework ที่มีน้ำหนักเบา (Lightweight) ทำให้การสร้าง RESTful API มีความรวดเร็วและอ่านโค้ดได้ง่าย
* **Frontend: Vue.js (Vue 3)**
  เหตุผล: ใช้ Composition API เพื่อการจัดการ State ที่ซับซ้อนได้อย่างมีระเบียบ ช่วยให้หน้า UI ผังที่นั่งสามารถตอบสนองต่อการอัปเดตแบบ Real-time ได้อย่างลื่นไหล
* **Database: MongoDB**
  เหตุผล: โครงสร้างแบบ Document-based มีความยืดหยุ่น เหมาะสำหรับการจัดเก็บข้อมูล Movie, Seat Map และ Booking Record ที่มีความสัมพันธ์เกี่ยวข้องกัน และรองรับการขยายตัว (Scalability) ได้ดี
* **Cache & Distributed Lock: Redis**
  เหตุผล: ใช้ความเร็วของ In-memory data store ในการทำ Distributed Lock เมื่อผู้ใช้อยู่ระหว่างการทำรายการ (5 นาที) เพื่อการันตีว่าจะไม่มีทางเกิด Double Booking อย่างเด็ดขาด
* **Message Queue: RabbitMQ**
  เหตุผล: เลือกใช้แทนที่ Pub/Sub ธรรมดา เพื่อให้ระบบมี Message Queue ที่แท้จริง รองรับการทำงานแบบ Asynchronous เช่น การส่ง Event ยืนยันการจองสำเร็จ (Async Logging/Notification) โดยรับประกันว่าข้อมูลจะไม่สูญหาย
* **Real-time Communication: WebSocket**
  เหตุผล: เป็นช่องทางการสื่อสารแบบสองทางที่เสถียร ทำให้ Client ทุกหน้าจอเห็นการอัปเดตสถานะที่นั่ง (AVAILABLE, LOCKED, BOOKED) พร้อมกันในทันทีโดยไม่ต้อง Refresh หน้าเว็บ
* **Authentication: Firebase Auth**
  เหตุผล: เป็นการลดความซับซ้อน (Decoupling) ในการจัดการระบบยืนยันตัวตน ระบบมีความปลอดภัยสูงและช่วยให้เราได้ `user_id` มาใช้งานร่วมกับ Backend ได้รวดเร็วและเป็นมาตรฐาน
* **Deployment: Docker & Docker Compose**
  เหตุผล: จัดการ Containerization เพื่อให้ทุก Services สามารถรันทำงานประสานกันได้อย่างสมบูรณ์ด้วยคำสั่ง `docker compose up --build` เพียงคำสั่งเดียว

## 3. Booking Flow

1. **User เลือกรอบฉายและที่นั่ง:** ฝั่ง Client ดึงผังที่นั่งจาก API `GET /api/showtimes/:showtime_id/seats` (สถานะ AVAILABLE)
2. **User กดล็อกที่นั่ง:** Client ยิง API `POST /api/bookings/lock` ส่ง `{showtime_id, seat_number}`
   - Backend ใช้ **Redis SETNX** สร้าง Key ล็อกที่นั่งเป็นเวลา 5 นาที
   - ส่งข้อความเข้า **RabbitMQ** แจ้งสถานะ LOCKED เพื่ออัปเดตหน้าจอผ่าน **WebSocket**
3. **การชำระเงินและการยืนยัน:**
   - **กรณีชำระเงินสำเร็จ (ภายใน 5 นาที):** Client ยิง API `POST /api/bookings/confirm`
     - Backend บันทึกข้อมูลลง MongoDB (สถานะ CONFIRMED)
     - ส่งข้อความเข้า RabbitMQ เพื่อแจ้งสถานะ BOOKED และทริกเกอร์ Worker ให้บันทึก Audit Log / จำลองส่งอีเมล
   - **กรณีชำระเงินไม่ทัน (Timeout):**
     - Key ใน Redis จะหมดอายุอัตโนมัติ (Expired)
     - **Timeout Listener** ฝั่ง Backend ดักจับ Event หมดอายุจาก Keyspace Notifications
     - ส่งข้อความสถานะ AVAILABLE กลับเข้า RabbitMQ และ WebSocket เพื่อปลดล็อกที่นั่งคืนสู่หน้าเว็บ

## 4. Redis Lock Strategy

เพื่อป้องกันปัญหาจองที่นั่งซ้ำซ้อน (Double Booking) ขณะที่ผู้ใช้หลายคนกดเลือกที่นั่งพร้อมกัน ระบบจึงเลือกใช้ **Redis Distributed Lock (SETNX - Set if Not Exists)**
- **ความเร็ว:** Redis ทำงานบน Memory การเช็กว่าที่นั่งถูกล็อกหรือยังทำได้รวดเร็วมากระดับมิลลิวินาที
- **Atomic Operation:** คำสั่ง SETNX การันตีได้ว่าถ้ามีคำสั่ง 2 ตัวเข้ามาพร้อมกัน ตัวหนึ่งจะสำเร็จ (True) และอีกตัวจะล้มเหลว (False) แน่นอน ช่วยตัดปัญหาสภาพการแข่งขัน (Race Condition)
- **TTL (Time To Live):** ล็อกมีการกำหนดเวลาตายชัดเจนที่ 5 นาที หากกระบวนการฝั่งแอปมีปัญหา (เช่น เน็ตหลุด แอปแครช) ล็อกก็จะถูกปลดคืนอัตโนมัติโดยไม่ค้างในระบบ

## 5. Message Queue

ระบบใช้ **RabbitMQ** เป็นศูนย์กลางในการสื่อสารเพื่อกระจาย Event โดยไม่ต้องพึ่งพา API หลัก (Decoupling) ประโยชน์หลัก 2 ข้อ:
1. **Real-time Seat Updates:** เมื่อมีการ Lock, Confirm, หรือ Timeout สถานะเก้าอี้จะถูก Publish ลงคิว `seat_updates` และฝั่ง WebSocket จะ Consume ออกมากระจายให้ User ที่หน้าจอทันที ทำให้เห็นที่นั่งเด้งสลับสถานะแบบ Real-time 
2. **Background Processing (Worker):** เมื่อยืนยันการจองสำเร็จ Worker จะมารับงานบันทึก Audit Logs ลง MongoDB และจำลองส่ง Email การันตีว่าข้อมูลจะไม่ตกหล่นแม้ยอดจองจะพุ่งสูง และช่วยให้ API ไม่โหลดหนักไปทำงานอื่น

## 6. Assumptions & Trade-offs

ผมทราบในเรื่องของ Best Practice (เช่น การลด Global Variables, การแยก Service/Repository Layer, การทำ Centralized Config, หรือการทำ Standardized Response) ว่ามีหนทางที่ดีกว่านี้ 
แต่ด้วยโค้ดที่มีความซับซ้อนเพิ่มมากขึ้นจนเกินขอบเขตความเข้าใจของผมในตอนนี้ ทำให้ผมเลือกที่จะใช้ทางที่ระดับความรู้ตอนนี้ของผมสามารถเข้าใจและสามารถอธิบายได้ดีที่สุดครับ

## 7. วิธีรันระบบ

ระบบได้ถูกตั้งค่า Docker Container ไว้เรียบร้อยแล้ว รันคำสั่งเดียวระบบจะผูก MongoDB, Redis, RabbitMQ และ Go Backend เข้าด้วยกัน:

```bash
docker compose up --build
```
ระบบ Backend จะรันอยู่ที่ `http://localhost:8080`

## 8. Test Credentials

**Admin Profile**
* Email: `admingoldenstage@gmail.com`
* Password: `admingoldenstage`

**User Profile**
* Email: `mockuser01@gmail.com`
* Password: `mockuser01`

## 9. API Reference

**Public Endpoints (ไม่ต้อง Login)**
* `GET /api/movies` : ดึงรายชื่อภาพยนตร์ทั้งหมดสำหรับหน้า Landing Page
* `GET /api/movies/:movie_id` : ดึงรายละเอียดของภาพยนตร์ 1 เรื่อง
* `GET /api/cinemas` : ดึงรายชื่อสาขาโรงภาพยนตร์ทั้งหมด (Master Data)
* `GET /api/cinemas/:cinema_id/halls` : ดึงรายชื่อโรงฉายของสาขาที่เลือก (Master Data)
* `GET /api/movies/:movie_id/showtimes` : ดึงรายการรอบฉายของภาพยนตร์ที่เลือก

**Secured Endpoints (ต้องแนบ Authorization: Bearer <Firebase_Token>)**
* `GET /api/bookings/me` : ดึงประวัติการจองตั๋วของผู้ใช้งาน
* `GET /api/showtimes/:showtime_id/seats` : ดึงผังที่นั่งและสถานะ (AVAILABLE, LOCKED, BOOKED)
* `POST /api/bookings/lock` : ส่งคำสั่งล็อกที่นั่งลง Redis (TTL 5 นาที) ป้องกัน Double Booking
* `POST /api/bookings/confirm` : ยืนยันการชำระเงิน บันทึกลง MongoDB และส่ง Event เข้า RabbitMQ

**Real-time Communication**
* `WS /ws/seats/:showtime_id` : WebSocket สำหรับรับข้อมูลอัปเดตสถานะที่นั่งแบบ Real-time

---

### วิธีการทดสอบ API ด้วย Postman

เนื่องจากเราปรับสถาปัตยกรรมไปใช้ `ObjectID` ของ MongoDB เป็น Best Practice แล้ว การทดสอบผ่าน Postman จะต้องทำตามขั้นตอนดังนี้:

**Note:** ก่อนทดสอบ API กรุณารันคำสั่ง `go run scripts/seed.go` เพื่อจำลองข้อมูล Master Data (Movies, Cinemas, Showtimes) ลงในฐานข้อมูลของท่านก่อนครับ

**1. หา ObjectID จากฐานข้อมูลจริง**
เปิดโปรแกรม **MongoDB Compass** เข้าไปที่ฐานข้อมูล `golden_stage_db` 
- เข้าไปที่ Collection `movies` ก๊อปปี้ค่า `_id` ของเรื่องที่ต้องการ (รหัส 24 ตัวอักษร เช่น `6a2be2...`) นำไปวางแทน `:movie_id`
- นำไปใช้เรียก `GET /api/movies/:movie_id/showtimes` เพื่อดูรอบฉายทั้งหมด
- จากผลลัพธ์รอบฉาย ให้ก๊อปปี้ค่า `id` ของรอบฉาย นำไปวางแทน `:showtime_id`
- นำไปใช้เรียก `GET /api/showtimes/:showtime_id/seats` เพื่อดูผังที่นั่ง

**2. การยิง API ที่ถูกล็อก (Secured Endpoints)**
สำหรับ API `/api/bookings/lock` และ `/api/bookings/confirm` ต้องใช้สิทธิ์ผู้ใช้:
- ใน Postman ให้ไปที่แท็บ **Authorization**
- เลือก Type เป็น **Bearer Token**
- นำ Firebase Token (ที่ได้จากการจำลองล็อกอิน) มาใส่ในช่อง Token
- ในแท็บ **Body** เลือก raw และเป็น `JSON` ตัวอย่างเช่น:
```json
{
  "showtime_id": "นำ _id ของรอบฉายมาใส่ตรงนี้",
  "seat_number": "A1"
}
```
