# Topup Online - ChatGPT充值平台

这是一个完整的ChatGPT充值平台项目，包含前台展示页面和后台管理系统。

## 项目结构

```
topup-online/
├── web/           # 前台充值页面（Vue 3 + Naive UI + Tailwind CSS）
├── admin/         # 后台管理系统（Vue 3 + Naive UI + Tailwind CSS）
├── backend/       # 后端服务（Go）
└── webbak/        # 备份文件
```

## 技术栈

### 前端
- **Vue 3** - 渐进式JavaScript框架
- **Vite** - 下一代前端构建工具
- **Naive UI** - Vue 3 组件库
- **Tailwind CSS** - 实用优先的CSS框架
- **TypeScript** - JavaScript的超集
- **Pinia** - 状态管理
- **Vue Router** - 路由管理

### 后端
- **Go** - 后端服务语言
- **Gin** - Web框架

## 项目说明

### 1. Web - 前台充值页面

**目录：** `web/`  
**端口：** 3000  
**访问：** http://localhost:3000

#### 功能特性
- ✅ ChatGPT充值展示页面
- ✅ 多语言支持（中文、英文、俄语）
- ✅ 响应式设计
- ✅ 粒子背景效果
- ✅ 功能特性展示
- ✅ 操作流程说明
- ✅ FAQ常见问题

#### 快速启动
```bash
cd web
npm install
npm run dev
```

#### 详细文档
请查看 [web/README.md](web/README.md)

---

### 2. Admin - 后台管理系统

**目录：** `admin/`  
**端口：** 3001  
**访问：** http://localhost:3001

#### 登录信息
- **首次使用：** 需要通过初始化页面创建管理员账号
- **已初始化：** 使用初始化时创建的管理员账号登录

#### 功能模块
- ✅ **系统初始化** - 首次使用引导，自动创建数据表结构
- ✅ **登录系统** - 用户认证和权限管理
- ✅ **控制台** - 数据统计和概览
- ✅ **用户管理** - 用户列表、添加、编辑、删除
- ✅ **订单管理** - 订单查询、详情、状态管理
- ✅ **系统设置** - 基础设置、支付设置、通知设置
- 🚧 **卡密管理** - 开发中
- 🚧 **角色管理** - 开发中
- 🚧 **操作日志** - 开发中

#### 快速启动
```bash
cd admin
npm install
npm run dev
```

#### 详细文档
请查看 [admin/README.md](admin/README.md)

---

### 3. Backend - 后端服务

**目录：** `backend/`  
**端口：** 8080  
**API：** http://localhost:8080/api

#### 快速启动
```bash
cd backend
go run main.go
```

---

## 开发流程

### 1. 准备数据库

首次使用需要准备MySQL数据库：

```sql
-- 创建数据库
CREATE DATABASE topup_online CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 创建数据库用户（可选，推荐）
CREATE USER 'topup_user'@'localhost' IDENTIFIED BY 'your_password';
GRANT ALL PRIVILEGES ON topup_online.* TO 'topup_user'@'localhost';
FLUSH PRIVILEGES;
```

### 2. 安装所有依赖

```bash
# 前台项目
cd web
npm install

# 后台项目
cd ../admin
npm install

# 后端项目
cd ../backend
go mod download
```

### 3. 启动所有服务

**推荐使用3个终端窗口：**

```bash
# 终端 1 - 前台项目（端口 3000）
cd web
npm run dev

# 终端 2 - 后台项目（端口 3001）
cd admin
npm run dev

# 终端 3 - 后端服务（端口 8080）
cd backend
go run main.go
```

### 4. 系统初始化

**首次使用必须完成系统初始化：**

1. 访问后台管理系统：http://localhost:3001
2. 系统会自动检测未初始化状态，跳转到初始化页面
3. 按照引导完成三个步骤：
   - **步骤1：** 配置数据库连接信息并测试连接
   - **步骤2：** 设置管理员账号和密码
   - **步骤3：** 系统自动创建表结构并完成初始化
4. 初始化完成后，使用创建的管理员账号登录

**详细初始化指南：** 请查看 [INITIALIZATION.md](INITIALIZATION.md)

### 5. 访问地址

- **前台页面：** http://localhost:3000
- **后台管理：** http://localhost:3001
- **API接口：** http://localhost:8080/api

---

## 项目特点

### 系统初始化向导
- ✅ 首次使用自动引导初始化
- ✅ 可视化配置数据库连接
- ✅ 自动创建表结构和管理员账号
- ✅ 防止重复初始化

### 完全独立的前后台
- ✅ 前台和后台是两个完全独立的Vue项目
- ✅ 可以单独开发、部署和维护
- ✅ 不同的端口，互不干扰
- ✅ 各自拥有独立的路由、状态管理

### 统一的技术栈
- ✅ 都使用 Vue 3 + Naive UI + Tailwind CSS
- ✅ 统一的开发体验和代码风格
- ✅ 组件和工具可以复用

### 现代化的开发方式
- ✅ TypeScript 类型支持
- ✅ Vite 极速热更新
- ✅ 组件化开发
- ✅ 响应式设计

---

## 生产部署

### 构建项目

```bash
# 构建前台
cd web
npm run build
# 输出目录：web/dist

# 构建后台
cd ../admin
npm run build
# 输出目录：admin/dist

# 构建后端
cd ../backend
go build -o topup-online
```

### Nginx 配置示例

```nginx
# 前台页面
server {
    listen 80;
    server_name your-domain.com;
    
    root /path/to/web/dist;
    index index.html;
    
    location / {
        try_files $uri $uri/ /index.html;
    }
    
    location /api {
        proxy_pass http://localhost:8080;
    }
}

# 后台管理
server {
    listen 80;
    server_name admin.your-domain.com;
    
    root /path/to/admin/dist;
    index index.html;
    
    location / {
        try_files $uri $uri/ /index.html;
    }
    
    location /api {
        proxy_pass http://localhost:8080;
    }
}
```

---

## 浏览器支持

- Chrome >= 87
- Firefox >= 78
- Safari >= 14
- Edge >= 88

---

## 开发规范

1. 使用 Vue 3 Composition API
2. 遵循 TypeScript 类型检查
3. 使用 Tailwind 实用类进行样式设计
4. 使用 Naive UI 组件构建界面
5. 保持代码风格一致

---

## License

MIT License

---

## 联系方式

如有问题，请联系：support@example.com

