# Multi-stage build para optimizar el tamaño de la imagen
FROM node:18-alpine AS build

# Establecer directorio de trabajo
WORKDIR /app

# Copiar archivos de dependencias
COPY package*.json ./

# Instalar dependencias
RUN npm ci

# Copiar código fuente
COPY . .

# Construir la aplicación para producción
RUN npm run build

# Etapa de producción con Nginx
FROM nginx:alpine

# Copiar archivos construidos desde la etapa anterior (ajustando la ruta)
COPY --from=build /app/dist/anb-frontend/browser /usr/share/nginx/html

# Eliminar la configuración por defecto de Nginx y el index.html predeterminado
RUN rm /etc/nginx/conf.d/default.conf && rm -f /usr/share/nginx/html/index.html

# Copiar nuestra configuración personalizada
COPY default.conf /etc/nginx/conf.d/default.conf

# Renombrar index.csr.html a index.html si existe (para Angular SSR)
RUN if [ -f /usr/share/nginx/html/index.csr.html ]; then mv /usr/share/nginx/html/index.csr.html /usr/share/nginx/html/index.html; fi

# Exponer puerto 80
EXPOSE 80

# Comando por defecto
CMD ["nginx", "-g", "daemon off;"]
