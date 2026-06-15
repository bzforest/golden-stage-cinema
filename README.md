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

## 3. Real-time Seat Booking Flow

ระบบการจองที่นั่งแบบ Real-time ถูกออกแบบมาเพื่อแก้ปัญหา Double Booking และเพิ่มประสบการณ์ผู้ใช้ที่ลื่นไหล (Seamless UI) โดยอาศัยการทำงานร่วมกันของ **Vue.js + Go + Redis + RabbitMQ + WebSockets** 

**Flow การทำงาน:**
1. **User เลือกรอบฉายและที่นั่ง (Single Source of Truth):** Frontend (Vue) ดึงผังที่นั่งจาก API `GET /api/showtimes/:showtime_id/seats` ซึ่ง Backend (Go) จะประมวลผลข้อมูลจาก 3 แหล่ง:
   - ผังที่นั่งดั้งเดิม (Showtime Seats)
   - ประวัติการจองจริง (Bookings) เพื่อหาสถานะ `BOOKED` อย่างแม่นยำ (ป้องกันปัญหาข้อมูลไม่ตรงกัน)
   - ข้อมูลการล็อกชั่วคราว (Redis) 
2. **กดเลือกเก้าอี้ (Optimistic Update):** 
   - ทันทีที่คลิก Vue จะเปลี่ยนสีเก้าอี้เป็นสีเหลืองและยิง API `POST /api/bookings/lock` หรือ `DELETE` ไปยัง Backend
   - หาก API ตอบกลับ `200 OK` (หรือการทำงานระดับ Local state สมบูรณ์) ระบบจะทำการอัปเดต UI ให้ทันทีโดยไม่ต้องรอ WebSocket ขาตั้งรับ (Optimistic Update) ให้ความรู้สึกรวดเร็ว
3. **การล็อคระดับระบบ (Redis Distributed Lock):**
   - Backend ใช้ **Redis SETNX** สร้าง Key ล็อคที่นั่ง (มี TTL 5 นาที) เพื่อกันคนอื่นแย่งกด
4. **กระจายข่าว (RabbitMQ Fanout):**
   - หลังล็อคสำเร็จ Backend จะ `Publish` ข้อความแจ้งสถานะ (เช่น `LOCKED`) พร้อมแนบ `user_id` เข้าสู่ Exchange แบบ **Fanout** (`seat_updates_ex`)
5. **แจ้งเตือนทุกคน (WebSockets):**
   - ฝั่ง WebSocket Hub จะสร้าง Exclusive Queue ขึ้นมาผูกกับ Connection เพื่อรับข้อความจาก Fanout และสั่ง Broadcast ข้อความนั้นไปยัง Client ทุกหน้าที่กำลังดูรอบฉายเดียวกัน
6. **Vue อัปเดต UI ทันที:**
   - เมื่อ `ws.onmessage` ฝั่ง Client ได้รับข้อความ จะทำการแยกแยะ `user_id`
   - หากเป็นเก้าอี้ที่ตนเองเพิ่งกด (Boomerang) จะเปลี่ยนสถานะเป็น `SELECTED`
   - หากเป็นของคนอื่น จะตีความเป็น `LOCKED` (เปลี่ยนเป็นสีแดง) ป้องกันความสับสน
   - หากสถานะเป็น `AVAILABLE` ก็จะปลดล็อคสีกลับเป็นปกติให้คนอื่นกดได้ทันที
   - **(New!) Reconnection Sync:** หากเน็ตผู้ใช้หลุดแล้วกลับมาเชื่อมต่อใหม่ (`ws.onopen`) ระบบจะยิง API ดึงข้อมูลผังที่นั่งทั้งหมดจากฐานข้อมูลมาทับหน้าจออีกครั้ง เพื่ออุดรอยรั่วจากการพลาด Message ช่วงเน็ตหลุด
7. **การคืนเก้าอี้กลับระบบ (Timeout & Explicit Unlock):**
   - **Explicit Unlock:** เมื่อผู้ใช้กดย้อนกลับ/เปลี่ยนหน้า (`onBeforeUnmount`) หรือกดเปลี่ยนใจ (Deselect) ระบบจะยิง `DELETE /api/bookings/lock` คืนเก้าอี้กลับสู่ระบบทันที
   - **Auto Expired:** หากผู้ใช้ปิดเบราว์เซอร์กะทันหัน Redis TTL จะหมดอายุอัตโนมัติใน 5 นาที และ `timeout_listener` จะทำงานเพื่อ Broadcast สถานะ `AVAILABLE` กลับให้ทุกคนรับทราบพร้อมกัน
8. **การชำระเงินและ Bulk Confirmation (All-or-Nothing):**
   - เมื่อผู้ใช้ยืนยันการชำระเงิน Frontend จะมัดรวมเก้าอี้ทั้งหมด (Array) แล้วส่ง Request ไปยัง `POST /api/bookings/confirm` เพียงครั้งเดียว
   - Backend จะทำ **Pre-validation** โดยเช็ก Lock ใน Redis ของทุกที่นั่งพร้อมกัน หากพบว่ามีแม้แต่ที่นั่งเดียวที่ "หลุด Lock" หรือ "ไม่ใช่ของผู้ใช้คนนี้" ระบบจะทำการ **Abort All** (ยกเลิกทั้งตะกร้า) ทันที
   - หากผ่าน Pre-validation ระบบจะทำการบันทึกแบบ **Atomic All-or-Nothing** ด้วยคำสั่ง `UpdateOne` โดยใช้ทริค `$nin` เพื่อเช็กซ้ำในระดับ Data Layer หากมีคนแย่งจองในเสี้ยววินาที ระบบจะตีตกและ Rollback ทั้งหมดทันที
   - หากปลอดภัย 100% ระบบจะสร้างใบเสร็จด้วย `InsertMany`, ปลด Lock ใน Redis คืน, และวนลูป Broadcast ผ่าน RabbitMQ เพื่ออัปเดตสถานะให้ทุก Client เป็น `BOOKED` พร้อมกัน
9. **Bot Protection & Rate Limiting:**
   - การยิง API เพื่อกดล็อกเก้าอี้แต่ละครั้งถูกคุมความประพฤติด้วย **Redis Rate Limiting** (จำกัด 20 ครั้ง/นาที)
   - การ Confirm จองถูกตีกรอบด้วย `Binding Validation` ป้องกันการจ่ายเงินเกิน 10 ที่นั่งต่อรอบบิล ช่วยรับมือคนป่วนระบบ

## 4. Redis Lock Strategy

เพื่อป้องกันปัญหาจองที่นั่งซ้ำซ้อน (Double Booking) ขณะที่ผู้ใช้หลายคนกดเลือกที่นั่งพร้อมกัน ระบบจึงเลือกใช้ **Redis Distributed Lock (SETNX - Set if Not Exists)**
- **ความเร็ว:** Redis ทำงานบน Memory การเช็กว่าที่นั่งถูกล็อกหรือยังทำได้รวดเร็วมากระดับมิลลิวินาที
- **Atomic Operation:** คำสั่ง SETNX การันตีได้ว่าถ้ามีคำสั่ง 2 ตัวเข้ามาพร้อมกัน ตัวหนึ่งจะสำเร็จ (True) และอีกตัวจะล้มเหลว (False) แน่นอน ช่วยตัดปัญหาสภาพการแข่งขัน (Race Condition)
- **TTL (Time To Live):** ล็อกมีการกำหนดเวลาตายชัดเจนที่ 5 นาที หากกระบวนการฝั่งแอปมีปัญหา (เช่น เน็ตหลุด แอปแครช) ล็อกก็จะถูกปลดคืนอัตโนมัติโดยไม่ค้างในระบบ

## 5. Message Queue

ระบบใช้ **RabbitMQ** เป็นศูนย์กลางในการสื่อสารเพื่อกระจาย Event โดยไม่ต้องพึ่งพา API หลัก (Decoupling) ประโยชน์หลัก 2 ข้อ:
1. **Real-time Seat Updates (Hub Pattern):** เมื่อมีการ Lock, Unlock, Confirm หรือ Timeout สถานะเก้าอี้จะถูก Publish ลงคิว `seat_updates` โดยฝั่ง WebSocket จะมี Consumer กลาง (Hub) รับข้อความและทำการกระจาย (Fan-out) ไปให้ Client ทุกการเชื่อมต่อที่ดูรอบฉายเดียวกันพร้อมๆ กัน ทำให้เห็นที่นั่งเด้งสลับสถานะแบบ Real-time สมบูรณ์แบบ
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

**การตั้งค่า Frontend:**
ก่อนเริ่มรัน Frontend กรุณาคัดลอกไฟล์ `golden-stage-cinema-client/.env.example` เป็น `.env` และกรอกค่า Firebase Configuration ให้ครบถ้วน จากนั้นใช้คำสั่ง:
```bash
cd golden-stage-cinema-client
npm install
npm run dev
```

## 8. Test Credentials & Authentication

ระบบใช้ **Firebase Authentication** แบบ Client-side สำหรับผู้ตรวจระบบ (Examiners) สามารถใช้บัญชีที่เตรียมไว้ให้ด้านล่างนี้ล็อกอินเข้าใช้งานได้ทันทีครับ:

**Admin Profile**
* Email: `admingoldenstage@gmail.com`
* Password: `admingoldenstage`

**User Profile**
* Email: `mockuser01@gmail.com`
* Password: `mockuser01`

*(หรือสามารถทดลองสร้างบัญชีใหม่ได้เองผ่านเมนู Register บนหน้าเว็บเพื่อใช้งานในสิทธิ์ผู้ใช้ทั่วไป)*

---

**สำหรับนักพัฒนา (Developer Tools):**
หากคุณเป็นผู้พัฒนาและต้องการให้สิทธิ์ Admin แก่บัญชีอื่นๆ เพิ่มเติม สามารถทำได้โดย:
1. ไปที่หน้าเมนู Authentication ใน Firebase Console ของคุณเพื่อคัดลอก `UID`
2. นำ `UID` ไปกำหนดในไฟล์ `golden-stage-cinema-server/.env` (ตัวแปร `ADMIN_UID=...`) 
   หรือ รันคำสั่งพร้อมแนบ UID ต่อท้าย: `go run scripts/set_admin.go <UID>`
ระบบจะทำการตั้งค่า Custom Claims (`role: "admin"`) ให้กับบัญชีนั้นทันที

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
* `POST /api/bookings/lock` : ส่งคำสั่งล็อกที่นั่งลง Redis (TTL 5 นาที) และส่ง Event LOCKED เข้า WebSocket
* `DELETE /api/bookings/lock/:showtime_id/:seat_number` : สั่งปลดล็อกที่นั่งคืนระบบและส่ง Event AVAILABLE เข้า WebSocket
* `POST /api/bookings/confirm` : ยืนยันการชำระเงิน บันทึกลง MongoDB และส่ง Event BOOKED เข้า RabbitMQ

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

## 10. Admin System & Global Search

ระบบหลังบ้านถูกออกแบบมาให้ผู้ดูแลระบบสามารถตรวจสอบประวัติการจองและบันทึกเหตุการณ์ที่สำคัญได้ (Audit Logs) อย่างละเอียดผ่านหน้า Admin Dashboard โดยมีโครงสร้างและฟีเจอร์เด่นดังนี้:

### 1. ระบบจัดการสถานะ & ตรวจสอบสิทธิ์ Admin (Custom Claims)
- อาศัย **Firebase Custom Claims** (`role: admin`) ที่ฝังอยู่ใน Token
- **Route Guard:** ฝั่ง Frontend มีการสร้าง `requiresAdmin` เพื่อป้องกันไม่ให้ผู้ใช้งานทั่วไปสามารถเข้าถึงหน้า Dashboard ของ Admin (`/admin`) ได้ โดยจะถูก Redirect กลับไปยังหน้าโฮมเพจทันที

### 2. โครงสร้าง Admin Dashboard
- **Tabs:** แบ่งหน้าจอเป็น 2 ส่วน: "All Bookings" (ประวัติการจองทั้งหมด) และ "Audit Logs" (บันทึกเหตุการณ์ต่างๆ ในระบบ)
- **Data Table & Pagination:** รองรับการแสดงผลข้อมูลจำนวนมหาศาลได้อย่างมีประสิทธิภาพโดยใช้ Server-Side Pagination
- **Audit Logs Injection:** ฐานข้อมูลจะถูกบันทึกประวัติการกระทำอัตโนมัติลงในคอลเล็กชัน `audit_logs` ทุกครั้งที่มีการเรียกใช้งานฟังก์ชันสำคัญ เช่น `LockSeat` หรือ `ConfirmBooking`

### 3. Server-Side Global Search (Manual Execution)
- **Global Search:** อัปเกรดช่องค้นหาให้สามารถค้นหาข้อมูลข้ามระบบได้ครอบคลุมทุกมิติ ทั้งผ่านอีเมลผู้ใช้, ชื่อผู้ใช้ (ดึงจาก Firebase สดๆ), ชื่อภาพยนตร์, รหัส Booking ID, สถานะ, และวันที่จอง แบบ Server-side เต็มรูปแบบ
- **Aggregation Pipeline:** ฝั่ง Backend มีการแปลงฟิลด์ `_id` และ `created_at` ให้กลายเป็น String ด้วย `$addFields` เพื่อใช้ค้นหาผ่าน `$regex` 
- **Manual Search Execution:** ค้นหาโดยกดปุ่ม "Search" หรือปุ่ม Enter บนคีย์บอร์ด เพื่อยิง Request ไปประมวลผลที่ Backend ในครั้งเดียว ช่วยลดทราฟฟิกที่ไม่จำเป็นและลดภาระเซิร์ฟเวอร์

### 4. Data Enrichment & Audit Logs UI
- **Smart Data Extraction:** ฝั่ง Backend มีการแกะ (Parse) ข้อความ Log ด้วย Regex เพื่อดึง `UID`, `Seat` และ `Showtime_ID` แล้วไปดึงชื่อและอีเมลผู้ใช้จาก Firebase และข้อมูลภาพยนตร์จาก MongoDB แบบอัตโนมัติ เพื่อประกอบเป็นข้อมูลที่สมบูรณ์ก่อนส่งให้ Frontend
- **Badges UI:** นำข้อมูลที่ผ่านการ Enrich มาแสดงผลเป็น Badge สีสันสวยงาม อ่านง่าย (เช่น `[User: Admin]`, `[Seat: B5]`) แทนที่การแสดงผลข้อความดิบ
- **Detailed Modal:** จัดทำหน้าต่าง Modal สำหรับดูรายละเอียดเชิงลึกของแต่ละ Log แบบเจาะลึกครบทุกมิติ (พร้อมข้อมูล Raw JSON/Text ต้นฉบับ)

### 5. การยกระดับประสิทธิภาพเซิร์ฟเวอร์ขั้นสูง (Performance & Optimization)
เพื่อเตรียมพร้อมรองรับจำนวนผู้ใช้งานและปริมาณข้อมูลมหาศาล ระบบ Admin ได้รับการผ่าตัดโครงสร้างครั้งใหญ่เพื่อขจัดคอขวด:
- **กำจัด N+1 Query Problem:** ยกเลิกการเรียก API ของ Firebase และ MongoDB แบบวนลูป (For Loop) ในขณะแสดงผลตาราง แต่อัปเกรดไปใช้การดึงข้อมูลแบบ **Batch Fetching** (`GetUsers()` สำหรับ Firebase และ `$in` สำหรับ MongoDB) ซึ่งช่วยลดจำนวน Request จากหลัก 100+ ครั้ง เหลือเพียง 2-3 ครั้งเท่านั้นต่อการโหลดหนึ่งหน้า
- **ระบบ Global Cache (Redis):** นำ Redis ที่มีอยู่แล้วมาทำหน้าที่เป็นหน่วยความจำแคชสำหรับโปรไฟล์ของผู้ใช้งาน (Email, Display Name) โดยกำหนดอายุ 24 ชั่วโมง (TTL) เพื่อลดภาระการร้องขอข้อมูลไปยังเซิร์ฟเวอร์ Firebase ทุกครั้งที่มีการโหลดหน้าใหม่
- **แก้ปัญหา Index Scan Defeat:** ยกเลิกการใช้ `$addFields` เพื่อแปลงข้อมูล Database เป็น Text แบบกว้างๆ ที่ทำให้ฟังก์ชันค้นหาช้าลง และเปลี่ยนมาเช็ก/แปลง Format ในฝั่ง Go (เช่น ตรวจสอบความยาว ObjectID หรือ Parse วันที่ Date Format) เพื่อส่งคำสั่งค้นหาแบบตรงเป๊ะไปให้ MongoDB ทำให้ใช้งาน **Database Index** ได้เต็ม 100% ตอบสนองเร็วทันใจแม้มีข้อมูลหลักแสน
- **Database Schema Optimization:** แก้ไขปัญหาโครงสร้างคอลเล็กชัน `audit_logs` ผิดรูป โดยแยกย่อยฟิลด์ต่างๆ (`UID`, `ShowtimeID`, `SeatNumber`) ไว้ตั้งแต่ตอนสร้าง Log แทนที่จะเก็บเป็น String ยาวๆ เพื่อให้ฐานข้อมูลสามารถ Query หาเป้าหมายได้โดยตรง ลดการใช้ CPU ของเซิร์ฟเวอร์จากการ Parse ข้อความด้วย Regular Expression (Regex) 

## 11. User Profile & Watch History

ระบบหน้าต่างโปรไฟล์ถูกสร้างขึ้นเพื่อมอบประสบการณ์การใช้งานส่วนตัว (Personalized Experience) ให้กับผู้ใช้แต่ละคน โดยเชื่อมต่อกับฐานข้อมูลเพื่อแสดงผลแบบเรียลไทม์

### 1. โครงสร้างหน้า User Profile
- **Upcoming Screenings (My Tickets):** แสดงรายการตั๋วหนังและที่นั่งที่ผู้ใช้ทำการจองล่วงหน้าซึ่งยังไม่ถึงเวลาฉายจริง โดยมีข้อมูลชื่อเรื่อง, โรงภาพยนตร์, เวลาฉาย และเบอร์เก้าอี้ ครบถ้วน
- **Past Movies (Watch History):** แสดงประวัติภาพยนตร์ที่ผู้ใช้เคยรับชมไปแล้ว (เวลาฉายในอดีต) 
- **Account Settings:** สำหรับตรวจสอบข้อมูลส่วนตัวเบื้องต้น เช่น Display Name และ Email

### 2. Logic การทำงาน
- **Group Bookings:** ทำการรวบรวม (Group) ตั๋วหนังหลายใบที่จองในรอบฉายเดียวกัน ให้อยู่ในรูปแบบ Card ใบเดียวที่มี `Seats: [A1, A2]` เพื่อลดความซ้ำซ้อนของการแสดงผล
- **Time Comparison:** ใช้ Computed Properties ของ Vue ในการเปรียบเทียบเวลา (Time Comparison) กับเวลาปัจจุบัน เพื่อคัดแยกตั๋วที่ยังไม่ฉาย (Upcoming) และตั๋วที่ฉายจบแล้ว (Past History) แยกไปตามแต่ละแท็บอย่างแม่นยำ
