# syntax=docker/dockerfile:1
# 自定义后端 Dockerfile：为二次开发保留，避免依赖项目原始 Dockerfile。
# The registry mirror does not publish the Go 1.25 Alpine tag. This digest is
# the verified Go 1.24.13 bootstrap image; GOTOOLCHAIN downloads the exact
# go1.25.13 toolchain through the pinned GOPROXY before compiling.
ARG GO_IMAGE=public.ecr.aws/docker/library/golang@sha256:8bee1901f1e530bfb4a7850aa7a479d17ae3a18beb6e09064ed54cfd245b7191
ARG RUNTIME_IMAGE=public.ecr.aws/docker/library/alpine@sha256:28bd5fe8b56d1bd048e5babf5b10710ebe0bae67db86916198a6eec434943f8b

FROM ${GO_IMAGE} AS builder
WORKDIR /src
ENV GO111MODULE=on \
    GOPROXY=https://goproxy.cn,direct \
    GOTOOLCHAIN=auto \
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

RUN apk add --no-cache poppler-utils

COPY --from=builder /out/gva-server /app/server
COPY server/go.mod /app/go.mod
COPY server/resource /app/resource
COPY deploy/docker-dev/config.init.yaml /app/config.yaml

EXPOSE 8888
ENTRYPOINT ["/app/server"]
CMD ["-c", "/app/config.yaml"]
