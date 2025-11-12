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
| Endpoint | Config | Total | Éxito | **Throughput real** | p50 (ms) | **p95 (ms)** | p99 (ms) | Max (ms) |
|---|---|---:|---:|---:|---:|---:|---:|---:|
| **POST /auth/login** | 60s @30 rps (conc 30) | 1,487 | **100%** | **24.78 rps** | 941.97 | **2,134.96** | 2,273.11 | 2,556.19 |

---

### 2.2 Escenario 1 – Carga sostenida moderada
| Endpoint | Config | Total | Éxito | **Throughput real** | p50 (ms) | **p95 (ms)** | p99 (ms) | Max (ms) |
|---|---|---:|---:|---:|---:|---:|---:|---:|
| **GET /public/videos** | 5m @90 rps (conc 60) | 26,962 | **≈100%** (1 error) | **89.88 rps** | 95.34 | **186.85** | 582.56 | 15,000.31 |
| **GET /videos** | 5m @110 rps (conc 80) | 32,987 | **100%** | **109.96 rps** | 96.24 | **179.71** | 557.11 | 2,842.49 |



---

### 2.3 Escenario 2 – Alta concurrencia (*plateau*, 7 min)
| Endpoint | Config | Total | Éxito | **Throughput real** | p50 (ms) | **p95 (ms)** | p99 (ms) | Max (ms) |
|---|---|---:|---:|---:|---:|---:|---:|---:|
| **GET /public/videos** | 7m @260 rps (conc 140) | 91,626 | **99.98%** (21 errores) | **218.16 rps** | 474.74 | **1,280.89** | 1,919.99 | 9,650.49 |
| **GET /videos** | 7m @220 rps (conc 120) | 92,222 | **99.92%** (73 errores) | **219.58 rps** | 96.17 | **352.45** | 1,048.27 | 5,838.58 |

---

### 2.4 Escenario 2 – *Burst* (pico corto, 60 s)
| Endpoint | Config | Total | Éxito | **Throughput real** | p50 (ms) | **p95 (ms)** | p99 (ms) | Max (ms) |
|---|---|---:|---:|---:|---:|---:|---:|---:|
| **GET /public/videos** | 60s @320 rps (conc 160) | 12,987 | **99.56%** (57 errores) | **216.45 rps** | 521.71 | **1,505.44** | 2,275.32 | 4,054.95 |
| **GET /videos** | 60s @280 rps (conc 150) | 16,773 | **99.58%** (70 errores) | **279.55 rps** | 96.94 | **982.98** | 1,522.85 | 7,119.12 |

---

### 2.5 Por-ID (ligero, estabilidad funcional)
| Endpoint | Config | Total | Éxito | **Throughput real** | p50 (ms) | **p95 (ms)** | p99 (ms) | Max (ms) |
|---|---|---:|---:|---:|---:|---:|---:|---:|
| **GET /videos/:id** | 1m @20 rps (conc 10) | 1,198 | **100%** | **19.97 rps** | 85.70 | **103.42** | 120.72 | 545.05 |
| **GET /videos/:id/download** | 1m @20 rps (conc 10) | 1,198 | **100%** | **19.97 rps** | 83.63 | **133.38** | 262.14 | 488.84 |
| **POST /videos/:id/mark-processed** | 30s @5 rps (conc 5) | 149 | **100%** | **4.97 rps** | 104.27 | **113.78** | 160.92 | 201.06 |
| **POST /public/videos/:id/vote** | 30s @5 rps (conc 5) | 149 | **100%** | **4.97 rps** | 99.27 | **104.82** | 115.85 | 214.52 |
| **DELETE /public/videos/:id/vote** | 30s @5 rps (conc 5) | 149 | **100%** | **4.97 rps** | 82.87 | **90.12** | 98.47 | 187.43 |

---

## 3) Notas operativas
- **Upload** no devuelve `id` directo; el `runner` resolvió el `video_id` vía *polling* por título (confirma flujo asíncrono).
- Todas las ejecuciones fueron **en AWS**; no se incluyen pruebas locales.
- Los CSV/JSON quedaron en la carpeta `resultados-YYYYMMDD-HHMMSS` generada por el `runner`.

## 4) Evidencias (archivos generados)
- `login_load.csv/json`, `public_videos_*.csv/json`, `videos_*.csv/json`
- `video_<id>_*.csv/json`, `vote_<id>_*.csv/json`
- `headers_auth.txt`, `headers_json.txt`, `body_login.json`, `upload_resp.json`

