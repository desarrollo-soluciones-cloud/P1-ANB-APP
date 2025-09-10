# Documentación del Proyecto

Este documento centraliza toda la información técnica requerida para la entrega del proyecto, incluyendo el modelo de datos, la documentación de la API, los diagramas arquitectónicos, los procesos de despliegue y los resultados de calidad de código.

---

## 1. Modelo de Datos

El modelo de datos de la aplicación se representa mediante un **Diagrama Entidad-Relación (ERD)**, el cual describe las entidades principales, sus atributos y las relaciones entre ellas.

### Diagrama Entidad-Relación

<img width="547" height="690" alt="diagrama-ERD" src="https://github.com/user-attachments/assets/721e0efa-a6bf-4ffe-b69e-da7aee26d331" />


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

## Descripción

Esta documentación contiene todos los endpoints de la Implementación de una API REST escalable con orquestación de tareas asíncronas para el procesamiento de archivos.

## Archivos Generados

1. **ANB API - Colección.postman_collection.json** - Colección principal con todos los endpoints
2. **ANB API - Local.postman_environment.json** - Entorno de desarrollo local


 [Ir a la carpeta backend](..backend/collectios/)

---

## 3. Diagrama de Componentes

<img width="725" height="721" alt="diagrama_componentes" src="https://github.com/user-attachments/assets/5bcee35e-ba98-4641-932e-ce0ff66ac6ec" />

---

## 4. Diagrama de Flujo de Procesos


## Diagrama de flujo de procesos

El siguiente diagrama representa el flujo completo de interacción dentro de la API. Resume cómo los usuarios y el público general pueden usar el sistema, desde el **registro de jugadores** hasta la **participación en votaciones** y la **consulta de rankings**.

![Diagrama de flujo del proceso](Diagrama%20de%20flujo%20de%20proceso%20API%20RESK.png)

###  Explicación del flujo

1. **Inicio del proceso**  
   El usuario puede optar por registrarse en la plataforma para participar activamente o, si no desea autenticarse, puede consultar directamente el ranking público de jugadores.

2. **Gestión de usuarios**  
   - **Registro de usuarios:** Los jugadores aficionados crean una cuenta. El sistema valida que el email no esté duplicado y que las contraseñas coincidan.  
   - **Login:** Una vez registrado, el jugador debe autenticarse con sus credenciales. Si la autenticación es correcta, obtiene un *token JWT* que será utilizado en todas las operaciones protegidas.

3. **Gestión de videos (requiere autenticación)**  
   - **Ver mis videos:** El usuario autenticado puede listar todos los videos que ha subido, junto con su estado de procesamiento.  
   - **Subir video:** Permite cargar un archivo en formato MP4. Inmediatamente se encola una tarea asíncrona de procesamiento (recorte, ajuste de formato y agregado de logos).  
   - **Eliminar video:** El sistema valida:  
     - Si el video ya fue procesado o publicado → *No se puede eliminar*.  
     - Si el video aún no está procesado → *Eliminado exitosamente*.  
   - **Marcar video como procesado:** Acción que actualiza el estado del archivo cuando el worker termina su tarea.

4. **Sistema de votación (requiere autenticación)**  
   - **Votar video:** El usuario registrado puede votar por un video público habilitado. Posibles resultados:  
     - *Voto registrado exitosamente*.  
     - *Ya has votado por este video*.  
     - *No está autenticado* (si falta el token).  
     - *Video no encontrado* (si el ID no existe o no pertenece a los videos públicos).  
   - **Quitar voto:** El jugador puede retirar su voto de un video. Resultado esperado: *Voto eliminado*.

5. **Ranking público (no requiere autenticación)**  
   - Cualquier usuario, sin necesidad de autenticarse, puede consultar la tabla de clasificación. Esta muestra a los jugadores ordenados según la cantidad de votos que recibieron sus videos.  
   - El sistema puede devolver error **400** si los parámetros de consulta del ranking son inválidos.


---

## 5. Despliegue y Documentación

*(pendiente de completar: infraestructura, contenedores Docker, servicios activos, guía de réplica del entorno)*

---

## 6. Reporte de Análisis de SonarQube

El siguiente reporte corresponde al análisis realizado con **SonarQube Community** sobre la rama `dev` del proyecto **P1-ANB-APP**.  

El resultado global indica que el proyecto **pasa el Quality Gate**, aunque con advertencias importantes que deben atenderse.

### Métricas principales

1. **Quality Gate**
   - **Estado:**  *Passed*  
   - El código cumple con las condiciones mínimas de calidad.  

2. **Security (Seguridad)**
   - **0 Open Issues** → Calificación **A**.  
   - No se detectaron vulnerabilidades.  
   - **Security Hotspots:** 7 → requieren revisión manual (uso de librerías, cifrado, credenciales).

3. **Reliability (Fiabilidad)**
   - **1 Open Issue** → Calificación **A**.  
   - Excelente nivel de fiabilidad.

4. **Maintainability (Mantenibilidad)**
   - **70 Open Issues** → Calificación **B**.  
   - Existen problemas de complejidad y legibilidad que afectan el mantenimiento del código.

5. **Coverage (Cobertura de pruebas)**
   - **0.0%** sobre **1.3k líneas a cubrir**.  
   - No se cuenta con pruebas unitarias, lo que incrementa el riesgo de errores no detectados.

6. **Duplications (Duplicación de código)**
   - **4.9%** sobre **8.6k líneas**.  
   - Se identificó casi un 5% de código duplicado, lo que dificulta el mantenimiento.

### Conclusiones
- **Fortalezas:**  
  - No hay vulnerabilidades críticas de seguridad.  
  - Muy buena fiabilidad (solo 1 issue).  
  - El Quality Gate está aprobado.  

- **Debilidades:**  
  - Cobertura de pruebas **0%** → urge implementar tests unitarios.  
  - Duplicación del **4.9%** → se recomienda refactorizar.  
  - **70 issues de mantenibilidad** → reducir complejidad y mejorar legibilidad.  
  - 7 *security hotspots* → revisión manual obligatoria.  

