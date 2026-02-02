---
title: Opus API
emoji: 🚀
colorFrom: blue
colorTo: green
sdk: docker
app_port: 7860
---

# Opus API

一个用于 API 消息格式转换的服务，将 Claude API 格式转换为其他格式。

## 功能特性

- ✨ Claude API 消息格式转换
- 🔄 流式响应支持
- 🛠️ 工具调用处理
- 📊 Token 计数
- 💾 请求/响应日志记录（调试模式）

## API 端点

### 1. 消息转换接口
```
POST /v1/messages
```

将 Claude API 格式的消息转换为目标格式。

**请求示例:**
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

### 2. 健康检查接口
```
GET /health
```

检查服务运行状态。

**响应示例:**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-01T00:00:00Z"
}
```

## 技术栈

- **后端**: Go 1.21
- **Web 框架**: Gin
- **Token 计数**: tiktoken-go
- **容器化**: Docker
- **部署平台**: Hugging Face Spaces

## 环境变量

- `PORT`: 服务端口（默认: 7860）
- `DEBUG_MODE`: 调试模式开关
- `LOG_DIR`: 日志目录路径

## 本地开发

### 使用 Go 直接运行
```bash
go run cmd/server/main.go
```

### 使用 Docker
```bash
docker build -t opus-api .
docker run -p 7860:7860 opus-api
```

## 项目结构

```
opus-api/
├── cmd/server/          # 主程序入口
├── internal/
│   ├── converter/       # 格式转换逻辑
│   ├── handler/         # HTTP 处理器
│   ├── logger/          # 日志管理
│   ├── parser/          # 消息解析
│   ├── stream/          # 流式处理
│   ├── tokenizer/       # Token 计数
│   └── types/           # 类型定义
├── app.py              # Python 启动脚本
├── Dockerfile          # Docker 构建文件
└── go.mod              # Go 依赖管理
```

## 许可证

本项目遵循开源协议。

## 联系方式

如有问题或建议，欢迎提交 Issue。