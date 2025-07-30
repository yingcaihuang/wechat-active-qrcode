# 微信二维码活码管理系统

一个基于Golang开发的微信二维码活码管理系统，支持动态二维码生成、扫描统计、后台管理等功能。

## 功能特性

- 🔐 **用户认证**: JWT认证，支持用户注册、登录
- 📱 **二维码管理**: 创建、编辑、删除二维码
- 📊 **统计分析**: 扫描次数统计、趋势分析、热门二维码
- 🖼️ **图片生成**: 自动生成二维码图片
- 📈 **数据可视化**: 提供详细的统计数据和图表
- 🔒 **权限控制**: 基于角色的权限管理
- 🚀 **高性能**: 基于Gin框架，SQLite数据库

## 技术栈

- **后端框架**: Gin
- **数据库**: SQLite
- **认证**: JWT
- **二维码生成**: go-qrcode
- **配置管理**: Viper
- **容器化**: Docker

## 快速开始

### 环境要求

- Go 1.21+
- SQLite 3.x

### 本地运行

1. 克隆项目
```bash
git clone https://github.com/your-username/wechat-active-qrcode.git
cd wechat-active-qrcode
```

2. 安装依赖
```bash
go mod download
```

3. 运行应用
```bash
go run cmd/server/main.go
```

4. 访问应用
```
http://localhost:8080
```

### Docker运行

1. 构建镜像
```bash
docker build -t wechat-qrcode .
```

2. 运行容器
```bash
docker run -p 8080:8080 -v $(pwd)/data:/root/data wechat-qrcode
```

### Docker Compose

```bash
docker-compose up -d
```

## API文档

### 认证相关

#### 用户登录
```http
POST /api/auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "password"
}
```

#### 用户注册
```http
POST /api/auth/register
Content-Type: application/json

{
  "username": "newuser",
  "password": "password123"
}
```

### 二维码管理

#### 创建二维码
```http
POST /api/qrcodes
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "我的二维码",
  "original_url": "https://example.com"
}
```

#### 获取二维码列表
```http
GET /api/qrcodes?page=1&page_size=10
Authorization: Bearer <token>
```

#### 获取二维码详情
```http
GET /api/qrcodes/{id}
Authorization: Bearer <token>
```

#### 更新二维码
```http
PUT /api/qrcodes/{id}
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "更新的名称",
  "original_url": "https://new-example.com",
  "status": 1
}
```

#### 删除二维码
```http
DELETE /api/qrcodes/{id}
Authorization: Bearer <token>
```

### 统计相关

#### 获取总览统计
```http
GET /api/statistics/overview
Authorization: Bearer <token>
```

#### 获取趋势数据
```http
GET /api/statistics/trends?days=7
Authorization: Bearer <token>
```

#### 获取二维码统计
```http
GET /api/statistics/qrcodes/{id}/stats
Authorization: Bearer <token>
```

### 公开接口

#### 记录扫描
```http
POST /api/public/scan/{id}
```

## 默认账户

系统启动时会自动创建默认管理员账户：

- **用户名**: admin
- **密码**: password

**注意**: 请在生产环境中修改默认密码！

## 配置说明

配置文件位于 `configs/config.yaml`：

```yaml
server:
  port: ":8080"
  mode: "debug"

database:
  sqlite_path: "./data/qrcode.db"

jwt:
  secret: "your-secret-key-change-in-production"
  expire: 24

cors:
  allowed_origins:
    - "*"
  allowed_methods:
    - "GET"
    - "POST"
    - "PUT"
    - "DELETE"
    - "OPTIONS"
  allowed_headers:
    - "Origin"
    - "Content-Type"
    - "Accept"
    - "Authorization"
```

## 项目结构

```
wechat-active-qrcode/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── api/
│   │   ├── handlers/
│   │   ├── middleware/
│   │   └── routes.go
│   ├── auth/
│   │   └── jwt.go
│   ├── config/
│   │   └── config.go
│   ├── database/
│   │   └── sqlite.go
│   ├── models/
│   │   └── models.go
│   └── services/
│       ├── auth.go
│       ├── qrcode.go
│       └── statistics.go
├── pkg/
│   ├── qrcode/
│   │   └── generator.go
│   └── utils/
│       └── helpers.go
├── configs/
│   └── config.yaml
├── data/
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── go.sum
└── README.md
```

## 开发指南

### 添加新功能

1. 在 `internal/models/` 中定义数据模型
2. 在 `internal/services/` 中实现业务逻辑
3. 在 `internal/api/handlers/` 中创建API处理器
4. 在 `internal/api/routes.go` 中添加路由

### 数据库迁移

系统使用GORM自动迁移，启动时会自动创建表结构。

### 测试

```bash
go test ./...
```

## 部署

### 生产环境部署

1. 修改配置文件中的敏感信息
2. 使用HTTPS
3. 配置反向代理（如Nginx）
4. 设置防火墙规则
5. 配置日志收集

### 性能优化

- 启用Gin的Release模式
- 配置数据库连接池
- 使用Redis缓存热点数据
- 配置CDN加速静态资源

## 贡献指南

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 更新日志

### v1.0.0 (2024-01-01)

- 初始版本发布
- 支持二维码管理
- 支持用户认证
- 支持统计分析
- 支持Docker部署

## 联系方式

- 项目主页: https://github.com/your-username/wechat-active-qrcode
- 问题反馈: https://github.com/your-username/wechat-active-qrcode/issues 