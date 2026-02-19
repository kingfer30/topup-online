# 第一阶段：构建应用
FROM node:20 AS builder

# 设置工作目录
WORKDIR /app

# 接收构建参数
ARG VITE_API_BASE_URL=/api
ENV VITE_API_BASE_URL=$VITE_API_BASE_URL

# 复制 package.json 和 package-lock.json
COPY package*.json ./

# 安装依赖
RUN npm install

# 复制整个项目（包括 .env.production）
COPY . .

# 构建生产版本（会自动加载 .env.production）
RUN npm run build

# 第二阶段：使用 Nginx 托管静态文件
FROM nginx:alpine

# 设置时区
ENV TZ=Asia/Shanghai
RUN apk add --no-cache tzdata

# 从构建阶段复制构建结果到 Nginx 的默认静态文件目录
COPY --from=builder /app/dist /usr/share/nginx/html

# 复制 Nginx 配置文件
COPY nginx.conf /etc/nginx/conf.d/default.conf

# 暴露端口 80
EXPOSE 80

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost/ || exit 1

# 启动 Nginx
CMD ["nginx", "-g", "daemon off;"]
