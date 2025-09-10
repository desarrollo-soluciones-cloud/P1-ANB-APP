# ANB - Asociación Nacional de Baloncesto

## Descripción

Plataforma web completa para la **Asociación Nacional de Baloncesto (ANB)** que permite a los usuarios gestionar y votar por contenido de video. El sistema incluye funcionalidades de autenticación, subida de videos, gestión personal de contenido, votación pública y rankings en tiempo real.

## Arquitectura del Sistema

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Frontend      │    │   Backend API   │    │   PostgreSQL    │
│   Angular 18    │───▶│   Go + Gin      │───▶│   Database      │
│   Port 3001     │    │   Port 9090     │    │   Port 5432     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │
                                ▼
┌─────────────────┐    ┌─────────────────┐
│   Redis Cache   │    │   Worker        │
│   Port 6379     │◀───│   Video Process │
└─────────────────┘    └─────────────────┘
```

## Tecnologías Principales

### Frontend
- **Angular 18** - Framework web progresivo
- **Angular Material** - Biblioteca de componentes UI
- **TypeScript** - Lenguaje tipado
- **Docker + Nginx** - Containerización y servidor web

### Backend
- **Go 1.24** - Lenguaje de programación principal
- **Gin Framework** - Framework web HTTP
- **PostgreSQL** - Base de datos principal
- **Redis + Asynq** - Cache y procesamiento asíncrono
- **JWT** - Autenticación
- **FFmpeg** - Procesamiento de videos

## Estructura del Proyecto

```
ANB-APP/
├── frontend/              # Aplicación Angular
│   ├── src/              # Código fuente
│   ├── Dockerfile        # Imagen Docker
│   ├── default.conf      # Configuración Nginx
│   └── README.md         # Documentación Frontend
├── backend/              # API en Go
│   ├── src/              # Código fuente organizado por módulos
│   ├── worker/           # Procesador asíncrono
│   ├── uploads/          # Archivos de video
│   ├── Dockerfile        # Imagen Docker
│   └── readme.md         # Documentación Backend
├── docker-compose.yml    # Orquestación de servicios
├── .env.example          # Variables de entorno
└── README.md            # Este archivo
```

## Inicio Rápido

### Prerrequisitos
- Docker y Docker Compose
- Git

### Instalación

```bash
# 1. Clonar el repositorio
git clone https://github.com/desarrollo-soluciones-cloud/P1-ANB-APP.git
cd P1-ANB-APP

# 2. Configurar variables de entorno
cp .env.example .env
# Editar .env con tus configuraciones

# 3. Levantar todos los servicios
docker-compose up -d

# 4. Verificar que todo esté funcionando
curl http://localhost:9090/health
```

### URLs de Acceso

- **Frontend**: http://localhost:3001
- **Backend API**: http://localhost:9090
- **Base de datos**: localhost:5433
- **Redis**: localhost:6379

## Funcionalidades Principales

### Autenticación y Usuarios
- Registro de nuevos usuarios
- Inicio de sesión con JWT
- Gestión de perfiles de usuario
- Protección de rutas y recursos

### Gestión de Videos
- Subida de videos (.mp4)
- Procesamiento asíncrono con FFmpeg
- Almacenamiento en sistema de archivos
- Gestión personal de contenido
- Eliminación de videos propios

### Sistema de Votación
- Catálogo público de videos
- Votación (un voto por usuario por video)
- Rankings dinámicos por popularidad
- Estadísticas en tiempo real

### Procesamiento Asíncrono
- Cola de trabajos con Redis y Asynq
- Conversión y optimización de videos
- Generación de miniaturas
- Notificaciones de estado

## Documentación Técnica

Para información detallada sobre cada componente:

### 📖 [Frontend - Documentación Completa](./frontend/README.md)
- Instalación y configuración de Angular
- Estructura de componentes y servicios
- Desarrollo local y testing
- Build y deployment
- Troubleshooting

### 📖 [Backend - Documentación Completa](./backend/readme.md)
- Arquitectura de la API
- Endpoints y autenticación
- Sistema asíncrono con Asynq
- Base de datos y migraciones
- Docker y deployment

## Desarrollo

### Desarrollo Local (Sin Docker)

**Backend:**
```bash
cd backend
go mod download
go run main.go
```

**Worker:**
```bash
cd backend/worker
go run main.go
```

**Frontend:**
```bash
cd frontend
npm install
ng serve
```

### Con Docker Compose

```bash
# Desarrollo con reconstrucción
docker-compose up --build

# Solo servicios específicos
docker-compose up postgres-anb redis-anb -d

# Ver logs
docker-compose logs -f api
docker-compose logs -f frontend

# Escalar workers
docker-compose up --scale worker=3 -d
```

## Variables de Entorno

Principales configuraciones en `.env`:

```bash
# Base de datos
DB_HOST=postgres-anb
DB_PORT=5432
DB_USER=anb_user
DB_PASSWORD=anb_password
DB_NAME=anb_db

# Redis
REDIS_ADDR=redis-anb:6379

# JWT
JWT_SECRET=tu_jwt_secret_muy_seguro

# Servidor
SERVER_PORT=9090

# Worker
WORKER_CONCURRENCY=10
```

## API Endpoints (Resumen)

```http
# Autenticación
POST /api/v1/auth/signup        # Registro
POST /api/v1/auth/login         # Login

# Videos (Autenticados)
POST /api/v1/videos/upload      # Subir video
GET  /api/v1/videos             # Mis videos
GET  /api/v1/videos/:id         # Detalle
DELETE /api/v1/videos/:id       # Eliminar

# Público
GET  /api/v1/public/videos      # Videos públicos
POST /api/v1/public/videos/:id/vote  # Votar
GET  /api/v1/public/rankings    # Rankings

# Sistema
GET  /health                    # Estado del sistema
```

## Deployment en Producción

### Docker Compose (Recomendado)

```bash
# Configurar variables de producción
cp .env.example .env.prod

# Levantar en modo producción
docker-compose -f docker-compose.yml up -d

# Monitoreo
docker-compose ps
docker-compose logs --tail=100 -f
```

### Servicios Individuales

Cada componente puede desplegarse independientemente siguiendo las guías específicas en sus respectivos READMEs.

## Monitoreo y Mantenimiento

### Health Checks
```bash
# API Health
curl http://localhost:9090/health

# Frontend
curl http://localhost:3001

# Base de datos
docker exec postgres-anb pg_isready
```

### Logs
```bash
# Ver logs de todos los servicios
docker-compose logs -f

# Logs específicos
docker-compose logs api worker frontend
```

### Backup
```bash
# Backup de base de datos
docker exec postgres-anb pg_dump -U anb_user anb_db > backup.sql

# Backup de videos
tar -czf videos_backup.tar.gz backend/uploads/
```

# SonarQube

Se encuentra en la wiki del repo.

### Estándares
- **Backend**: Convenciones de Go, tests unitarios
- **Frontend**: Angular style guide, ESLint, Prettier
- **Docker**: Multi-stage builds, optimización de imágenes
- **Git**: Conventional commits

**ANB - Asociación Nacional de Baloncesto v1.0**  
*Sistema completo de gestión y votación de videos*

