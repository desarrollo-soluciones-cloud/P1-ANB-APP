FROM nginx:alpine

# Copiar template de configuración de nginx
COPY default.conf.template /etc/nginx/templates/default.conf.template

# Copiar archivos del build del frontend
COPY dist/anb-frontend/browser /usr/share/nginx/html

# Exponer puerto 80
EXPOSE 80

# Nginx automáticamente procesa archivos .template y sustituye variables de entorno
CMD ["nginx", "-g", "daemon off;"]
