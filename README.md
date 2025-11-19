# 智汇圈微信小程序后端项目

基于 Go + Gin + MySQL 的微信小程序后端服务

## 项目简介

这是一个为微信小程序提供后端服务的项目，使用 Go 语言的 Gin 框架开发，提供用户认证、数据管理等核心功能。

## 技术栈

- **后端框架**: Gin (Go)
- **数据库**: MySQL 8.0
- **ORM**: GORM
- **配置管理**: Viper
- **认证**: JWT
- **前端**: UniApp (微信小程序)

## 项目结构

```
ZHQ_backend/
├── main.go              # 程序入口
├── config/              # 配置管理
├── controllers/         # 控制器层
├── models/             # 数据模型
├── routes/             # 路由配置
├── middleware/         # 中间件
├── utils/              # 工具函数
├── database/           # 数据库连接
├── docs/               # 项目文档
├── .env.example        # 环境变量模板
└── .gitignore         # Git忽略文件
```

## 快速开始

### 环境要求

- Go 1.19+
- MySQL 8.0+
- Git

### 本地开发设置

1. **克隆项目**
   ```bash
   git clone https://github.com/James-FS/ZHQ_backend.git
   cd ZHQ_backend
   ```

2. **安装依赖**
   ```bash
   go mod tidy
   ```

3. **配置环境变量**
   ```bash
   cp .env.example .env
   # 编辑 .env 文件，填入实际的数据库连接信息
   ```

4. **创建数据库**
   ```sql-norun
   CREATE DATABASE zhq CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
   ```

5. **运行项目**
   ```bash
   go run main.go
   ```

6. **测试API**
   ```bash
   curl http://localhost:8080/health
   ```

## API 文档

### 基础接口

- `GET /health` - 健康检查

### 认证接口

- `POST /api/v1/auth/login` - 微信登录

### 用户接口

- `GET /api/v1/user/profile` - 获取用户信息
- `PUT /api/v1/user/profile` - 更新用户信息

## 开发规范

### 分支管理

- `main` - 主分支，部署到生产环境
- `develop` - 开发分支，集成测试
- `feature/*` - 功能分支，新功能开发
- `hotfix/*` - 修复分支，紧急修复

### 提交规范
- `feat:` 新功能
- `fix:` 修复bug
- `docs:` 文档更新
- `test:` 测试相关

示例：
```
feat: 添加用户注册功能

- 实现微信小程序用户注册
- 添加手机号验证
- 更新用户数据模型
```

### 代码规范

- 遵循 Go 官方代码规范
- 使用 `gofmt` 格式化代码
- 添加必要的注释和文档

## 部署说明

### 环境变量

生产环境需要配置以下环境变量：

- `DB_HOST` - 数据库主机
- `DB_USER` - 数据库用户名
- `DB_PASSWORD` - 数据库密码
- `JWT_SECRET` - JWT密钥
- `WECHAT_APPID` - 微信小程序AppID
- `WECHAT_SECRET` - 微信小程序Secret

## 团队成员

- [@James-FS](https://github.com/James-FS) - 项目负责人
- 待添加其他团队成员...


## 更新日志

### v1.0.0 (2025-11-17)
- 初始化项目