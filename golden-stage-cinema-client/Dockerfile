# ==========================================
# Stage 1: Build (Compile Vue App)
# ==========================================
FROM node:22-alpine AS builder

WORKDIR /app

# Copy package files ก่อนเพื่อ Cache Layer ของ Dependencies
COPY package.json package-lock.json* ./
RUN npm ci

# Copy Source Code ทั้งหมด
COPY . .

# รับค่า Firebase Config ผ่าน Build Arguments
# ARG VITE_API_BASE_URL=http://localhost:8080/api
# ARG VITE_FIREBASE_API_KEY
# ARG VITE_FIREBASE_AUTH_DOMAIN
# ARG VITE_FIREBASE_PROJECT_ID
# ARG VITE_FIREBASE_STORAGE_BUCKET
# ARG VITE_FIREBASE_MESSAGING_SENDER_ID
# ARG VITE_FIREBASE_APP_ID

# ตั้งค่า Environment Variables สำหรับ Vite Build
# ENV VITE_API_BASE_URL=$VITE_API_BASE_URL
# ENV VITE_FIREBASE_API_KEY=$VITE_FIREBASE_API_KEY
# ENV VITE_FIREBASE_AUTH_DOMAIN=$VITE_FIREBASE_AUTH_DOMAIN
# ENV VITE_FIREBASE_PROJECT_ID=$VITE_FIREBASE_PROJECT_ID
# ENV VITE_FIREBASE_STORAGE_BUCKET=$VITE_FIREBASE_STORAGE_BUCKET
# ENV VITE_FIREBASE_MESSAGING_SENDER_ID=$VITE_FIREBASE_MESSAGING_SENDER_ID
# ENV VITE_FIREBASE_APP_ID=$VITE_FIREBASE_APP_ID

# Build Production Bundle (ข้าม type-check เพื่อความเร็ว)
RUN npm run build-only

# ==========================================
# Stage 2: Serve (Nginx Alpine)
# ==========================================
FROM nginx:alpine

# Copy ไฟล์ Static ที่ Build ได้จาก Stage 1
COPY --from=builder /app/dist /usr/share/nginx/html

# Copy Nginx Config สำหรับ SPA Routing
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
