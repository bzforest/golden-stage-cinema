# ==========================================
# Stage 1: Build (Compile Go Binary)
# ==========================================
FROM golang:1.26-alpine AS builder

WORKDIR /app

# ติดตั้ง Git สำหรับ Go modules ที่ดึงจาก Git repository
RUN apk add --no-cache git

# Copy go.mod และ go.sum ก่อนเพื่อ Cache Layer ของ Dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy Source Code ทั้งหมด
COPY . .

# Compile เป็น Binary ขนาดเล็ก (ปิด CGO เพื่อให้รันบน Alpine ได้)
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server .

# Compile Seed Script แยกต่างหาก
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/seed ./scripts/seed.go

# ==========================================
# Stage 2: Run (Lightweight Alpine Image)
# ==========================================
FROM alpine:latest

WORKDIR /app

# ติดตั้ง CA Certificates สำหรับ HTTPS (Firebase SDK ต้องใช้)
RUN apk add --no-cache ca-certificates tzdata

# Copy Binary จาก Builder Stage
COPY --from=builder /app/server .
COPY --from=builder /app/seed .

# Copy Firebase Service Account Key (Secret)
COPY firebase-service-account.json .

# Copy Entrypoint Script
COPY entrypoint.sh .
RUN sed -i 's/\r$//' entrypoint.sh && chmod +x entrypoint.sh

EXPOSE 8080

CMD ["./entrypoint.sh"]
