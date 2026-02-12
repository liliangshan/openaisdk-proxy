FROM golang:1.25-alpine AS builder

WORKDIR /app

# 复制依赖文件
COPY go.mod go.mod

# 下载依赖
RUN go mod download

# 复制源代码
COPY main.go .

# 构建
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o cert-tool main.go

# 最终镜像
FROM alpine:latest

RUN apk --no-cache add ca-certificates certbot

WORKDIR /app

# 从构建阶段复制可执行文件
COPY --from=builder /app/cert-tool .

# 创建证书目录
RUN mkdir -p /app/cert

# 设置时区
RUN apk add --no-cache tzdata

# 设置默认时区
ENV TZ=Asia/Shanghai

# 创建便捷脚本
COPY <<'EOF' /usr/local/bin/cert-tool
#!/bin/sh
exec /app/cert-tool "$@"
EOF

RUN chmod +x /usr/local/bin/cert-tool

ENTRYPOINT ["/usr/local/bin/cert-tool"]
