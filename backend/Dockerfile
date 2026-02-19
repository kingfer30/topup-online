# 使用多阶段构建

# 阶段1: 构建阶段
FROM golang:1.23-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的构建工具
RUN apk add --no-cache git

# 复制 go.mod 和 go.sum 并下载依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制整个项目
COPY . .

# 构建 Go 应用（禁用 CGO 以生成静态二进制文件）
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o topupOnline .

# 阶段2: 运行阶段
FROM alpine:latest

# 安装必要的运行时依赖
RUN apk --no-cache add ca-certificates tzdata

# 设置时区为上海
ENV TZ=Asia/Shanghai

# 创建非 root 用户
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/topupOnline .

# 复制必要的配置文件（如果存在）
COPY --from=builder /app/db.sql ./db.sql

# 更改所有权
RUN chown -R appuser:appuser /app

# 切换到非 root 用户
USER appuser

# 暴露端口（默认 8080）
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/health || exit 1

# 启动应用
CMD ["./topupOnline"]

