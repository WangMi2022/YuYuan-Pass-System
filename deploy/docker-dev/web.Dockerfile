# syntax=docker/dockerfile:1
# 自定义前端 Dockerfile：为二次开发保留，避免依赖项目原始 Dockerfile。
ARG NODE_IMAGE=public.ecr.aws/docker/library/node@sha256:2cf067cfed83d5ea958367df9f966191a942351a2df77d6f0193e162b5febfc0
ARG NGINX_IMAGE=registry.cn-zhangjiakou.aliyuncs.com/yunli_mid_platform/nginx@sha256:97a145fb5809fd90ebdf66711f69b97e29ea99da5403c20310dcc425974a14f9

FROM ${NODE_IMAGE} AS deps
WORKDIR /app
ENV npm_config_registry=https://registry.npmmirror.com
COPY web/package*.json ./
RUN npm ci --legacy-peer-deps

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
