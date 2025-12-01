# 构建阶段
FROM golang:1.25.2-alpine AS builder

# 安装必要的构建工具
RUN apk add --no-cache git make

# 设置工作目录
WORKDIR /build

# 复制 go.work 和 go.work.sum
COPY go.work go.work.sum ./

# 复制所有模块的 go.mod 和 go.sum
COPY go.mod go.sum ./
COPY api/go.mod api/go.sum ./api/
COPY auth/go.mod auth/go.sum ./auth/
COPY cmd/go.mod cmd/go.sum ./cmd/
COPY common/go.mod common/go.sum ./common/
COPY gateway/go.mod gateway/go.sum ./gateway/
COPY user/go.mod user/go.sum ./user/

# 下载依赖（利用 Docker 缓存层）
RUN go work sync

# 复制所有源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/boot ./cmd/boot

# 运行阶段
FROM alpine:latest

# 安装必要的运行时依赖（如 ca-certificates 用于 HTTPS，curl 用于健康检查）
RUN apk --no-cache add ca-certificates tzdata curl

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/boot /app/boot

# 复制配置文件（保持目录结构）
COPY gateway/etc/ /app/gateway/etc/
COPY auth/etc/ /app/auth/etc/
COPY user/etc/ /app/user/etc/

# 复制 docs 目录（gateway 需要 docs/openapi）
COPY docs/ /app/docs/

# 修改 user.yaml 中的 Kafka brokers 为 Docker 服务名（如果存在 localhost:9092）
RUN sed -i 's/localhost:9092/kafka:9092/g' /app/user/etc/user.yaml || true

# 设置时区
ENV TZ=Europe/Berlin

# 暴露端口
# Gateway: 28256, Auth RPC: 7777, User RPC: 7778
EXPOSE 28256 7777 7778

# 运行应用（使用相对路径，工作目录是 /app）
CMD ["/app/boot", "-gateway", "gateway/etc/gateway-api.yaml", "-auth", "auth/etc/auth.yaml", "-user", "user/etc/user.yaml"]

