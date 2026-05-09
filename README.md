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

## 学习路径建议

1. **第一阶段**: 密码生成器 → 掌握 Go 基础语法
2. **第二阶段**: URL 检查器 → 掌握 Go 并发
3. **第三阶段**: Todo API → 掌握 Web API
4. **第四阶段**: 博客系统 → 掌握 Web 开发

## 练习题

每个项目都包含练习题，可以在学完基础知识后尝试挑战！

## 相关资源

- [Go 语言官方文档](https://golang.org/doc/)
- [Go 语言中文文档](https://go.zhljj.com/doc/)
- [Go Web 编程](https://github.com/Unknwon/the-way-to-go_ZH_CN)
