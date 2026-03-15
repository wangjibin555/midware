# Midware - 基础中间件库

> 一套完整的 Go 语言基础中间件库，提供认证、日志、缓存等常用功能

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

## 📦 模块列表

| 模块 | 状态 | 说明 |
|------|------|------|
| [**Auth**](./Auth) | ✅ 完成 | 认证授权（JWT、权限、Session） |
| [**Logger**](./Logger) | ✅ 完成 | 结构化日志系统 |
| **Cache** | 🚧 开发中 | 缓存抽象层 |
| **RateLimit** | 🚧 开发中 | 限流中间件 |
| **ErrorHandler** | 🚧 开发中 | 统一错误处理 |
| **Validator** | 🚧 开发中 | 数据验证 |
| **Compression** | 🚧 开发中 | 数据压缩 |

## 🚀 快速开始

### 安装

```bash
# 安装 Auth 模块
go get github.com/wangjibin555/midware/Auth

# 安装 Logger 模块
go get github.com/wangjibin555/midware/Logger
```

### 使用示例

```go
package main

import (
    "github.com/wangjibin555/midware/Auth"
    "github.com/wangjibin555/midware/Logger"
    "time"
)

func main() {
    // 1. 初始化 Logger
    logger := Logger.New(Logger.DefaultConfig())
    logger.Info("Application started")

    // 2. 初始化 Auth
    auth, _ := Auth.New(
        Auth.DefaultConfig(),
        Auth.WithJWTSecret("your-32-bytes-secret-key-here!"),
    )

    // 3. 创建用户并生成 Token
    user := &Auth.User{
        ID:       "user-001",
        Username: "admin",
        Roles:    []string{"admin"},
        Status:   Auth.UserStatusNormal,
    }

    tokens, _ := auth.GenerateTokenPair(user)
    logger.Info("Token generated", "user_id", user.ID)

    // 4. 验证 Token
    claims, _ := auth.Verify(tokens.AccessToken)
    logger.Info("Token verified", "username", claims.Username)
}
```

## 💡 特性

### Auth 模块
- ✅ JWT Token 生成和验证（零第三方依赖）
- ✅ 角色权限控制（RBAC）
- ✅ Token 黑名单（主动登出）
- ✅ Argon2id 密码加密
- ✅ Session 管理
- ✅ HTTP 中间件
- ✅ API Key、验证码等工具

### Logger 模块
- ✅ 结构化日志
- ✅ 多级别输出
- ✅ JSON/Text 格式
- ✅ 文件轮转
- ✅ 调用者信息

## 🛠️ 开发指南

### 克隆项目

```bash
git clone git@github.com:wangjibin555/midware.git
cd midware
```

### 运行测试

```bash
# 测试 Auth 模块
cd Auth
go test -v

# 测试 Logger 模块
cd Logger
go test -v
```

### 构建

```bash
# 构建所有模块
cd Auth && go build
cd ../Logger && go build
```

## 📝 项目结构

```
midware/
├── Auth/              # 认证授权模块
│   ├── README.md      # 模块文档
│   ├── auth.go        # 核心实现
│   ├── jwt.go         # JWT 实现
│   ├── middleware.go  # HTTP 中间件
│   └── crypto/        # 加密工具
├── Logger/            # 日志模块
│   ├── README.md
│   └── logger.go
├── Cache/             # 缓存模块（待开发）
├── RateLimit/         # 限流模块（待开发）
├── USAGE.md           # 使用指南
└── README.md          # 本文件
```

## 🤝 贡献

欢迎贡献代码、报告 Bug、提出建议！

1. Fork 本仓库
2. 创建新分支 (`git checkout -b feature/amazing`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing`)
5. 创建 Pull Request

## 📄 许可证

MIT License - 详见 [LICENSE](LICENSE) 文件

## 🙏 致谢

感谢所有贡献者！

---

**如有问题，欢迎提 Issue！**
