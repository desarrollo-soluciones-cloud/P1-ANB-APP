# ANB App (Frontend)

Este es el frontend para la Asociación Nacional de Baloncesto (ANB). La aplicación está construida con **Angular 18**, **Angular Material** y se conecta al backend a través de una API REST. Todo el entorno está containerizado con **Docker**.

## Tecnologías Utilizadas
* **Framework:** Angular 18
* **UI Library:** Angular Material
# ANB App (Frontend)

Frontend de la Asociación Nacional de Baloncesto (ANB). Esta versión está enfocada en la gestión y consumo de contenidos de video: subida, listado personal, detalle, listado público, votación y tabla de clasificación.

La aplicación está construida con Angular (v18) y Angular Material. Se comunica con el backend REST que corre en el puerto 9090 en desarrollo.

## Tecnologías
- Angular 18
- Angular Material
- TypeScript
- SCSS
- Angular HttpClient
- Docker & Docker Compose (para orquestar frontend + backend)

---
## Prerrequisitos
Instalaciones recomendadas para desarrollo:
- Docker y Docker Compose (recomendado para levantar todo el stack)
- Node.js 18+ y npm (para desarrollo local)
- Angular CLI (opcional): `npm i -g @angular/cli`

---
## Levantar todo con Docker (recomendado)
Desde la raíz del repositorio (donde está `docker-compose.yml`):

```bash
docker-compose up --build
```

Después de iniciarse:

- Frontend: http://localhost:3000
- Backend API: http://localhost:9090

Para detener:

```bash
docker-compose down
```

---
## Desarrollo local (sin Docker)

1) Instalar dependencias

```bash
cd frontend
npm install
```

2) Levantar servidor de desarrollo

```bash
ng serve --open
```

Por defecto la app queda en `http://localhost:4200/`.

3) Build de producción

```bash
ng build
```

---
## Estructura principal (resumen)

```
frontend/
├── src/
│   └── app/
│       ├── components/
│       │   ├── auth/            # login / register
│       │   ├── dashboard/       # shell / sidebar
│       │   └── videos/          # upload, list, detail, public, ranking
│       ├── services/            # api services (video.service, auth.service, api.service)
│       └── guards/              # auth guard
```

---
## Funcionalidades principales

- Autenticación (registro / login)
- Subida de videos (multipart/form-data)
- Listado de mis videos
- Detalle de video (muestra processed_url, votos, metadata)
- Eliminación de video (el backend aplica reglas de negocio)
- Listado público de videos para votar
- Votar / quitar voto en un video (un voto por usuario)
- Tabla de clasificación (rankings)

---
## Contrato API (resumen)

NOTA: El backend no debe modificarse desde este repo; la app asume que la API corre en `http://localhost:9090`.

- Autenticación
  - `POST /api/v1/auth/signup`  - Registro
  - `POST /api/v1/auth/login`   - Login (devuelve JWT)

- Videos (autenticado)
  - `POST /api/v1/videos/upload`           - Subir video (multipart/form-data)
      - Campos: `video` (file, .mp4), `title` (string)
      - Restricción en frontend: archivo .mp4 y tamaño máximo recomendado 100MB
  - `GET /api/v1/videos`                    - Listar mis videos
  - `GET /api/v1/videos/:video_id`         - Detalle de un video (incluye `processed_url`, `votes`)
  - `DELETE /api/v1/videos/:video_id`      - Eliminar video

- Public (público / votación)
  - `GET /api/v1/public/videos`            - Listado público de videos
  - `POST /api/v1/public/videos/:id/vote`  - Votar por video (autenticado)
  - `DELETE /api/v1/public/videos/:id/vote`- Quitar voto (autenticado)
  - `GET /api/v1/public/rankings`          - Tabla de clasificación

### Notas sobre autenticación
- La app utiliza JWT en `localStorage` (AuthService gestiona cabeceras Authorization).
- Los endpoints protegidos requieren el header `Authorization: Bearer <token>`.

---
## Pruebas rápidas / flujo de trabajo

1. Levanta backend (por ejemplo con `docker-compose`) en `:9090`.
2. Levanta frontend (`ng serve`) o usa Docker.
3. Regístrate / loguea.
4. Sube un video desde la sección "Upload" (campo `video` y `title`).
5. Revisa "Mis videos" y abre el detalle para ver `processed_url` y votos.
6. Revisa "Public videos" y prueba votar / quitar voto.
7. Abre "Tabla de clasificación" para ver el ranking.

---
## Observaciones y próximos pasos

- El listado público actualmente no incluye, de forma consistente, un flag `voted_by_current_user` en todas las respuestas; si necesitas mostrar el estado de voto al cargar la lista, lo ideal es que el backend devuelva ese campo o un endpoint adicional que liste los IDs votados por el usuario autenticado.
- Se han eliminado las antiguas funcionalidades de gestión de tareas/categorías para dejar la app enfocada en videos.

---
Si necesitas que añada ejemplos de llamadas curl o scripts de pruebas automáticas para upload/voto/rankings, lo agrego.
