# Documentación API ANB - Postman Collection

## Descripción

Colección de Postman para la API de la Academia Nacional de Baloncesto (ANB). Incluye todos los endpoints para registro de jugadores, autenticación, gestión de videos y sistema de votación.

## Configuración Rápida

### 1. Importar en Postman
1. Abre Postman
2. Click en **Import**
3. Arrastra o selecciona los archivos:
   - `ANB_Collection.json`
   - `ANB_Environment.json`

### 2. Configurar Ambiente
- Selecciona el ambiente **"ANB Environment"**
- Base URL configurada: `http://localhost:9090`

## Endpoints Disponibles

### Authentication
- `POST /api/v1/auth/signup` - Registro de jugadores
- `POST /api/v1/auth/login` - Inicio de sesión

### Videos
- `POST /api/v1/videos/upload` - Subir video
- `GET /api/v1/videos` - Mis videos
- `GET /api/v1/videos/{id}` - Detalle de video
- `GET /api/v1/videos/{id}/download` - Descargar video
- `DELETE /api/v1/videos/{id}` - Eliminar video
- `POST /api/v1/videos/{id}/mark-processed` - Marcar como procesado

### Public Videos
- `GET /api/v1/public/videos` - Videos públicos para votación

### Voting
- `POST /api/v1/public/videos/{id}/vote` - Votar por video
- `DELETE /api/v1/public/videos/{id}/vote` - Remover voto

### Rankings
- `GET /api/v1/public/rankings` - Tabla de clasificación

## Guía de Uso

### 1. Registro e Inicio de Sesión
```json
// 1. Registro (POST /api/v1/auth/signup)
{
  "first_name": "John",
  "last_name": "Doe",
  "email": "john@example.com",
  "password": "StrongPass123",
  "password2": "StrongPass123",
  "city": "Bogotá",
  "country": "Colombia"
}

// 2. Login (POST /api/v1/auth/login) 
{
  "email": "john@example.com",
  "password": "StrongPass123"
}
```

### 2. Gestión de Videos
- **Upload**: Usar form-data con `video` (archivo MP4) y `title`
- **List**: Obtener todos tus videos
- **Detail**: Ver detalles específicos de un video
- **Download**: Descargar archivo original o procesado

### 3. Sistema de Votación
- Ver videos públicos disponibles para votación
- Votar una vez por video (requiere autenticación)
- Ver rankings ordenados por votos

## Autenticación

La API usa tokens JWT. El token se guarda automáticamente después del login exitoso.

**Header requerido para endpoints protegidos:**
```
Authorization: Bearer {token}
```

## Tests Incluidos

Cada endpoint incluye tests que verifican:
- Códigos de respuesta correctos
- Estructura de datos válida  
- Manejo de errores
- Tiempos de respuesta

## Ejecución con Postman

### Orden Recomendado:
1. **Sign Up** (primera vez) o **Login**
2. **Upload Video** (con archivo MP4)
3. **List My Videos**
4. **Vote for Video** (usando video público)
5. **Get Rankings**

### Runner de Postman:
1. Click derecho en la colección
2. "Run collection"
3. Configurar delay de 500ms entre requests
4. Ejecutar

## Ejecución con Newman CLI

Newman es la herramienta de línea de comandos para ejecutar colecciones de Postman.

### Instalación
```bash
# Instalar Newman globalmente
npm install -g newman

# Opcional: Instalar reporter HTML para reportes más bonitos
npm install -g newman-reporter-html
```

### Comandos Básicos

#### Ejecutar colección completa
```bash
newman run ANB_Collection.json -e ANB_Environment.json
```

#### Ejecutar con reporte HTML
```bash
newman run ANB_Collection.json -e ANB_Environment.json --reporters html --reporter-html-export reporte.html
```

#### Ejecutar con delay entre requests
```bash
newman run ANB_Collection.json -e ANB_Environment.json --delay-request 500
```

#### Ejecutar solo una carpeta específica
```bash
newman run ANB_Collection.json -e ANB_Environment.json --folder "Authentication"
```

#### Ejecutar con múltiples reporters
```bash
newman run ANB_Collection.json -e ANB_Environment.json --reporters cli,html,json --reporter-html-export reporte.html --reporter-json-export resultados.json
```

### Opciones Útiles

| Opción | Descripción | Ejemplo |
|--------|-------------|---------|
| `--delay-request` | Delay en ms entre requests | `--delay-request 1000` |
| `--timeout` | Timeout por request en ms | `--timeout 10000` |
| `--reporters` | Tipos de reporte | `--reporters cli,html` |
| `--folder` | Ejecutar solo una carpeta | `--folder "Videos"` |
| `--iteration-count` | Número de iteraciones | `--iteration-count 3` |
| `--verbose` | Salida detallada | `--verbose` |
| `--silent` | Salida mínima | `--silent` |

### Ejemplos Prácticos

#### Test rápido de conectividad
```bash
newman run ANB_Collection.json -e ANB_Environment.json --folder "Authentication" --verbose
```

#### Test completo con reporte
```bash
newman run ANB_Collection.json -e ANB_Environment.json --delay-request 1000 --reporters html,cli --reporter-html-export test-results.html
```

#### Test de CI/CD (sin colores, formato JSON)
```bash
newman run ANB_Collection.json -e ANB_Environment.json --reporters json --reporter-json-export ci-results.json --color off
```

### Interpretación de Resultados

Newman mostrará:
- **Total requests**: Número total de requests ejecutados
- **Failed requests**: Requests que fallaron
- **Test scripts**: Tests ejecutados y su estado
- **Assertions**: Validaciones que pasaron/fallaron

Ejemplo de salida:
```
┌─────────────────────────┬──────────────────┬──────────────────┐
│                         │         executed │           failed │
├─────────────────────────┼──────────────────┼──────────────────┤
│              iterations │                1 │                0 │
├─────────────────────────┼──────────────────┼──────────────────┤
│                requests │               11 │                0 │
├─────────────────────────┼──────────────────┼──────────────────┤
│            test-scripts │               11 │                0 │
├─────────────────────────┼──────────────────┼──────────────────┤
│      prerequest-scripts │                0 │                0 │
├─────────────────────────┼──────────────────┼──────────────────┤
│              assertions │               25 │                0 │
└─────────────────────────┴──────────────────┴──────────────────┘
```

### Troubleshooting Newman

#### Error: Command not found
```bash
# Verificar instalación
newman --version

# Si no está instalado
npm install -g newman
```

#### Error de permisos en Windows
```cmd
# Ejecutar como administrador o usar:
npx newman run ANB_Collection.json -e ANB_Environment.json
```

#### Error de SSL (desarrollo local)
```bash
newman run ANB_Collection.json -e ANB_Environment.json --insecure
```

## Variables de Ambiente

- `base_url`: `http://localhost:9090`
- `access_token`: Se llena automáticamente
- `video_id`: ID para pruebas (cambiar según necesidad)

## Códigos de Respuesta

| Código | Significado |
|--------|-------------|
| 200 | OK - Operación exitosa |
| 201 | Created - Recurso creado |
| 400 | Bad Request - Error de validación |
| 401 | Unauthorized - No autenticado |
| 403 | Forbidden - Sin permisos |
| 404 | Not Found - Recurso no encontrado |

## Troubleshooting

**Error 401**: Token expirado → Ejecutar login nuevamente
**Error 404**: Video no encontrado → Verificar `video_id`
**Conexión falla**: Servidor no ejecutándose → Iniciar backend en puerto 9090
