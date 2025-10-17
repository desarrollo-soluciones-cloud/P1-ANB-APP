FROM nginx:alpine

# Copiar configuración de nginx
COPY default.conf /etc/nginx/conf.d/default.conf

# Copiar archivos del build del frontend
COPY dist/anb-frontend/browser /usr/share/nginx/html

# Exponer puerto 80
EXPOSE 80

# Comando para iniciar nginx
CMD ["nginx", "-g", "daemon off;"]
