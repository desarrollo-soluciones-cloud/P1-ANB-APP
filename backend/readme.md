# ANB Backend API

## Descripción

API REST para la plataforma ANB (Asociación Nacional de Baloncesto) - un sistema de gestión y votación de videos desarrollado con Go, Gin Framework, PostgreSQL y Redis. La aplicación permite a los usuarios registrarse, subir videos, votar por ellos y ver rankings en tiempo real.

## Arquitectura

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Frontend      │    │   Backend API   │    │   PostgreSQL    │
│   Angular       │───▶│   Go + Gin      │───▶│   Database      │
│   (Port 3001)   │    │   (Port 9090)   │    │   (Port 5432)   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │
                                ▼
┌─────────────────┐    ┌─────────────────┐
│   Redis Cache   │    │   Worker        │
│   (Port 6379)   │◀───│   Video Process │
└─────────────────┘    └─────────────────┘
```

### Componentes Principales

- **API Server**: Servidor principal que maneja las peticiones HTTP
- **Worker**: Procesador asíncrono para tareas de videos (conversión, validación)
- **PostgreSQL**: Base de datos principal para persistencia
- **Redis**: Cache y cola de trabajos asíncronos
- **FFmpeg**: Procesamiento de videos

## Tecnologías

- **Go 1.24.6**: Lenguaje de programación principal
- **Gin**: Framework web HTTP
- **GORM**: ORM para PostgreSQL
- **Asynq**: Cola de trabajos asíncronos con Redis
- **JWT**: Autenticación basada en tokens
- **FFmpeg**: Procesamiento de videos
- **Docker**: Containerización

## Estructura del Proyecto

```
backend/
├── main.go                 # Punto de entrada de la API
├── go.mod                  # Dependencias de Go
├── go.sum                  # Checksums de dependencias
├── Dockerfile              # Imagen Docker para API
├── Dockerfile.test         # Imagen Docker para tests
├── .env.example            # Variables de entorno de ejemplo
├── worker/
│   └── main.go            # Worker asíncrono
├── src/
│   ├── auth/              # Autenticación JWT
│   │   ├── auth.go
│   │   ├── auth.middleware.go
│   │   └── auth_test.go
│   ├── database/          # Configuración de base de datos
│   │   └── database.go
│   ├── user/              # Gestión de usuarios
│   │   ├── user.entity.go
│   │   ├── user.dto.go
│   │   ├── user.repository.go
│   │   ├── user.service.go
│   │   ├── user.controller.go
│   │   ├── user.routes.go
│   │   └── *_test.go
│   ├── video/             # Gestión de videos
│   │   ├── video.entity.go
│   │   ├── video.dto.go
│   │   ├── video.repository.go
│   │   ├── video.service.go
│   │   ├── video.controller.go
│   │   ├── video.routes.go
│   │   └── *_test.go
│   └── vote/              # Sistema de votación
│       ├── vote.entity.go
│       ├── vote.dto.go
│       ├── vote.repository.go
│       ├── vote.service.go
│       ├── vote.controller.go
│       ├── vote.routes.go
│       └── *_test.go
└── uploads/               # Archivos subidos
    ├── originals/         # Videos originales
    ├── processed/         # Videos procesados
    └── temp/             # Archivos temporales
```

## Autenticación

El sistema utiliza **JWT (JSON Web Tokens)** para la autenticación:

- **Registro**: `POST /api/v1/auth/signup`
- **Login**: `POST /api/v1/auth/login`
- **Middleware**: Protege rutas que requieren autenticación
- **Expiración**: Tokens configurables vía variable de entorno

### Flujo de Autenticación

1. Usuario se registra o hace login
2. API genera JWT token
3. Cliente incluye token en header `Authorization: Bearer <token>`
4. Middleware valida token en cada request protegido

## API Endpoints

### Autenticación

```http
POST /api/v1/auth/signup
Content-Type: application/json

{
  "username": "string",
  "email": "string", 
  "password": "string",
  "password2": "string",
  "city": "string",
  "country": "string"
}
```

```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "string",
  "password": "string"
}
```

### Usuarios

```http
GET /api/v1/users/profile
Authorization: Bearer <token>

GET /api/v1/users/:user_id
Authorization: Bearer <token>
```

### Videos (Protegidos)

```http
# Subir video
POST /api/v1/videos/upload
Authorization: Bearer <token>
Content-Type: multipart/form-data

file: <video_file>
title: "string"
description: "string"

# Mis videos
GET /api/v1/videos
Authorization: Bearer <token>

# Video por ID
GET /api/v1/videos/:video_id
Authorization: Bearer <token>

# Descargar video
GET /api/v1/videos/:video_id/download
Authorization: Bearer <token>

# Eliminar video
DELETE /api/v1/videos/:video_id
Authorization: Bearer <token>

# Marcar como procesado
POST /api/v1/videos/:video_id/mark-processed
Authorization: Bearer <token>
```

### Videos Públicos

```http
# Listar videos públicos
GET /api/v1/public/videos
GET /api/public/videos  # Endpoint compatible

# Rankings de videos
GET /api/v1/public/rankings
```

### Votación

```http
POST /api/v1/public/videos/:video_id/vote
Authorization: Bearer <token>
```

### Salud del Sistema

```http
GET /health
```

## Sistema Asíncrono con Asynq

### Worker de Videos

El sistema utiliza **Asynq** para el procesamiento asíncrono de videos:

```go
// Tareas asíncronas disponibles
const (
    TypeVideoProcessing = "video:processing"  // Procesamiento de video
    TypeVideoValidation = "video:validation"  // Validación de formato
    TypeVideoThumbnail  = "video:thumbnail"   // Generación de miniaturas
)
```

### Flujo de Procesamiento

1. **Upload**: Usuario sube video → Se guarda en `/uploads/originals/`
2. **Queue**: Se crea tarea asíncrona en Redis
3. **Worker**: Procesa video con FFmpeg
4. **Output**: Video procesado se guarda en `/uploads/processed/`
5. **Status**: Se actualiza estado en base de datos

### Configuración del Worker

```bash
# Iniciar worker
cd worker/
go run main.go
```

## Docker

### Desarrollo Local

```bash
# Construir imagen
docker build -t anb-backend .

# Solo base de datos y Redis
docker-compose up postgres-anb redis-anb -d

# Ejecutar API localmente
go run main.go

# Ejecutar worker localmente  
cd worker && go run main.go
```

## Docker

### Desarrollo Local

```bash
# Construir imagen
docker build -t anb-backend .

# Solo base de datos y Redis
docker-compose up postgres-anb redis-anb -d

# Ejecutar API localmente
go run main.go

# Ejecutar worker localmente  
cd worker && go run main.go
```

### Producción Completa

```bash
# Levantar todo el stack
docker-compose up -d

# Ver logs
docker-compose logs api
docker-compose logs worker

# Escala worker (múltiples instancias)
docker-compose up --scale worker=3 -d
```
```

### Producción Completa

```bash
# Levantar todo el stack
docker-compose up -d

# Ver logs
docker-compose logs api
docker-compose logs worker

# Escala worker (múltiples instancias)
docker-compose up --scale worker=3 -d
```

### Dockerfile Multi-etapa

```dockerfile
# Etapa de construcción
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /api_server ./main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o /worker_server ./worker/main.go

# Etapa de producción
FROM alpine:latest
RUN apk add --no-cache ffmpeg
WORKDIR /app
COPY --from=builder /api_server /api_server
COPY --from=builder /worker_server /worker_server
COPY ./intro ./intro
EXPOSE 9090
CMD ["/api_server"]
```

## Variables de Entorno

Crear archivo `.env` basado en `.env.example`:

```bash
# Base de datos
DB_HOST=localhost
DB_PORT=5432
DB_USER=anb_user
DB_PASSWORD=anb_password
DB_NAME=anb_db

# Redis
REDIS_ADDR=localhost:6379

# JWT
JWT_SECRET=tu_jwt_secret_muy_seguro

# Servidor
SERVER_PORT=9090

# Worker
WORKER_CONCURRENCY=10
```

## Testing

```bash
# Ejecutar todos los tests
go test ./...

# Tests con cobertura
go test -cover ./...

# Tests de un módulo específico
go test ./src/video -v

### Estructura de Tests

- **Unit Tests**: `*_test.go` para cada componente  
- **Integration Tests**: Tests de base de datos y API
- **Mocks**: Para servicios externos y dependencias

## Base de Datos

### Modelos Principales

```go
// Usuario
type User struct {
    ID       uint   `gorm:"primaryKey"`
    Username string `gorm:"unique;not null"`
    Email    string `gorm:"unique;not null"`
    Password string `gorm:"not null"`
    City     string
    Country  string
    Videos   []Video `gorm:"foreignKey:UserID"`
    Votes    []Vote  `gorm:"foreignKey:UserID"`
}

// Video
type Video struct {
    ID          uint   `gorm:"primaryKey"`
    Title       string `gorm:"not null"`
    Description string
    Filename    string `gorm:"not null"`
    UserID      uint   `gorm:"not null"`
    Status      string `gorm:"default:pending"` // pending, processed, failed
    VoteCount   int    `gorm:"default:0"`
    User        User   `gorm:"foreignKey:UserID"`
    Votes       []Vote `gorm:"foreignKey:VideoID"`
}

// Voto
type Vote struct {
    ID      uint `gorm:"primaryKey"`
    UserID  uint `gorm:"not null"`
    VideoID uint `gorm:"not null"`
    User    User  `gorm:"foreignKey:UserID"`
    Video   Video `gorm:"foreignKey:VideoID"`
}
```

### Migraciones

```bash
# Las migraciones se ejecutan automáticamente al iniciar
# Ver: database/database.go -> MigrateTables()
```

## Instalación y Ejecución

### Prerrequisitos

- Go 1.24+
- PostgreSQL 15+
- Redis 7+
- FFmpeg
- Docker y Docker Compose (opcional)

### Instalación Local

```bash
# 1. Clonar repositorio
git clone <repository-url>
cd backend

# 2. Instalar dependencias
go mod download

# 3. Configurar variables de entorno
cp .env.example .env
# Editar .env con tus configuraciones

# 4. Ejecutar base de datos (Docker)
docker-compose up postgres-anb redis-anb -d

# 5. Ejecutar API
go run main.go

# 6. Ejecutar Worker (en otra terminal)
cd worker
go run main.go
```

### Con Docker Compose

```bash
# Levantar todo el stack
docker-compose up -d

# Verificar que todo esté funcionando
curl http://localhost:9090/health
```

### Logs Estructurados

```go
// Configuración de logs en main.go
log.Printf("Server starting on port %s", serverPort)
log.Printf("Connected to database: %s", dbName)
log.Printf("Redis connected: %s", redisAddr)
```

## Seguridad

### Medidas Implementadas

- **Autenticación JWT**: Tokens seguros con expiración
- **CORS**: Configurado para permitir frontend
- **Validación**: Sanitización de inputs con validator
- **Hash de Contraseñas**: bcrypt para almacenamiento seguro
- **Rate Limiting**: (Recomendado implementar)

### Headers de Seguridad

```go
// CORS configurado en main.go
router.Use(func(c *gin.Context) {
    c.Header("Access-Control-Allow-Origin", "*")
    c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
    c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
    // ... más headers
})
```

**ANB Backend API v1.0**  
*Desarrollado en Go*
