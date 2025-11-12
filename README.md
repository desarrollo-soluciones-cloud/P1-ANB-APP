# ANB - Asociación Nacional de Baloncesto

## Descripción

Plataforma web completa para la **Asociación Nacional de Baloncesto (ANB)** que permite a los usuarios gestionar y votar por contenido de video. El sistema incluye funcionalidades de autenticación, subida de videos, gestión personal de contenido, votación pública y rankings en tiempo real.

## Integrantes

* Tomas Acosta Bernal - 20201127 - t.acosta@uniandes.edu.co
* Samuel Romero Yepez - 201518954 - sj.romero10@uniandes.edu.co
* Alejandro Herrera Jiménez - 201915788 - a.herrera20@uniandes.edu.co
* Mauricio Ramírez Montilla -202522791 - m.ramirezm23@uniandes.edu.co

## Link Video 
https://drive.google.com/file/d/1jyZLgtK4Ha-CYfT9oR9ODp-jG0nPtS9B/view?usp=drive_link

## Link Video Entrega 2

https://drive.google.com/drive/folders/1KWCLaTHANyOGnqqDDBSkBzEqoW-9WYe6?usp=sharing

## Link Video Entrega 3

https://drive.google.com/drive/folders/1nFnT1uNvYGAMnevXR_JsidbAXz_FDDlK?usp=sharing

## Link Video Entrega 4

https://drive.google.com/drive/folders/1YuRrVT7tPMPV8mcHdCkdooBG0yrHawEI?usp=sharing

## Arquitectura del Sistema

<img width="904" height="563" alt="image" src="https://github.com/user-attachments/assets/ad6ae9cf-a1b3-4863-ad7d-685d2ed8bc0d" />



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

## Despliegue en AWS

### 1. Frontend en EC2

#### Requisitos previos
- EC2 con Docker y Docker Compose instalados
- Puerto 80 abierto en Security Group
- Acceso SSH a la instancia

#### Paso 1: Generar el paquete de despliegue
En tu máquina local (Windows PowerShell):

```powershell
# Navegar al directorio del frontend
cd frontend

# Compilar el frontend
npm run build

# Crear carpeta temporal
mkdir temp-package

# Copiar archivos necesarios
Copy-Item deploy\Dockerfile temp-package\
Copy-Item deploy\default.conf.template temp-package\
Copy-Item deploy\docker-compose.yml temp-package\
Copy-Item deploy\.env.example temp-package\
Copy-Item dist temp-package\ -Recurse

# Crear el ZIP
Compress-Archive -Path temp-package\* -DestinationPath frontend-deploy.zip -Force

# Limpiar
Remove-Item temp-package -Recurse -Force
```

#### Paso 2: Subir a EC2
```bash
# Opción A: Usar SCP
scp -i tu-llave.pem frontend-deploy.zip ec2-user@TU-IP-PUBLICA:/home/ec2-user/

```

#### Paso 3: Desplegar en EC2
```bash
# Conectarse a EC2
ssh -i tu-llave.pem ec2-user@TU-IP-PUBLICA

# Descomprimir
unzip frontend-deploy.zip
cd frontend-deploy

# Opción A: Usar IP del backend por defecto
docker-compose up -d

# Opción B: Especificar URL del backend
export BACKEND_URL=http://44.198.15.64:9090
docker-compose up -d

# Opción C: Crear archivo .env
cp .env.example .env
nano .env  # Editar BACKEND_URL
docker-compose up -d

# Verificar
docker-compose ps
```

#### Cambiar URL del backend (Load Balancer)
```bash
cd frontend-deploy
export BACKEND_URL=http://tu-load-balancer.elb.amazonaws.com
docker-compose down
docker-compose up -d
```

#### Acceder a la aplicación
```
http://TU-IP-PUBLICA-EC2
```

### 2. Backend API en EC2

#### Requisitos previos
- EC2 con Docker y Docker Compose instalados
- Puerto 9090 abierto en Security Group (API)
- Acceso a PostgreSQL (RDS o EC2)
- Acceso a Redis (ElastiCache o EC2)
- Bucket S3 creado: `anb-app-videos-prod`
- IAM Role con permisos S3 (o credenciales AWS)
- Acceso SSH a la instancia

#### Paso 1: Compilar el backend
En tu máquina local (Windows PowerShell):

```powershell
# Navegar al directorio del backend
cd backend

# Verificar que compila correctamente
go build -o api_server ./main.go

# Crear carpeta temporal
mkdir temp-package

# Copiar archivos necesarios
Copy-Item deploy\Dockerfile temp-package\
Copy-Item deploy\docker-compose.yml temp-package\
Copy-Item deploy\.env.example temp-package\
Copy-Item go.mod temp-package\
Copy-Item go.sum temp-package\
Copy-Item main.go temp-package\
Copy-Item -Path src temp-package\src -Recurse
Copy-Item -Path intro temp-package\intro -Recurse

# Crear el ZIP
Compress-Archive -Path temp-package\* -DestinationPath backend-deploy.zip -Force

# Limpiar
Remove-Item temp-package -Recurse -Force
```

#### Paso 2: Subir a EC2
```bash
# Usando SCP
scp -i tu-llave.pem backend-deploy.zip ec2-user@44.198.15.64:/home/ec2-user/
```

#### Paso 3: Configurar credenciales AWS en EC2
```bash
# Conectarse a EC2
ssh -i tu-llave.pem ec2-user@44.198.15.64

# Opción A: Si tienes IAM Role configurado (RECOMENDADO)
# No necesitas hacer nada, AWS SDK usará el IAM Role automáticamente

# Opción B: Configurar credenciales manualmente
aws configure
# Ingresar: AWS Access Key ID, Secret Access Key, Region (us-east-1)

# Opción C: Para AWS Academy Learner Lab
# Copiar credenciales desde AWS Academy > AWS Details > AWS CLI
nano ~/.aws/credentials
# Pegar las credenciales incluyendo aws_session_token
```

#### Paso 4: Desplegar en EC2
```bash
# Descomprimir
unzip backend-deploy.zip
cd backend-deploy

# Crear archivo .env con tu configuración
cp .env.example .env
nano .env

# Configurar las variables (ejemplo):
# DB_HOST=tu-rds-endpoint.us-east-1.rds.amazonaws.com
# DB_PORT=5432
# DB_USER=anb_user
# DB_PASSWORD=tu-password-seguro
# DB_NAME=anb_db
# REDIS_ADDR=tu-redis-endpoint.cache.amazonaws.com:6379
# JWT_SECRET=un-secret-muy-seguro-y-largo
# S3_BUCKET_NAME=anb-app-videos-prod
# AWS_REGION=us-east-1

# Construir y levantar el contenedor del API
docker-compose build
docker-compose up -d

# Verificar que está corriendo
docker-compose ps

# Ver logs
docker-compose logs -f api
```

#### Paso 5: Verificar el despliegue
```bash
# Verificar API
curl http://localhost:9090/health

# Verificar desde fuera de la EC2
curl http://44.198.15.64:9090/health

# Ver logs en tiempo real
docker-compose logs -f
```

#### Cambiar a Load Balancer
Cuando tengas un Load Balancer configurado:

```bash
# En la EC2 del FRONTEND, actualizar la URL del backend:
cd /home/ec2-user/frontend-deploy
export BACKEND_URL=http://tu-load-balancer.elb.amazonaws.com
docker-compose down
docker-compose up -d

# El backend NO necesita cambios, solo asegúrate de que:
# - El Load Balancer apunte a las instancias EC2 del backend (puerto 9090)
# - Los Security Groups permitan tráfico del Load Balancer a las EC2
# - El Health Check del Load Balancer use: /health
```

#### Comandos útiles

**Reiniciar API:**
```bash
docker-compose restart api
```

**Actualizar código:**
```bash
# Subir nuevo backend-deploy.zip
# Luego en EC2:
docker-compose down
rm -rf backend-deploy
unzip backend-deploy.zip
cd backend-deploy
cp ../backend-deploy/.env .  # Reutilizar configuración
docker-compose build --no-cache
docker-compose up -d
```

**Ver uso de recursos:**
```bash
docker stats
```

**Limpiar logs:**
```bash
docker-compose logs --tail=100 api
```

### 3. Worker en EC2 (Separado)

#### Requisitos previos
- EC2 con Docker y Docker Compose instalados
- Puerto 6379 abierto en Security Group (Redis)
- Acceso a PostgreSQL (RDS o EC2) - misma base de datos que el backend
- Bucket S3: `anb-app-videos-prod` (mismo que el backend)
- IAM Role con permisos S3 (o credenciales AWS)
- Acceso SSH a la instancia
- FFmpeg (incluido en el Dockerfile)

#### Paso 1: Compilar el worker
En tu máquina local (Windows PowerShell):

```powershell
# Navegar al directorio del worker
cd backend\worker

# Verificar que compila correctamente
go build -o worker_server ./main.go

# Crear carpeta temporal
mkdir temp-package

# Copiar archivos necesarios
Copy-Item deploy\Dockerfile temp-package\
Copy-Item deploy\docker-compose.yml temp-package\
Copy-Item deploy\.env.example temp-package\
Copy-Item ..\go.mod temp-package\
Copy-Item ..\go.sum temp-package\
Copy-Item main.go temp-package\
Copy-Item -Path ..\src temp-package\src -Recurse
Copy-Item -Path ..\intro temp-package\intro -Recurse
Copy-Item -Path . temp-package\worker -Recurse -Include main.go

# Crear el ZIP
Compress-Archive -Path temp-package\* -DestinationPath ..\worker-deploy.zip -Force

# Limpiar
Remove-Item temp-package -Recurse -Force

# El archivo worker-deploy.zip estará en: backend\worker-deploy.zip
```

#### Paso 2: Subir a EC2
```bash
# Usando SCP
scp -i tu-llave.pem worker-deploy.zip ec2-user@34.233.29.202:/home/ec2-user/
```

#### Paso 3: Configurar credenciales AWS en EC2
```bash
# Conectarse a EC2
ssh -i tu-llave.pem ec2-user@34.233.29.202

# Opción A: Si tienes IAM Role configurado (RECOMENDADO)
# No necesitas hacer nada, AWS SDK usará el IAM Role automáticamente

# Opción B: Configurar credenciales manualmente
aws configure
# Ingresar: AWS Access Key ID, Secret Access Key, Region (us-east-1)

# Opción C: Para AWS Academy Learner Lab
# Copiar credenciales desde AWS Academy > AWS Details > AWS CLI
nano ~/.aws/credentials
# Pegar las credenciales incluyendo aws_session_token
```

#### Paso 4: Desplegar Worker + Redis en EC2
```bash
# Descomprimir
unzip worker-deploy.zip
cd worker-deploy

# Crear archivo .env con tu configuración
cp .env.example .env
nano .env

# Configurar las variables (ejemplo):
# DB_HOST=postgres-anb.cd6qswmk4njt.us-east-1.rds.amazonaws.com
# DB_PORT=5432
# DB_USER=anb_user
# DB_PASSWORD=admin1234
# DB_NAME=anb_db
# DB_SSLMODE=require
# REDIS_ADDR=redis:6379  # NO cambiar, siempre es redis:6379
# S3_BUCKET_NAME=anb-app-videos-prod
# AWS_REGION=us-east-1
# AWS_ACCESS_KEY_ID=tu-access-key
# AWS_SECRET_ACCESS_KEY=tu-secret-key
# AWS_SESSION_TOKEN=tu-session-token  # Solo para AWS Academy

# Construir y levantar Redis + Worker
docker-compose build --no-cache
docker-compose up -d

# Verificar que están corriendo
docker-compose ps

# Ver logs del worker
docker-compose logs -f worker

# Ver logs de Redis
docker-compose logs -f redis
```

#### Paso 5: Configurar el Backend para conectarse al Worker
En la EC2 del **BACKEND**, actualizar el `.env` para apuntar a este Worker:

```bash
# En la EC2 del BACKEND (44.198.15.64)
ssh -i tu-llave.pem ec2-user@44.198.15.64
cd backend-deploy
nano .env

# Cambiar la línea de REDIS_ADDR:
REDIS_ADDR=34.233.29.202:6379  # IP pública de la EC2 del Worker

# Reiniciar el backend
docker-compose restart api

# Verificar logs
docker-compose logs -f api
```

#### Paso 6: Verificar el despliegue
```bash
# En la EC2 del Worker, verificar que Redis está escuchando
docker-compose exec redis redis-cli ping
# Debe responder: PONG

# Verificar logs del worker esperando tareas
docker-compose logs worker | tail -20
# Debe ver: "ANB Worker is running and connected to PostgreSQL..."
# Debe ver: "Waiting for video processing tasks..."

# Verificar conectividad desde el Backend
# En la EC2 del Backend:
telnet 34.233.29.202 6379
# O con nc:
nc -zv 34.233.29.202 6379

# O probar conexión completa con redis-cli:
redis-cli -h 34.233.29.202 -p 6379 ping
# Debe responder: PONG

# Si redis-cli no está instalado:
sudo yum install redis -y  # Amazon Linux
# sudo apt install redis-tools -y  # Ubuntu
```

#### Paso 7: Probar el flujo completo
```bash
# Desde el frontend, subir un video
# Verificar en logs del Backend que se encola la tarea:
docker-compose logs -f api | grep "enqueued"

# Verificar en logs del Worker que procesa el video:
docker-compose logs -f worker
# Debe ver:
# "Received task for video ID: X"
# "Processing video..."
# "Successfully processed video ID: X"
```

#### Cambiar a Load Balancer (Redis)
Cuando tengas un Load Balancer para el Worker:

```bash
# En la EC2 del BACKEND, actualizar Redis:
cd /home/ec2-user/backend-deploy
nano .env
# Cambiar: REDIS_ADDR=worker-lb.elb.amazonaws.com:6379
docker-compose restart api

# NOTA: El Worker NO necesita cambios en su REDIS_ADDR
# El Worker siempre usa: REDIS_ADDR=redis:6379 (local en docker-compose)
```

#### Comandos útiles

**Reiniciar Worker:**
```bash
docker-compose restart worker
```

**Reiniciar Redis:**
```bash
docker-compose restart redis
```

**Ver tareas en Redis:**
```bash
docker-compose exec redis redis-cli
> KEYS *
> LLEN asynq:default
> exit
```

**Actualizar código del Worker:**
```bash
# Subir nuevo worker-deploy.zip
# Luego en EC2:
docker-compose down
rm -rf worker-deploy
unzip worker-deploy.zip
cd worker-deploy
cp ../worker-deploy/.env .  # Reutilizar configuración
docker-compose build --no-cache
docker-compose up -d
docker-compose logs -f worker
```

**Ver uso de recursos:**
```bash
docker stats
```

**Limpiar archivos temporales del worker:**
```bash
docker-compose exec worker ls -lh /tmp/video-processing/
docker-compose exec worker rm -rf /tmp/video-processing/*
```

#### Troubleshooting

**Error: "no such host" en el Worker**
```bash
# Verificar que DB_HOST en .env tiene el hostname correcto
cat .env | grep DB_HOST
# Debe ser: DB_HOST=postgres-anb.cd6qswmk4njt.us-east-1.rds.amazonaws.com
# NO debe ser: DB_HOST=anb-app-db...

# Si está mal, corregir y reiniciar:
nano .env
docker-compose restart worker
```

**El Worker no procesa videos**
```bash
# 1. Verificar que el Backend puede conectarse a Redis
# En EC2 del Backend:
telnet 34.233.29.202 6379

# 2. Verificar Security Group del Worker
# Debe permitir entrada en puerto 6379 desde la IP del Backend

# 3. Ver cola de Redis
docker-compose exec redis redis-cli LLEN asynq:default
# Si retorna > 0, hay tareas pendientes

# 4. Reiniciar worker
docker-compose restart worker
```

**Credenciales AWS expiradas (AWS Academy)**
```bash
# Obtener nuevas credenciales de AWS Academy
# Actualizar .env:
nano .env
# Actualizar: AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, AWS_SESSION_TOKEN
docker-compose restart worker
```

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







