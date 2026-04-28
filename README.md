# Blog Service

基于 Go 语言的博客后端服务，跟随[《Go 语言编程之旅》](https://golang2.eddycjy.com/)第二章实践开发。\
项目构建了一个简单的，单体的，GO后台服务。实现了对博客业务的增删改查。

## 技术栈

| 组件 | 技术选型                                                                    | 版本 |
|------|-------------------------------------------------------------------------|------|
| Web 框架 | [Gin](https://github.com/gin-gonic/gin)                                 | v1.6.3 |
| ORM | [GORM](https://github.com/jinzhu/gorm)                                  | v1.9.12 |
| 配置管理 | [Viper](https://github.com/spf13/viper)                                 | v1.4.0 |
| 日志 | 自定义 JSON Logger + [Lumberjack](https://github.com/natefinch/lumberjack) | - |
| API 文档 | [Swagger](https://github.com/swaggo/gin-swagger)                        | v1.2.0 |
| 参数校验 | [Validator](https://github.com/go-playground/validator) (Gin 内置)        | v10 |
| 鉴权 | [JWT](https://github.com/dgrijalva/jwt-go)                               | v3.2.0 |
| 数据库 | MySQL                                                                   | 8.0+ |
| 语言 | Go                                                                      | 1.23+ |

## 项目结构

```
blog-service/
├── main.go                 程序入口
├── configs/
│   └── config.yaml         配置文件
├── global/                 全局变量（配置、日志、数据库连接）
├── internal/
│   ├── dao/                数据访问层
│   ├── middleware/          HTTP 中间件
│   ├── model/              数据模型
│   ├── routers/            路由 + Handler
│   └── service/            业务逻辑层
├── pkg/
│   ├── app/                统一响应 + 参数校验 + 分页
│   ├── convert/            类型转换工具
│   ├── errcode/            错误码定义
│   ├── logger/             日志组件
│   ├── setting/            配置读取
│   ├── upload/             文件上传工具（校验 + 存储）
│   └── util/               通用工具（MD5）
├── docs/                   Swagger 文档 + 架构说明
└── storage/
    ├── logs/               日志输出目录
    └── uploads/            上传文件存储目录
```

## 快速开始

### 前置条件

- Go 1.23+
- MySQL 8.0+

### 1. 克隆项目

```bash
git clone https://github.com/harker-xm/goLearning.git
cd goLearning
```

### 2. 创建数据库

```sql
CREATE DATABASE IF NOT EXISTS blog_service DEFAULT CHARACTER SET utf8mb4 DEFAULT COLLATE utf8mb4_general_ci;

USE blog_service;

CREATE TABLE `blog_tag` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(100) DEFAULT '' COMMENT '标签名称',
  `created_on` int(10) unsigned DEFAULT '0' COMMENT '创建时间',
  `created_by` varchar(100) DEFAULT '' COMMENT '创建人',
  `modified_on` int(10) unsigned DEFAULT '0' COMMENT '修改时间',
  `modified_by` varchar(100) DEFAULT '' COMMENT '修改人',
  `deleted_on` int(10) unsigned DEFAULT '0' COMMENT '删除时间',
  `is_del` tinyint(3) unsigned DEFAULT '0' COMMENT '是否删除 0 为未删除、1 为已删除',
  `state` tinyint(3) unsigned DEFAULT '1' COMMENT '状态 0 为禁用、1 为启用',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='标签管理';

CREATE TABLE `blog_article` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `title` varchar(100) DEFAULT '' COMMENT '文章标题',
  `desc` varchar(255) DEFAULT '' COMMENT '文章简述',
  `cover_image_url` varchar(255) DEFAULT '' COMMENT '封面图片地址',
  `content` longtext COMMENT '文章内容',
  `created_on` int(10) unsigned DEFAULT '0' COMMENT '创建时间',
  `created_by` varchar(100) DEFAULT '' COMMENT '创建人',
  `modified_on` int(10) unsigned DEFAULT '0' COMMENT '修改时间',
  `modified_by` varchar(100) DEFAULT '' COMMENT '修改人',
  `deleted_on` int(10) unsigned DEFAULT '0' COMMENT '删除时间',
  `is_del` tinyint(3) unsigned DEFAULT '0' COMMENT '是否删除 0 为未删除、1 为已删除',
  `state` tinyint(3) unsigned DEFAULT '1' COMMENT '状态 0 为禁用、1 为启用',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='文章管理';

CREATE TABLE `blog_article_tag` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `article_id` int(11) NOT NULL COMMENT '文章 ID',
  `tag_id` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '标签 ID',
  `created_on` int(10) unsigned DEFAULT '0' COMMENT '创建时间',
  `created_by` varchar(100) DEFAULT '' COMMENT '创建人',
  `modified_on` int(10) unsigned DEFAULT '0' COMMENT '修改时间',
  `modified_by` varchar(100) DEFAULT '' COMMENT '修改人',
  `deleted_on` int(10) unsigned DEFAULT '0' COMMENT '删除时间',
  `is_del` tinyint(3) unsigned DEFAULT '0' COMMENT '是否删除 0 为未删除、1 为已删除',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='文章标签关联';

CREATE TABLE `blog_auth` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `app_key` varchar(20) DEFAULT '' COMMENT 'Key',
  `app_secret` varchar(50) DEFAULT '' COMMENT 'Secret',
  `created_on` int(10) unsigned DEFAULT 0,
  `created_by` varchar(100) DEFAULT '',
  `modified_on` int(10) unsigned DEFAULT 0,
  `modified_by` varchar(100) DEFAULT '',
  `deleted_on` int(10) unsigned DEFAULT 0,
  `is_del` tinyint(1) unsigned DEFAULT 0,
  PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='认证管理';

-- 插入测试认证数据
INSERT INTO `blog_auth`(`app_key`, `app_secret`, `created_on`, `created_by`, `is_del`)
VALUES ('eddycjy', 'go-programming-tour-book', 0, 'eddycjy', 0);
```

### 3. 修改配置

编辑 `configs/config.yaml`，修改数据库连接信息：

```yaml
Database:
  Username: root
  Password: your_password    # 改成你的密码
  Host: 127.0.0.1:3306
  DBName: blog_service
```

### 4. 安装依赖并运行

```bash
go mod tidy
go run main.go
```

服务启动后监听 `http://localhost:8000`。

### 5. 验证

```bash
# 获取 JWT token
curl -X POST http://127.0.0.1:8000/auth \
  -F app_key=eddycjy -F app_secret=go-programming-tour-book
# 返回 {"token": "eyJhbG..."}

# 用 token 访问 API（替换 <TOKEN> 为上一步拿到的值）
curl 'http://127.0.0.1:8000/api/v1/tags?state=1&token=<TOKEN>'
```

## API 接口

启动服务后访问 Swagger 文档：http://localhost:8000/swagger/index.html

> `/api/v1/` 下的所有接口需要 JWT token 鉴权，通过 `/auth` 获取 token 后，在 Query 参数 `?token=xxx` 或 Header `token: xxx` 中携带。

### 认证

| 方法 | 路径 | 说明 | 鉴权 |
|------|------|------|------|
| POST | `/auth` | 获取 JWT token | 无需 |

### 标签管理

| 方法 | 路径 | 说明 | 状态 |
|------|------|------|------|
| GET | `/api/v1/tags` | 获取标签列表 | ✅ |
| POST | `/api/v1/tags` | 创建标签 | ✅ |
| PUT | `/api/v1/tags/:id` | 更新标签 | ✅ |
| DELETE | `/api/v1/tags/:id` | 删除标签 | ✅ |

### 文章管理

| 方法 | 路径 | 说明 | 状态 |
|------|------|------|------|
| GET | `/api/v1/articles` | 获取文章列表 | 🔲 |
| GET | `/api/v1/articles/:id` | 获取单篇文章 | 🔲 |
| POST | `/api/v1/articles` | 创建文章 | 🔲 |
| PUT | `/api/v1/articles/:id` | 更新文章 | 🔲 |
| DELETE | `/api/v1/articles/:id` | 删除文章 | 🔲 |

### 文件上传

| 方法 | 路径 | 说明 | 状态 |
|------|------|------|------|
| POST | `/upload/file` | 上传文件（图片） | ✅ |
| GET | `/static/*filepath` | 访问已上传的文件 | ✅ |

## 请求/响应示例

### 获取 token

```bash
curl -X POST http://127.0.0.1:8000/auth \
  -F app_key=eddycjy -F app_secret=go-programming-tour-book
```

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### 创建标签（需带 token）

```bash
curl -X POST http://127.0.0.1:8000/api/v1/tags \
  -F 'name=Go' -F 'created_by=test' -F 'state=1'
```

```json
{}
```

### 查询标签列表

```bash
curl http://127.0.0.1:8000/api/v1/tags?state=1
```

```json
{
  "list": [
    {
      "id": 1,
      "name": "Go",
      "state": 1,
      "created_by": "test",
      "created_on": 1745836601,
      "modified_on": 1745836601
    }
  ],
  "pager": {
    "page": 1,
    "page_size": 10,
    "total_rows": 1
  }
}
```

### 参数校验错误

```bash
curl http://127.0.0.1:8000/api/v1/tags?state=6
```

```json
{
  "code": 10000001,
  "msg": "入参错误",
  "details": ["State必须是[0 1]中的一个"]
}
```

### 上传图片

```bash
curl -X POST http://127.0.0.1:8000/upload/file \
  -F file=@photo.png -F type=1
```

```json
{
  "file_access_url": "http://127.0.0.1:8000/static/2c26b46b68ffc68ff99b453c1d304134.png"
}
```

上传后可通过返回的 URL 直接访问文件。支持 `.jpg`、`.jpeg`、`.png`，单文件最大 5MB。

## 架构说明

详见 [docs/architecture.md](docs/architecture.md)

```
请求 → gin.Logger → gin.Recovery → Translations → Handler
                                                      │
                                                      ▼
                                                   Service
                                                      │
                                                      ▼
                                                     DAO
                                                      │
                                                      ▼
                                                    Model
                                                      │
                                                      ▼
                                                    MySQL

/api/v1/* 路由额外经过 JWT 中间件:
请求 → Logger → Recovery → Translations → JWT → Handler → ...
```

## 学习参考

- [《Go 语言编程之旅》](https://golang2.eddycjy.com/) — 煎鱼
- [Gin 文档](https://gin-gonic.com/docs/)
- [GORM 文档](https://gorm.io/docs/)

## License

MIT
