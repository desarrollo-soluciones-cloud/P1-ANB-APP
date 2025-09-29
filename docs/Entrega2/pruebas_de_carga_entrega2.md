# Análisis de capacidad – ANB API (AWS **solo 44.198.15.64:9090**)

  
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
 
  
**Conclusión:** En AWS, la API **sostiene** hasta **120 rps** (público) y **100 rps** (privado) en lectura con **p95≈100 ms y 0% errores**. Se recomienda **ramp-up** (150–200 rps) y análisis dirigido de **/public/rankings**.

---

## 2) Escenarios y metodología
**Escenario 1 (moderado):** 30 rps, 2 min, conc. 20.  
**Escenario 2 (estrés):** 120 rps público / 100 rps privado, 3 min, conc. 60.  
**Herramientas:** `loadtest.go` (Go) con `run_local.ps1 -RunEsc2` (solo como orquestador, apuntando a AWS).  
**Métrica de éxito:** status < 500 y sin fallo de transporte.  
**Métricas recolectadas:** throughput, % éxito, latencias (mean/p50/**p95**/p99/max).

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

## 6) Recomendaciones para escalar
**Inmediatas (operación):**  
- Activar **tracing** (OpenTelemetry) y **slow logs** focalizados en **/public/rankings** (parámetros, duración, plan/índices).  
- Ajustar **timeouts** y circuit breakers solo en rutas seguras (idempotentes).  
- Monitorear **p95** y **tasa 5xx** por endpoint; alertas específicas para `rankings`.

**Base de datos / caché:**  
- Verificar e indexar consultas de `rankings` (joins/sorts); evaluar **materialized view** o **precomputo + Redis** (TTL 5–30s).  
- Revisar el **pool de conexiones** (límites y saturación) en hora pico.

**Infraestructura:**  
- **Auto Scaling** por p95/CPU/5xx; mínimo **2 réplicas** detrás de **ALB** para HA.  
- **CDN** para respuestas públicas calientes (rankings/listados).  
- Servir descargas desde **S3 + pre-signed URL** si el payload crece.


---

## 7) Conclusiones
- **Escenario 1 y 2 (AWS) cumplidos** con 0% errores y **p95 ≈ 95–100 ms** en los **listados principales**.  
- El sistema tiene **margen** para subir a **150–200 rps**.  
- **/public/rankings** requiere **optimización/observabilidad** antes de aumentar su carga.
