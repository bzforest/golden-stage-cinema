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

(รอเติมข้อมูล)

## 4. Redis Lock Strategy

(รอเติมข้อมูล)

## 5. Message Queue

(รอเติมข้อมูล)

## 6. Assumptions & Trade-offs

ผมทราบในเรื่องของ Best Practice (เช่น การลด Global Variables, การแยก Service/Repository Layer, การทำ Centralized Config, หรือการทำ Standardized Response) ว่ามีหนทางที่ดีกว่านี้ 
แต่ด้วยโค้ดที่มีความซับซ้อนเพิ่มมากขึ้นจนเกินขอบเขตความเข้าใจของผมในตอนนี้ ทำให้ผมเลือกที่จะใช้ทางที่ระดับความรู้ตอนนี้ของผมสามารถเข้าใจและสามารถอธิบายได้ดีที่สุดครับ

## 7. วิธีรันระบบ

(รอเติมข้อมูล)

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
* `GET /api/movies/:movie_id/showtimes` : ดึงรายการรอบฉายของภาพยนตร์ที่เลือก

**Secured Endpoints (ต้องแนบ Authorization: Bearer <Firebase_Token>)**
* `GET /api/showtimes/:showtime_id/seats` : ดึงผังที่นั่งและสถานะ (AVAILABLE, LOCKED, BOOKED)
* `POST /api/bookings/lock` : ส่งคำสั่งล็อกที่นั่งลง Redis (TTL 5 นาที) ป้องกัน Double Booking
* `POST /api/bookings/confirm` : ยืนยันการชำระเงิน บันทึกลง MongoDB และส่ง Event เข้า RabbitMQ

**Real-time Communication**
* `WS /ws/seats/:showtime_id` : WebSocket สำหรับรับข้อมูลอัปเดตสถานะที่นั่งแบบ Real-time
