FROM nginx:1.27-alpine
# API_URL is substituted into the nginx template at container start
ENV API_URL=https://apilayer.aisupportmm.com
COPY nginx.conf.template /etc/nginx/templates/default.conf.template
COPY index.html /usr/share/nginx/html/index.html
EXPOSE 80
