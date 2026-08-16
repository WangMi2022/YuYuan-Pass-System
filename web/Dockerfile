# 如果需要用 cicd ，请设置环境变量：
# variables:
#     DOCKER_BUILDKIT: 1

FROM public.ecr.aws/docker/library/node@sha256:2cf067cfed83d5ea958367df9f966191a942351a2df77d6f0193e162b5febfc0 AS base
ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
RUN corepack enable
COPY . /app
WORKDIR /app


FROM base AS prod-deps
RUN --mount=type=cache,id=pnpm,target=/pnpm/store pnpm install --prod

FROM base AS build
COPY --from=prod-deps /app/node_modules /app/node_modules
RUN --mount=type=cache,id=pnpm,target=/pnpm/store pnpm install && pnpm run build


FROM registry.cn-zhangjiakou.aliyuncs.com/yunli_mid_platform/nginx@sha256:97a145fb5809fd90ebdf66711f69b97e29ea99da5403c20310dcc425974a14f9
LABEL MAINTAINER="bypanghu@163.com"
COPY --from=base  /app/.docker-compose/nginx/conf.d/my.conf /etc/nginx/conf.d/my.conf
COPY --from=build  /app/dist /usr/share/nginx/html
RUN ls -al /usr/share/nginx/html
