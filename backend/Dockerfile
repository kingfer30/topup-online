# 使用多阶段构建

# 阶段1: 构建阶段
FROM golang:1.25-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制 go.mod 和 go.sum 并下载依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制整个项目
COPY . .

# 构建 Go 应用（禁用 CGO 以生成静态二进制文件）
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o guo-depot .

# 阶段2: 运行阶段
FROM alpine:latest

# 安装必要的运行时依赖
RUN apk --no-cache add ca-certificates tzdata

# 设置时区为上海
ENV TZ=Asia/Shanghai

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/guo-depot .

# 复制必要的配置文件（如果存在）
COPY --from=builder /app/db.sql ./db.sql

# 创建数据目录，用于存放 .initialized、.env 等持久化文件
# Docker 部署时将此目录映射到宿主机即可保持初始化状态
RUN mkdir -p /app/data

# 设置数据目录环境变量
ENV DATA_DIR=/app/data

# 暴露端口（默认 8080）
EXPOSE 8080

# 启动应用
CMD ["./guo-depot"]

