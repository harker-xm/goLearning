# Blog Service 系统架构总结

## 一、项目全景目录

```
blog-service/
│
├── main.go                          ← 程序入口：init() 初始化 + main() 启动服务
├── configs/
│   └── config.yaml                  ← 唯一配置文件：Server/App/Database 三个区段
│
├── global/                          ← 全局变量仓库（各模块都从这里读取）
│   ├── setting.go                      ServerSetting / AppSetting / DatabaseSetting / Logger
│   └── db.go                           DBEngine（gorm 数据库连接）
│
├── internal/                        ← 内部业务代码（不对外暴露）
│   ├── dao/
│   │   ├── dao.go                      Dao 结构体（封装 *gorm.DB）+ New 构造函数
│   │   └── tag.go                      Tag 数据访问：CountTag/GetTagList/CreateTag/UpdateTag/DeleteTag
│   ├── middleware/
│   │   └── translations.go             翻译中间件：把 validator 错误信息翻译成中文
│   ├── model/
│   │   ├── model.go                    公共 Model 结构体 + NewDBEngine + 3 个 gorm 回调（时间戳/软删除）
│   │   ├── tag.go                      Tag 模型 + 5 个 CRUD 方法 + TagSwagger
│   │   ├── article.go                  Article 模型 + ArticleSwagger
│   │   └── article_tag.go             ArticleTag 关联模型
│   ├── routers/
│   │   ├── router.go                   路由注册中心：中间件 + API 分组 + Swagger
│   │   └── api/v1/
│   │       ├── tag.go                  Tag Handler：List/Create/Update/Delete 完整实现
│   │       └── article.go             Article Handler（空实现，待开发）
│   └── service/
│       ├── service.go                  Service 结构体（ctx + dao）+ New 构造函数
│       ├── tag.go                      Tag 请求结构体 + 5 个业务方法
│       └── article.go                 Article 请求结构体（业务方法待实现）
│
├── pkg/                             ← 可复用公共包（理论上可以被其他项目引用）
│   ├── app/
│   │   ├── app.go                      统一响应封装（ToResponse/ToResponseList/ToErrorResponse）
│   │   ├── form.go                     BindAndValid 参数绑定+校验+翻译
│   │   └── pagination.go              分页工具（GetPage/GetPageSize/GetPageOffset）
│   ├── convert/
│   │   └── convert.go                 类型转换工具（StrTo）
│   ├── errcode/
│   │   ├── errcode.go                  Error 结构体 + 错误码 → HTTP 状态码映射
│   │   ├── common_code.go            预定义 9 个通用错误码
│   │   └── module_code.go            业务模块错误码（标签 5 个）
│   ├── logger/
│   │   └── logger.go                  JSON 日志组件（6 级日志 + 文件切割）
│   └── setting/
│       ├── setting.go                  viper 配置读取
│       └── section.go                 配置结构体定义 + ReadSection
│
├── docs/                            ← swag init 自动生成 + 架构文档
│   ├── docs.go / swagger.json / swagger.yaml
│   └── architecture.md              本文件
│
└── storage/logs/                    ← 日志输出目录
    └── app.log
```

---

## 二、启动流程

```
$ go run main.go
         │
         ▼
    Go 运行时按 import 顺序初始化各包
         │
         ├── errcode 包初始化
         │   └── common_code.go: 注册 9 个通用错误码
         │   └── module_code.go: 注册 5 个标签业务错误码
         │
         ├── docs 包的 init()
         │   └── 注册 Swagger 文档信息到 swag 全局注册表
         │
         ▼
    main 包的 init() 执行 —— 三步初始化链
         │
    ①    setupSetting()
         │  viper.New() → 读 configs/config.yaml → 解析成内存 map
         │  ReadSection("Server", ...)   → global.ServerSetting
         │  ReadSection("App", ...)      → global.AppSetting
         │  ReadSection("Database", ...) → global.DatabaseSetting
         │  ReadTimeout *= time.Second   (60 → 60秒)
         │
    ②    setupLogger()
         │  lumberjack.Logger{ Filename: "storage/logs/app.log", MaxSize: 600MB, MaxAge: 10天 }
         │  logger.NewLogger(lumberjack, ...) → global.Logger
         │
    ③    setupDBEngine()
         │  model.NewDBEngine(global.DatabaseSetting)
         │  → gorm.Open("mysql", dsn)
         │  → 注册 3 个回调：创建时间戳 / 更新时间戳 / 软删除
         │  → SetMaxIdleConns(10), SetMaxOpenConns(30)
         │  → global.DBEngine = db 连接
         │
         ▼
    main() 执行
         │  router := routers.NewRouter()   构建路由
         │  http.Server{ Addr: ":8000", ReadTimeout: 60s, WriteTimeout: 60s }
         │  s.ListenAndServe()
         │
         ▼
    🟢 服务启动完毕，监听 :8000
```

任何一步 setupXxx 返回 error，程序直接 `log.Fatalf` 退出，不会带着残缺状态运行。

---

## 三、NewRouter() 路由注册

```go
func NewRouter() *gin.Engine {
    r := gin.New()                    // 空引擎
    r.Use(gin.Logger())               // 中间件1：请求日志
    r.Use(gin.Recovery())             // 中间件2：panic 捕获
    r.Use(middleware.Translations())  // 中间件3：初始化翻译器
    r.GET("/swagger/*any", ...)       // Swagger 文档路由

    apiv1 := r.Group("/api/v1")       // API v1 路由组
    {
        // 标签 5 条路由（CRUD 已完整实现）
        // 文章 6 条路由（handler 待实现）
    }
    return r
}
```

路由树：

```
gin.Engine
├── GET 树
│   ├── /swagger/*any
│   ├── /api/v1/tags           → [Logger, Recovery, Translations, Tag.List]       ✅ 已实现
│   ├── /api/v1/articles       → [Logger, Recovery, Translations, Article.List]   🔲 待实现
│   └── /api/v1/articles/:id   → [Logger, Recovery, Translations, Article.Get]    🔲 待实现
├── POST 树
│   ├── /api/v1/tags           → [Logger, Recovery, Translations, Tag.Create]     ✅ 已实现
│   └── /api/v1/articles       → [Logger, Recovery, Translations, Article.Create] 🔲 待实现
├── PUT 树
│   ├── /api/v1/tags/:id       → [Logger, Recovery, Translations, Tag.Update]     ✅ 已实现
│   └── /api/v1/articles/:id   → [Logger, Recovery, Translations, Article.Update] 🔲 待实现
├── PATCH 树
│   ├── /api/v1/tags/:id/state    ✅ 已实现
│   └── /api/v1/articles/:id/state 🔲 待实现
└── DELETE 树
    ├── /api/v1/tags/:id          ✅ 已实现
    └── /api/v1/articles/:id      🔲 待实现
```

每条路由都是 **4 个 handler**（3 个中间件 + 1 个业务 handler）。

---

## 四、请求链路 — 完整 CRUD 流程

### 场景 A：校验失败（GET /api/v1/tags?state=6）

```
客户端 → http.Server → gin.ServeHTTP
  → ① Logger 中间件（记录开始时间）
  → ② Recovery 中间件（defer recover）
  → ③ Translations 中间件
       c.Set("trans", 中文翻译器)
  → ④ Tag.List handler
       ├─ BindAndValid(c, &param)
       │    ├─ ShouldBind: state=6 → param.State=6
       │    └─ validator: oneof=0 1 → 6 不合法 ❌
       │       → Translate(trans) → "State必须是[0 1]中的一个"
       └─ ToErrorResponse(InvalidParams)
            → HTTP 400 {"code":10000001, "msg":"入参错误", "details":["..."]}
```

### 场景 B：正常查询（GET /api/v1/tags?state=1）

```
客户端 → http.Server → gin.ServeHTTP
  → ① Logger → ② Recovery → ③ Translations
  → ④ Tag.List handler
       ├─ BindAndValid → ✅ 校验通过
       ├─ svc := service.New(ctx)
       │    └─ svc.dao = dao.New(global.DBEngine)
       ├─ svc.CountTag(&CountTagRequest{State: 1})
       │    └─ dao.CountTag → tag.Count(db)
       │         → SQL: SELECT count(*) FROM blog_tag WHERE state=1 AND is_del=0
       │         → 返回 totalRows = 1
       ├─ svc.GetTagList(&param, &pager)
       │    └─ dao.GetTagList → tag.List(db, offset, limit)
       │         → SQL: SELECT * FROM blog_tag WHERE state=1 AND is_del=0 LIMIT 10 OFFSET 0
       │         → 返回 []*Tag
       └─ ToResponseList(tags, totalRows)
            → HTTP 200 {"list":[...], "pager":{"page":1,"page_size":10,"total_rows":1}}
```

### 场景 C：创建标签（POST /api/v1/tags）

```
客户端 → ... → Tag.Create handler
  ├─ BindAndValid → ✅ name="Rust", created_by="test", state=1
  ├─ svc.CreateTag(&param)
  │    └─ dao.CreateTag → tag.Create(db)
  │         → gorm 回调: updateTimeStampForCreateCallback
  │              自动设置 created_on = now, modified_on = now
  │         → SQL: INSERT INTO blog_tag (name,state,created_by,created_on,modified_on,...) VALUES (...)
  └─ ToResponse(gin.H{})
       → HTTP 200 {}
```

### 场景 D：更新标签（PUT /api/v1/tags/1）

```
客户端 → ... → Tag.Update handler
  ├─ ID = convert.StrTo(c.Param("id")).MustUInt32()   从 URL 路径取 id
  ├─ BindAndValid → ✅
  ├─ svc.UpdateTag(&param)
  │    └─ dao.UpdateTag → 构建 map{"state":1, "modified_by":"test", "name":"Golang"}
  │         → tag.Update(db, values)
  │              → gorm 回调: updateTimeStampForUpdateCallback
  │                   自动设置 modified_on = now
  │              → SQL: UPDATE blog_tag SET name='Golang',state=1,modified_by='test',modified_on=... WHERE id=1 AND is_del=0
  └─ HTTP 200 {}
```

**为什么 Update 用 map 而不用 struct？** GORM 的 Update(struct) 会跳过零值字段。如果你想把 state 改成 0，GORM 认为 0 是零值会忽略它。用 map 传入则不会有这个问题。

### 场景 E：删除标签（DELETE /api/v1/tags/1）

```
客户端 → ... → Tag.Delete handler
  ├─ BindAndValid → ✅
  ├─ svc.DeleteTag(&param)
  │    └─ dao.DeleteTag → tag.Delete(db)
  │         → gorm 回调: deleteCallback（软删除）
  │              发现 Model 有 DeletedOn + IsDel 字段
  │              → SQL: UPDATE blog_tag SET deleted_on=now, is_del=1 WHERE id=1 AND is_del=0
  │              （不是 DELETE FROM，而是 UPDATE 标记删除）
  └─ HTTP 200 {}
```

之后查询时 `WHERE is_del=0` 会自动过滤掉已删除的记录。

---

## 五、分层架构与各层职责

```
┌─────────────┐
│   main.go   │  程序入口：初始化配置/日志/数据库 → 启动 HTTP 服务
└──────┬──────┘
       │
┌──────┴──────┐
│   global/   │  全局变量仓库：所有模块的共享状态
│             │  ServerSetting / AppSetting / DatabaseSetting / Logger / DBEngine
└──────┬──────┘
       │ 被以下各层读取
       ▼
┌──────────────────────────────────────────────────────────────┐
│                    internal/ (业务层)                          │
│                                                               │
│  routers/router.go      路由注册：URL → Handler 的映射关系    │
│       │                                                       │
│       ▼                                                       │
│  middleware/             中间件：请求到达 handler 之前的       │
│  translations.go        预处理（翻译器初始化）                 │
│       │                                                       │
│       ▼                                                       │
│  routers/api/v1/        Handler 层：接收请求、校验参数、      │
│  tag.go / article.go    调用 service、返回响应                │
│       │                                                       │
│       ▼                                                       │
│  service/               Service 层：业务逻辑编排              │
│  tag.go / article.go    请求结构体 + 调用 DAO 方法            │
│       │                                                       │
│       ▼                                                       │
│  dao/                   DAO 层：数据访问对象                   │
│  tag.go                 封装 Model 方法，处理分页/参数组装     │
│       │                                                       │
│       ▼                                                       │
│  model/                 Model 层：数据库表结构映射            │
│  tag.go / article.go    GORM CRUD 方法 + 回调（时间戳/软删除）│
└──────────────────────────────────────────────────────────────┘
       │ 依赖
       ▼
┌──────────────────────────────────────────────────────────────┐
│                    pkg/ (公共工具层)                            │
│                                                               │
│  pkg/setting/       配置读取（viper 封装）                     │
│  pkg/logger/        日志组件（JSON 格式 + 文件切割）           │
│  pkg/errcode/       错误码定义 + HTTP 状态码映射               │
│  pkg/app/           统一响应 + 参数校验 + 分页                 │
│  pkg/convert/       类型转换工具                               │
└──────────────────────────────────────────────────────────────┘
```

### 各层数据流向

```
Handler 接收请求
  │
  ├─ 接收 *gin.Context，从中提取参数
  ├─ 用 service.XXXRequest 结构体做参数校验
  ├─ 创建 Service 实例
  │
  ▼
Service 编排业务
  │
  ├─ 接收 Request 结构体
  ├─ 调用 DAO 方法（可能调多个）
  ├─ 返回结果给 Handler
  │
  ▼
DAO 数据访问
  │
  ├─ 接收基础类型参数（string, uint8, int）
  ├─ 组装 Model 对象
  ├─ 调用 Model 的 CRUD 方法
  │
  ▼
Model 数据库操作
  │
  ├─ 接收 *gorm.DB
  ├─ 拼接 gorm 查询链（Where/Offset/Limit/Find/Create/Update/Delete）
  ├─ gorm 回调自动处理时间戳和软删除
  ├─ 返回结果
  │
  ▼
MySQL 数据库
```

### internal/ vs pkg/ 的区别

| | internal/ | pkg/ |
|--|-----------|------|
| **含义** | 这个项目**私有**的业务代码 | **可复用**的公共工具代码 |
| **Go 的约束** | Go 编译器禁止外部项目 import internal/ 下的包 | 任何项目都能 import |
| **举例** | Tag 的路由、模型、service — 只属于这个博客系统 | 日志、错误码、分页 — 换个项目也能用 |

---

## 六、GORM 回调机制

在 `NewDBEngine` 中注册了 3 个回调，替换 gorm 默认行为：

| 回调 | 触发时机 | 作用 |
|------|----------|------|
| `updateTimeStampForCreateCallback` | INSERT 之前 | 自动填充 `created_on` 和 `modified_on` 为当前时间戳 |
| `updateTimeStampForUpdateCallback` | UPDATE 之前 | 自动更新 `modified_on` 为当前时间戳 |
| `deleteCallback` | DELETE 时 | 如果 Model 有 `DeletedOn`+`IsDel` 字段，执行**软删除**（UPDATE SET is_del=1），否则执行真删除 |

---

## 七、错误码体系

### 通用错误码（common_code.go）

| 错误码 | 含义 | HTTP 状态码 |
|--------|------|------------|
| 0 | 成功 | 200 |
| 10000000 | 服务内部错误 | 500 |
| 10000001 | 入参错误 | 400 |
| 10000002 | 找不到 | 500 |
| 10000003~06 | 鉴权失败 | 401 |
| 10000007 | 请求过多 | 429 |

### 标签业务错误码（module_code.go）

| 错误码 | 含义 | HTTP 状态码 |
|--------|------|------------|
| 20010001 | 获取标签列表失败 | 500 |
| 20010002 | 创建标签失败 | 500 |
| 20010003 | 更新标签失败 | 500 |
| 20010004 | 删除标签失败 | 500 |
| 20010005 | 统计标签失败 | 500 |

---