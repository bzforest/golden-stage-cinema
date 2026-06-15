#!/bin/sh

# รันคำสั่ง Seed เพื่อเช็กและจำลองข้อมูลเบื้องต้น (จะข้ามการทำงานอัตโนมัติหากมีข้อมูลอยู่แล้ว)
echo "Running auto-seeder..."
./seed

# เริ่มการทำงานของ Server ตัวจริง
echo "Starting Go Backend Server..."
exec ./server
