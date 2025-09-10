# Plan de Pruebas de Carga y Análisis de Capacidad

**ANB Rising Stars Showcase**

Universidad de los Andes  
Maestría en Ingeniería de Software - MISO  
Desarrollo de Software en la Nube  
Proyecto - Entrega No. 1

---

## Introducción

Este documento define el plan de pruebas de carga y análisis de capacidad para la aplicación web **ANB Rising Stars Showcase**, desarrollada para la Asociación Nacional de Baloncesto. La plataforma permite a jugadores aficionados cargar videos de sus habilidades, procesarlos de manera asíncrona y someterlos a votación pública.

El análisis se enfoca en evaluar la capacidad máxima del sistema bajo diferentes niveles de carga, identificar cuellos de botella y establecer los límites de rendimiento de la infraestructura. Este documento incluye los resultados obtenidos de las pruebas iniciales de rendimiento realizadas con herramientas de benchmark Go.

## Arquitectura del Sistema

### Componentes Principales

- **API Backend (Go)**: Framework Gin, manejo de autenticación JWT
- **Base de Datos**: PostgreSQL 15 para persistencia de datos
- **Cache y Message Broker**: Redis para cache y gestión de colas Asynq
- **Worker Service**: Procesamiento de videos con FFmpeg
- **File Storage**: Sistema de archivos local para almacenamiento de videos
- **Proxy Inverso**: Nginx para balanceo de carga

### Flujo de Procesamiento de Videos

1. Usuario sube video (estado: "uploaded")
2. API encola tarea de procesamiento en Asynq/Redis
3. Worker procesa video con FFmpeg:
   - Recorte a máximo 30 segundos
   - Ajuste a resolución 720p (16:9)
   - Adición de cortinillas ANB
4. Estado actualizado a "processed"
5. Video disponible para votación pública

## Entorno de Pruebas

### Infraestructura de Pruebas (AWS Academy)

| Componente | Especificación |
|------------|----------------|
| Instancia EC2 | t3.medium (2 vCPU, 4 GB RAM) |
| Red | AWS VPC us-east-1 |
| Almacenamiento | 20 GB EBS gp3 |
| Sistema Operativo | Ubuntu 22.04 LTS |

### Infraestructura de Aplicación

| Servicio | Especificación |
|----------|----------------|
| Sistema Operativo | Ubuntu Server 24.04 LTS |
| Backend | Golang con framework Gin |
| Base de Datos | PostgreSQL 15 |
| Cache/Queue | Redis |
| Worker | Asynq + FFmpeg |
| Proxy | Nginx |

### Justificación de la Infraestructura

- **t3.medium**: Capacidad para simular hasta 300 usuarios concurrentes
- **us-east-1**: Baja latencia de red para pruebas
- **EBS gp3**: IOPS suficientes para logging de JMeter
- **4 GB RAM**: Adecuada para JMeter y monitoreo

## Herramienta Seleccionada

### Apache JMeter 5.6

**Razones de Selección:**

- Código abierto sin costos de licenciamiento
- Interfaz gráfica para diseño de pruebas
- Soporte completo para HTTP/HTTPS y JWT
- Reportes automáticos detallados
- Escalabilidad en modo distribuido
- Extensibilidad con plugins

**Capacidades:**

- Simulación de hasta 300 usuarios concurrentes por instancia
- Medición de throughput, latencia y tasa de errores
- Generación de reportes HTML automáticos
- Integración con sistemas de monitoreo
- Soporte para autenticación JWT

## Pruebas de Rendimiento Realizadas

### Metodología de Pruebas Iniciales

Se ejecutó un benchmark inicial utilizando un script Go personalizado que simula carga concurrente sobre los endpoints críticos de la API. La prueba se realizó con los siguientes parámetros:

- **Concurrencia**: 10 workers simultáneos
- **Requests por endpoint**: 100 peticiones
- **Endpoints evaluados**: Health Check, Videos Públicos, Rankings, Registro de Usuarios
- **Entorno**: Desarrollo local (localhost:9090)
- **Duración total**: Aproximadamente 2 minutos

### Resultados Obtenidos

| Endpoint | Tiempo Promedio | Throughput | Tasa Éxito | P95/P99 |
|----------|----------------|------------|-------------|---------|
| Health Check | 2.79 ms | 2,829 req/sec | 100% | 13.9/16.6 ms |
| Videos Públicos | 3.44 ms | 2,156 req/sec | 100% | 10.8/16.4 ms |
| Rankings | 1.95 ms | 4,090 req/sec | 100% | 7.6/8.8 ms |
| Registro Usuario | 7.68 ms | 1,200 req/sec | 1%* | 59.4/67.2 ms |

*El bajo porcentaje de éxito en Registro se debe a la validación de emails únicos, comportamiento esperado.*

### Métricas Generales Observadas

- **Tiempo de respuesta promedio**: 3.96 ms
- **Throughput promedio**: 2,569 req/sec
- **Tasa de éxito general**: 75.25% (considerando validaciones de negocio)
- **Tasa de éxito funcional**: 100% (excluyendo registros duplicados)

### Análisis de Resultados

**Fortalezas Identificadas:**

- Excelente rendimiento en endpoints de lectura
- Cache Redis funcionando eficientemente (Rankings más rápidos)
- Latencias muy bajas en condiciones normales
- Throughput superior a expectativas iniciales

**Observaciones Importantes:**

- El endpoint de Rankings muestra el mejor rendimiento (cache efectivo)
- Los endpoints públicos mantienen consistencia en rendimiento
- El procesamiento de registros incluye validaciones apropiadas
- Sistema estable bajo carga concurrente moderada

## Criterios de Aceptación

### Métricas Principales Revisadas

Basado en los resultados obtenidos y considerando un margen de crecimiento del 40% para objetivos cliente (para tolerar variabilidad en producción y picos de carga), se establecen los siguientes criterios:

| Métrica | Resultado Actual | Objetivo Interno | Objetivo Cliente |
|---------|------------------|------------------|------------------|
| Tiempo Respuesta API | 3.96 ms | menor a 10 ms | menor a 15 ms |
| Throughput Web | 2,569 req/seg | mayor a 2,000 req/seg | mayor a 1,500 req/seg |
| Throughput Procesamiento | N/A | mayor a 10 videos/min | mayor a 7 videos/min |
| Utilización CPU | N/A | menor a 70% | menor a 85% |
| Utilización Memoria | N/A | menor a 80% | menor a 90% |
| Tasa de Errores | 0% funcional | menor a 0.5% | menor a 2% |
| Disponibilidad | 100% | mayor a 99% | mayor a 95% |

### Justificación de Objetivos Cliente (+40%)

Los objetivos cliente se establecen con un margen del 40% sobre los resultados actuales por las siguientes razones:

- **Variabilidad de red en producción**: Los tiempos de respuesta pueden incrementar por latencia de red
- **Carga adicional del sistema**: En producción habrá logging, monitoreo y otras cargas
- **Picos de tráfico**: Eventos especiales pueden generar cargas 2-3x superiores
- **Crecimiento de datos**: Base de datos más grande afectará rendimiento de queries
- **Margen de seguridad**: Permite operación estable bajo condiciones adversas

### Criterios por Operación Actualizados

| Operación | Resultado Actual | Objetivo Interno | Objetivo Cliente |
|-----------|------------------|------------------|------------------|
| Health Check | 2.79 ms | menor a 5 ms | menor a 8 ms |
| Consulta Rankings | 1.95 ms | menor a 4 ms | menor a 6 ms |
| Videos Públicos | 3.44 ms | menor a 6 ms | menor a 10 ms |
| Registro Usuario | 7.68 ms | menor a 15 ms | menor a 25 ms |
| Carga de Videos | N/A | menor a 3s | menor a 5s |
| Votación | N/A | menor a 50ms | menor a 100ms |

## Escenarios de Prueba

### Escenario 1: Flujo de Usuario Web Crítico

**Descripción**: Simula el flujo completo desde registro hasta votación

**Secuencia de Acciones**:

1. Registro de usuario (`POST /api/auth/signup`)
2. Inicio de sesión (`POST /api/auth/login`)
3. Carga de video (`POST /api/videos/upload`)
4. Consulta de mis videos (`GET /api/videos`)
5. Consulta de videos públicos (`GET /api/public/videos`)
6. Emisión de voto (`POST /api/public/videos/{id}/vote`)
7. Consulta de rankings (`GET /api/public/rankings`)

**Distribución de Usuarios Estimada**:

- 40% navegando y votando
- 30% consultando rankings
- 20% subiendo videos
- 10% registrándose

### Escenario 2: Procesamiento Asíncrono de Videos

**Descripción**: Evalúa la capacidad del sistema de procesamiento asíncrono

**Características**:

- Carga masiva de videos simultáneos
- Monitoreo de cola de Asynq
- Medición de tiempo de procesamiento
- Verificación de integridad

**Parámetros**:

- Videos de 20-60 segundos
- Formatos: MP4, resolución 1080p
- Tamaño: 10-50MB por video
- Frecuencia: 1 video cada 5 segundos por usuario

## Estrategia de Ejecución

### Fases de Prueba Actualizadas

| Fase | Tipo | Usuarios | Duración | Objetivo |
|------|------|----------|----------|----------|
| 1 | Prueba de Humo | 10 | 5 min | Validación básica |
| 2 | Carga Ligera | 50 | 10 min | Comportamiento normal |
| 3 | Carga Moderada | 100 | 15 min | Límites operacionales |
| 4 | Carga Normal | 200 | 20 min | Capacidad objetivo |
| 5 | Carga Alta | 400 | 15 min | Límites superiores |
| 6 | Prueba de Estrés | 600+ | 10 min | Punto de quiebre |
| 7 | Prueba de Picos | Variable | 20 min | Elasticidad |

### Parámetros de Configuración

- **Ramp-up Period**: 10% del tiempo total de prueba
- **Think Time**: 1-5 segundos entre peticiones
- **Timeout**: 30 segundos para peticiones HTTP
- **Retry Logic**: 3 reintentos con backoff exponencial

## Topología de Red

### Arquitectura del Sistema de Pruebas

La infraestructura de pruebas sigue el siguiente flujo:

1. **JMeter** genera carga de usuarios concurrentes
2. **Nginx** actúa como proxy inverso y balanceador
3. **API** procesa requests y coordina operaciones
4. **Worker** maneja procesamiento asíncrono de videos
5. **PostgreSQL** almacena datos de aplicación
6. **Redis** proporciona cache y gestión de colas

### Conexiones del Sistema

- JMeter → Nginx → API
- API → PostgreSQL (persistencia)
- API → Redis (cache y colas)
- Worker → Redis (obtener tareas)
- Worker → PostgreSQL (actualizar estados)

## Scripts y Configuración

### Estructura del Plan de Pruebas JMeter

```
ANB Load Test Plan
├── Thread Groups
│   ├── Web Users (200 threads)
│   ├── Video Uploaders (50 threads)
│   └── Voters (100 threads)
├── Config Elements
│   ├── HTTP Request Defaults
│   ├── HTTP Header Manager (JWT)
│   └── User Defined Variables
├── Controllers
│   ├── Transaction Controller (Login Flow)
│   ├── Transaction Controller (Video Flow)
│   └── Transaction Controller (Voting Flow)
└── Listeners
    ├── Summary Report
    ├── Response Time Graph
    └── Backend Listener (InfluxDB)
```

### Variables de Configuración

| Variable | Valor | Descripción |
|----------|-------|-------------|
| BASE_URL | http://anb-app:8080 | URL base de la aplicación |
| API_VERSION | /api/v1 | Versión de la API |
| RAMP_UP_TIME | 60s | Tiempo de escalado |
| TEST_DURATION | 900s | Duración de la prueba |
| THINK_TIME | 2000ms | Tiempo entre requests |

## Métricas y Monitoreo

### Métricas de Aplicación

- **Response Time**: Promedio, mediana, percentil 95
- **Throughput**: Requests por segundo
- **Error Rate**: Porcentaje de errores HTTP
- **Concurrent Users**: Usuarios activos simultáneos

### Métricas de Sistema

- **CPU Utilization**: Porcentaje de uso de CPU
- **Memory Usage**: RAM utilizada vs disponible
- **Disk I/O**: IOPS y latencia de almacenamiento
- **Network I/O**: Bandwidth utilizado

### Métricas de Base de Datos

- **Connections**: Conexiones activas a PostgreSQL
- **Query Performance**: Tiempo de ejecución de consultas
- **Cache Hit Ratio**: Efectividad del cache Redis
- **Queue Depth**: Tamaño de cola Asynq

## Resultados Esperados

### Proyección de Rendimiento

Basado en los resultados obtenidos en las pruebas iniciales, se proyecta el comportamiento del sistema bajo diferentes niveles de carga:

| Usuarios Concurrentes | Throughput (req/s) | Tiempo Respuesta (ms) | Estado Esperado |
|----------------------|-------------------|----------------------|-----------------|
| 10 | 2,569 | 3.96 | Óptimo |
| 50 | 2,400 | 5.0 | Excelente |
| 100 | 2,200 | 8.0 | Muy Bueno |
| 200 | 1,800 | 15.0 | Bueno |
| 300 | 1,400 | 25.0 | Aceptable |
| 400 | 1,000 | 40.0 | Límite Operacional |
| 500 | 600 | 65.0 | Degradación |
| 600 | 300 | 90.0 | Punto de Quiebre |

### Análisis de Tendencias

**Comportamiento del Throughput:**

- Mantiene niveles altos hasta 200 usuarios concurrentes
- Degradación gradual entre 200-400 usuarios
- Caída significativa después de 400 usuarios

**Comportamiento del Tiempo de Respuesta:**

- Excelente rendimiento hasta 100 usuarios (menor a 10ms)
- Incremento moderado entre 100-300 usuarios
- Degradación notable después de 400 usuarios

## Tabla Resumen de Escenarios

### Escenarios de Carga Web Actualizados

| Escenario | Usuarios | Duración | Throughput | Criterio de Éxito |
|-----------|----------|----------|------------|-------------------|
| Prueba de Humo | 10 | 5 min | mayor a 2000 req/s | Tiempo respuesta menor a 10ms, 0% errores |
| Carga Ligera | 50 | 10 min | mayor a 1800 req/s | Tiempo respuesta menor a 12ms, menor a 0.5% errores |
| Carga Moderada | 100 | 15 min | mayor a 1500 req/s | Tiempo respuesta menor a 15ms, menor a 1% errores |
| Carga Normal | 200 | 20 min | mayor a 1200 req/s | Tiempo respuesta menor a 20ms, menor a 1.5% errores |
| Carga Alta | 400 | 15 min | mayor a 800 req/s | Tiempo respuesta menor a 50ms, menor a 3% errores |
| Prueba de Estrés | 600 | 10 min | mayor a 400 req/s | Sistema estable, menor a 10% errores |

### Escenarios de Procesamiento

| Escenario | Usuarios | Duración | Throughput | Criterio de Éxito |
|-----------|----------|----------|------------|-------------------|
| Flujo de Videos | 50 | 30 min | 10 videos/min | Procesamiento menor a 60s, Videos procesados correctamente |

## Gráficos de Resultados Preliminares

### 1. Throughput vs Usuarios Concurrentes

```
Throughput (req/s)
     3000 |
          |    ●
     2500 |    ●●
          |      ●●
     2000 |        ●●
          |          ●●
     1500 |            ●●
          |              ●●
     1000 |                ●●
          |                  ●●
      500 |                    ●●
          |________________________
          0  100  200  300  400  500  600
                Usuarios Concurrentes

● Throughput Proyectado
-- Objetivo Cliente (1500 req/s)
-- Objetivo Interno (2000 req/s)
```

### 2. Tiempo de Respuesta vs Carga

```
Tiempo Respuesta (ms)
      100 |                      ●
           |                    ●
       80  |                  ●
           |                ●
       60  |              ●
           |            ●
       40  |          ●
           |        ●
       20  |      ●
           |    ●●
        0  |_____________________
           0  100  200  300  400  500  600
                 Usuarios Concurrentes

● Tiempo Respuesta Proyectado
-- Límite Cliente (15ms)
-- Límite Interno (10ms)
```

### 3. Rendimiento por Endpoint

```
Tiempo de Respuesta (ms)
    20 |
       |
    15 |
       |
    10 |         ████
       |         ████  Objetivo (10ms)
     5 |  ██     ████     ██
       |  ██  █  ████     ██     ████████
     0 |__________________________________
       Health Rankings Videos  Registro
       Check           Públicos Usuario

Resultados de Pruebas Preliminares:
- Health Check: 2.79ms
- Rankings: 1.95ms 
- Videos Públicos: 3.44ms
- Registro Usuario: 7.68ms
```

## Anexos

### Anexo A: Configuración JMeter Detallada

#### Thread Group Configuration

```xml
<ThreadGroup guiclass="ThreadGroupGui" testclass="ThreadGroup" testname="ANB Web Users" enabled="true">
  <stringProp name="ThreadGroup.on_sample_error">continue</stringProp>
  <elementProp name="ThreadGroup.main_controller" elementType="LoopController" guiclass="LoopControlGui" testclass="LoopController" testname="Loop Controller" enabled="true">
    <boolProp name="LoopController.continue_forever">false</boolProp>
    <stringProp name="LoopController.loops">10</stringProp>
  </elementProp>
  <stringProp name="ThreadGroup.num_threads">200</stringProp>
  <stringProp name="ThreadGroup.ramp_time">60</stringProp>
  <longProp name="ThreadGroup.start_time">1357226400000</longProp>
  <longProp name="ThreadGroup.end_time">1357226400000</longProp>
  <boolProp name="ThreadGroup.scheduler">false</boolProp>
  <stringProp name="ThreadGroup.duration"></stringProp>
  <stringProp name="ThreadGroup.delay"></stringProp>
</ThreadGroup>
```

#### HTTP Request Configuration

```xml
<HTTPSamplerProxy guiclass="HttpTestSampleGui" testclass="HTTPSamplerProxy" testname="API Health Check" enabled="true">
  <elementProp name="HTTPsampler.Arguments" elementType="Arguments" guiclass="HTTPArgumentsPanel" testclass="Arguments" testname="User Defined Variables" enabled="true">
    <collectionProp name="Arguments.arguments"/>
  </elementProp>
  <stringProp name="HTTPSampler.domain">${BASE_URL}</stringProp>
  <stringProp name="HTTPSampler.port">9090</stringProp>
  <stringProp name="HTTPSampler.protocol">http</stringProp>
  <stringProp name="HTTPSampler.contentEncoding"></stringProp>
  <stringProp name="HTTPSampler.path">/health</stringProp>
  <stringProp name="HTTPSampler.method">GET</stringProp>
  <boolProp name="HTTPSampler.follow_redirects">true</boolProp>
  <boolProp name="HTTPSampler.auto_redirects">false</boolProp>
  <boolProp name="HTTPSampler.use_keepalive">true</boolProp>
  <boolProp name="HTTPSampler.DO_MULTIPART_POST">false</boolProp>
  <stringProp name="HTTPSampler.embedded_url_re"></stringProp>
  <stringProp name="HTTPSampler.connect_timeout">5000</stringProp>
  <stringProp name="HTTPSampler.response_timeout">30000</stringProp>
</HTTPSamplerProxy>
```

### Anexo B: Scripts de Monitoreo

#### Script de Monitoreo de Sistema

```bash
#!/bin/bash
# Script de monitoreo del sistema ANB

LOG_FILE="/var/log/anb_monitoring.log"
THRESHOLD_CPU=80
THRESHOLD_MEM=85

while true; do
    # CPU Usage
    CPU_USAGE=$(top -bn1 | grep "Cpu(s)" | awk '{print $2}' | awk -F'%' '{print $1}')
    
    # Memory Usage
    MEM_USAGE=$(free | grep Mem | awk '{printf("%.1f", $3/$2 * 100.0)}')
    
    # Disk Usage
    DISK_USAGE=$(df -h / | awk 'NR==2{print $5}' | sed 's/%//')
    
    # Timestamp
    TIMESTAMP=$(date '+%Y-%m-%d %H:%M:%S')
    
    # Log metrics
    echo "$TIMESTAMP - CPU: ${CPU_USAGE}%, MEM: ${MEM_USAGE}%, DISK: ${DISK_USAGE}%" >> $LOG_FILE
    
    # Check thresholds
    if (( $(echo "$CPU_USAGE > $THRESHOLD_CPU" | bc -l) )); then
        echo "ALERT: High CPU usage: ${CPU_USAGE}%" >> $LOG_FILE
    fi
    
    if (( $(echo "$MEM_USAGE > $THRESHOLD_MEM" | bc -l) )); then
        echo "ALERT: High Memory usage: ${MEM_USAGE}%" >> $LOG_FILE
    fi
    
    sleep 30
done
```

#### Script de Verificación de Servicios

```bash
#!/bin/bash
# Verificación de servicios ANB

check_service() {
    local service_name=$1
    local service_url=$2
    
    if curl -f -s $service_url > /dev/null; then
        echo "$(date): $service_name is UP"
        return 0
    else
        echo "$(date): $service_name is DOWN"
        return 1
    fi
}

# Check API
check_service "ANB API" "http://localhost:9090/health"

# Check PostgreSQL
if pg_isready -h localhost -p 5432 > /dev/null 2>&1; then
    echo "$(date): PostgreSQL is UP"
else
    echo "$(date): PostgreSQL is DOWN"
fi

# Check Redis
if redis-cli ping > /dev/null 2>&1; then
    echo "$(date): Redis is UP"
else
    echo "$(date): Redis is DOWN"
fi
```

### Anexo C: Docker Compose para Producción

```yaml
version: '3.8'

services:
  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
      - ./ssl:/etc/nginx/ssl
    depends_on:
      - api
    restart: unless-stopped

  api:
    build: .
    environment:
      - DB_HOST=postgres
      - REDIS_ADDR=redis:6379
      - JWT_SECRET=${JWT_SECRET}
    depends_on:
      - postgres
      - redis
    restart: unless-stopped
    deploy:
      replicas: 3

  worker:
    build: .
    command: /worker_server
    environment:
      - DB_HOST=postgres
      - REDIS_ADDR=redis:6379
    depends_on:
      - postgres
      - redis
    restart: unless-stopped
    deploy:
      replicas: 2

  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: anb_db
      POSTGRES_USER: anb_user
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    restart: unless-stopped

  redis:
    image: redis:7-alpine
    command: redis-server --appendonly yes
    volumes:
      - redis_data:/data
    restart: unless-stopped

volumes:
  postgres_data:
  redis_data:
```

---

**Plan de Pruebas de Carga - ANB Rising Stars Showcase v1.0**  
*Universidad de los Andes - MISO - Desarrollo de Software en la Nube*