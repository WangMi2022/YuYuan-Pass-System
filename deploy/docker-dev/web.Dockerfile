# syntax=docker/dockerfile:1
# 自定义前端 Dockerfile：为二次开发保留，避免依赖项目原始 Dockerfile。
ARG NODE_IMAGE=public.ecr.aws/docker/library/node:20-slim
ARG NGINX_IMAGE=nginx:alpine

FROM ${NODE_IMAGE} AS deps
WORKDIR /app
ENV npm_config_registry=https://registry.npmmirror.com
COPY web/package*.json ./
RUN npm install --legacy-peer-deps

FROM deps AS build
WORKDIR /app
COPY web/ ./
RUN npm run build

FROM ${NGINX_IMAGE} AS runtime
WORKDIR /usr/share/nginx/html
COPY deploy/docker-dev/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/dist /usr/share/nginx/html
EXPOSE 8080
CMD ["nginx", "-g", "daemon off;"]
