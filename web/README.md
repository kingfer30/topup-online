# Topup Online Web

这是一个基于 Vue 3 + Naive UI + Tailwind CSS 的前端项目框架。

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
│   ├── components/      # 通用组件
│   │   └── AppContent.vue
│   ├── views/           # 页面组件
│   │   ├── Home.vue
│   │   └── About.vue
│   ├── router/          # 路由配置
│   │   └── index.ts
│   ├── stores/          # 状态管理
│   │   └── user.ts
│   ├── utils/           # 工具函数
│   │   └── http.ts
│   ├── styles/          # 全局样式
│   │   └── index.css
│   ├── App.vue          # 根组件
│   └── main.ts          # 入口文件
├── public/              # 公共资源
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

## 功能特性

- ✅ Vue 3 Composition API
- ✅ TypeScript 支持
- ✅ Vite 极速热更新
- ✅ Naive UI 组件库集成
- ✅ Tailwind CSS 实用样式
- ✅ Vue Router 路由管理
- ✅ Pinia 状态管理
- ✅ Axios HTTP 请求封装
- ✅ 响应式布局设计

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

## 浏览器支持

- Chrome >= 87
- Firefox >= 78
- Safari >= 14
- Edge >= 88

