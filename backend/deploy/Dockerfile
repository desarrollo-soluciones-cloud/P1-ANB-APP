# --- Etapa 1: Builder ---
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Compilar SOLO el API
RUN CGO_ENABLED=0 GOOS=linux go build -o /api_server ./main.go

# --- Etapa 2: Runtime ---
FROM alpine:latest

# Instalar certificados SSL para AWS SDK
RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copiar SOLO el binario del API
COPY --from=builder /api_server /app/api_server

# Copiar recursos necesarios
COPY intro ./intro

# Exponer puerto del API
EXPOSE 9090

# El comando se especifica en docker-compose
CMD ["/app/api_server"]
