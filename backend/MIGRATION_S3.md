# 🚀 Migración Completa a S3 - Backend

## ✅ Cambios Realizados

### **1. Storage Layer (`src/storage/storage.go`)**
- ❌ Eliminado: `localStorageService`
- ✅ Implementado: `s3StorageService` con:
  - `Upload()`: Subir archivos a S3
  - `Delete()`: Eliminar objetos de S3
  - `GetPresignedURL()`: Generar URLs firmadas temporales (1 hora)

### **2. Main Backend (`main.go`)**
- ✅ Inicialización de S3 con variables de entorno
- ❌ Eliminado: `router.Static("/uploads", "./uploads")`
- ✅ Validación de `S3_BUCKET_NAME` requerido

### **3. Video Service (`src/video/video.service.go`)**
#### Método `Upload()`
- Ahora guarda en S3 con key: `originals/{timestamp}-{userId}.mp4`
- Almacena S3 key en DB (no URL pública)

#### Método `ListByUserID()`, `GetByID()`, `ListPublic()`
- Generan presigned URLs on-demand para cada video
- Frontend recibe URLs completas de S3

#### Método `Delete()`
- Elimina de S3 usando `storageSvc.Delete()`
- Elimina tanto original como procesado

#### Método `MarkAsProcessed()`
- Actualiza `ProcessedURL` con S3 key: `processed/{basename}.mp4`

### **4. Worker (`worker/main.go`)**
#### Nuevo flujo híbrido:
1. **Descarga** video original desde S3 a `/tmp`
2. **Procesa** con FFmpeg en filesystem local
3. **Sube** video procesado a S3
4. **Limpia** archivos temporales
5. **Actualiza** DB con S3 key del video procesado

#### Nuevos métodos helper:
- `downloadFromS3()`: Descarga objeto S3 a archivo local
- `uploadToS3()`: Sube archivo local a S3

### **5. Video Controller (`src/video/video.controller.go`)**
- Endpoint `Download()` marcado como **deprecated**
- Retorna HTTP 410 Gone con mensaje explicativo
- Frontend usa presigned URLs directamente

---

## 📋 Variables de Entorno Requeridas

```bash
# OBLIGATORIAS
S3_BUCKET_NAME=anb-app-videos-prod
AWS_REGION=us-east-1

# Opcionales (si no usas IAM Role)
AWS_ACCESS_KEY_ID=AKIA...
AWS_SECRET_ACCESS_KEY=...
AWS_SESSION_TOKEN=...  # Solo AWS Academy
```

---

## 🔄 Flujo Completo

### **Upload de Video**
```
Usuario → Frontend → API Backend
                       ↓
                   S3 Upload (originals/123.mp4)
                       ↓
                   DB: Store S3 key
                       ↓
                   Enqueue processing task
```

### **Procesamiento de Video**
```
Worker Poll Task
    ↓
Download from S3 → /tmp/123.mp4
    ↓
FFmpeg Processing (local)
    ↓
Upload to S3 → processed/123.mp4
    ↓
Update DB with S3 key
    ↓
Clean /tmp files
```

### **Ver Video (Frontend)**
```
Usuario → API GET /videos/{id}
            ↓
        Generate Presigned URLs (1h expiry)
            ↓
        Return JSON with S3 URLs
            ↓
    Frontend <video> loads directly from S3
```

---

## 🧪 Testing

### **1. Verificar compilación**
```bash
cd backend
go build ./...
```

### **2. Verificar worker compila**
```bash
cd backend/worker
go build .
```

### **3. Test local (requiere AWS credentials)**
```bash
# Configurar .env
cp .env.example .env
# Editar .env con tus valores

# Iniciar backend
go run main.go

# Iniciar worker (en otra terminal)
cd worker
go run main.go
```

### **4. Test upload**
```bash
curl -X POST http://localhost:9090/api/v1/videos/upload \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -F "title=Test Video" \
  -F "video=@test.mp4"
```

### **5. Verificar en S3**
```bash
aws s3 ls s3://anb-app-videos-prod/originals/
```

---

## 🐳 Docker / Deployment

### **Dockerfile Updates Needed**
No cambios necesarios en Dockerfile, pero asegurar:

1. **Backend Dockerfile**
   - Variables de entorno S3 inyectadas
   - IAM Role asignado al task/instancia

2. **Worker Dockerfile**
   - Variables de entorno S3 inyectadas
   - IAM Role asignado al task/instancia
   - `/tmp` con suficiente espacio para videos

### **Docker Compose Updates**
```yaml
services:
  api:
    environment:
      - S3_BUCKET_NAME=anb-app-videos-prod
      - AWS_REGION=us-east-1
      # No incluir credenciales hardcoded!
  
  worker:
    environment:
      - S3_BUCKET_NAME=anb-app-videos-prod
      - AWS_REGION=us-east-1
```

### **ECS Task Definition**
```json
{
  "taskRoleArn": "arn:aws:iam::ACCOUNT:role/ANB-App-Backend-Role",
  "containerDefinitions": [{
    "environment": [
      {"name": "S3_BUCKET_NAME", "value": "anb-app-videos-prod"},
      {"name": "AWS_REGION", "value": "us-east-1"}
    ]
  }]
}
```

---

## ⚠️ Notas Importantes

### **Presigned URLs Expiran**
- Duración: **1 hora**
- Si usuario reporta video no carga después de 1h → Es normal
- Solución: Refrescar página para obtener nueva URL

### **Costos S3**
- **Upload**: $0.005 por 1,000 requests
- **Storage**: $0.023/GB/mes
- **Bandwidth**: $0.09/GB transferencia
- **Presigned URLs**: Gratis (no consume API calls)

### **Performance**
- ✅ Videos se sirven directo desde S3 (más rápido)
- ✅ Backend no maneja streaming de video
- ✅ Worker procesa localmente (rápido) y sube resultado

### **Seguridad**
- ✅ Bucket privado (no public access)
- ✅ Acceso solo vía presigned URLs
- ✅ URLs temporales (1h)
- ✅ IAM Role en producción (no hardcode credentials)

---

## 🔧 Troubleshooting

### **Error: "S3_BUCKET_NAME environment variable is required"**
- Solución: Agregar `S3_BUCKET_NAME` a `.env` o variables de entorno

### **Error: "failed to upload to S3: AccessDenied"**
- Causa: IAM Role sin permisos
- Solución: Verificar policy `ANB-App-S3-Access` en IAM Role

### **Error: "failed to download from S3: NoSuchKey"**
- Causa: S3 key no existe en bucket
- Solución: Verificar que upload fue exitoso, revisar logs

### **Worker no procesa videos**
- Verificar: Worker tiene acceso a S3
- Verificar: `/tmp` tiene espacio suficiente
- Verificar: FFmpeg instalado en contenedor worker

### **Frontend muestra "Error loading video"**
- Causa: Presigned URL expiró
- Solución: Refrescar página
- Prevención: Implementar auto-refresh de URLs en frontend

---

## ✅ Checklist de Deployment

### Pre-deployment:
- [ ] Bucket S3 creado
- [ ] IAM Policy creada
- [ ] IAM Role asignado a ECS Task / EC2
- [ ] Variables de entorno configuradas
- [ ] Código compila sin errores
- [ ] Tests locales pasaron

### Post-deployment:
- [ ] Upload de video funciona
- [ ] Worker procesa videos
- [ ] Videos se visualizan en frontend
- [ ] Presigned URLs funcionan
- [ ] Delete elimina de S3
- [ ] No errores en logs

---

## 📚 Referencias

- [AWS SDK Go v2 Docs](https://aws.github.io/aws-sdk-go-v2/docs/)
- [S3 Presigned URLs](https://docs.aws.amazon.com/AmazonS3/latest/userguide/PresignedUrlUploadObject.html)
- [IAM Roles for ECS](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-iam-roles.html)
