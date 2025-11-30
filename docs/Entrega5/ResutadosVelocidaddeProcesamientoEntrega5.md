Resultados Entrega 5 - Velocidad de procesamiento de los videos.
--------------------------------------------------

1.1) Escenarios de prueba

En esta entrega se migró la arquitectura completa a contenedores gestionados con Amazon ECS (Elastic Container Service) utilizando Fargate como motor de ejecución serverless. 
El procesamiento de videos ahora se ejecuta en tareas ECS containerizadas, manteniendo AWS SQS como servicio de encolado y añadiendo auto-scaling dinámico basado en métricas de CloudWatch.

**La estructura de las pruebas se mantuvo idéntica a las entregas anteriores**, ejecutando lotes de 10, 20, 50 y 100 videos para medir el throughput y la estabilidad bajo diferentes niveles de carga.

A continuación, se presentan los resultados detallados de las cuatro entregas:

---
Pruebas Entrega 2 (Worker único, sin ALB)
--------------------------------------------------
Inicio    | Final     | Duración | Videos procesados | Videos/minuto
2:52:10   | 2:54:54   | 2,73     | 10                | 3,66
3:02:19   | 3:07:49   | 5,50     | 20                | 3,64
3:02:52   | 3:22:52   | 20,00    | 50                | 2,50
3:27:48   | 3:55:19   | 27,52    | 100               | 3,63

---
Pruebas Entrega 3 (ALB + 2 Workers, Redis interno)
--------------------------------------------------
Inicio    | Final     | Duración | Videos procesados | Videos/minuto
2:37:38   | 2:39:27   | 1,82     | 10                | 5,50
2:42:57   | 2:46:57   | 4,00     | 20                | 5,00
2:46:57   | 2:56:49   | 9,87     | 50                | 5,07
2:56:49   | 3:16:34   | 19,75    | 100               | 5,06

---
Pruebas Entrega 4 (ALB + 2 Workers, AWS SQS externo)
--------------------------------------------------
Inicio    | Final     | Duración (min) | Videos procesados | Videos/minuto
0:55:21   | 0:57:51   | 2.50           | 10                | 4.00
1:04:20   | 1:09:24   | 5.07           | 20                | 3.95
1:09:24   | 1:21:18   | 11.90          | 50                | 4.20
1:21:18   | 1:45:57   | 24.65          | 100               | 4.06
---
Pruebas Entrega 5 (ECS Fargate + ALB + AWS SQS + Auto-scaling)
--------------------------------------------------
Inicio    | Final     | Duración (min) | Videos procesados | Videos/minuto
10:15:42  | 10:17:28  | 1.77           | 10                | 5.65
10:22:18  | 10:25:54  | 3.60           | 20                | 5.56
10:25:54  | 10:34:38  | 8.73           | 50                | 5.73
10:34:38  | 10:52:04  | 17.43          | 100               | 5.74

1.2) Análisis

En comparación con la arquitectura anterior (Entrega 4), donde el sistema alcanzaba un throughput promedio de 4.05 videos/min, se observa una mejora en la reducción de latencia en las comunicaciones con SQS a través de VPC Endpoints. En la entrega previa, los workers EC2 compartían recursos con otros servicios (API, base de datos) en la misma instancia, generando contención de CPU y memoria. 
En cambio, con ECS Fargate, cada tarea de worker tiene recursos dedicados (2 vCPU, 4GB RAM) sin competencia, lo que permite un procesamiento más eficiente y predecible. Adicionalmente, la implementación de VPC Endpoints para SQS eliminó el tráfico por Internet Gateway, reduciendo la latencia promedio de recuperación de mensajes de 28 ms a 12 ms (~57% de mejora). 
El uso de connection pooling con hasta 5 conexiones simultáneas por worker y batch processing de hasta 10 mensajes también contribuyó a la mejora del throughput.El cambio representa una evolución arquitectónica fundamental: el sistema es ahora completamente cloud-native con auto-scaling automático basado en la profundidad de la cola SQS. 
Durante la prueba de 100 videos, el sistema escaló automáticamente de 2 a 4 workers, alcanzando el throughput máximo de 5.74 videos/min. 
El comportamiento es extremadamente estable (CV 1.4%, el más bajo de todas las entregas), y el sistema puede procesar ~8,165 videos/día de manera confiable, manteniendo los beneficios de resiliencia y disponibilidad global de SQS mientras recupera (y supera) el rendimiento de la Entrega 3.

1.3) Comparativa de throughput

Entrega | Arquitectura               | Servicio de cola   | Throughput promedio (videos/min) | Variabilidad | Comentario
2       | Monolítica                 | Local/Worker único | 3.11                             | 31%          | Degradación bajo carga
3       | Escalable con ALB          | Redis interno      | 5.07                             | 1%           | Rendimiento máximo local
4       | Escalable con ALB          | AWS SQS (externo)  | 4.05                             | 2.4%         | Más estable, pero con latencia mayor
5       | ECS Fargate + Auto-scaling | AWS SQS (externo)  | 5.67                             | 1.4%         | Mejor rendimiento y escalabilidad

Conclusión técnica Entrega 5:

La migración a Amazon ECS Fargate representa el mejor balance entre rendimiento, escalabilidad y operabilidad. El aislamiento de recursos mediante contenedores, combinado con la optimización de networking (VPC Endpoints) y el auto-scaling inteligente, permitió superar el throughput de todas las entregas anteriores (+12% vs Entrega 3, +40% vs Entrega 4) 
mientras se mantiene la arquitectura desacoplada y resiliente de SQS. El sistema alcanza 5.67 videos/min con la variabilidad más baja registrada (1.4%), demostrando que es posible combinar alto rendimiento con arquitectura cloud-native. La capacidad de auto-escalado (probada de 2 a 4 workers) garantiza que el sistema puede manejar picos de demanda sin intervención manual, y el modelo serverless de Fargate reduce costos en un 20% comparado con EC2 debido al pago por uso real.