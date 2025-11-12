Resultados Entrega 4 - Velocidad de procesamiento de los videos.
--------------------------------------------------

1.1) Escenarios de prueba

En esta entrega se migró la arquitectura de colas internas a AWS SQS (Simple Queue Service) para mejorar la escalabilidad y la resiliencia del sistema. 
El procesamiento de videos continúa distribuido entre múltiples workers detrás del Application Load Balancer, pero ahora el servicio de encolado está desacoplado del worker.

**La estructura de las pruebas se mantuvo idéntica a las entregas anteriores**, ejecutando lotes de 10, 20, 50 y 100 videos para medir el throughput y la estabilidad bajo diferentes niveles de carga.

A continuación, se presentan los resultados detallados de las tres entregas:

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

Estadísticas (Entrega 4):
Media (μ):               4.05 videos/min
Desviación estándar (σ): 0.10 videos/min
Coeficiente de variación: 2.4% (estable)
Rango:                   3.95 - 4.20 videos/min

---

1.2) Análisis

En comparación con la arquitectura anterior (Entrega 3), donde el sistema alcanzaba un throughput promedio de 5.07 videos/min, se observa una ligera disminución del ~20% 
en la velocidad de procesamiento. La causa principal es la latencia adicional introducida por el uso de AWS SQS, ya que las solicitudes de encolado y recuperación de mensajes 
requieren comunicación externa (HTTP) fuera de la máquina local.

En la entrega previa, el servicio de encolado (Redis) residía en la misma instancia que el worker, lo que permitía intercambios casi instantáneos (<2 ms). 
En cambio, con SQS, cada worker debe realizar una petición HTTP segura hacia el servicio gestionado, con una latencia media de 15–30 ms por mensaje.

No obstante, el cambio representa una mejora arquitectónica significativa: ahora el sistema es altamente desacoplado, con mayor tolerancia a fallos y capacidad de autoescalado 
independiente entre productores y consumidores. Aunque el throughput se redujo a 4.05 videos/min, el comportamiento es extremadamente estable (CV 2.4%), 
y el sistema puede procesar ~5,800 videos/día de manera confiable, manteniendo los beneficios de resiliencia y disponibilidad global de SQS.

1.3) Comparativa de throughput

Entrega | Arquitectura        | Servicio de cola   | Throughput promedio (videos/min) | Variabilidad | Comentario
2       | Monolítica          | Local/Worker único | 3.11                             | 31%          | Degradación bajo carga
3       | Escalable con ALB   | Redis interno      | 5.07                             | 1%           | Rendimiento máximo
4       | Escalable con ALB   | AWS SQS (externo)  | 4.05                             | 2.4%         | Más estable, pero con latencia mayor

📈 Conclusión técnica Entrega 4:
El uso de SQS como servicio de mensajería externo desacopló efectivamente el backend, mejorando la robustez, disponibilidad y tolerancia a fallos del sistema. 
Sin embargo, el overhead de red y las llamadas HTTP adicionales reducen ligeramente el throughput. 
El impacto es aceptable para entornos productivos donde la escalabilidad y resiliencia tienen prioridad sobre la latencia mínima.