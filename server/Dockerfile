FROM public.ecr.aws/docker/library/golang@sha256:8bee1901f1e530bfb4a7850aa7a479d17ae3a18beb6e09064ed54cfd245b7191 as builder

WORKDIR /app/server
COPY . .

RUN go env -w GO111MODULE=on \
    && go env -w GOPROXY=https://goproxy.cn,direct \
    && go env -w GOTOOLCHAIN=auto \
    && go env -w CGO_ENABLED=0 \
    && go env \
    && go mod download \
    && go build -o server .

FROM public.ecr.aws/docker/library/alpine@sha256:28bd5fe8b56d1bd048e5babf5b10710ebe0bae67db86916198a6eec434943f8b

LABEL MAINTAINER="MiTo"
# 设置时区
ENV TZ=Asia/Shanghai
RUN apk update && apk add --no-cache tzdata openntpd poppler-utils \
    && ln -sf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

WORKDIR /app/server

COPY --from=0 /app/server/server ./
COPY --from=0 /app/server/resource ./resource/
COPY --from=0 /app/server/config.docker.yaml ./

# 挂载目录：如果使用了sqlite数据库，容器命令示例：docker run -d -v /宿主机路径/maa.db:/app/server/maa.db -p 8888:8888 --name maa-server-v1 maa-server:1.0
# VOLUME ["/app/server"]

EXPOSE 8888
ENTRYPOINT ./server -c config.docker.yaml
