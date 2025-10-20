# File: Dockerfile (Versi yang sudah disesuaikan)

# --- Tahap 1: Build/Compile Aplikasi Go ---
FROM golang:1.24-alpine AS builder

WORKDIR /src

# Salin file dependensi terlebih dahulu untuk optimasi cache Docker
# Ini akan menyalin go.mod dan go.sum dari root proyek Anda
COPY go.mod go.sum ./
RUN go mod download

# Salin semua sisa kode sumber dari proyek Anda ke dalam container
# Ini akan menyalin folder cmd, internal, modules, dll.
COPY . .

# Compile aplikasi Go Anda dengan menunjuk langsung ke file main.go
# Perhatikan path di akhir: ./cmd/app/main.go
# Ini memberitahu Go di mana letak entry point aplikasi Anda.
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main ./cmd/app/main.go

# --- Tahap 2: Final Image yang Ramping ---
FROM alpine:latest

WORKDIR /app

# Salin HANYA binary yang sudah dicompile dari tahap 'builder'
COPY --from=builder /app/main .

# Expose port yang digunakan aplikasi (ambil dari .env)
# Pastikan GOWA_PORT Anda di .env adalah 8000 atau sesuaikan di sini
EXPOSE 8000

# Perintah untuk menjalankan aplikasi Anda
CMD ["./main"]