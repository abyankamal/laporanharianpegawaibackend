FROM golang:alpine AS builder

WORKDIR /app

# Meng-copy file modul Go dan download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Meng-copy seluruh source code
COPY . .

# Build aplikasi Go
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/api-server ./cmd/api

# Stage 2: Image yang lebih kecil untuk menjalankan aplikasi
FROM alpine:latest

# Install tzdata untuk timezone Asia/Jakarta dan ca-certificates
RUN apk add --no-cache tzdata ca-certificates

ENV TZ=Asia/Jakarta

WORKDIR /app

# Meng-copy binary dari stage builder
COPY --from=builder /app/api-server .

# Meng-copy file pendukung
COPY --from=builder /app/.env .
COPY --from=builder /app/serviceAccountKey.json .
COPY --from=builder /app/images ./images

# Membuat direktori uploads jika belum ada
RUN mkdir -p ./uploads/photos ./uploads/reports ./uploads/attendance ./uploads/leave

# Expose port aplikasi
EXPOSE 5000

# Perintah untuk menjalankan aplikasi
CMD ["./api-server"]
