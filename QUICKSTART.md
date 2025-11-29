# 快速开始指南

## 5分钟快速启动

### 前置要求

- Node.js >= 16
- Go >= 1.16
- MySQL >= 5.7

### 1. 克隆项目

```bash
git clone <your-repo-url>
cd topup-online
```

### 2. 准备数据库

```sql
CREATE DATABASE topup_online CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

### 3. 安装依赖

```bash
# 前台
cd web && npm install && cd ..

# 后台
cd admin && npm install && cd ..

# 后端
cd backend && go mod download && cd ..
```

### 4. 启动服务（三个终端）

**终端1 - 后端：**
```bash
cd backend
go run main.go
# 运行在 http://localhost:8080
```

**终端2 - 前台：**
```bash
cd web
npm run dev
# 运行在 http://localhost:3000
```

**终端3 - 后台：**
```bash
cd admin
npm run dev
# 运行在 http://localhost:3001
```

### 5. 初始化系统

1. 打开浏览器访问：http://localhost:3001
2. 自动跳转到初始化页面
3. 填写数据库信息：
   - 数据库地址：`localhost`
   - 数据库端口：`3306`
   - 数据库名称：`topup_online`
   - 数据库用户：`root`
   - 数据库密码：`你的密码`
4. 点击"测试连接"确保连接成功
5. 点击"下一步"
6. 设置管理员信息：
   - 用户名：`admin`
   - 密码：`设置一个强密码`
   - 邮箱：`admin@example.com`
7. 点击"开始初始化"
8. 等待完成后点击"前往登录"

### 6. 开始使用

**前台页面（用户端）：**
- 访问：http://localhost:3000
- 功能：查看充值信息、多语言支持

**后台管理（管理端）：**
- 访问：http://localhost:3001
- 登录：使用初始化时创建的管理员账号
- 功能：用户管理、订单管理、系统设置等

## 目录结构

```
topup-online/
├── web/          # 前台（端口3000）
├── admin/        # 后台（端口3001）
├── backend/      # 后端（端口8080）
├── README.md     # 项目说明
├── INITIALIZATION.md  # 初始化详细指南
└── QUICKSTART.md # 本文件
```

## 常用命令

### 开发模式

```bash
# 前台
cd web && npm run dev

# 后台
cd admin && npm run dev

# 后端
cd backend && go run main.go
```

### 生产构建

```bash
# 前台
cd web && npm run build

# 后台
cd admin && npm run build

# 后端
cd backend && go build
```

## 默认端口

| 服务 | 端口 | 访问地址 |
|------|------|----------|
| 前台 | 3000 | http://localhost:3000 |
| 后台 | 3001 | http://localhost:3001 |
| API  | 8080 | http://localhost:8080/api |

## 重要API

| API | 方法 | 说明 |
|-----|------|------|
| /api/system/init/status | GET | 检查初始化状态 |
| /api/system/init/test-db | POST | 测试数据库连接 |
| /api/system/init | POST | 执行系统初始化 |

## 常见问题

### Q: 数据库连接失败？
A: 检查MySQL是否运行，数据库信息是否正确

### Q: 端口被占用？
A: 修改vite.config.ts中的port配置

### Q: 如何重新初始化？
A: 删除backend/.initialized文件，清空数据库

## 下一步

- 查看 [README.md](README.md) 了解完整功能
- 查看 [INITIALIZATION.md](INITIALIZATION.md) 了解初始化详情
- 查看各项目的README了解具体模块

## 技术支持

- GitHub Issues
- Email: support@example.com

