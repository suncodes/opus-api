---
title: Opus API
emoji: 🚀
colorFrom: blue
colorTo: green
sdk: docker
app_port: 7860
---

# Opus API - 账号管理系统

一个用于 API 消息格式转换的服务，将 Claude API 格式转换为其他格式，支持多账号 Cookie 管理和轮询。

## ✨ 功能特性

- 🔄 Claude API 消息格式转换
- 📊 Token 计数统计
- 🎯 **账号管理系统**
  - 用户登录认证（JWT）
  - 多 Morph 账号 Cookie 管理
  - Cookie 有效性检测（手动/自动）
  - Cookie 轮询策略（轮询/优先级/最少使用）
  - Web 管理界面
- 🛠️ 工具调用处理
- 💾 请求/响应日志记录（调试模式）
- 🔌 支持客户端 Cookie/Authorization 请求头覆盖

## 🚀 快速开始

### 部署到 Hugging Face Spaces

1. **创建 Space**
   - 在 Hugging Face 上创建新 Space
   - 选择 **Docker SDK**
   - Space 名称例如：`your-username/opus-api`

2. **配置环境变量**
   
   在 Space 的 **Settings → Repository secrets** 中添加：
   
   | Secret 名称 | 说明 | 示例值 |
   |-------------|------|--------|
   | `DATABASE_URL` | PostgreSQL 数据库连接 | `postgresql://user:password@host:5432/dbname` |
   | `JWT_SECRET` | JWT 签名密钥 | `your-random-secret-key-here` |
   | `DEFAULT_ADMIN_PASSWORD` | 管理员密码 | `changeme123` |

3. **推送代码**
   ```bash
   git remote add hf https://huggingface.co/spaces/YOUR_USERNAME/YOUR_SPACE_NAME
   git push hf main
   ```

4. **访问管理界面**
   - 访问 `https://YOUR_USERNAME-YOUR_SPACE_NAME.hf.space`
   - 使用默认账号登录：
     - 用户名：`admin`
     - 密码：你设置的 `DEFAULT_ADMIN_PASSWORD`

## 📡 API 端点

### 认证 API

```
POST /api/auth/login       # 用户登录
POST /api/auth/logout      # 用户登出
GET  /api/auth/me          # 获取当前用户信息
```

### Cookie 管理 API（需要认证）

```
GET    /api/cookies                # 获取 Cookie 列表
POST   /api/cookies                # 添加 Cookie
GET    /api/cookies/:id            # 获取单个 Cookie
PUT    /api/cookies/:id            # 更新 Cookie
DELETE /api/cookies/:id            # 删除 Cookie
POST   /api/cookies/:id/validate   # 验证单个 Cookie
POST   /api/cookies/validate/all   # 批量验证所有 Cookie
GET    /api/cookies/stats          # 获取统计信息
```

### 消息转换 API

```
POST /v1/messages     # 消息转换接口（支持客户端 Cookie 覆盖）
GET  /health          # 健康检查接口
```

## 🔧 使用方法

### 1. Web 管理界面

访问 `/dashboard` 或 `/static/dashboard.html`，登录后可以：
- 查看 Cookie 统计信息
- 添加/编辑/删除 Cookie
- 手动或批量验证 Cookie 有效性
- 设置 Cookie 优先级

### 2. 调用消息 API

**基础请求：**
```bash
curl -X POST https://your-space.hf.space/v1/messages \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-opus-4-20250514",
    "max_tokens": 1024,
    "messages": [
      {"role": "user", "content": "Hello!"}
    ]
  }'
```

**使用自定义 Cookie（覆盖轮询）：**
```bash
curl -X POST https://your-space.hf.space/v1/messages \
  -H "Content-Type: application/json" \
  -H "Cookie: _gcl_aw=GCL.17692..." \
  -d '{
    "model": "claude-opus-4-20250514",
    "max_tokens": 1024,
    "messages": [
      {"role": "user", "content": "Hello!"}
    ]
  }'
```

## 🗄️ 数据库结构

### users 表
```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### morph_cookies 表
```sql
CREATE TABLE morph_cookies (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    api_key TEXT NOT NULL,
    session_key TEXT,
    is_valid BOOLEAN DEFAULT true,
    last_validated TIMESTAMP,
    last_used TIMESTAMP,
    priority INTEGER DEFAULT 0,
    usage_count BIGINT DEFAULT 0,
    error_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### user_sessions 表
```sql
CREATE TABLE user_sessions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## 🔄 Cookie 轮询策略

系统支持三种轮询策略，通过环境变量 `ROTATION_STRATEGY` 配置：

| 策略 | 说明 |
|------|------|
| `priority` | 按优先级（数字越大越优先，默认） |
| `round_robin` | 轮询（按顺序循环使用） |
| `least_used` | 使用次数最少的优先 |

## 🔐 环境变量配置

| 变量名 | 说明 | 默认值 | 必需 |
|--------|------|--------|------|
| `DATABASE_URL` | PostgreSQL 连接字符串 | - | ✅ |
| `JWT_SECRET` | JWT 签名密钥 | - | ✅ |
| `DEFAULT_ADMIN_USERNAME` | 默认管理员用户名 | `admin` | ❌ |
| `DEFAULT_ADMIN_PASSWORD` | 默认管理员密码 | `changeme123` | ❌ |
| `COOKIE_MAX_ERROR_COUNT` | Cookie 失败阈值 | `3` | ❌ |
| `ROTATION_STRATEGY` | 轮询策略 | `priority` | ❌ |
| `DEBUG_MODE` | 调试模式 | `false` | ❌ |

## 🛠️ 本地开发

### 前置要求
- Go 1.21+
- PostgreSQL 14+
- Python 3.9+（用于 Hugging Face 部署）

### 安装依赖
```bash
go mod download
```

### 配置环境变量
```bash
cp .env.example .env
# 编辑 .env 文件，设置数据库连接等
```

### 运行服务
```bash
# 使用 Go 直接运行
go run cmd/server/main.go

# 或使用 Docker
docker build -t opus-api .
docker run -p 7860:7860 \
  -e DATABASE_URL="postgresql://..." \
  -e JWT_SECRET="your-secret" \
  opus-api
```

## 📁 项目结构

```
opus-api/
├── cmd/server/              # 主程序入口
│   └── main.go
├── internal/
│   ├── converter/           # 格式转换逻辑
│   ├── handler/             # HTTP 处理器
│   │   ├── messages.go      # 消息处理
│   │   ├── health.go        # 健康检查
│   │   ├── auth.go          # 认证处理
│   │   └── cookies.go       # Cookie 管理
│   ├── middleware/          # 中间件
│   │   └── auth.go          # JWT 认证
│   ├── model/               # 数据模型
│   │   ├── db.go            # 数据库连接
│   │   ├── user.go          # 用户模型
│   │   └── cookie.go        # Cookie 模型
│   ├── service/             # 业务逻辑
│   │   ├── auth_service.go  # 认证服务
│   │   ├── cookie_service.go# Cookie 服务
│   │   ├── validator.go     # Cookie 验证
│   │   └── rotator.go       # Cookie 轮询
│   ├── logger/              # 日志管理
│   ├── parser/              # 消息解析
│   ├── stream/              # 流式处理
│   ├── tokenizer/           # Token 计数
│   ├── types/               # 类型定义
│   └── converter/           # 格式转换
├── web/static/              # 前端静态文件
│   ├── index.html           # 登录页
│   ├── dashboard.html       # 管理面板
│   ├── styles.css           # 样式
│   └── app.js               # 前端逻辑
├── migrations/              # 数据库迁移（可选）
├── .env.example             # 环境变量示例
├── app.py                   # Python 启动脚本
├── Dockerfile               # Docker 构建文件
└── go.mod                   # Go 依赖管理
```

## 📝 技术栈

- **后端**: Go 1.21, Gin
- **数据库**: PostgreSQL, GORM
- **认证**: JWT
- **前端**: HTML, CSS, Vanilla JavaScript
- **容器化**: Docker
- **部署平台**: Hugging Face Spaces

## 🔐 安全建议

1. **首次登录后立即修改默认密码**
2. **生产环境使用强 JWT_SECRET**
3. **定期更新 Cookie**（Morph Cookie 可能会过期）
4. **使用 HTTPS**（Hugging Face Spaces 自动提供）

## 📄 许可证

本项目遵循开源协议。

## 🤝 贡献

欢迎提交 Issue 和 Pull Request。

## 📞 联系方式

如有问题或建议，欢迎提交 Issue。