// Advanced Syntax
// Go 语言高级语法介绍 - 适合已掌握基础语法的学习者
// 学习目标：接口、泛型、反射、Context、错误处理、并发模式、defer/panic/recover、
//          结构体嵌入、闭包与高阶函数、函数选项模式
//
// 运行方式：
//   go run advanced_syntax.go
//   go run advanced_syntax.go -topic interface
//   go run advanced_syntax.go -topic generics
//   go run advanced_syntax.go -topic all
//
// 支持的主题 (topic)：
//   interface  - 接口与多态
//   generics   - 泛型 (需要 Go 1.18+)
//   reflect    - 反射
//   context    - 上下文与取消
//   errors     - 错误处理
//   concurrent - 高级并发模式
//   defer      - defer / panic / recover
//   embed      - 结构体嵌入
//   closure    - 闭包与高阶函数
//   options    - 函数选项模式
//   all        - 运行所有示例
//
// @author 初学者学习项目

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"
)

// ============================================
// 入口函数
// ============================================

// 演示函数表，主题名 -> 演示函数
// 通过映射实现命令分发，是一种非常常见的设计模式
var demos = map[string]func(){
	"interface":  demoInterface,
	"generics":   demoGenerics,
	"reflect":    demoReflect,
	"context":    demoContext,
	"errors":     demoErrors,
	"concurrent": demoConcurrent,
	"defer":      demoDefer,
	"embed":      demoEmbed,
	"closure":    demoClosure,
	"options":    demoOptions,
}

// 推荐运行顺序
var demoOrder = []string{
	"interface",
	"embed",
	"errors",
	"defer",
	"closure",
	"options",
	"generics",
	"reflect",
	"context",
	"concurrent",
}

func main() {
	topic := flag.String("topic", "all", "要演示的高级语法主题")
	flag.Parse()

	fmt.Println("====================================")
	fmt.Println("    Go 语言高级语法介绍 v1.0")
	fmt.Println("====================================")
	fmt.Println()

	if *topic == "all" {
		for _, name := range demoOrder {
			runDemo(name)
		}
		return
	}

	runDemo(*topic)
}

// runDemo 运行单个主题演示
func runDemo(name string) {
	fn, ok := demos[name]
	if !ok {
		fmt.Printf("未知主题：%s\n", name)
		fmt.Println("支持的主题：")
		for _, k := range demoOrder {
			fmt.Printf("  - %s\n", k)
		}
		return
	}

	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("主题：%s\n", name)
	fmt.Println(strings.Repeat("=", 50))
	fn()
	fmt.Println()
}

// ============================================
// 1. 接口与多态 (Interface)
// ============================================
//
// 接口是 Go 实现多态的核心机制。
// 与 Java/C++ 不同，Go 的接口是 **隐式实现** 的：
// 任何类型只要实现了接口中定义的所有方法，就自动实现了该接口。
// 这种方式被称为 "structural typing"（结构化类型），
// 让代码更加灵活、解耦。

// Shape 接口：定义形状的通用行为
type Shape interface {
	Area() float64
	Perimeter() float64
}

// Stringer 接口：约定字符串表示形式
// 注意：fmt 包中已有同名接口，这里命名为 Describer 避免冲突
type Describer interface {
	Describe() string
}

// Circle 圆形
type Circle struct {
	Radius float64
}

// Area 实现 Shape 接口
func (c Circle) Area() float64 {
	return 3.14159 * c.Radius * c.Radius
}

// Perimeter 实现 Shape 接口
func (c Circle) Perimeter() float64 {
	return 2 * 3.14159 * c.Radius
}

// Describe 实现 Describer 接口
func (c Circle) Describe() string {
	return fmt.Sprintf("圆形(半径=%.2f)", c.Radius)
}

// Rectangle 矩形
type Rectangle struct {
	Width, Height float64
}

func (r Rectangle) Area() float64      { return r.Width * r.Height }
func (r Rectangle) Perimeter() float64 { return 2 * (r.Width + r.Height) }
func (r Rectangle) Describe() string {
	return fmt.Sprintf("矩形(宽=%.2f, 高=%.2f)", r.Width, r.Height)
}

// printShapeInfo 多态：传入任何实现了 Shape 的类型都可以
func printShapeInfo(s Shape) {
	fmt.Printf("  面积=%.2f, 周长=%.2f\n", s.Area(), s.Perimeter())
}

func demoInterface() {
	shapes := []Shape{
		Circle{Radius: 3},
		Rectangle{Width: 4, Height: 5},
	}

	// 多态遍历
	for _, s := range shapes {
		// 类型断言：把接口值还原为具体类型
		if d, ok := s.(Describer); ok {
			fmt.Println(d.Describe())
		}
		printShapeInfo(s)
	}

	// 类型 switch：对接口值进行类型判断
	var x interface{} = 42
	switch v := x.(type) {
	case int:
		fmt.Printf("整数：%d (两倍=%d)\n", v, v*2)
	case string:
		fmt.Printf("字符串：%s\n", v)
	case nil:
		fmt.Println("空值")
	default:
		fmt.Printf("未知类型：%T\n", v)
	}

	// 空接口 interface{} 可以承载任意类型
	// Go 1.18+ 推荐使用 any，它是 interface{} 的别名
	var anything any = "hello"
	fmt.Printf("any 持有的值：%v (类型=%T)\n", anything, anything)
}

// ============================================
// 2. 泛型 (Generics, Go 1.18+)
// ============================================
//
// 泛型让我们写出适用于多种类型的函数和数据结构，
// 避免为每种类型重复编写相同逻辑。
// 类型参数写在方括号 [] 中，常见约束有 any、comparable，
// 或者通过类型集合自定义约束。

// Number 数字约束：可以是整数或浮点数
// "~" 表示底层类型为 int/float64 的自定义类型也满足
type Number interface {
	~int | ~int64 | ~float32 | ~float64
}

// Sum 泛型求和函数：对任何 Number 类型的切片求和
func Sum[T Number](nums []T) T {
	var total T
	for _, n := range nums {
		total += n
	}
	return total
}

// Map 泛型 map 函数：对切片中每个元素应用函数 f
func Map[T, U any](items []T, f func(T) U) []U {
	result := make([]U, len(items))
	for i, item := range items {
		result[i] = f(item)
	}
	return result
}

// Filter 泛型 filter 函数：保留满足谓词 pred 的元素
func Filter[T any](items []T, pred func(T) bool) []T {
	result := make([]T, 0, len(items))
	for _, item := range items {
		if pred(item) {
			result = append(result, item)
		}
	}
	return result
}

// Stack 泛型栈：可以存储任意类型的元素
type Stack[T any] struct {
	items []T
}

func (s *Stack[T]) Push(v T) {
	s.items = append(s.items, v)
}

func (s *Stack[T]) Pop() (T, bool) {
	var zero T
	if len(s.items) == 0 {
		return zero, false
	}
	top := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return top, true
}

func (s *Stack[T]) Len() int { return len(s.items) }

func demoGenerics() {
	// 整数求和
	ints := []int{1, 2, 3, 4, 5}
	fmt.Printf("Sum(%v) = %d\n", ints, Sum(ints))

	// 浮点数求和
	floats := []float64{1.1, 2.2, 3.3}
	fmt.Printf("Sum(%v) = %.2f\n", floats, Sum(floats))

	// Map: int -> string
	strs := Map(ints, func(n int) string {
		return fmt.Sprintf("数字%d", n)
	})
	fmt.Printf("Map 结果：%v\n", strs)

	// Filter：保留偶数
	evens := Filter(ints, func(n int) bool { return n%2 == 0 })
	fmt.Printf("偶数：%v\n", evens)

	// 泛型栈
	s := &Stack[string]{}
	s.Push("第一")
	s.Push("第二")
	s.Push("第三")
	fmt.Printf("栈大小：%d\n", s.Len())
	if v, ok := s.Pop(); ok {
		fmt.Printf("出栈：%s\n", v)
	}
}

// ============================================
// 3. 反射 (Reflection)
// ============================================
//
// 反射让程序在运行时检查类型和值，并动态操作它们。
// 主要用于序列化、ORM、依赖注入等通用框架。
// 注意：反射性能较低，且失去类型安全，应在确实需要时才使用。

// User 用于反射演示的结构体
type User struct {
	Name  string `json:"name" validate:"required"`
	Age   int    `json:"age" validate:"min=0,max=150"`
	Email string `json:"email" validate:"email"`
}

// inspect 通过反射打印结构体字段信息
func inspect(v interface{}) {
	t := reflect.TypeOf(v)
	val := reflect.ValueOf(v)

	// 如果传入的是指针，取其指向的值
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		val = val.Elem()
	}

	if t.Kind() != reflect.Struct {
		fmt.Printf("非结构体类型：%s\n", t.Kind())
		return
	}

	fmt.Printf("结构体名：%s（共 %d 个字段）\n", t.Name(), t.NumField())
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := val.Field(i)
		jsonTag := field.Tag.Get("json")
		validateTag := field.Tag.Get("validate")
		fmt.Printf("  字段：%-6s 类型：%-7s 值：%-15v json=%-7s validate=%s\n",
			field.Name, field.Type, value.Interface(), jsonTag, validateTag)
	}
}

// setField 通过反射动态修改结构体字段（必须传入指针）
func setField(obj interface{}, name string, value interface{}) error {
	v := reflect.ValueOf(obj).Elem()
	field := v.FieldByName(name)
	if !field.IsValid() {
		return fmt.Errorf("字段不存在：%s", name)
	}
	if !field.CanSet() {
		return fmt.Errorf("字段不可修改：%s", name)
	}
	field.Set(reflect.ValueOf(value))
	return nil
}

func demoReflect() {
	u := User{Name: "小明", Age: 25, Email: "ming@example.com"}

	// 检查类型和字段
	inspect(u)

	// 动态修改字段
	if err := setField(&u, "Age", 30); err != nil {
		fmt.Println("修改失败：", err)
	}
	fmt.Printf("修改后：%+v\n", u)
}

// ============================================
// 4. 上下文与取消 (Context)
// ============================================
//
// context 包是 Go 中处理超时、取消和请求作用域数据的标准方式。
// 几乎所有处理外部请求 / IO 的函数都应接受 context.Context 作为第一个参数。
// 通过 ctx.Done() 通道感知取消信号，通过 ctx.Err() 获取取消原因。

// slowOperation 模拟一个耗时操作，支持通过 context 取消
func slowOperation(ctx context.Context, name string) error {
	select {
	case <-time.After(2 * time.Second):
		fmt.Printf("  [%s] 操作完成\n", name)
		return nil
	case <-ctx.Done():
		// 上下文被取消（超时或主动取消）
		fmt.Printf("  [%s] 操作被取消：%v\n", name, ctx.Err())
		return ctx.Err()
	}
}

func demoContext() {
	// 1. 超时控制：500ms 后自动取消
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel() // 一定要调用 cancel 释放资源
	_ = slowOperation(ctx, "超时演示")

	// 2. 手动取消
	ctx2, cancel2 := context.WithCancel(context.Background())
	go func() {
		time.Sleep(200 * time.Millisecond)
		cancel2() // 200ms 后主动取消
	}()
	_ = slowOperation(ctx2, "手动取消")

	// 3. 携带请求作用域的值（应仅用于横切关注点，如 traceID）
	type keyType string
	const reqIDKey keyType = "requestID"
	ctx3 := context.WithValue(context.Background(), reqIDKey, "req-12345")
	if id, ok := ctx3.Value(reqIDKey).(string); ok {
		fmt.Printf("  携带的 requestID = %s\n", id)
	}
}

// ============================================
// 5. 错误处理 (Errors)
// ============================================
//
// Go 把错误作为值（value）传递，而不是异常。
// Go 1.13+ 引入了错误包装、errors.Is、errors.As，
// 让我们能在保留错误链的同时进行类型 / 值匹配。

// ErrNotFound 哨兵错误：用作可比较的错误值
var ErrNotFound = errors.New("资源未找到")

// ValidationError 自定义错误类型：携带额外字段信息
type ValidationError struct {
	Field   string
	Message string
}

// Error 实现 error 接口
func (e *ValidationError) Error() string {
	return fmt.Sprintf("字段 %s 校验失败：%s", e.Field, e.Message)
}

// findUser 模拟查找用户，可能返回 ErrNotFound
func findUser(id int) (*User, error) {
	if id == 0 {
		return nil, fmt.Errorf("findUser(%d): %w", id, ErrNotFound) // 包装错误
	}
	if id < 0 {
		return nil, &ValidationError{Field: "id", Message: "必须为正数"}
	}
	return &User{Name: "测试用户", Age: 20}, nil
}

func demoErrors() {
	// 1. errors.Is：判断错误链中是否包含某个哨兵错误
	_, err := findUser(0)
	if errors.Is(err, ErrNotFound) {
		fmt.Println("  errors.Is 匹配：用户不存在")
	}

	// 2. errors.As：把错误链中的某个具体类型提取出来
	_, err = findUser(-1)
	var ve *ValidationError
	if errors.As(err, &ve) {
		fmt.Printf("  errors.As 匹配：字段=%s 错误=%s\n", ve.Field, ve.Message)
	}

	// 3. 错误包装
	wrapped := fmt.Errorf("处理请求失败: %w", ErrNotFound)
	fmt.Printf("  包装后的错误：%v\n", wrapped)
	fmt.Printf("  Unwrap 后：%v\n", errors.Unwrap(wrapped))
}

// ============================================
// 6. 高级并发模式 (Concurrent)
// ============================================
//
// Go 的并发原语主要是 goroutine + channel。
// 这里演示几种常见模式：扇出/扇入、select 多路复用、超时控制。
// 记住口号："Don't communicate by sharing memory; share memory by communicating."

// pipeline 生成 1..n 的数字
func generateNumbers(n int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for i := 1; i <= n; i++ {
			out <- i
		}
	}()
	return out
}

// square 流水线阶段：对输入通道的每个数取平方
func square(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			out <- n * n
		}
	}()
	return out
}

// fanIn 扇入：把多个通道合并为一个
func fanIn(channels ...<-chan int) <-chan int {
	out := make(chan int)
	var wg sync.WaitGroup
	for _, ch := range channels {
		wg.Add(1)
		go func(c <-chan int) {
			defer wg.Done()
			for v := range c {
				out <- v
			}
		}(ch)
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func demoConcurrent() {
	// 1. 流水线（pipeline）模式
	nums := generateNumbers(5)
	squared := square(nums)
	var collected []int
	for v := range squared {
		collected = append(collected, v)
	}
	fmt.Printf("  流水线结果：%v\n", collected)

	// 2. 扇入：合并多个通道
	a := generateNumbers(3)
	b := generateNumbers(3)
	merged := fanIn(a, b)
	count := 0
	for range merged {
		count++
	}
	fmt.Printf("  扇入收到 %d 个元素\n", count)

	// 3. select + 超时
	ch := make(chan string)
	go func() {
		time.Sleep(300 * time.Millisecond)
		ch <- "结果"
	}()
	select {
	case v := <-ch:
		fmt.Printf("  收到：%s\n", v)
	case <-time.After(100 * time.Millisecond):
		fmt.Println("  超时（100ms 内未收到结果）")
	}
}

// ============================================
// 7. defer / panic / recover
// ============================================
//
// defer 用于推迟执行（常用于资源清理）。
// panic 用于报告不可恢复的错误（如空指针、越界）。
// recover 必须在 defer 中调用，用于拦截 panic、避免程序崩溃。
// 经验法则：在 Go 中，常规错误用 error 传递，panic 仅用于真正异常的情况。

// safeDivide 安全除法：捕获 panic 并以 error 返回
func safeDivide(a, b int) (result int, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("捕获到 panic：%v", r)
		}
	}()
	return a / b, nil // 当 b=0 时会触发运行时 panic
}

func demoDefer() {
	// 1. defer 的执行顺序：后进先出（LIFO）
	fmt.Println("  defer 顺序演示：")
	func() {
		defer fmt.Println("    defer 1（最先注册，最后执行）")
		defer fmt.Println("    defer 2")
		defer fmt.Println("    defer 3（最后注册，最先执行）")
		fmt.Println("    函数体执行")
	}()

	// 2. defer 与参数求值：参数在 defer 语句执行时立即求值
	i := 10
	func() {
		defer fmt.Printf("  defer 看到的 i = %d（立即求值）\n", i)
		i = 20
		fmt.Printf("  函数内 i 被改为 %d\n", i)
	}()

	// 3. recover 拦截 panic
	if _, err := safeDivide(10, 0); err != nil {
		fmt.Printf("  safeDivide 错误：%v\n", err)
	}
	if v, err := safeDivide(10, 2); err == nil {
		fmt.Printf("  safeDivide(10,2) = %d\n", v)
	}
}

// ============================================
// 8. 结构体嵌入 (Embedding)
// ============================================
//
// Go 没有传统意义上的 "继承"，而是通过 **组合 / 嵌入** 来复用代码。
// 内层类型的字段与方法会被外层类型 "提升"，可直接访问。
// 这是 Go 实现 "is-a" 关系的惯用方式。

// Animal 基础类型
type Animal struct {
	Name string
	Age  int
}

func (a Animal) Greet() string {
	return fmt.Sprintf("我是 %s，今年 %d 岁", a.Name, a.Age)
}

// Dog 通过嵌入 Animal 获得其字段与方法
type Dog struct {
	Animal // 匿名字段：嵌入
	Breed  string
}

// Dog 可以覆盖 Animal 的方法
func (d Dog) Greet() string {
	return fmt.Sprintf("汪！%s（%s 品种）", d.Animal.Greet(), d.Breed)
}

// ReadWriter 接口嵌入：把多个接口组合成一个
type Reader interface {
	Read() string
}
type Writer interface {
	Write(s string)
}
type ReadWriter interface {
	Reader
	Writer
}

// memBuffer 同时实现 Reader 与 Writer
type memBuffer struct {
	data string
}

func (m *memBuffer) Read() string   { return m.data }
func (m *memBuffer) Write(s string) { m.data = s }

func demoEmbed() {
	d := Dog{
		Animal: Animal{Name: "旺财", Age: 3},
		Breed:  "金毛",
	}
	// 直接访问嵌入字段的字段
	fmt.Printf("  名字：%s\n", d.Name)
	// 调用被覆盖的方法
	fmt.Printf("  打招呼：%s\n", d.Greet())

	// 接口嵌入示例
	var rw ReadWriter = &memBuffer{}
	rw.Write("Hello, Go!")
	fmt.Printf("  ReadWriter 读到：%s\n", rw.Read())
}

// ============================================
// 9. 闭包与高阶函数 (Closure)
// ============================================
//
// 函数在 Go 中是一等公民：可以作为参数、返回值，也可以赋值给变量。
// 闭包是 "捕获了外层变量" 的函数，可以维持自身状态。

// counter 返回一个闭包，每次调用计数器加 1
func counter() func() int {
	count := 0
	return func() int {
		count++
		return count
	}
}

// applyTwice 高阶函数：把函数 f 对 x 应用两次
func applyTwice(f func(int) int, x int) int {
	return f(f(x))
}

func demoClosure() {
	// 闭包维持状态
	c1 := counter()
	c2 := counter() // 与 c1 互相独立
	fmt.Printf("  c1: %d %d %d\n", c1(), c1(), c1())
	fmt.Printf("  c2: %d %d\n", c2(), c2())

	// 高阶函数 + 匿名函数
	double := func(x int) int { return x * 2 }
	fmt.Printf("  applyTwice(double, 5) = %d\n", applyTwice(double, 5))

	// 闭包捕获循环变量的常见陷阱：
	// Go 1.22 之前需要显式拷贝变量；Go 1.22+ 已自动为每次迭代创建新变量
	fns := make([]func() int, 0, 3)
	for i := 1; i <= 3; i++ {
		i := i // 显式拷贝，兼容老版本
		fns = append(fns, func() int { return i })
	}
	for _, f := range fns {
		fmt.Printf("  捕获的 i = %d\n", f())
	}
}

// ============================================
// 10. 函数选项模式 (Functional Options)
// ============================================
//
// 函数选项模式用一组 "Option" 函数来构造对象，
// 避免构造函数参数爆炸、同时保持向后兼容。
// 是 Go 社区中非常流行的 API 设计风格（grpc、zap、kafka-go 等都使用）。

// Server 一个示例服务器配置
type Server struct {
	Host    string
	Port    int
	Timeout time.Duration
	TLS     bool
}

// Option 选项函数类型
type Option func(*Server)

// WithPort 设置端口
func WithPort(p int) Option {
	return func(s *Server) { s.Port = p }
}

// WithTimeout 设置超时
func WithTimeout(d time.Duration) Option {
	return func(s *Server) { s.Timeout = d }
}

// WithTLS 启用 TLS
func WithTLS() Option {
	return func(s *Server) { s.TLS = true }
}

// NewServer 使用函数选项创建服务器
// 必传字段（host）通过普通参数，可选字段通过 opts
func NewServer(host string, opts ...Option) *Server {
	// 默认值
	s := &Server{
		Host:    host,
		Port:    8080,
		Timeout: 30 * time.Second,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func demoOptions() {
	// 使用默认值
	s1 := NewServer("localhost")
	fmt.Printf("  默认配置：%+v\n", s1)

	// 按需覆盖部分配置
	s2 := NewServer(
		"example.com",
		WithPort(443),
		WithTimeout(10*time.Second),
		WithTLS(),
	)
	fmt.Printf("  自定义配置：%+v\n", s2)
}

// ============================================
// 练习题（思考题）
// ============================================
//
// 1. 给 Stack[T] 添加 Peek 方法（查看栈顶但不出栈）。
// 2. 给 Filter 写一个 "FilterMap" 版本：同时进行过滤和映射。
// 3. 实现一个带最大并发数限制的 "worker pool"，使用 channel 控制信号量。
// 4. 实现 mustGet[T any](m map[string]T, key string) T，
//    若 key 不存在则 panic，并用 recover 编写对应的安全版本。
// 5. 用反射实现一个简单的 JSON 校验器：
//    根据 struct tag 中的 "validate" 规则检查字段是否合法。
