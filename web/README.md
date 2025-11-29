# Topup Online Web

这是一个基于 Vue 3 + Naive UI + Tailwind CSS 的前端项目框架，包含前台页面和后台管理系统。

## 技术栈

- **Vue 3** - 渐进式JavaScript框架
- **Vite** - 下一代前端构建工具
- **Naive UI** - Vue 3 组件库
- **Tailwind CSS** - 实用优先的CSS框架
- **TypeScript** - JavaScript的超集
- **Pinia** - Vue状态管理库
- **Vue Router** - Vue官方路由
- **Axios** - HTTP客户端

## 项目结构

```
web/
├── src/
│   ├── assets/          # 静态资源
│   │   └── images/
│   ├── components/      # 通用组件
│   │   ├── AppContent.vue
│   │   └── Layout/
│   ├── layouts/         # 布局组件
│   │   └── AdminLayout.vue  # 后台管理布局
│   ├── views/           # 页面组件
│   │   ├── Home.vue         # 前台首页
│   │   ├── About.vue        # 关于页面
│   │   └── admin/           # 后台页面
│   │       ├── Login.vue
│   │       ├── Dashboard.vue
│   │       ├── Users.vue
│   │       ├── Orders.vue
│   │       └── Settings.vue
│   ├── router/          # 路由配置
│   │   └── index.ts
│   ├── stores/          # 状态管理
│   │   └── user.ts
│   ├── utils/           # 工具函数
│   │   └── http.ts
│   ├── types/           # TypeScript类型
│   │   └── index.ts
│   ├── api/             # API接口
│   │   └── user.ts
│   ├── lang/            # 多语言
│   │   ├── zh.ts
│   │   ├── en.ts
│   │   └── ru.ts
│   ├── styles/          # 全局样式
│   │   ├── index.css
│   │   └── custom.css
│   ├── App.vue          # 根组件
│   └── main.ts          # 入口文件
├── public/              # 公共资源
│   ├── flags/           # 国旗图标
│   └── favicon.ico
├── index.html           # HTML模板
├── vite.config.ts       # Vite配置
├── tailwind.config.js   # Tailwind配置
├── postcss.config.js    # PostCSS配置
├── tsconfig.json        # TypeScript配置
└── package.json         # 项目配置
```

## 快速开始

### 安装依赖

```bash
npm install
```

### 启动开发服务器

```bash
npm run dev
```

项目将在 http://localhost:3000 运行

### 构建生产版本

```bash
npm run build
```

### 预览生产版本

```bash
npm run preview
```

## 项目功能

### 前台页面

#### 首页 (`/`)
- ✅ 多语言支持（中文、英文、俄语）
- ✅ 响应式设计
- ✅ 粒子背景效果
- ✅ 功能特性展示
- ✅ 操作流程说明
- ✅ FAQ常见问题
- ✅ 自动语言检测

#### 关于页面 (`/about`)
- ✅ 技术栈介绍
- ✅ 项目结构说明
- ✅ 快速开始指南

### 后台管理系统

#### 访问地址
- 登录页面：`/admin/login`
- 管理后台：`/admin`（自动重定向到 `/admin/dashboard`）

#### 默认登录信息（演示）
- 用户名：`admin`
- 密码：`admin`

#### 后台功能模块

##### 1. 控制台 (`/admin/dashboard`)
- ✅ 数据统计卡片（订单、收入、用户、卡密）
- ✅ 图表展示区域
- ✅ 最近订单列表
- ✅ 实时数据更新

##### 2. 用户管理
- **用户列表** (`/admin/users`)
  - ✅ 用户搜索和筛选
  - ✅ 用户状态管理
  - ✅ 角色分配
  - ✅ 添加/编辑/删除用户
- **角色管理** (`/admin/roles`)
  - 开发中...

##### 3. 订单管理
- **订单列表** (`/admin/orders`)
  - ✅ 订单搜索和筛选
  - ✅ 日期范围筛选
  - ✅ 订单状态管理
  - ✅ 订单详情查看
  - ✅ 订单取消功能
  - ✅ 数据导出
- **退款管理** (`/admin/refunds`)
  - 开发中...

##### 4. 卡密管理
- **卡密列表** (`/admin/cards`)
  - 开发中...
- **生成卡密** (`/admin/card-generate`)
  - 开发中...

##### 5. 系统设置 (`/admin/settings`)
- ✅ 基础设置（网站名称、标题、描述等）
- ✅ 支付设置（支付方式、金额限制、手续费等）
- ✅ 通知设置（邮件、短信、SMTP配置等）
- ✅ 多标签页设计
- ✅ 表单验证

##### 6. 操作日志 (`/admin/logs`)
- 开发中...

#### 后台布局特性
- ✅ 左侧菜单栏（可折叠）
- ✅ 顶部导航栏
- ✅ 面包屑导航
- ✅ 主题切换（亮色/暗色）
- ✅ 用户信息下拉菜单
- ✅ 返回首页按钮
- ✅ 响应式设计

## 配置说明

### API 代理

在 `vite.config.ts` 中配置了 API 代理，将 `/api` 请求代理到后端服务器：

```typescript
server: {
  proxy: {
    '/api': {
      target: 'http://localhost:8080',
      changeOrigin: true,
    },
  },
}
```

### Tailwind CSS

在 `tailwind.config.js` 中禁用了 `preflight`，以避免与 Naive UI 的样式冲突：

```javascript
corePlugins: {
  preflight: false,
}
```

## 开发建议

1. 使用 Composition API 编写组件
2. 遵循 TypeScript 类型检查
3. 使用 Tailwind 实用类进行样式设计
4. 使用 Naive UI 组件构建界面
5. 使用 Pinia 进行状态管理
6. 使用封装的 http 工具进行 API 请求

## 路由说明

### 前台路由
- `/` - 首页
- `/about` - 关于页面

### 后台路由
- `/admin/login` - 后台登录
- `/admin/dashboard` - 控制台
- `/admin/users` - 用户列表
- `/admin/roles` - 角色管理
- `/admin/orders` - 订单列表
- `/admin/refunds` - 退款管理
- `/admin/cards` - 卡密列表
- `/admin/card-generate` - 生成卡密
- `/admin/settings` - 系统设置
- `/admin/logs` - 操作日志

## 多语言支持

项目支持三种语言：
- 中文 (zh)
- 英文 (en)
- 俄语 (ru)

语言文件位于 `src/lang/` 目录。

## 浏览器支持

- Chrome >= 87
- Firefox >= 78
- Safari >= 14
- Edge >= 88

## 待开发功能

- [ ] 角色管理页面
- [ ] 退款管理页面
- [ ] 卡密列表页面
- [ ] 生成卡密页面
- [ ] 操作日志页面
- [ ] 个人设置页面
- [ ] 修改密码页面
- [ ] 图表集成（建议使用 ECharts）
- [ ] 权限控制系统
- [ ] 数据导出功能
- [ ] 批量操作功能
