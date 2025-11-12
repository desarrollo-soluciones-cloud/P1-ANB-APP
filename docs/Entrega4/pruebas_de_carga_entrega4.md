# Análisis de capacidad – ANB API (AWS)

**Ámbito:** Backend ANB API (auth, listados públicos/privados de videos, operaciones por-id, votos).  
**Ambiente:** `http://34.207.169.60:9090/api/v1`  
**Carpeta de salida:** `resultados-20251111-130015`

---


## 1) Resumen de endpoints cubiertos
- **Autenticación:** `POST /auth/login` 
- **Lecturas públicas:** `GET /public/videos` 
- **Lecturas privadas (listado):** `GET /videos` 
- **Operaciones por ID:** `GET /videos/:id`, `GET /videos/:id/download`  
  **Acciones:** `POST /videos/:id/mark-processed`, `POST/DELETE /public/videos/:id/vote`

> Métricas clave por bloque: **Total**, **% Éxito**, **Throughput real**, **p50**, **p95**, **p99**, **Max** (ms).

---

## 2) Resultados detallados

### 2.1 Autenticación (cobertura ligera)

![Autenticacion](Autenticacion.png)

---

### 2.2 Escenario 1 – Carga sostenida moderada


![Escenario 1](E1%20Carga%20sostenida%20moderada.png)

---

### 2.3 Escenario 2 – Alta concurrencia (*plateau*, 7 min)

![Escenario 1](E2%20Alta%20concurrencia.png)

---

### 2.4 Escenario 2 – *Burst* (pico corto, 60 s)

![Escenario 1](E2%20pico%20corto.png)

---

### 2.5 Por-ID (ligero, estabilidad funcional)

![Escenario 1](Por-ID.png)

---

## 3) Notas operativas
- Todas las ejecuciones fueron **en AWS**
- Los CSV/JSON quedaron en la carpeta `resultados-YYYYMMDD-HHMMSS` generada por el `runner`.

## 4) Evidencias (archivos generados)
- `login_load.csv/json`, `public_videos_*.csv/json`, `videos_*.csv/json`
- `video_<id>_*.csv/json`, `vote_<id>_*.csv/json`
- `headers_auth.txt`, `headers_json.txt`, `body_login.json`, `upload_resp.json`

