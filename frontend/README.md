# ANB Frontend

## Descripción

Frontend para la plataforma ANB (Asociación Nacional de Baloncesto) - una aplicación web moderna para gestión y votación de videos. Desarrollada con Angular 18, Angular Material y completamente containerizada con Docker.

La aplicación permite a los usuarios registrarse, autenticarse, subir videos, gestionar su contenido personal, votar por videos públicos y visualizar rankings en tiempo real.

## Arquitectura

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Angular App   │    │   Nginx Proxy   │    │   Backend API   │
│   (Development) │───▶│   (Production)  │───▶│   Go + Gin      │
│   Port 4200     │    │   Port 3001     │    │   Port 9090     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Componentes Principales

- **Angular 18**: Framework principal con standalone components
- **Angular Material**: Biblioteca de componentes UI
- **Nginx**: Servidor web para producción con proxy reverso
- **TypeScript**: Tipado estático y desarrollo moderno
- **SCSS**: Preprocesador CSS para estilos avanzados

## Tecnologías

- **Angular 18**: Framework web progresivo
- **Angular Material**: Componentes UI siguiendo Material Design
- **TypeScript**: Lenguaje de programación tipado
- **SCSS**: Preprocesador CSS
- **RxJS**: Programación reactiva
- **Angular HttpClient**: Cliente HTTP para API REST
- **Docker**: Containerización
- **Nginx**: Servidor web de producción

## Estructura del Proyecto

```
frontend/
├── src/
│   ├── app/
│   │   ├── components/
│   │   │   ├── auth/
│   │   │   │   ├── login/          # Inicio de sesión
│   │   │   │   └── register/       # Registro de usuarios
│   │   │   ├── dashboard/          # Layout principal
│   │   │   └── videos/
│   │   │       ├── detail/         # Detalle de video
│   │   │       ├── list/           # Lista personal de videos
│   │   │       ├── public/         # Videos públicos
│   │   │       ├── ranking/        # Tabla de clasificación
│   │   │       └── upload/         # Subida de videos
│   │   ├── guards/
│   │   │   └── auth.guard.ts       # Protección de rutas
│   │   ├── models/
│   │   │   ├── user.model.ts       # Modelo de usuario
│   │   │   └── video.model.ts      # Modelo de video
│   │   ├── services/
│   │   │   ├── auth.service.ts     # Servicio de autenticación
│   │   │   ├── video.service.ts    # Servicio de videos
│   │   │   └── api.service.ts      # Servicio base de API
│   │   ├── app.component.*         # Componente raíz
│   │   ├── app.config.ts           # Configuración de la app
│   │   └── app.routes.ts           # Configuración de rutas
│   ├── environments/
│   │   ├── environment.ts          # Configuración desarrollo
│   │   └── environment.prod.ts     # Configuración producción
│   ├── styles.scss                 # Estilos globales
│   └── index.html                  # Página principal
├── public/
│   ├── favicon.ico                 # Icono de la aplicación
│   └── anb.jpg                     # Logo ANB
├── angular.json                    # Configuración de Angular
├── package.json                    # Dependencias npm
├── tsconfig.json                   # Configuración TypeScript
├── Dockerfile                      # Imagen Docker
├── default.conf                    # Configuración Nginx
└── README.md                       # Este archivo
```

## Funcionalidades

### Autenticación
- **Registro de usuarios**: Formulario completo con validaciones
- **Inicio de sesión**: Autenticación JWT
- **Gestión de sesiones**: Almacenamiento seguro de tokens
- **Protección de rutas**: Guard para rutas autenticadas

### Gestión de Videos
- **Subida de videos**: Upload con validación de formato y tamaño
- **Lista personal**: Gestión de videos propios
- **Detalle de video**: Visualización completa con metadatos
- **Eliminación**: Borrado de videos propios

### Sistema Público
- **Videos públicos**: Catálogo de todos los videos
- **Sistema de votación**: Un voto por usuario por video
- **Rankings**: Clasificación por popularidad
- **Búsqueda y filtros**: Navegación eficiente

### Interfaz de Usuario
- **Material Design**: Diseño moderno y consistente
- **Responsive**: Adaptado a dispositivos móviles y desktop
- **Animaciones**: Transiciones suaves y feedback visual
- **Temas**: Soporte para temas claros y oscuros

## Instalación y Configuración

### Prerrequisitos

- Node.js 18+ y npm
- Angular CLI (opcional): `npm install -g @angular/cli`
- Docker y Docker Compose (para producción)

### Desarrollo Local

```bash
# 1. Clonar repositorio
git clone <repository-url>
cd frontend

# 2. Instalar dependencias
npm install

# 3. Configurar entorno
# Verificar src/environments/environment.ts

# 4. Iniciar servidor de desarrollo
ng serve

# 5. Abrir navegador
# http://localhost:4200
```

### Comandos de Desarrollo

```bash
# Servidor de desarrollo
ng serve --open

# Build de desarrollo
ng build

# Build de producción
ng build --configuration production

# Linter
ng lint

# Análisis de bundle
ng build --source-map
npx webpack-bundle-analyzer dist/anb-frontend/stats.json
```

## Docker

### Desarrollo con Docker

```bash
# Desde la raíz del proyecto
docker-compose up frontend -d

# Ver logs
docker-compose logs frontend

# Reconstruir imagen
docker-compose build frontend
```

### Producción

```bash
# Levantar stack completo
docker-compose up -d

# URLs de acceso
# Frontend: http://localhost:3001
# Backend: http://localhost:9090
```

### Dockerfile Multi-etapa

```dockerfile
# Etapa de construcción
FROM node:18-alpine AS build
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

# Etapa de producción
FROM nginx:alpine
COPY --from=build /app/dist/anb-frontend/browser /usr/share/nginx/html
COPY default.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

## Configuración de Entornos

### Development (environment.ts)

```typescript
export const environment = {
  production: false,
  apiUrl: 'http://localhost:9090'
};
```

### Production (environment.prod.ts)

```typescript
export const environment = {
  production: true,
  apiUrl: ''  // Usa nginx proxy
};
```

## API Integration

### Endpoints Principales

```typescript
// Autenticación
POST /api/v1/auth/signup     // Registro
POST /api/v1/auth/login      // Login

// Usuarios  
GET /api/v1/users/profile    // Perfil usuario
GET /api/v1/users/:user_id   // Usuario por ID

// Videos (Autenticados)
POST /api/v1/videos/upload          // Subir video
GET /api/v1/videos                  // Mis videos
GET /api/v1/videos/:video_id        // Detalle video
DELETE /api/v1/videos/:video_id     // Eliminar video

// Videos Públicos
GET /api/v1/public/videos           // Lista pública
GET /api/v1/public/rankings         // Rankings

// Votación
POST /api/v1/public/videos/:id/vote // Votar
```

### Servicios Angular

```typescript
// AuthService - Gestión de autenticación
login(credentials): Observable<any>
register(userData): Observable<any>
logout(): void
isAuthenticated(): boolean

// VideoService - Gestión de videos
uploadVideo(formData): Observable<any>
getMyVideos(): Observable<any>
getPublicVideos(): Observable<any>
voteVideo(videoId): Observable<any>
```

## Deployment

### Build de Producción

```bash
# Build optimizado
ng build --configuration production

# Verificar salida
ls -la dist/anb-frontend/browser/
```

### Nginx Configuration

```nginx
server {
    listen 80;
    server_name localhost;
    root /usr/share/nginx/html;
    index index.html;

    # SPA Configuration
    location / {
        try_files $uri $uri/ /index.html;
    }

    # API Proxy
    location /api/ {
        proxy_pass http://api:9090;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    # Static files caching
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }
}
```

## Flujo de Trabajo

### Usuario Típico

1. **Registro/Login**: Crear cuenta o iniciar sesión
2. **Dashboard**: Acceder al panel principal
3. **Subir Video**: Upload de contenido con metadatos
4. **Gestionar**: Ver y administrar videos personales
5. **Explorar**: Navegar videos públicos
6. **Votar**: Participar en el sistema de votación
7. **Rankings**: Consultar clasificaciones

### Desarrollador

1. **Setup**: Configurar entorno de desarrollo
2. **Development**: Implementar nuevas funcionalidades
3. **Testing**: Ejecutar pruebas unitarias y E2E
4. **Build**: Generar versión de producción
5. **Deploy**: Desplegar con Docker Compose

**ANB Frontend v1.0**  
*Desarrollado con Angular*
