# syntax=docker/dockerfile:1.4

# 第一阶段：构建（自动适配平台）
FROM --platform=$BUILDPLATFORM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 启用构建缓存（兼容Docker BuildKit）
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=bind,source=go.mod,target=go.mod \
    --mount=type=bind,source=go.sum,target=go.sum \
    go mod download

# 复制源码并构建（自动处理换行符）
COPY . .
ARG TARGETOS TARGETARCH
RUN --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -ldflags="-w -s" -o /app/main .

# 第二阶段：运行（按平台选择基础镜像）
FROM alpine:3.19 AS linux-amd64
FROM alpine:3.19 AS linux-arm64
FROM mcr.microsoft.com/windows/nanoserver:1809 AS windows-amd64

# 最终镜像选择器
FROM ${TARGETOS}-${TARGETARCH} AS final

# 平台特定配置
WORKDIR /app
COPY --from=builder /app/main ./main

# Linux配置
RUN if [ "$TARGETOS" = "linux" ]; then \
      apk add --no-cache ca-certificates tzdata && \
      addgroup -S appgroup && adduser -S appuser -G appgroup && \
      chown appuser:appgroup /app/main; \
    fi

# Windows配置
USER ContainerAdministrator
RUN if [ "$TARGETOS" = "windows" ]; then \
      icacls "C:\app" /grant "Everyone:(OI)(CI)F"; \
    fi
USER ContainerUser

EXPOSE 8080

# 跨平台启动命令
CMD if [ "$TARGETOS" = "windows" ]; then \
      .\main.exe; \
    else \
      ./main; \
    fi
