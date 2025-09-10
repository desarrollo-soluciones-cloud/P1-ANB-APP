# Documentación del Proyecto

Este documento centraliza toda la información técnica requerida para la entrega del proyecto, incluyendo el modelo de datos, la documentación de la API, los diagramas arquitectónicos, los procesos de despliegue y los resultados de calidad de código.

---

## 1. Modelo de Datos

El modelo de datos de la aplicación se representa mediante un **Diagrama Entidad-Relación (ERD)**, el cual describe las entidades principales, sus atributos y las relaciones entre ellas.

### Diagrama Entidad-Relación
[Ver diagrama ERD](./diagrama-ERD.png)

### Descripción de entidades y relaciones
- **Users**  
  Contiene la información básica de los usuarios registrados en la aplicación.  
  **Atributos principales**: `id`, `first_name`, `last_name`, `email`, `password`, `city`, `country`, `created_at`, `updated_at`.

- **Videos**  
  Representa los videos subidos por los usuarios.  
  **Atributos principales**: `id`, `user_id`, `title`, `status`, `original_url`, `processed_url`, `vote_count`, `uploaded_at`, `processed_at`.  
  **Relación**: cada video pertenece a un usuario (`users 1—N videos`).

- **Votes**  
  Registra los votos que realizan los usuarios sobre los videos.  
  **Atributos principales**: `id`, `video_id`, `user_id`, `voted_at`, `created_at`.  
  **Relaciones**:  
  - Un usuario puede emitir múltiples votos (`users 1—N votes`).  
  - Un video puede recibir múltiples votos (`videos 1—N votes`).

---

## 2. Documentación de la API

*(pendiente de completar: descripción de endpoints, colección Postman, ejemplos de request/response, pruebas ejecutadas)*

---

## 3. Diagrama de Componentes

*(pendiente de completar: representación de backend, worker, broker y base de datos)*

---

## 4. Diagrama de Flujo de Procesos

*(pendiente de completar: descripción detallada de carga, procesamiento y entrega de archivos)*

---

## 5. Despliegue y Documentación

*(pendiente de completar: infraestructura, contenedores Docker, servicios activos, guía de réplica del entorno)*

---

## 6. Reporte de Análisis de SonarQube

*(pendiente de completar: métricas de bugs, vulnerabilidades, code smells, cobertura de pruebas unitarias, duplicación de código, estado del Quality Gate)*

---
