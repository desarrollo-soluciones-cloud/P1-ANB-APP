# --- Etapa 1: Builder ---
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ffmpeg
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /api_server ./main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o /worker_server ./worker/main.go


FROM alpine:latest

RUN apk add --no-cache ffmpeg
WORKDIR /app

COPY --from=builder /api_server /api_server
COPY --from=builder /worker_server /worker_server

COPY ./intro ./intro