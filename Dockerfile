# syntax=docker/dockerfile:1

# =====================================================================
# Stage 1 — build
# Binary dikompilasi di dalam image golang resmi. CGO dimatikan supaya
# binary statis bisa berjalan di runtime alpine tanpa libc tambahan.
# =====================================================================
FROM golang:1.26-alpine AS build

WORKDIR /src

# go.mod/go.sum dicopy lebih dulu supaya layer cache modul dipakai ulang
# setiap kali hanya kode aplikasi yang berubah.
COPY go.mod go.sum ./
RUN go mod download

# Sisanya: seluruh sumber aplikasi.
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/app .

# =====================================================================
# Stage 2 — runtime
# Image akhir hanya berisi binary aplikasi dan folder logs. Aplikasi
# berjalan sebagai user non-root "app".
# =====================================================================
FROM alpine:3.21

# ca-certificates: koneksi TLS keluar. tzdata: tanggal lokal WIB.
RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S app \
    && adduser -S -G app app

WORKDIR /app

# Binary hasil build + folder log milik user app.
COPY --from=build /out/app /app/app
RUN mkdir -p /app/logs && chown -R app:app /app

USER app

EXPOSE 3000

# Health check didefinisikan per service di docker-compose.yml:
# api memakai endpoint HTTP, worker memakai pemeriksaan proses.

ENTRYPOINT ["/app/app"]
