# syntax=docker/dockerfile:1
# 自定义后端 Dockerfile：为二次开发保留，避免依赖项目原始 Dockerfile。
ARG GO_IMAGE=public.ecr.aws/docker/library/golang:1.24-alpine
ARG RUNTIME_IMAGE=public.ecr.aws/docker/library/alpine:latest

FROM ${GO_IMAGE} AS builder
WORKDIR /src
ENV GO111MODULE=on \
    GOPROXY=https://goproxy.cn,direct \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

COPY server/go.mod server/go.sum ./
RUN go mod download

COPY server/ ./
RUN go build -tags timetzdata -trimpath -ldflags="-s -w" -o /out/gva-server .

FROM ${RUNTIME_IMAGE} AS runtime
WORKDIR /app
ENV TZ=Asia/Shanghai \
    GIN_MODE=release

COPY --from=builder /out/gva-server /app/server
COPY server/go.mod /app/go.mod
COPY server/resource /app/resource
COPY deploy/docker-dev/config.init.yaml /app/config.yaml

EXPOSE 8888
ENTRYPOINT ["/app/server"]
CMD ["-c", "/app/config.yaml"]
