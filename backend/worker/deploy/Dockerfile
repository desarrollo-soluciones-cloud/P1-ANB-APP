# --- Etapa 1: Builder ---
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Compilar SOLO el Worker (main.go está en la raíz del paquete)
RUN CGO_ENABLED=0 GOOS=linux go build -o /worker_server ./main.go

# --- Etapa 2: Runtime ---
FROM alpine:latest

# Instalar FFmpeg para procesamiento de videos y certificados SSL para AWS
RUN apk add --no-cache ffmpeg ca-certificates

WORKDIR /app

# Copiar SOLO el binario del Worker
COPY --from=builder /worker_server /app/worker_server

# Copiar recursos necesarios (intro videos)
COPY intro ./intro

# Crear directorio temporal para procesamiento de videos
RUN mkdir -p /tmp

# El comando se especifica en docker-compose
CMD ["/app/worker_server"]
