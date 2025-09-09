# 📘 ANB API – Documentación Integral

> Última actualización: 2025-09-08 00:03:50 (hora UTC)




---

## 3) Documentación de la API (Postman)
### Archivos
- `ANB_API_Postman_Collection.json` – colección con todos los endpoints.
- `ANB_API_Local.postman_environment.json` – entorno local (`{{base_url}} = http://localhost:9090`).

### Base URL
```
http://localhost:9090/api/v1
```

### Endpoints
#### Auth
- `POST /auth/signup` – Registro de jugadores (valida `email` único y contraseñas).
- `POST /auth/login` – Retorna `access_token` (JWT).

#### Videos (requiere JWT)
- `POST /videos/upload` – form-data (`title`, `video`).
- `GET /videos` – lista los videos del usuario.
- `GET /videos/:video_id` – detalle.
- `DELETE /videos/:video_id` – elimina si procede.
- `POST /videos/:video_id/mark-processed` – marca como procesado.

#### Público
- `GET /public/videos` – lista videos públicos.
- `GET /public/rankings` – ranking por votos.

#### Votes (requiere JWT)
- `POST /public/videos/:video_id/vote`
- `DELETE /public/videos/:video_id/vote`

### Pruebas sugeridas en Postman
1. **Login** y guardado automático del token (`{{token}}`) vía script de tests (incluido).
2. **Upload** con `form-data` (campo `video` como **File**).
3. **Flujo**: signup → login → upload → listar → consultar detalle → votar → ranking.

---



## 8) Anexos: ejemplos de requests

#### Ejemplo de request
```json
{
  "first_name": "John",
  "last_name": "Doe",
  "email": "john@example.com",
  "password": "StrongPass123",
  "password2": "StrongPass123",
  "city": "Bogotá",
  "country": "Colombia"
}
```

#### Códigos de respuesta
| Código | Descripción |
|--------|-------------|
| 201 | Usuario creado exitosamente |
| 400 | Error de validación (email duplicado, contraseñas no coinciden) |

---

### 2. Inicio de sesión
Autentica al usuario con email y contraseña. Devuelve un **JWT** que debe usarse en las solicitudes protegidas.

#### Ejemplo de request
```json
{
  "email": "john@example.com",
  "password": "StrongPass123"
}
```

#### Ejemplo de respuesta exitosa
```json
{
  "access_token": "eyJ0eXAiOiJKV1QiLCJhbGci...",
  "token_type": "Bearer",
  "expires_in": 3600
}
```

#### Códigos de respuesta
| Código | Descripción |
|--------|-------------|
| 200 | Autenticación exitosa, retorna token |
| 401 | Credenciales inválidas |

---

## Gestión de videos (carga, procesamiento y acceso)

### 1. Carga de video
Permite a un usuario autenticado subir un video. El archivo se almacena y se encola automáticamente una tarea de procesamiento asíncrono (recorte, ajuste 16:9, logos institucionales).

#### Parámetros (form-data)
| Nombre | Tipo | Requerido | Descripción |
|--------|------|-----------|-------------|
| `video` | archivo | Sí | Archivo MP4, máx. 100MB |
| `title` | string | Sí | Título del video |

#### Ejemplo de respuesta
```json
{
  "message": "Video subido correctamente. Procesamiento en curso.",
  "task_id": "123456"
}
```

#### Códigos de respuesta
| Código | Descripción |
|--------|-------------|
| 201 | Video subido, tarea creada |
| 400 | Error en archivo (tipo o tamaño inválido) |
| 401 | Falta de autenticación |

---

### 2. Consultar mis videos
Devuelve el listado de videos del usuario junto con estado de procesamiento.

#### Ejemplo de respuesta
```json
[
  {
    "video_id": "123456",
    "title": "Mi mejor tiro de 3",
    "status": "processed",
    "uploaded_at": "2025-03-10T14:30:00Z",
    "processed_at": "2025-03-10T14:35:00Z",
    "processed_url": "https://anb.com/videos/processed/123456.mp4"
  },
  {
    "video_id": "654321",
    "title": "Habilidades de dribleo",
    "status": "uploaded",
    "uploaded_at": "2025-03-11T10:15:00Z"
  }
]
```

#### Códigos de respuesta
| Código | Descripción |
|--------|-------------|
| 200 | Lista de videos obtenida |
| 401 | Falta de autenticación |

---

### 3. Consultar detalle de un video
Devuelve la información de un video específico.

#### Ejemplo de respuesta
```json
{
  "video_id": "a1b2c3d4",
  "title": "Tiros de tres en movimiento",
  "status": "processed",
  "uploaded_at": "2025-03-15T14:22:00Z",
  "processed_at": "2025-03-15T15:10:00Z",
  "original_url": "https://anb.com/uploads/a1b2c3d4.mp4",
  "processed_url": "https://anb.com/processed/a1b2c3d4.mp4",
  "votes": 125
}
```

#### Códigos de respuesta
| Código | Descripción |
|--------|-------------|
| 200 | Consulta exitosa |
| 401 | No autenticado o token inválido |
| 403 | No autorizado (no es propietario) |
| 404 | Video no encontrado |

---

### 4. Eliminar video
Permite borrar un video si aún no fue publicado o procesado.

#### Ejemplo de respuesta
```json
{
  "message": "El video ha sido eliminado exitosamente.",
  "video_id": "a1b2c3d4"
}
```

#### Códigos de respuesta
| Código | Descripción |
|--------|-------------|
| 200 | Video eliminado correctamente |
| 400 | No puede eliminarse (ya publicado) |
| 401 | No autenticado |
| 403 | No autorizado |
| 404 | Video no encontrado |

---

## Sistema de votación pública

### 1. Listar videos para votar
`GET /api/v1/public/videos`  
Devuelve todos los videos públicos disponibles.

---

### 2. Emitir voto
`POST /api/v1/public/videos/:video_id/vote`  
Registra un voto de un usuario autenticado.

#### Ejemplo de respuesta
```json
{
  "message": "Voto registrado exitosamente."
}
```

#### Códigos de respuesta
| Código | Descripción |
|--------|-------------|
| 200 | Voto exitoso |
| 400 | Ya votaste este video |
| 401 | Falta de autenticación |
| 404 | Video no encontrado |

---

## Ranking de jugadores

### 1. Consultar ranking
`GET /api/v1/public/rankings`  
Devuelve ranking de competidores según número de votos.

#### Ejemplo de respuesta
```json
[
  { "position": 1, "username": "superplayer", "city": "Bogotá", "votes": 1530 },
  { "position": 2, "username": "nextstar", "city": "Bogotá", "votes": 1495 }
]
```

#### Códigos de respuesta
| Código | Descripción |
|--------|-------------|
| 200 | Lista obtenida |
| 400 | Parámetro inválido |