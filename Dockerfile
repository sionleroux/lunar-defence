FROM nginx
COPY web /usr/share/nginx/html
RUN sed -i '/mp4/a application\/wasm wasm;' /etc/nginx/mime.types
