// Blog System
// 一个适合 Go 语言初学者的简易博客系统
// 学习目标：HTML 模板、静态文件服务、Session、Cookie、表单处理
//
// 运行方式：
//   go run blog_system.go
//   启动后访问 http://localhost:8080
//
// 页面路由：
//   GET  /               首页（文章列表）
//   GET  /post/:id        文章详情
//   GET  /write          写文章（需登录）
//   POST /api/post       创建文章（API）
//   GET  /login          登录页
//   POST /api/login      登录（API）
//   GET  /logout        登出
//
// @author 初学者学习项目

package main

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	sync "sync"
	"time"
)

// 常量定义
// ============================================

const (
	// ServerAddr 服务器地址
	ServerAddr = ":8080"
	// UploadDir 上传文件目录
	UploadDir = "./uploads"
	// TemplateDir 模板目录
	TemplateDir = "./templates"
	// MaxFileSize 最大文件大小（10MB）
	MaxFileSize = 10 << 20
)

// 全局变量
// ============================================

var (
	// postStore 文章存储
	postStore *PostStore
	// userStore 用户存储
	userStore *UserStore
	// templates 模板缓存
	templates *template.Template
)

// 结构体定义
// ============================================

// Post 文章
type Post struct {
	ID        int       `json:"id"`        // ID
	Title     string    `json:"title"`    // 标题
	Content   string    `json:"content"`  // 内容（Markdown）
	Author    string    `json:"author"`   // 作者
	Tags      []string  `json:"tags"`    // 标签
	Views     int       `json:"views"`    // 浏览量
	CreatedAt time.Time `json:"created_at"` // 创建时间
	UpdatedAt time.Time `json:"updated_at"` // 更新时间
}

// PostStore 文章存储
type PostStore struct {
	Posts  map[int]*Post
	NextID int
	mu    *sync.RWMutex
}

// NewPostStore 创建文章存储
func NewPostStore() *PostStore {
	return &PostStore{
		Posts:  make(map[int]*Post),
		NextID: 1,
		mu:    &sync.RWMutex{},
	}
}

// User 用户
type User struct {
	ID       int       // ID
	Username string   // 用户名
	Password string   // 密码（SHA256 哈希）
	Salt     string   // 盐
	Avatar   string   // 头像
	Bio      string   // 个人简介
}

// UserStore 用户存储
type UserStore struct {
	Users  map[string]*User // username -> User
	ByID   map[int]*User   // id -> User
	NextID int
	mu     *sync.RWMutex
}

// NewUserStore 创建用户存储
func NewUserStore() *UserStore {
	return &UserStore{
		Users: make(map[string]*User),
		ByID:  make(map[int]*User),
		NextID: 1,
		mu:    &sync.RWMutex{},
	}
}

// PageData 页面数据
// 用于模板渲染
type PageData struct {
	Title   string
	User   *User
	Data   interface{}
	Flashes []string
}

// Flash 闪现消息
type Flash struct {
	Type    string
	Message string
}

// 函数定义
// ============================================

// main 主函数
func main() {
	fmt.Println("==================================")
	fmt.Println("      Go 简易博客系统 v1.0")
	fmt.Println("==================================")
	fmt.Println()

	// 1. 创建必要的目录
	createDirectories()

	// 2. 初始化存储
	postStore = NewPostStore()
	userStore = NewUserStore()

	// 3. 添加示例用户和文章
	initSampleData()

	// 4. 加载模板
	templates = loadTemplates()

	// 5. 设置路由
	mux := http.NewServeMux()

	// 静态文件
	mux.HandleFunc("/static/", handleStatic)
	mux.HandleFunc("/uploads/", handleUploads)

	// 页面路由
	mux.HandleFunc("/", handleHome)
	mux.HandleFunc("/post/", handlePost)
	mux.HandleFunc("/write", handleWrite)
	mux.HandleFunc("/login", handleLoginPage)
	mux.HandleFunc("/register", handleRegisterPage)
	mux.HandleFunc("/profile", handleProfile)
	mux.HandleFunc("/logout", handleLogout)

	// API 路由
	mux.HandleFunc("/api/post", handlePostAPI)
	mux.HandleFunc("/api/login", handleLoginAPI)
	mux.HandleFunc("/api/register", handleRegisterAPI)
	mux.HandleFunc("/api/upload", handleUploadAPI)

	// 6. 创建服务器
	server := &http.Server{
		Addr:    ServerAddr,
		Handler: addContextMiddleware(mux),
	}

	// 7. 启动服务器
	fmt.Printf("博客系统启动: http://localhost%s\n", ServerAddr)
	fmt.Println("==================================")
	fmt.Println()
	fmt.Println("示例账户:")
	fmt.Println("  用户名: admin")
	fmt.Println("  密码: admin123")
	fmt.Println()

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("服务器启动失败: %v\n", err)
	}
}

// createDirectories 创建必要的目录
func createDirectories() {
	dirs := []string{UploadDir, TemplateDir}
	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err := os.MkdirAll(dir, 0755)
			if err != nil {
				log.Fatalf("创建目录失败: %v\n", err)
			}
			fmt.Printf("创建目录: %s\n", dir)
		}
	}

	// 创建示例模板
	if _, err := os.Stat(TemplateDir + "/index.html"); os.IsNotExist(err) {
		createSampleTemplates()
	}
}

// initSampleData 初始化示例数据
func initSampleData() {
	// 添加管理员用户
	userStore.mu.Lock()
	userStore.Users["admin"] = &User{
		ID:       1,
		Username: "admin",
		Password: hashPassword("admin123", "admin_salt"),
		Salt:     "admin_salt",
		Avatar:   "",
		Bio:      "博客管理员",
	}
	userStore.ByID[1] = userStore.Users["admin"]
	userStore.NextID = 2
	userStore.mu.Unlock()

	// 添加示例文章
	postStore.mu.Lock()
	samplePosts := []*Post{
		{
			ID:      1,
			Title:   "欢迎来到 Go 博客系统",
			Content: "# 欢迎来到 Go 博客系统\n\n这是一个用 Go 语言编写的简易博客系统。\n\n## 功能特性\n\n-Markdown 支持\n-代码高亮\n-响应式设计\n-用户系统\n\n```go\npackage main\n\nfunc main() {\n\tfmt.Println(\"Hello, World!\")\n}\n```\n\n开始你的博客之旅吧！",
			Author:  "admin",
			Tags:    []string{"Go", "博客", "教程"},
			Views:   100,
			CreatedAt: time.Now(),
		},
		{
			ID:      2,
			Title:   "Go 语言并发编程",
			Content: "# Go 语言并发编程\n\nGo 语言原生支持并发编程，非常强大。\n\n## Goroutine\n\n```go\ngo func() {\n\t// 并发执行\n}()\n```\n\n## Channel\n\n```go\nch := make(chan int)\nch <- 1\n<-ch\n```\n\n学习 Go 并发编程，让你的程序更高效！",
			Author:  "admin",
			Tags:    []string{"Go", "并发"},
			Views:   50,
			CreatedAt: time.Now(),
		},
		{
			ID:      3,
			Title:   "Go Web 开发指南",
			Content: "# Go Web 开发指南\n\nGo 语言的 net/http 包非常强大，可以快速构建 Web 应用。\n\n## 简单示例\n\n```go\nhttp.HandleFunc(\"/\", func(w http.ResponseWriter, r *http.Request) {\n\tfmt.Fprint(w, \"Hello, World!\")\n})\nhttp.ListenAndServe(\":8080\", nil)\n```\n\n开始你的 Go Web 开发之旅吧！",
			Author:  "admin",
			Tags:    []string{"Go", "Web"},
			Views:   30,
			CreatedAt: time.Now(),
		},
	}

	for _, post := range samplePosts {
		postStore.Posts[post.ID] = post
	}
	postStore.NextID = 4
	postStore.mu.Unlock()
}

// loadTemplates 加载模板
func loadTemplates() *template.Template {
	// 加载所有 .html 文件
	templates, err := template.ParseGlob(TemplateDir + "/*.html")
	if err != nil {
		log.Fatalf("加载模板失败: %v\n", err)
	}

	// 添加函数
	templates = templates.Funcs(template.FuncMap{
		"date": func(t time.Time) string {
			return t.Format("2006-01-02 15:04:05")
		},
		"markdown": func(content string) string {
			// 简单的 Markdown 转换（实际应使用库）
			return strings.ReplaceAll(content, "\n", "<br>")
		},
		"truncate": func(s string, length int) string {
			if len(s) > length {
				return s[:length] + "..."
			}
			return s
		},
	})

	return templates
}

// 路由处理器
// ============================================

// handleHome 处理首页
func handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// 获取分页参数
	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		page, _ = strconv.Atoi(p)
	}

	// 获取文章列表
	postStore.mu.RLock()
	posts := make([]*Post, 0, len(postStore.Posts))
	for _, post := range postStore.Posts {
		posts = append(posts, post)
	}
	postStore.mu.RUnlock()

	// 简单的分页
	pageSize := 10
	start := (page - 1) * pageSize
	end := start + pageSize
	if start > len(posts) {
		posts = posts[:0]
	} else {
		if end > len(posts) {
			end = len(posts)
		}
		posts = posts[start:end]
	}

	// 获取当前用户
	user := getCurrentUser(r)

	// 渲染模板
	data := PageData{
		Title: "首页",
		User:  user,
		Data: map[string]interface{}{
			"Posts":    posts,
			"Page":    page,
			"HasMore": end < len(postStore.Posts),
		},
	}

	err := templates.ExecuteTemplate(w, "index.html", data)
	if err != nil {
		log.Printf("渲染模板失败: %v\n", err)
	}
}

// handlePost 处理文章详情页
func handlePost(w http.ResponseWriter, r *http.Request) {
	// 提取文章 ID
	idStr := strings.TrimPrefix(r.URL.Path, "/post/")
	if idStr == r.URL.Path {
		http.NotFound(w, r)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// 获取文章
	postStore.mu.RLock()
	post, ok := postStore.Posts[id]
	postStore.mu.RUnlock()

	if !ok {
		http.NotFound(w, r)
		return
	}

	// 增加浏览量
	postStore.mu.Lock()
	post.Views++
	postStore.mu.Unlock()

	// 获取当前用户
	user := getCurrentUser(r)

	// 渲染模板
	data := PageData{
		Title: post.Title,
		User:  user,
		Data: map[string]interface{}{
			"Post": post,
		},
	}

	err = templates.ExecuteTemplate(w, "post.html", data)
	if err != nil {
		log.Printf("渲染模板失败: %v\n", err)
	}
}

// handleWrite 处理写文章页
func handleWrite(w http.ResponseWriter, r *http.Request) {
	// 检查是否登录
	user := getCurrentUser(r)
	if user == nil {
		http.Redirect(w, r, "/login?redirect=/write", http.StatusFound)
		return
	}

	// 渲染模板
	data := PageData{
		Title: "写文章",
		User:  user,
	}

	templates.ExecuteTemplate(w, "write.html", data)
}

// handleLoginPage 处理登录页
func handleLoginPage(w http.ResponseWriter, r *http.Request) {
	// 如果已登录，跳转到首页
	if user := getCurrentUser(r); user != nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	data := PageData{
		Title: "登录",
	}

	templates.ExecuteTemplate(w, "login.html", data)
}

// handleRegisterPage 处理注册页
func handleRegisterPage(w http.ResponseWriter, r *http.Request) {
	// 如果已登录，跳转到首页
	if user := getCurrentUser(r); user != nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	data := PageData{
		Title: "注册",
	}

	templates.ExecuteTemplate(w, "register.html", data)
}

// handleProfile 处理个人资料页
func handleProfile(w http.ResponseWriter, r *http.Request) {
	user := getCurrentUser(r)
	if user == nil {
		http.Redirect(w, r, "/login?redirect=/profile", http.StatusFound)
		return
	}

	data := PageData{
		Title: "个人资料",
		User:  user,
	}

	templates.ExecuteTemplate(w, "profile.html", data)
}

// handleLogout 处理登出
func handleLogout(w http.ResponseWriter, r *http.Request) {
	// 清除 session
	http.SetCookie(w, &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	http.Redirect(w, r, "/", http.StatusFound)
}

// API 处理器
// ============================================

// handlePostAPI 处理文章 API
func handlePostAPI(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// 获取文章列表
		postStore.mu.RLock()
		posts := make([]*Post, 0, len(postStore.Posts))
		for _, post := range postStore.Posts {
			posts = append(posts, post)
		}
		postStore.mu.RUnlock()

		writeJSON(w, posts)

	case http.MethodPost:
		// 检查是否登录
		user := getCurrentUser(r)
		if user == nil {
			writeError(w, "未登录", http.StatusUnauthorized)
			return
		}

		// 解析请求
		var post Post
		if err := parseJSON(r, &post); err != nil {
			writeError(w, "请求解析失败")
			return
		}

		// 验证
		if post.Title == "" {
			writeError(w, "标题不能为空")
			return
		}

		// 保存
		postStore.mu.Lock()
		post.ID = postStore.NextID
		post.Author = user.Username
		post.CreatedAt = time.Now()
		postStore.Posts[postStore.NextID] = &post
		postStore.NextID++
		postStore.mu.Unlock()

		writeJSON(w, map[string]interface{}{
			"id": post.ID,
		})

	default:
		writeError(w, "方法不允许", http.StatusMethodNotAllowed)
	}
}

// handleLoginAPI 处理登录 API
func handleLoginAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	// 解析请求
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := parseJSON(r, &req); err != nil {
		writeError(w, "请求解析失败")
		return
	}

	// 验证用户
	userStore.mu.RLock()
	user, ok := userStore.Users[req.Username]
	userStore.mu.RUnlock()

	if !ok {
		writeError(w, "用户名或密码错误")
		return
	}

	if user.Password != hashPassword(req.Password, user.Salt) {
		writeError(w, "用户名或密码错误")
		return
	}

	// 创建 session
	session := fmt.Sprintf("%d:%s", user.ID, user.Username)
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    session,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   3600 * 24 * 7, // 7 天
	})

	writeJSON(w, map[string]interface{}{
		"success": true,
		"user":   user,
	})
}

// handleRegisterAPI 处理注册 API
func handleRegisterAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	// 解析请求
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Bio     string `json:"bio"`
	}

	if err := parseJSON(r, &req); err != nil {
		writeError(w, "请求解析失败")
		return
	}

	// 验证
	if req.Username == "" || req.Password == "" {
		writeError(w, "用户名和密码不能为空")
		return
	}

	// 检查用户名是否已存在
	userStore.mu.RLock()
	_, exists := userStore.Users[req.Username]
	userStore.mu.RUnlock()

	if exists {
		writeError(w, "用户名已存在")
		return
	}

	// 创建用户
	salt := generateSalt()
	user := &User{
		ID:       userStore.NextID,
		Username: req.Username,
		Password: hashPassword(req.Password, salt),
		Salt:     salt,
		Bio:      req.Bio,
	}

	userStore.mu.Lock()
	userStore.Users[req.Username] = user
	userStore.ByID[user.ID] = user
	userStore.NextID++
	userStore.mu.Unlock()

	writeJSON(w, map[string]interface{}{
		"success": true,
	})
}

// handleUploadAPI 处理文件上传 API
func handleUploadAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	// 检查是否登录
	user := getCurrentUser(r)
	if user == nil {
		writeError(w, "未登录", http.StatusUnauthorized)
		return
	}

	// 解析 multipart 表单
	err := r.ParseMultipartForm(MaxFileSize)
	if err != nil {
		writeError(w, "解析表单失败")
		return
	}

	// 获取文件
	file, handler, err := r.FormFile("file")
	if err != nil {
		writeError(w, "获取文件失败")
		return
	}
	defer file.Close()

	// 生成文件名
	ext := filepath.Ext(handler.Filename)
	filename := fmt.Sprintf("%d%s", time.Now().Unix(), ext)
	filepath := filepath.Join(UploadDir, filename)

	// 创建文件
	dst, err := os.Create(filepath)
	if err != nil {
		writeError(w, "创建文件失败")
		return
	}
	defer dst.Close()

	// 复制内容
	_, err = io.Copy(dst, file)
	if err != nil {
		writeError(w, "保存文件失败")
		return
	}

	// 返回文件路径
	writeJSON(w, map[string]interface{}{
		"success": true,
		"url":    "/uploads/" + filename,
	})
}

// 静态文件处理
// ============================================

// handleStatic 处理静态文件
func handleStatic(w http.ResponseWriter, r *http.Request) {
	filename := strings.TrimPrefix(r.URL.Path, "/static/")
	filepath := filepath.Join(".", r.URL.Path)

	http.ServeFile(w, r, filepath)
}

// handleUploads 处理上传文件
func handleUploads(w http.ResponseWriter, r *http.Request) {
	filename := strings.TrimPrefix(r.URL.Path, "/uploads/")
	filepath := filepath.Join(UploadDir, filename)

	http.ServeFile(w, r, filepath)
}

// 辅助函数
// ============================================

// getCurrentUser 获取当前用户
func getCurrentUser(r *http.Request) *User {
	cookie, err := r.Cookie("session")
	if err != nil {
		return nil
	}

	// 解析 session
	var userID int
	var username string
	if _, err := fmt.Sscanf(cookie.Value, "%d:%s", &userID, &username); err != nil {
		return nil
	}

	userStore.mu.RLock()
	user := userStore.ByID[userID]
	userStore.mu.RUnlock()

	return user
}

// hashPassword 哈希密码
func hashPassword(password, salt string) string {
	h := sha256.New()
	io.WriteString(h, password+salt)
	return hex.EncodeToString(h.Sum(nil))
}

// generateSalt 生成盐
func generateSalt() string {
	h := md5.New()
	io.WriteString(h, strconv.FormatInt(time.Now().UnixNano(), 10))
	return hex.EncodeToString(h.Sum(nil))[:16]
}

// writeJSON 写入 JSON 响应
func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// parseJSON 解析 JSON 请求
func parseJSON(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

// writeError 写入错误响应
func writeError(w http.ResponseWriter, message string, status ...int) {
	code := http.StatusBadRequest
	if len(status) > 0 {
		code = status[0]
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

// 中间件
// ============================================

// addContextMiddleware 添加上下文中间件
func addContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 可以在这里添加请求 ID、计时等
		next.ServeHTTP(w, r)
	})
}

// createSampleTemplates 创建示例模板
func createSampleTemplates() {
	// index.html
	indexHTML := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}} - Go 博客</title>
    <link rel="stylesheet" href="/static/style.css">
</head>
<body>
    <header>
        <nav>
            <a href="/">首页</a>
            {{if .User}}
                <a href="/write">写文章</a>
                <a href="/profile">个人资料</a>
                <a href="/logout">登出</a>
            {{else}}
                <a href="/login">登录</a>
                <a href="/register">注册</a>
            {{end}}
        </nav>
    </header>

    <main>
        <h1>{{.Title}}</h1>

        <div class="posts">
            {{range .Data.Posts}}
            <article class="post">
                <h2><a href="/post/{{.ID}}">{{.Title}}</a></h2>
                <p class="meta">
                    作者: {{.Author}} | 
                    浏览: {{.Views}} | 
                    时间: {{date .CreatedAt}}
                </p>
                <p>{{truncate .Content 200}}</p>
                <div class="tags">
                    {{range .Tags}}<span class="tag">{{.}}</span>{{end}}
                </div>
            </article>
            {{end}}
        </div>

        {{if .Data.HasMore}}
        <a href="/?page={{add .Data.Page 1}}" class="more">加载更多</a>
        {{end}}
    </main>

    <footer>
        <p>&copy; 2024 Go 博客系统</p>
    </footer>
</body>
</html>`

	os.WriteFile(TemplateDir+"/index.html", []byte(indexHTML), 0644)

	// post.html
	postHTML := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}} - Go 博客</title>
    <link rel="stylesheet" href="/static/style.css">
</head>
<body>
    <header>
        <nav>
            <a href="/">首页</a>
            {{if .User}}
                <a href="/write">写文章</a>
                <a href="/profile">个人资料</a>
                <a href="/logout">登出</a>
            {{else}}
                <a href="/login">登录</a>
            {{end}}
        </nav>
    </header>

    <main>
        <article class="post-detail">
            <h1>{{.Data.Post.Title}}</h1>
            <p class="meta">
                作者: {{.Data.Post.Author}} | 
                浏览: {{.Data.Post.Views}} | 
                时间: {{date .Data.Post.CreatedAt}}
            </p>
            <div class="content">
                {{markdown .Data.Post.Content}}
            </div>
        </article>
    </main>

    <footer>
        <p>&copy; 2024 Go 博客系统</p>
    </footer>
</body>
</html>`

	os.WriteFile(TemplateDir+"/post.html", []byte(postHTML), 0644)
}

// 代码要点总结
// ============================================
//
// HTML 模板：
//
// 1. template.ParseGlob
//    - 批量解析模板文件
//    - 支持通配符
//
// 2. template.FuncMap
//    - 添加自定义函数
//    - 扩展模板能力
//
// 3. ExecuteTemplate
//    - 执行指定模板
//    - 传入数据
//
// 4. 模板语法
//    - {{.Field}} 字段
//    - {{range}} 循环
//    - {{if}} 条件
//    - {{template}} 模板包含
//
// 文件上传：
//
// 1. r.ParseMultipartForm
//    - 解析 multipart 表单
//    - 包含文件
//
// 2. r.FormFile
//    - 获取上传的文件
//    - 返回 file 和 header
//
// 3. os.Create
//    - 创建文件
//    - 用于保存上传内容
//
// Session/Cookie：
//
// 1. http.Cookie
//    - 创建 Cookie
//    - 设置属性
//
// 2. r.Cookie
//    - 读取 Cookie
//
// 3. http.SetCookie
//    - 设置 Cookie 到响应

// 练习题
// ============================================
//
// 练习 1: 添加 Markdown 渲染
//   问题: 如何渲染 Markdown？
//   提示: 使用 blackfriday 或 goldmark 库
//
// 练习 2: 添加评论功能
//   问题: 如何添加评论？
//   提示: 创建评论存储结构
//
// 练习 3: 添加搜索功能
//   问题: 如何搜索文章？
//   提示: 使用正则或索引
//
// 练习 4: 添加标签分类
//   问题: 如何按标签筛选？
//   提示: 使用 query 参数
//
// 练习 5: 添加数据库
//   问题: 如何使用数据库？
//   提示: 使用 gorm 或 sqlx