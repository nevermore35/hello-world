// Todo API
// 一个适合 Go 语言初学者的 RESTful Todo 列表 API
// 学习目标：HTTP 服务器、RESTful API、JSON 编解码、路由、中间件
//
// 运行方式：
//   go run todo_api.go
//   启动后访问 http://localhost:8080
//
// API 端点：
//   GET    /api/todos          获取所有待办事项
//   GET    /api/todos/:id     获取单个待办事项
//   POST   /api/todos         创建待办事项
//   PUT    /api/todos/:id     更新待办事项
//   DELETE /api/todos/:id    删除待办事项
//
// @author 初学者学习项目

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// 常量定义
// ============================================

const (
	// ServerAddr 服务器地址
	ServerAddr = ":8080"
	// APIPrefix API 前缀
	APIPrefix = "/api"
	// StaticDir 静态文件目录
	StaticDir = "./static"
)

// 全局变量
// ============================================

var (
	// todoStore 待办事项存储
	// 使用内存存储，生产环境应使用数据库
	todoStore *TodoStore
)

// 结构体定义
// ============================================

// Todo 待办事项
// 表示一个待办事项
type Todo struct {
	ID          int       `json:"id"`          // ID（JSON 序列化时使用蛇形命名）
	Title       string    `json:"title"`       // 标题
	Content    string    `json:"content"`     // 内容
	Completed   bool      `json:"completed"`  // 是否完成
	Priority   int       `json:"priority"`    // 优先级（1-5）
	Tags       []string  `json:"tags"`        // 标签
	CreatedAt   time.Time `json:"created_at"`   // 创建时间
	UpdatedAt   time.Time `json:"updated_at"`  // 更新时间
}

// TodoStore 待办事项存储
// 内存中的待办事项存储
type TodoStore struct {
	Todos  map[int]*Todo // 使用 map 存储，ID 作为键
	NextID int          // 下一个 ID
	mu    *sync.RWMutex // 读写锁
}

// NewTodoStore 创建新的待办事项存储
func NewTodoStore() *TodoStore {
	return &TodoStore{
		Todos:  make(map[int]*Todo),
		NextID: 1,
		mu:    &sync.RWMutex{},
	}
}

// Request 请求结构体
// API 请求体
type Request struct {
	Title     string   `json:"title"`      // 标题
	Content   string   `json:"content"`    // 内容
	Completed bool     `json:"completed"`  // 是否完成
	Priority  int      `json:"priority"`   // 优先级
	Tags      []string `json:"tags"`       // 标签
}

// Response 响应结构体
// API 响应体
type Response struct {
	Success bool        `json:"success"` // 是否成功
	Message string      `json:"message"` // 消息
	Data    interface{} `json:"data,omitempty"` // 数据
	Error   string      `json:"error,omitempty"` // 错误信息
}

// ErrorResponse 错误响应
// 便捷的错误响应创建
func ErrorResponse(message string) Response {
	return Response{
		Success: false,
		Error:   message,
	}
}

// SuccessResponse 成功响应
// 便捷的成功响应创建
func SuccessResponse(message string, data interface{}) Response {
	return Response{
		Success: true,
		Message: message,
		Data:    data,
	}
}

// 函数定义
// ============================================

// main 主函数
func main() {
	fmt.Println("==================================")
	fmt.Println("      Go Todo API v1.0")
	fmt.Println("==================================")
	fmt.Println()

	// 1. 初始化存储
	todoStore = NewTodoStore()

	// 添加一些示例数据
	initSampleData()

	// 2. 设置路由
	mux := http.NewServeMux()

	// API 路由
	mux.HandleFunc(APIPrefix+"/todos", handleTodos)
	mux.HandleFunc(APIPrefix+"/todos/", handleTodoByID)

	// 静态文件路由（可选）
	// mux.HandleFunc("/", handleStatic)

	// 3. 创建服务器
	server := &http.Server{
		Addr:         ServerAddr,
		Handler:      loggingMiddleware(mux), // 添加日志中间件
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// 4. 启动服务器
	fmt.Printf("服务器启动: http://localhost%s\n", ServerAddr)
	fmt.Printf("API 文档: http://localhost%s/api/todos\n", ServerAddr)
	fmt.Println("==================================")
	fmt.Println()

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("服务器启动失败: %v\n", err)
	}
}

// initSampleData 初始化示例数据
func initSampleData() {
	samples := []*Todo{
		{
			ID:        1,
			Title:     "学习 Go 语言",
			Content:  "学习 Go 基础语法",
			Completed: false,
			Priority:  5,
			Tags:      []string{"学习", "Go"},
			CreatedAt: time.Now(),
		},
		{
			ID:        2,
			Title:     "完成项目",
			Content:  "完成 Todo API 项目",
			Completed: false,
			Priority:  4,
			Tags:      []string{"项目", "Go"},
			CreatedAt: time.Now(),
		},
		{
			ID:        3,
			Title:     "编写文档",
			Content:  "编写项目文档",
			Completed: true,
			Priority:  3,
			Tags:      []string{"文档"},
			CreatedAt: time.Now(),
		},
	}

	for _, todo := range samples {
		todoStore.Todos[todo.ID] = todo
	}
	todoStore.NextID = 4
}

// 路由处理器
// ============================================

// handleTodos 处理 /api/todos 请求
func handleTodos(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// 获取所有待办事项
		getAllTodos(w, r)
	case http.MethodPost:
		// 创建待办事项
		createTodo(w, r)
	default:
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
	}
}

// handleTodoByID 处理 /api/todos/:id 请求
func handleTodoByID(w http.ResponseWriter, r *http.Request) {
	// 提取 ID
	idStr := strings.TrimPrefix(r.URL.Path, APIPrefix+"/todos/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse("无效的 ID"))
		return
	}

	switch r.Method {
	case http.MethodGet:
		// 获取单个待办事项
		getTodoByID(w, r, id)
	case http.MethodPut:
		// 更新待办事项
		updateTodo(w, r, id)
	case http.MethodDelete:
		// 删除待办事项
		deleteTodo(w, r, id)
	default:
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
	}
}

// API 处理函数
// ============================================

// getAllTodos 获取所有待办事项
func getAllTodos(w http.ResponseWriter, r *http.Request) {
	todoStore.mu.RLock()
	defer todoStore.mu.RUnlock()

	// 转换为切片
	todos := make([]*Todo, 0, len(todoStore.Todos))
	for _, todo := range todoStore.Todos {
		todos = append(todos, todo)
	}

	writeJSON(w, http.StatusOK, SuccessResponse("获取成功", todos))
}

// getTodoByID 获取单个待办事项
func getTodoByID(w http.ResponseWriter, r *http.Request, id int) {
	todoStore.mu.RLock()
	defer todoStore.mu.RUnlock()

	todo, ok := todoStore.Todos[id]
	if !ok {
		writeJSON(w, http.StatusNotFound, ErrorResponse("待办事项不存在"))
		return
	}

	writeJSON(w, http.StatusOK, SuccessResponse("获取成功", todo))
}

// createTodo 创建待办事项
func createTodo(w http.ResponseWriter, r *http.Request) {
	// 1. 解析请求体
	var req Request
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse("读取请求体失败"))
		return
	}

	err = json.Unmarshal(body, &req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse("解析请求体失败"))
		return
	}

	// 2. 验证必填字段
	if req.Title == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse("标题不能为空"))
		return
	}

	// 3. 创建待办事项
	todo := &Todo{
		Title:     req.Title,
		Content:  req.Content,
		Completed: req.Completed,
		Priority: req.Priority,
		Tags:     req.Tags,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 4. 保存到存储
	todoStore.mu.Lock()
	todo.ID = todoStore.NextID
	todoStore.Todos[todoStore.NextID] = todo
	todoStore.NextID++
	todoStore.mu.Unlock()

	// 5. 返回响应
	writeJSON(w, http.StatusCreated, SuccessResponse("创建成功", todo))
}

// updateTodo 更新待办事项
func updateTodo(w http.ResponseWriter, r *http.Request, id int) {
	// 1. 检查待办事项是否存在
	todoStore.mu.RLock()
	_, ok := todoStore.Todos[id]
	todoStore.mu.RUnlock()

	if !ok {
		writeJSON(w, http.StatusNotFound, ErrorResponse("待办事项不存在"))
		return
	}

	// 2. 解析请求体
	var req Request
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse("读取请求体失败"))
		return
	}

	err = json.Unmarshal(body, &req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse("解析请求体失败"))
		return
	}

	// 3. 更新待办事项
	todoStore.mu.Lock()
	todo := todoStore.Todos[id]

	// 只更新提供的字段
	if req.Title != "" {
		todo.Title = req.Title
	}
	if req.Content != "" {
		todo.Content = req.Content
	}
	todo.Completed = req.Completed
	if req.Priority > 0 {
		todo.Priority = req.Priority
	}
	if req.Tags != nil {
		todo.Tags = req.Tags
	}
	todo.UpdatedAt = time.Now()

	todoStore.mu.Unlock()

	// 4. 返回响应
	writeJSON(w, http.StatusOK, SuccessResponse("更新成功", todo))
}

// deleteTodo 删除待办事项
func deleteTodo(w http.ResponseWriter, r *http.Request, id int) {
	// 检查待办事项是否存在
	todoStore.mu.RLock()
	_, ok := todoStore.Todos[id]
	todoStore.mu.RUnlock()

	if !ok {
		writeJSON(w, http.StatusNotFound, ErrorResponse("待办事项不存在"))
		return
	}

	// 删除待办事项
	todoStore.mu.Lock()
	delete(todoStore.Todos, id)
	todoStore.mu.Unlock()

	// 返回响应
	writeJSON(w, http.StatusOK, SuccessResponse("删除成功", nil))
}

// 辅助函数
// ============================================

// writeJSON 写入 JSON 响应
func writeJSON(w http.ResponseWriter, status int, response Response) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// 处理 OPTIONS 请求
	if 0 == 0 { // 简化处理
		_ = "" // 占位
	}

	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}

// 中间件
// ============================================

// loggingMiddleware 日志中间件
// 记录每个请求的日志
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// 记录请求
		log.Printf("[%s] %s %s", r.Method, r.URL.Path, r.RemoteAddr)

		// 调用下一个处理器
		next.ServeHTTP(w, r)

		// 记录响应时间
		duration := time.Since(start)
		log.Printf("[%d] %s %s (%v)", 200, r.Method, r.URL.Path, duration)
	})
}

// 代码要点总结
// ============================================
//
// HTTP 服务器核心概念：
//
// 1. http.Handler 接口
//    - 实现 ServeHTTP(http.ResponseWriter, *http.Request) 方法
//    - 用于处理 HTTP 请求
//
// 2. http.ListenAndServe
//    - 启动 HTTP 服务器
//    - 阻塞直到服务器关闭
//
// 3. http.ServeMux
//    - HTTP 路由多路复用器
//    - 根据路径匹配处理器
//
// 4. http.Request
//    - 表示 HTTP 请求
//    - 包含方法、路径、头等信息
//
// 5. http.ResponseWriter
//    - 写入响应
//    - Header() 设置响应头
//    - Write() 写入响应体
//    - WriteHeader() 设置状态码
//
// RESTful API 设计原则：
//
// 1. 资源命名
//    - 使用名词复数
//    - 如 /todos, /users
//
// 2. HTTP 方法
//    - GET: 获取资源
//    - POST: 创建资源
//    - PUT: 更新资源（完整）
//    - PATCH: 更新资源（部分）
//    - DELETE: 删除资源
//
// 3. 状态码
//    - 200 OK
//    - 201 Created
//    - 400 Bad Request
//    - 404 Not Found
//    - 500 Internal Server Error
//
// 4. 响应格式
//    - 使用 JSON
//    - 包含 success、data、error 字段
//
// JSON 处理：
//
// 1. json.Marshal
//    - 将结构体序列化为 JSON
//
// 2. json.Unmarshal
//    - 将 JSON 反序列化为结构体
//
// 3. json.Decoder
//    - 流式解码
//    - 适合大文件
//
// 4. json.Encoder
//    - 流式编码
//    - 适合大文件
//
// 并发安全：
//
// 1. sync.RWMutex
//    - 读写锁
//    - 多个读锁同时持有
//    - 写锁独占
//
// 2. 锁的粒度
//    - 尽量减少锁的范围
//    - 复制数据后再加锁

// 练习题
// ============================================
//
// 练习 1: 添加分页
//   问题: 如何实现分页获取？
//   提示: 添加 limit 和 offset 参数
//
// 练习 2: 添加过滤
//   问题: 如何按状态过滤？
//   提示: 使用 query 参数
//
// 练习 3: 添加排序
//   问题: 如何排序结果？
//   提示: 使用 sort.Slice
//
// 练习 4: 添加数据库
//   问题: 如何使用数据库存储？
//   提示: 使用 gorm 或 sqlx
//
// 练习 5: 添加认证
//   问题: 如何添加 JWT 认证？
//   提示: 实现 JWT 中间件