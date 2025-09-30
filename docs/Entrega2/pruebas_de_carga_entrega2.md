# Análisis de capacidad – ANB API 

  
**Ámbito:** Backend ANB API (auth, videos públicos/privados, rankings, votos).  
 corridas en AWS contra `http://44.198.15.64:9090/api/v1` (última corrida: carpeta `resultados-20250927-192706`).

---

## 1) Resumen ejecutivo
- Se ejecutaron los **dos escenarios** solicitados por el enunciado sobre **/public/videos** y **/videos** (privado):  
  - **Escenario 1 (moderado):** 30 rps, 2 min, concurrencia 20.  
  - **Escenario 2 (estrés):** 120 rps (público) / 100 rps (privado), 3 min, concurrencia 60.  
- **Estabilidad de lectura (core):** 0% errores en ambos endpoints críticos para ambos escenarios. **Throughput real ≈ plan**.  
- **Latencia (lecturas):** p50≈82–85 ms y **p95≈95–100 ms** (AWS). Picos máx. aislados ~396 ms sin impacto en throughput sostenido.  
- **Hallazgos adicionales:**  
  - **/public/rankings** presentó **89.57% de éxito** (125/1198 errores) y **p95≈171 ms** ⇒ investigar.  
  - **Endpoints por id** (`/videos/:id`, `download`, `mark-processed`, votos **POST/DELETE**) estables al ritmo ligero (5–20 rps).
- **Escenario 3 (Upload + Worker):** Se evaluó el comportamiento del endpoint **/videos/upload** y el procesamiento asíncrono del worker. Con 1000 requests de prueba, el sistema mostró **0% errores** con latencia promedio de **~1732 ms**. El worker procesó videos a una tasa de **2.5–3.66 videos/minuto** bajo diferentes niveles de concurrencia

---

## 2) Escenarios y metodología
**Escenario 1 (moderado):** 30 rps, 2 min, conc. 20.  
**Escenario 2 (estrés):** 120 rps público / 100 rps privado, 3 min, conc. 60.  
**Herramientas:** `loadtest.go` (Go) con `runner.go -RunEsc2` (solo como orquestador, apuntando a AWS).  
**Métrica de éxito:** status < 500 y sin fallo de transporte.  
**Métricas recolectadas:** throughput, % éxito, latencias (mean/p50/**p95**/p99/max).
**Escenario 3 (Upload + Procesamiento):** Evaluación del endpoint `/videos/upload` con Apache JMeter y medición del tiempo de procesamiento del worker mediante logs de AWS.
**Apache JMeter** para upload de videos, logs de AWS CloudWatch para análisis del worker.  
---

## 3) Resultados – Lecturas críticas (AWS)

![Escenarios 1 y 2](Resultados%20Lecturas%20criticas%20(AWS).png)

---

## 4) Cobertura adicional (ligera)

![Carga Ligera](Cobertura%20Adicional.png)

---

## 5) Análisis
1) **Capacidad de lectura (core):** Se sostienen **120 rps** público y **100 rps** privado con **p95≈100 ms** y **0% errores**.  
2) **Latencias:** p50≈83 ms (RTT+proc corto) y **p95 < 100–101 ms** en estrés; picos máx. aislados ~396 ms sin impacto material.  
3) **Variabilidad en /public/rankings:** La caída a **89.57% éxito** sugiere **consulta lenta, dependencia externa o lock**. Contrasta con `/public/videos` (100%).  
4) **Endpoints por id:** Comportamiento estable en esta corrida; mantener vigilancia si se incrementa el tráfico por id.

---

*Análisis del endpoint:**
- **Tiempo de respuesta (Load time):** El endpoint responde en promedio en **1.7 segundos** para aceptar el upload del video de 3.63 MB. Este tiempo incluye la transferencia completa del archivo, su almacenamiento en disco y el encolamiento de la tarea de procesamiento en Redis/Asynq.
- **Latencia:** **1732 ms** indica que el servidor tarda ese tiempo en enviar el primer byte de respuesta, lo cual es consistente con el Load time dado que la respuesta es pequeña (375 bytes de JSON con `task_id`).
- **Tiempo de conexión:** **94 ms** refleja una latencia de red estable entre el cliente JMeter y el servidor AWS.
- **Tasa de éxito:** **100%** (0 errores) en la prueba inicial de 1000 requests, demostrando estabilidad del endpoint bajo carga moderada.
- **Throughput observado:** El endpoint es capaz de recibir y encolar videos de manera confiable, delegando el procesamiento pesado al worker asíncrono.

![Response Time Graph](response_time_graph.png)

El gráfico de tiempos de respuesta muestra estabilidad general alrededor de **1000 ms** con picos aislados de hasta **1400–1500 ms**, sin degradación progresiva durante la prueba.

---

### 5.4) Resultados del Worker de procesamiento

Para medir la capacidad real de procesamiento del worker, se monitorearon los logs de AWS CloudWatch durante pruebas con diferentes niveles de concurrencia de uploads. El worker procesa videos aplicando transformaciones con FFmpeg (recorte, escalado, concatenación de intro/outro).

#### Tabla de resultados

| Inicio  | Final   | Duración | Videos procesados | Videos/minuto |
|---------|---------|----------|-------------------|---------------|
| 2:52:10 | 2:54:54 | 2:44     | 10                | 3.66          |
| 3:02:19 | 3:07:49 | 5:30     | 20                | 3.64          |
| 3:02:52 | 3:22:52 | 20:00    | 50                | 2.50          |
| 3:27:48 | 3:55:19 | 17:31    | 100               | 3.63          |

**Análisis del worker:**

1. **Capacidad nominal:** El worker procesa entre **2.5 y 3.66 videos por minuto** (promedio **~3.1 videos/min** o **0.052 videos/segundo**).

2. **Consistencia bajo carga ligera:** Con 10 y 20 videos simultáneos, el worker mantiene una tasa estable de **~3.65 videos/min**, indicando que opera dentro de su capacidad.

3. **Degradación con carga alta:** Al procesar 50 videos, la tasa cae a **2.5 videos/min** (31% de reducción), sugiriendo:
   - Saturación de CPU durante el procesamiento con FFmpeg
   - Posible contención en I/O de disco
   - Cola de tareas acumulándose en Redis

4. **Tiempo de procesamiento por video:** Basado en la tasa observada, cada video tarda aproximadamente **19–24 segundos** en completar su transformación (recorte a 30s + escalado + concatenación de intro/outro).

5. **Bottleneck identificado:** El worker es el cuello de botella del sistema para la escritura. Mientras el endpoint de upload puede recibir decenas de requests por segundo, el worker solo procesa **~3 videos/minuto**.

---

## 6) Análisis integrado

### 6.1) Estabilidad general
1) **Capacidad de lectura (core):** Se sostienen **120 rps** público y **100 rps** privado con **p95≈100 ms** y **0% errores**.  
2) **Latencias de lectura:** p50≈83 ms (RTT+proc corto) y **p95 < 100–101 ms** en estrés; picos máx. aislados ~396 ms sin impacto material.  
3) **Variabilidad en /public/rankings:** La caída a **89.57% éxito** sugiere **consulta lenta, dependencia externa o lock**. Contrasta con `/public/videos` (100%).  
4) **Endpoints por id:** Comportamiento estable en esta corrida; mantener vigilancia si se incrementa el tráfico por id.
5) **Upload de videos:** Endpoint estable con **0% errores** y latencia de **~1.7s** para videos de 3.63 MB. Maneja carga moderada sin degradación.
6) **Procesamiento asíncrono:** Worker limitado a **2.5–3.7 videos/min** por instancia. Este es el principal limitante para escritura bajo carga sostenida.

### 6.2) Arquitectura de escritura vs lectura

**Lecturas (GET):**
- Alta capacidad: **120+ rps** sostenidos
- Baja latencia: **p95 < 100 ms**
- Escalamiento horizontal trivial

**Escrituras (POST /upload + Worker):**
- Endpoint de upload: Alta capacidad de recepción (**>50 rps** estimado)
- Worker: Baja capacidad de procesamiento (**~3 videos/min/instancia**)
- **Escalamiento:** Requiere múltiples instancias del worker para aumentar throughput de procesamiento

Esta arquitectura asíncrona desacopla correctamente la recepción de videos (rápida) del procesamiento (lento e intensivo en CPU), permitiendo que la API permanezca responsiva incluso cuando la cola de procesamiento crece.

---

## 7) Recomendaciones para escalar

### 7.1) Inmediatas (operación)
- Activar **tracing** (OpenTelemetry) y **slow logs** focalizados en **/public/rankings** (parámetros, duración, plan/índices).  
- Ajustar **timeouts** y circuit breakers solo en rutas seguras (idempotentes).  
- Monitorear **p95** y **tasa 5xx** por endpoint; alertas específicas para `rankings`.

### 7.2) Base de datos / caché
- Verificar e indexar consultas de `rankings` (joins/sorts); evaluar **materialized view** o **precomputo + Redis** (TTL 5–30s).  
- Revisar el **pool de conexiones** (límites y saturación) en hora pico.

### 7.3) Infraestructura
- **Auto Scaling** por p95/CPU/5xx; mínimo **2 réplicas** detrás de **ALB** para HA.  
- **CDN** para respuestas públicas calientes (rankings/listados).

### 7.4) Procesamiento de videos (Escenario 3)
- **Escalar workers horizontalmente:** Desplegar **múltiples instancias del worker** (3–5 réplicas) para alcanzar **9–18 videos/min** de throughput agregado.
- **Optimizar FFmpeg:** Evaluar parámetros de preset (`-preset ultrafast` vs `fast`) para balance entre velocidad y calidad.
- **Priorización de cola:** Implementar colas con prioridad en Asynq para videos críticos o usuarios premium.
- **Monitoreo de cola:** Alertas cuando el tamaño de la cola en Redis supere umbrales (ej: >100 tareas pendientes).
- **Límite de uploads:** Considerar rate limiting en el endpoint de upload (ej: 10 uploads/min por usuario) para prevenir saturación de la cola.
- **Procesamiento paralelo:** Si el hardware lo permite, configurar el worker para procesar **2 videos simultáneamente** por instancia (requiere CPU multicore y verificar contención de I/O).

---

## 8) Conclusiones
- **Escenarios 1 y 2 (lecturas en AWS):** Cumplidos con 0% errores y **p95 ≈ 95–100 ms** en los **listados principales**.  
- El sistema tiene **margen** para subir a **150–200 rps** en operaciones de lectura.  
- **/public/rankings** requiere **optimización/observabilidad** antes de aumentar su carga.
- **Escenario 3 (Upload + Worker):**
  - El endpoint de upload es **robusto y estable**, aceptando videos con **0% errores** y latencia de **~1.7s**.
  - El worker es el **cuello de botella** con capacidad de **2.5–3.7 videos/min** por instancia.
  - Para alcanzar **10+ videos/min**, se requieren **3–4 instancias del worker** en paralelo.
  - La arquitectura asíncrona (upload inmediato + procesamiento en background) es **adecuada** para este tipo de carga.
- **Recomendación general:** El sistema está listo para producción en lecturas; para escrituras, escalar workers antes de campañas de alto volumen de uploads.
