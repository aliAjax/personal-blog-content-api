# Go Blog API

从 0 实现的 Go 个人博客后台 API，数据存储使用 MySQL。项目遵循 Go 企业级工程目录规范，将文章、分类、标签、评论、用户等业务模块拆分到 `internal` 下，每个模块均包含 `handler`、`service`、`repository`、`model`。

## 功能

- 用户注册、登录，JWT 认证，角色区分 `admin` / `user`
- 博主登录后管理文章、分类和标签
- 文章支持 `draft` 草稿和 `published` 发布状态
- 发布后前台按分类、标签浏览文章，并支持关键词搜索
- 访客查看已发布文章并提交评论
- 博主审核评论（`pending` / `approved` / `rejected`）或删除评论
- 启动时自动等待 MySQL 就绪、执行 `migrations/*.sql`，并创建演示数据

## 目录结构

```text
.
├── api/                 # OpenAPI 描述
├── cmd/server/          # 程序入口
├── configs/             # 配置样例
├── internal/
│   ├── article/         # 文章模块
│   ├── category/        # 分类模块
│   ├── tag/             # 标签模块
│   ├── comment/         # 评论模块
│   ├── user/            # 用户模块
│   ├── router/          # 路由装配
│   └── seed/            # 演示数据
├── migrations/          # SQL 迁移
├── pkg/                 # 可复用基础包
├── Dockerfile
└── docker-compose.yml
```

## 运行

```bash
docker compose up -d --build
```

服务监听容器 `8080` 端口，映射宿主机 `18095`：

```bash
curl http://127.0.0.1:18095/healthz
```

演示博主账号由环境变量控制，默认：

```text
用户名：admin
密码：Admin123!
```

如需自定义，复制 `.env.example` 为 `.env` 后修改。

## 主要接口

公开接口：

```text
POST   /api/v1/auth/register
POST   /api/v1/auth/login
GET    /api/v1/categories
GET    /api/v1/tags
GET    /api/v1/articles
GET    /api/v1/articles/{id}
GET    /api/v1/articles/{id}/comments
POST   /api/v1/articles/{id}/comments
```

博主接口（需要 `Authorization: Bearer <token>`）：

```text
GET    /api/v1/auth/me
GET    /api/v1/admin/articles
GET    /api/v1/admin/articles/{id}
POST   /api/v1/admin/articles
PUT    /api/v1/admin/articles/{id}
PATCH  /api/v1/admin/articles/{id}/status
DELETE /api/v1/admin/articles/{id}

GET    /api/v1/admin/categories
POST   /api/v1/admin/categories
PUT    /api/v1/admin/categories/{id}
DELETE /api/v1/admin/categories/{id}

GET    /api/v1/admin/tags
POST   /api/v1/admin/tags
PUT    /api/v1/admin/tags/{id}
DELETE /api/v1/admin/tags/{id}

GET    /api/v1/admin/comments
PATCH  /api/v1/admin/comments/{id}
DELETE /api/v1/admin/comments/{id}
```

## 清理

```bash
docker compose down -v
```
