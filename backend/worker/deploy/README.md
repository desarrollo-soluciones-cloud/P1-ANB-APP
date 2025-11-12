# 🎬 ANB Worker - Procesamiento de Videos con SQS

Este directorio contiene la configuración de deployment del **Worker** que procesa videos de forma asíncrona usando **Amazon SQS** como cola de mensajes.

## 📋 Descripción

El worker es un servicio que:
- 🎯 Escucha mensajes de **Amazon SQS** para procesar videos
- 🎬 Convierte videos a formato MP4 usando **FFmpeg**
- 📦 Guarda videos procesados en **Amazon S3**
- 🗄️ Actualiza metadata en **PostgreSQL (RDS)**
- ♻️ Maneja reintentos automáticos vía SQS visibility timeout

## 🏗️ Arquitectura

```
Backend API → Amazon SQS → Worker (este servicio)
                              ↓
                    [FFmpeg Processing]
                              ↓
                    Amazon S3 + PostgreSQL RDS
```

**Cambio importante:** Ya **NO se usa Redis**. El sistema de mensajería ahora es **Amazon SQS**, un servicio completamente administrado por AWS.

## 🚀 Deployment Rápido

### Opción 1: Script Automático (Recomendado)

**Windows:**
```cmd
cd backend\worker\deploy
deploy.bat
```

**Linux/Mac:**
```bash
cd backend/worker/deploy
chmod +x deploy.sh
./deploy.sh
```

### Opción 2: Manual con Docker Compose

```bash
# 1. Copiar variables de entorno
cp .env.example .env

# 2. Editar .env con tus credenciales
nano .env

# 3. Levantar el worker
docker-compose up -d

# 4. Ver logs
docker-compose logs -f worker
```

## ⚙️ Variables de Entorno Requeridas

Edita el archivo `.env` con estos valores:

### 🗄️ Base de Datos (PostgreSQL RDS)
```bash
DB_HOST=tu-rds-endpoint.us-east-1.rds.amazonaws.com
DB_PORT=5432
DB_USER=anb_user
DB_PASSWORD=tu_password_seguro
DB_NAME=anb_db
DB_SSLMODE=require
```

### 📨 Amazon SQS (Cola de Mensajes)
```bash
SQS_QUEUE_URL=https://sqs.us-east-1.amazonaws.com/ACCOUNT_ID/anb-video-processing-queue
```

**IMPORTANTE:** Este `SQS_QUEUE_URL` debe ser el **mismo** que usa el backend API.

### 📦 Amazon S3 (Almacenamiento)
```bash
S3_BUCKET_NAME=anb-app-videos-prod
AWS_REGION=us-east-1
```

### 🔐 Credenciales AWS (AWS Academy)
```bash
AWS_ACCESS_KEY_ID=ASIAQ...
AWS_SECRET_ACCESS_KEY=gXTK...
AWS_SESSION_TOKEN=IQoJb3JpZ2luX2V...
```

**Nota:** Si usas **IAM Roles en EC2**, NO necesitas estas variables. AWS las maneja automáticamente.

## 🛠️ Comandos Útiles

### Ver logs en tiempo real
```bash
docker-compose logs -f worker
```

### Reiniciar worker
```bash
docker-compose restart worker
```

### Detener worker
```bash
docker-compose down
```

### Ver estado de contenedores
```bash
docker-compose ps
```

### Entrar al contenedor (debug)
```bash
docker exec -it anb-worker-prod sh
```

### Limpiar todo y volver a construir
```bash
docker-compose down
docker-compose build --no-cache
docker-compose up -d
```

## 📊 Verificación del Deployment

### 1. Verificar que el worker está corriendo
```bash
docker-compose ps
```

Deberías ver:
```
NAME                 IMAGE                   STATUS
anb-worker-prod      worker-deploy_worker    Up X seconds
```

### 2. Verificar logs del worker
```bash
docker-compose logs --tail=50 worker
```

Deberías ver:
```
✅ Database connection successful
✅ S3 Storage initialized successfully
✅ SQS Consumer initialized: queue=https://sqs...
🚀 Worker started - Listening for tasks from SQS...
```

### 3. Probar procesamiento de video

Desde el frontend o Postman, sube un video. El worker debería:
1. Recibir mensaje de SQS automáticamente
2. Procesar el video con FFmpeg
3. Subir a S3
4. Actualizar la base de datos

Verifica en los logs:
```
--- WORKER: Received task for video ID: 123 (Retry: 0/3) ---
📹 Processing video ID: 123
✅ Video 123 processed successfully in 15.23s
```

## 🔧 Troubleshooting

### El worker no recibe mensajes

1. **Verificar SQS_QUEUE_URL:**
   ```bash
   docker exec anb-worker-prod env | grep SQS
   ```

2. **Verificar permisos IAM:** El worker necesita permisos para:
   - `sqs:ReceiveMessage`
   - `sqs:DeleteMessage`
   - `sqs:ChangeMessageVisibility`

3. **Verificar que el backend esté enviando mensajes:**
   - Ve a AWS Console → SQS
   - Revisa "Messages Available"

### FFmpeg falla al procesar

1. **Verificar que el video está en S3:**
   ```bash
   aws s3 ls s3://anb-app-videos-prod/originals/
   ```

2. **Verificar logs detallados:**
   ```bash
   docker-compose logs -f worker | grep ERROR
   ```

### Credenciales AWS expiradas (AWS Academy)

Las credenciales de AWS Academy expiran cuando cierras la sesión. Actualízalas:

1. Ve a **AWS Academy → AWS Details → AWS CLI**
2. Copia las nuevas credenciales
3. Actualiza `.env`
4. Reinicia el worker:
   ```bash
   docker-compose restart worker
   ```

## 📁 Estructura del Proyecto

```
backend/worker/deploy/
├── docker-compose.yml      # Configuración de Docker (SIN Redis)
├── Dockerfile             # Imagen del worker con FFmpeg
├── .env.example           # Template de variables de entorno
├── .env                   # TUS credenciales (no commitear)
├── deploy.sh              # Script de deployment (Linux/Mac)
├── deploy.bat             # Script de deployment (Windows)
├── README.md              # Este archivo
└── main.go                # Código del worker (en ../main.go)
```

## 🔄 Diferencias con la Versión Anterior (Redis)

| Aspecto | Redis/Asynq (Anterior) | Amazon SQS (Actual) |
|---------|------------------------|---------------------|
| **Servicio** | Redis en Docker | SQS administrado por AWS |
| **Puerto** | 6379 (expuesto) | No necesita puerto |
| **Dependencias** | `depends_on: redis` | Sin dependencias |
| **Load Balancer** | NLB requerido | No necesario |
| **Escalabilidad** | Manual | Automática |
| **Protocolo** | TCP | HTTPS |
| **Variable env** | `REDIS_ADDR` | `SQS_QUEUE_URL` |

## 🎯 Ventajas de SQS

✅ **Sin mantenimiento:** AWS maneja todo  
✅ **Alta disponibilidad:** Multi-AZ automático  
✅ **Escalabilidad infinita:** No hay límites prácticos  
✅ **Menor costo:** Pay-per-request en lugar de EC2 24/7  
✅ **Dead Letter Queue:** Manejo nativo de mensajes fallidos  
✅ **Sin Load Balancer:** SQS es una API HTTP, no necesita NLB  

## 📚 Referencias

- [Amazon SQS Documentation](https://docs.aws.amazon.com/sqs/)
- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [FFmpeg Documentation](https://ffmpeg.org/documentation.html)

## 🆘 Soporte

Si tienes problemas:
1. Revisa los logs: `docker-compose logs -f worker`
2. Verifica las variables de entorno: `docker exec anb-worker-prod env`
3. Revisa la consola de AWS SQS para ver mensajes en la cola
4. Consulta el archivo principal de documentación: `backend/MIGRATION_SQS.md`
