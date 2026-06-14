# 🛠️ Code Walkthrough: Real-time Seat Booking & UI Enhancements

เอกสารนี้รวบรวมรายละเอียดการอัปเดตโค้ดล่าสุด เพื่อแก้ไขปัญหาเรื่อง Redis Lock Leak, การทำงานของ WebSocket ที่ไม่สมบูรณ์ และการปรับปรุงสี UI ของผังที่นั่งให้ตรงตาม Design System

---

## 1. การแก้ไข Redis Lock Leak (ปลดล็อกที่นั่งคืนระบบ)

ปัญหาเดิมคือเมื่อผู้ใช้อยู่ในหน้าจองที่นั่งและล็อกที่นั่งไว้ (สถานะ `LOCKED` บน Redis) แต่กดย้อนกลับ หรือกดยกเลิกการเลือก ระบบไม่ได้คืนที่นั่งนั้นทันที ทำให้ต้องรอจนกว่า Redis จะ Timeout (5 นาที)

**การแก้ไข:**
*   **Backend (`features/bookings/controller.go`):** 
    *   สร้างฟังก์ชัน `UnlockSeat` เพื่อรับ Request ปลดล็อก โดยตรวจสอบก่อนว่า `userID` ที่ส่งมาตรงกับคนที่ถือ Lock ไว้หรือไม่ ถ้าตรงจะสั่ง `RedisClient.Del` เพื่อลบ Key ออกจาก Redis ทันที
    *   หลังจากลบสำเร็จ จะมีการส่ง Message (สถานะ `AVAILABLE`) เข้า RabbitMQ ทันทีเพื่อให้ WebSocket นำไปกระจายต่อ
*   **Backend (`features/bookings/routes.go`):** 
    *   เพิ่ม Route `DELETE /api/bookings/lock/:showtime_id/:seat_number`
*   **Frontend (`SeatMapView.vue`):**
    *   **การยกเลิกการเลือก (Deselect):** ในฟังก์ชัน `toggleSeat` เมื่อผู้ใช้กดที่เก้าอี้ที่เลือกไว้แล้ว ระบบจะยิง API `DELETE /lock/...` เพื่อคืนที่นั่ง
    *   **การออกจากหน้าจอ:** ในฟังก์ชัน `releaseSelectedSeats` ที่ผูกกับ Lifecycle `onBeforeUnmount` จะมีการ Loop ดูรายการเก้าอี้ที่ผู้ใช้เลือกค้างไว้ และยิง API `DELETE /lock/...` เพื่อคืนที่นั่งให้ทั้งหมดก่อนที่ Component จะถูกทำลาย

---

## 2. การอัปเดต UI สีเก้าอี้และ Legend

ผู้ใช้ต้องการให้สีของเก้าอี้และคำอธิบาย (Legend) ตรงกับ Design System ที่กำหนดไว้

**การแก้ไขใน `SeatMapView.vue`:**
*   อัปเดต Tailwind classes ในส่วนการ Render ปุ่มเก้าอี้ (`<button>`) ให้ตรงตามนี้:
    *   `AVAILABLE`: `bg-muted/60` (สีเทาเข้ม)
    *   `SELECTED`: `bg-yellow-500` (สีเหลือง/ทอง)
    *   `LOCKED`: `bg-red-500 text-white` (สีแดง)
    *   `BOOKED`: `bg-muted opacity-50` (สีเทาทึบแสง)
*   ปรับปรุงส่วน **Legend** ด้านล่างให้ครอบคลุมสถานะทั้งหมด 5 แบบ (Available, Selected, Locked, Booked, Premium) เพื่อให้ผู้ใช้เข้าใจสถานะที่นั่งได้อย่างชัดเจน

---

## 3. สถาปัตยกรรม Real-time WebSocket (Hub Pattern Fan-out)

ปัญหาเดิมคือเมื่อเปิดหน้าจอทดสอบ 2 หน้าต่างแล้วกดจองที่นั่ง หน้าต่างที่ 2 ไม่เกิดการเปลี่ยนสีแบบ Real-time เนื่องจาก RabbitMQ กระจาย Message แบบ Round-Robin ให้ Consumer แค่ตัวเดียว และปัญหาอื่นๆ ระหว่างทาง

**การแก้ไข Backend (Go):**
*   **API `POST/DELETE /lock`:** เพิ่มโค้ดส่วนของการ Publish Message แจ้งสถานะ (`LOCKED` หรือ `AVAILABLE`) เข้า RabbitMQ ทันทีเมื่อ Redis เคลียร์ล็อกหรือสร้างล็อกสำเร็จ โดยในจังหวะ `POST` มีการแนบตัวแปร `user_id` เข้าไปใน Payload แจ้งเตือนด้วย
*   **RabbitMQ `Fanout Exchange`:** รื้อโครงสร้าง Queue ธรรมดา และเปลี่ยนมาประกาศเป็น `Fanout Exchange` (`seat_updates_ex`) แทน เพื่อรับประกันการ Broadcast สู่ทุก Node โดยไม่สนว่าจะมีใครรอรับกี่คน
*   **WebSocket Hub (`websocket.go`):** 
    *   เปลี่ยนมาใช้สถาปัตยกรรม **Hub Pattern** แบ่งกลุ่ม Client ที่ต่อเข้ามาตาม `showtimeID`
    *   มี Goroutine ทำหน้าที่เป็น Consumer คอยสร้าง Temporary/Exclusive Queue ออกมาดักรอข้อความจาก Fanout เพื่อกระจาย (`WriteJSON`) กลับไปหาทุกคนในห้องรอบฉายนั้นๆ

**การแก้ไข Frontend (Vue):**
*   **การจัดการ `ws.onmessage`:** เมื่อมีข้อความเข้ามา ระบบจะเช็ก `user_id` (ที่เพิ่งเพิ่มเข้าไป) ว่าตรงกับ `authStore.user.uid` ปัจจุบันหรือไม่ หากข้อความเป็น `LOCKED` แต่เป็นของตัวเราเอง ระบบจะแปลงสถานะเป็น `SELECTED` เพื่อแก้ปัญหา WebSocket Boomerang (ข้อความกลับมาทับทำให้ปุ่มโดน Disable)
*   **Optimistic Update ทันทีทันใด:** ในฟังก์ชัน `toggleSeat` มีการปรับให้เปลี่ยน State ใน Store ทันทีที่ผู้ใช้คลิกโดยไม่รอให้ API Response (ทำคู่กับ `await api.delete()`) หากเกิดข้อผิดพลาดค่อย Revert ข้อมูลกลับ ทำให้การตอบสนองของ UI รวดเร็วที่สุดในมุมมองของผู้ใช้งาน
*   **Reactivity Optimization:** ปรับแต่งโครงสร้าง `movieStore.updateSeatStatus` ใน `useMovieStore.ts` ให้ทำการ Re-assign ค่า Object ภายในอาร์เรย์ (`...seats.value[index]`) เพื่อให้ Vue ดักจับ Deep Reactivity ได้แม่นยำ รวมถึงเซ็ตค่าเริ่มต้นของ `selectedSeatIds` ทันทีตั้งแต่ `onMounted`

ด้วยการปรับปรุงทั้ง 3 ส่วนนี้ ระบบการจองที่นั่งของ Golden Stage Cinema จึงมีความสมบูรณ์แบบ ทั้งในด้าน UI/UX และระบบ Backend Real-time Sync แบบมืออาชีพครับ 🚀
