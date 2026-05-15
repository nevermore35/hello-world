# Go 语言初学者学习项目

一套适合 Go 语言初学者学习的项目，从简单到复杂，循序渐进。

## 项目列表

### 1. 密码生成器 (password_generator.go)
- **难度**: ⭐ 入门
- **学习目标**: Go 基本语法、字符串处理、命令行参数
- **功能**: 生成随机密码，支持自定义长度和字符类型
- **运行**:
  ```bash
  go run password_generator.go
  go run password_generator.go -length 16 -numbers -symbols
  ```
- **知识点**:
  - Go 程序基本结构
  - 变量和常量
  - 结构体
  - 函数定义和调用
  - 条件语句和循环
  - 格式化输出

### 2. 并发URL检查器 (url_checker.go)
- **难度**: ⭐⭐ 初级
- **学习目标**: Go 并发编程、goroutine、channel
- **功能**: 并发检查多个 URL 的可用性
- **运行**:
  ```bash
  go run url_checker.go
  go run url_checker.go -urls https://google.com,https://github.com
  ```
- **知识点**:
  - goroutine 并发执行
  - channel 通道通信
  - WaitGroup 等待组
  - Mutex 互斥锁
  - atomic 原子操作
  - Worker Pool 工作池模式

### 3. Todo 列表 API (todo_api.go)
- **难度**: ⭐⭐⭐ 中级
- **学习目标**: HTTP 服务器、RESTful API、JSON
- **功能**: RESTful API 服务端
- **运行**:
  ```bash
  go run todo_api.go
  # 然后访问 http://localhost:8080/api/todos
  ```
- **API 端点**:
  - `GET /api/todos` - 获取所有待办
  - `GET /api/todos/:id` - 获取单个待办
  - `POST /api/todos` - 创建待办
  - `PUT /api/todos/:id` - 更新待办
  - `DELETE /api/todos/:id` - 删除待办
- **知识点**:
  - net/http 服务器
  - RESTful API 设计
  - JSON 编解码
  - 路由处理
  - 读写锁

### 4. 简易博客系统 (blog_system.go)
- **难度**: ⭐⭐⭐⭐ 中高级
- **学习目标**: HTML 模板、静态文件、Cookie、文件上传
- **功能**: 带用户系统的博客
- **运行**:
  ```bash
  go run blog_system.go
  # 访问 http://localhost:8080
  ```
- **登录账户**:
  - 用户名: admin
  - 密码: admin123
- **知识点**:
  - html/template 模板
  - 静态文件服务
  - Cookie 和 Session
  - 文件上传
  - 表单处理

### 5. 协程专题 (goroutines.go)
- **难度**: ⭐⭐⭐ 中级
- **学习目标**: 系统、深入地掌握 Go 协程（goroutine）与并发原语
- **运行**:
  ```bash
  # 运行所有主题
  go run goroutines.go

  # 只看某个主题
  go run goroutines.go -topic basic
  go run goroutines.go -topic channels
  go run goroutines.go -topic patterns
  ```
- **覆盖的主题**:
  - `basic`      goroutine 启动、参数求值、GOMAXPROCS
  - `channels`   无缓冲 / 带缓冲 / 单向 channel、关闭、range
  - `select`     多路复用、超时、非阻塞收发、退出信号
  - `waitgroup`  `sync.WaitGroup` 的正确用法与常见误用
  - `mutex`      `sync.Mutex` / `sync.RWMutex` 保护共享数据
  - `once`       `sync.Once` 实现惰性初始化 / 单例
  - `atomic`     `sync/atomic` 原子计数与 CAS
  - `context`    `context` 的取消、超时与传值
  - `patterns`   Worker Pool / Pipeline / Fan-out + Fan-in / Semaphore
  - `pitfalls`   协程泄漏、循环变量捕获、死锁、并发 map、errgroup 模式
- **知识点**:
  - goroutine 生命周期与调度
  - channel 通信、关闭与方向控制
  - sync 包常用原语（Mutex / RWMutex / WaitGroup / Once / atomic）
  - context 在请求链路上的取消传播
  - 经典并发模式与典型 bug 案例

### 6. 高级语法介绍 (advanced_syntax.go)
- **难度**: ⭐⭐⭐⭐⭐ 高级
- **学习目标**: 系统了解 Go 语言中常用的高级特性与惯用法
- **运行**:
  ```bash
  # 运行所有主题
  go run advanced_syntax.go

  # 只看某个主题
  go run advanced_syntax.go -topic generics
  go run advanced_syntax.go -topic concurrent
  ```
- **覆盖的主题**:
  - `interface`  接口、类型断言、类型 switch、`any`
  - `embed`      结构体嵌入与接口嵌入（组合优于继承）
  - `errors`     哨兵错误、自定义错误、`errors.Is` / `errors.As`、错误包装
  - `defer`      `defer` 执行顺序、参数求值时机、`panic` / `recover`
  - `closure`    闭包、高阶函数、循环变量捕获陷阱
  - `options`    函数选项模式（Functional Options Pattern）
  - `generics`   泛型（Go 1.18+）：类型参数、类型集合、泛型容器
  - `reflect`    反射：检查与修改字段、读取 struct tag
  - `context`    `context` 超时、取消、传值
  - `concurrent` 高级并发：pipeline、fan-in、`select` + 超时
- **适合人群**: 已完成前 4 个项目，希望理解 Go 惯用法与现代特性的学习者
- **建议**: 按 `interface → embed → errors → defer → closure → options → generics → reflect → context → concurrent` 的顺序逐个阅读，并完成文件末尾的 5 道练习题

### 7. GORM 使用示例 (gorm_example.go)
- **难度**: ⭐⭐⭐⭐ 中高级
- **学习目标**: 掌握 Go 中最流行的 ORM 框架 GORM 的使用方法
- **前置要求**: 已完成 Todo API，了解数据库基础
- **运行**:
  ```bash
  # 初始化模块并安装依赖
  go mod init example
  go get gorm.io/gorm gorm.io/driver/sqlite

  # 运行所有主题
  go run gorm_example.go

  # 只看某个主题
  go run gorm_example.go -topic crud
  go run gorm_example.go -topic query
  go run gorm_example.go -topic associations
  go run gorm_example.go -topic transaction
  go run gorm_example.go -topic advanced
  go run gorm_example.go -topic pitfalls
  ```
- **覆盖的主题**:
  - `model`       模型定义、字段标签、约定优于配置
  - `connect`     数据库连接、驱动选择、连接池配置
  - `migrate`     自动迁移、索引、Migrator API
  - `crud`        Create / Read / Update / Delete 基本操作
  - `query`       条件查询、排序分页、聚合、子查询、Pluck
  - `associations` Has One / Has Many / Many to Many、Preload、Joins
  - `transaction`  自动事务、手动事务、嵌套事务（SavePoint）
  - `advanced`    原生 SQL、Scopes、批量操作、软删除
  - `pitfalls`    零值更新、N+1 查询、连接池耗尽、事务遗漏
- **适合人群**: 已完成 Todo API 和博客系统，希望将内存存储替换为数据库的学习者
- **建议**: 按 `model → connect → migrate → crud → query → associations → transaction → advanced → pitfalls` 的顺序逐个阅读，然后尝试完成文件末尾的 5 道练习题

## GORM 框架介绍

GORM 是 Go 语言中最流行的 ORM（对象关系映射）库，基于 `database/sql` 构建，提供了模型定义、自动迁移、链式查询、关联管理、事务等丰富的功能。对于已经从 `todo_api.go` 和 `blog_system.go` 掌握了 Web 开发基础的学习者来说，GORM 是将内存存储升级为数据库存储的最佳选择。

### GORM 的特点

- **开发效率高**: 链式 API 简洁直观，减少大量样板 SQL 代码
- **自动迁移**: 根据 struct 定义自动创建/更新表结构
- **关联管理**: Has One / Has Many / Belongs To / Many To Many 一站式处理
- **事务支持**: 自动/手动事务、嵌套事务（SavePoint）
- **多数据库支持**: SQLite、MySQL、PostgreSQL、SQL Server

### GORM 和 `database/sql` 的关系

- `database/sql` 是 Go 标准库，提供基础的数据库操作接口
- GORM 是建立在 `database/sql` 之上的 ORM 层，提供更高层次的抽象
- 学习建议是先了解标准库，再学习 GORM，这样更容易理解 ORM 解决了什么问题

### 最小示例

```go
package main

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Product struct {
	gorm.Model
	Code  string
	Price uint
}

func main() {
	db, _ := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	db.AutoMigrate(&Product{})

	db.Create(&Product{Code: "D42", Price: 100})

	var product Product
	db.First(&product, 1)                 // 按主键查询
	db.First(&product, "code = ?", "D42") // 按条件查询

	db.Model(&product).Update("Price", 200)
	db.Delete(&product)
}
```

### 运行方式

```bash
go mod init example
go get gorm.io/gorm gorm.io/driver/sqlite
go run main.go
```

### 适合用 GORM 练习的方向

- 把 `todo_api.go` 的内存存储改为 GORM + SQLite
- 把 `blog_system.go` 的 Posts / Users 改为 GORM 模型
- 实现用户之间的关注/粉丝关系（多对多自引用）
- 为博客添加评论系统（使用 GORM 的 Has Many 关联）
- 实现文章标签系统（使用 GORM 的 Many To Many 关联）

## Gin 框架介绍

Gin 是 Go 语言中非常流行的 Web 框架，底层基于标准库 `net/http`，但在路由组织、中间件处理、参数绑定和 JSON 响应等方面做了更高层的封装。对于已经学完 `todo_api.go` 的学习者来说，Gin 是从标准库 Web 开发迈向工程化开发的一个自然下一步。

### Gin 的特点

- **性能高**: 路由匹配效率高，适合构建高并发 Web 服务
- **API 简洁**: 路由、中间件、请求参数处理写法直观
- **中间件机制清晰**: 便于统一处理日志、鉴权、跨域、异常恢复
- **JSON 开发友好**: 非常适合构建 RESTful API
- **生态完善**: 社区中有大量认证、校验、文档等扩展方案

### Gin 和 `net/http` 的关系

- `net/http` 是 Go 标准库，适合打基础，帮助理解 HTTP 服务本质
- Gin 是建立在 `net/http` 之上的封装，减少重复样板代码
- 学习建议是先掌握标准库，再学习 Gin，这样更容易理解框架做了什么

### Gin 中常见的核心概念

- **Engine**: Gin 的核心实例，负责路由注册和请求分发
- **Context (`*gin.Context`)**: 封装请求和响应，常用于取参数、返回 JSON
- **Router Group**: 用于给一组接口统一加前缀，例如 `/api`
- **Middleware**: 在请求处理前后执行的逻辑，例如日志、鉴权、异常恢复
- **Binding**: 把 JSON、表单、Query 参数绑定到结构体

### 最小示例

```go
package main

import "github.com/gin-gonic/gin"

func main() {
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.Run(":8080")
}
```

### 运行方式

```bash
go mod init example
go get github.com/gin-gonic/gin
go run main.go
```

启动后访问 `http://localhost:8080/ping`，会得到 JSON 响应：

```json
{"message":"pong"}
```

### 适合用 Gin 练习的方向

- 把 `todo_api.go` 改写成 Gin 版本
- 为 Todo API 增加路由分组和中间件
- 使用结构体验证请求参数
- 增加统一错误返回格式
- 尝试接入 JWT 登录认证

## 学习路径建议

1. **第一阶段**: 密码生成器 → 掌握 Go 基础语法
2. **第二阶段**: URL 检查器 → 上手 Go 并发的常见用法
3. **第三阶段**: 协程专题 → 系统掌握 goroutine / channel / sync / context
4. **第四阶段**: Todo API → 掌握 Web API
5. **第五阶段**: 博客系统 → 掌握 Web 开发
6. **第六阶段**: GORM 示例 → 掌握数据库 ORM 操作
7. **第七阶段**: 高级语法介绍 → 掌握 Go 惯用法与现代特性

## 练习题

每个项目都包含练习题，可以在学完基础知识后尝试挑战！

- GORM 的练习题尤其推荐**将 Todo API 或博客系统改用 GORM + SQLite**，这是最实用的综合练习。

## 相关资源

- [Go 语言官方文档](https://golang.org/doc/)
- [Go 语言中文文档](https://go.zhljj.com/doc/)
- [GORM 官方文档](https://gorm.io/zh_CN/)
- [Gin 官方文档](https://gin-gonic.com/)
- [Go Web 编程](https://github.com/Unknwon/the-way-to-go_ZH_CN)
