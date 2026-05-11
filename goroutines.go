// Goroutines
// Go 协程（goroutine）专题教程 - 由浅入深，覆盖常用并发原语与典型模式
// 学习目标：goroutine 启动与生命周期、channel、select、sync 包、context、
//          常见并发模式（worker pool / pipeline / fan-out / fan-in / semaphore），
//          以及竞态、死锁、协程泄漏等常见陷阱及其规避方法。
//
// 运行方式：
//   # 运行全部主题
//   go run goroutines.go
//
//   # 只看某个主题
//   go run goroutines.go -topic basic
//   go run goroutines.go -topic channels
//   go run goroutines.go -topic patterns
//
// 支持的主题 (topic)：
//   basic      - goroutine 基础：启动、参数、调度、GOMAXPROCS
//   channels   - channel：无缓冲、带缓冲、单向、关闭、range
//   select     - select 多路复用：阻塞 / 非阻塞 / 超时 / done
//   waitgroup  - sync.WaitGroup：等待一组协程结束
//   mutex      - sync.Mutex / sync.RWMutex：共享数据保护
//   once       - sync.Once：单次初始化
//   atomic     - sync/atomic：无锁计数与状态
//   context    - context：取消、超时、传值
//   patterns   - 高阶模式：worker pool / pipeline / fan-out / fan-in / semaphore / generator
//   pitfalls   - 常见陷阱：协程泄漏、循环变量捕获、死锁、竞态
//   all        - 运行所有示例（默认）
//
// @author 初学者学习项目

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// ============================================
// 入口函数
// ============================================

// 演示函数表，主题名 -> 演示函数
var goroutineDemos = map[string]func(){
	"basic":     demoGoroutineBasic,
	"channels":  demoChannels,
	"select":    demoSelect,
	"waitgroup": demoWaitGroup,
	"mutex":     demoMutex,
	"once":      demoOnce,
	"atomic":    demoAtomic,
	"context":   demoContextCancel,
	"patterns":  demoPatterns,
	"pitfalls":  demoPitfalls,
}

// 推荐运行顺序：先掌握基础原语，再看模式，最后看陷阱
var goroutineOrder = []string{
	"basic",
	"channels",
	"select",
	"waitgroup",
	"mutex",
	"once",
	"atomic",
	"context",
	"patterns",
	"pitfalls",
}

func main() {
	topic := flag.String("topic", "all", "要演示的协程主题")
	flag.Parse()

	fmt.Println("====================================")
	fmt.Println("   Go 协程（Goroutine）专题 v1.0")
	fmt.Println("====================================")
	fmt.Println()

	if *topic == "all" {
		for _, name := range goroutineOrder {
			runGoroutineDemo(name)
		}
		return
	}

	runGoroutineDemo(*topic)
}

func runGoroutineDemo(name string) {
	fn, ok := goroutineDemos[name]
	if !ok {
		fmt.Printf("未知主题：%s\n", name)
		fmt.Println("支持的主题：")
		for _, k := range goroutineOrder {
			fmt.Printf("  - %s\n", k)
		}
		return
	}
	fmt.Printf(">>> 主题：%s\n", name)
	fn()
	fmt.Println()
}

// ============================================
// 1. goroutine 基础 (basic)
// ============================================
//
// goroutine 是 Go 运行时调度的轻量级执行单元（不是操作系统线程）。
// 启动一个 goroutine 只需要在函数调用前加 `go` 关键字。
//
// 关键认知：
//   - goroutine 启动开销很小（栈初始 2KB，按需增长），单机开数十万都很正常。
//   - main 函数返回，整个程序就会结束，**不会等待**未完成的 goroutine。
//   - GOMAXPROCS 控制并行执行 goroutine 的 OS 线程数（默认 = CPU 核心数）。

func sayHello(name string, delay time.Duration) {
	time.Sleep(delay)
	fmt.Printf("  你好，我是 %s\n", name)
}

func demoGoroutineBasic() {
	fmt.Printf("  当前 CPU 核心数：%d，GOMAXPROCS=%d\n", runtime.NumCPU(), runtime.GOMAXPROCS(0))
	fmt.Printf("  当前 goroutine 数（演示前）：%d\n", runtime.NumGoroutine())

	// 1. 启动一个匿名 goroutine
	go func() {
		fmt.Println("  [匿名 goroutine] 我并发执行")
	}()

	// 2. 启动多个具名函数 goroutine
	go sayHello("Alice", 20*time.Millisecond)
	go sayHello("Bob", 10*time.Millisecond)
	go sayHello("Carol", 30*time.Millisecond)

	// 3. 错误示范：直接退出而不等待，goroutine 可能根本没机会执行
	//    所以这里我们简单 sleep 一下；真实项目应使用 sync.WaitGroup 或 channel 同步。
	time.Sleep(100 * time.Millisecond)

	fmt.Printf("  当前 goroutine 数（sleep 后）：%d\n", runtime.NumGoroutine())

	// 4. 参数求值时机
	//    `go f(x)` 中的 x 是在 go 语句执行时立即求值，而不是在 goroutine 运行时求值
	x := 1
	go func(v int) {
		fmt.Printf("  传值 v=%d（go 语句执行时已固定）\n", v)
	}(x)
	x = 2
	time.Sleep(20 * time.Millisecond)
}

// ============================================
// 2. channel（通道）
// ============================================
//
// channel 是 goroutine 之间最常用的通信方式。
// 谚语："不要通过共享内存来通信，要通过通信来共享内存。"
//
// 三种 channel：
//   - 无缓冲：发送和接收必须同时就绪（同步交接）
//   - 带缓冲：缓冲区未满可发，未空可收（异步队列）
//   - 单向：只读或只写，常用于函数签名以表达意图
//
// 关闭规则：
//   - 只有发送方应该关闭 channel
//   - 不要重复关闭、不要向已关闭的 channel 发送（会 panic）
//   - 从已关闭且空的 channel 接收会立即返回零值；用 `v, ok := <-ch` 区分

func produceInts(out chan<- int, n int) {
	for i := 1; i <= n; i++ {
		out <- i
	}
	close(out) // 发送方负责 close
}

func demoChannels() {
	// 1. 无缓冲 channel：握手式通信
	ping := make(chan string)
	go func() {
		ping <- "ping" // 发送方阻塞，直到有人接收
	}()
	fmt.Printf("  无缓冲 channel 收到：%s\n", <-ping)

	// 2. 带缓冲 channel：可异步发送 N 条消息再阻塞
	buf := make(chan int, 3)
	buf <- 1
	buf <- 2
	buf <- 3
	fmt.Printf("  缓冲 channel 长度=%d 容量=%d\n", len(buf), cap(buf))
	close(buf)
	// 即使关闭，已入队的数据仍可正常读出
	for v := range buf {
		fmt.Printf("  从缓冲 channel 读出：%d\n", v)
	}

	// 3. 单向 channel + range：典型生产者-消费者
	out := make(chan int)
	go produceInts(out, 5)
	sum := 0
	for v := range out { // range 自动在 channel 关闭后退出
		sum += v
	}
	fmt.Printf("  range 累加结果：%d\n", sum)

	// 4. 用 v, ok 判断 channel 是否已关闭
	done := make(chan struct{})
	close(done)
	if v, ok := <-done; !ok {
		fmt.Printf("  done 已关闭，接收到零值：%v\n", v)
	}

	// 5. 用 channel 做信号量（轻量同步）
	finished := make(chan struct{})
	go func() {
		time.Sleep(20 * time.Millisecond)
		close(finished) // 关闭等价于"广播"信号
	}()
	<-finished
	fmt.Println("  收到完成信号")
}

// ============================================
// 3. select 多路复用
// ============================================
//
// select 像 channel 版的 switch：在多个 channel 操作中选一个就绪的执行。
// 常见用法：超时控制、退出信号、非阻塞收发、多 channel 合并。
//
// 注意：所有 case 都没就绪且没有 default 时，select 会一直阻塞。
//      所有 case 都没就绪但有 default 时，立即执行 default（非阻塞）。

func demoSelect() {
	// 1. 多 channel 等谁先来
	a := make(chan string)
	b := make(chan string)
	go func() { time.Sleep(30 * time.Millisecond); a <- "来自 A" }()
	go func() { time.Sleep(10 * time.Millisecond); b <- "来自 B" }()

	for i := 0; i < 2; i++ {
		select {
		case msg := <-a:
			fmt.Printf("  收到 A：%s\n", msg)
		case msg := <-b:
			fmt.Printf("  收到 B：%s\n", msg)
		}
	}

	// 2. 超时控制：避免被无响应的 channel 永久阻塞
	slow := make(chan string)
	go func() {
		time.Sleep(80 * time.Millisecond)
		slow <- "结果"
	}()
	select {
	case v := <-slow:
		fmt.Printf("  按时收到：%s\n", v)
	case <-time.After(20 * time.Millisecond):
		fmt.Println("  超时：20ms 内未收到结果")
	}

	// 3. 非阻塞收发：default 让 select 立即返回
	c := make(chan int, 1)
	select {
	case v := <-c:
		fmt.Printf("  非阻塞收到 %d\n", v)
	default:
		fmt.Println("  channel 空，非阻塞 select 走 default")
	}

	c <- 99
	select {
	case c <- 100:
		fmt.Println("  非阻塞发送成功")
	default:
		fmt.Println("  channel 满，非阻塞发送被丢弃")
	}

	// 4. done channel 退出循环
	done := make(chan struct{})
	tick := time.NewTicker(15 * time.Millisecond)
	defer tick.Stop()
	go func() {
		time.Sleep(45 * time.Millisecond)
		close(done)
	}()
loop:
	for {
		select {
		case <-tick.C:
			fmt.Println("  tick...")
		case <-done:
			fmt.Println("  done，跳出循环")
			break loop
		}
	}
}

// ============================================
// 4. sync.WaitGroup
// ============================================
//
// WaitGroup 用于等待一组 goroutine 结束。
// 三个方法：
//   Add(n)  : 等待计数 +n（必须在 goroutine 启动前调用）
//   Done()  : 等待计数 -1（goroutine 退出时调用，通常配合 defer）
//   Wait()  : 阻塞直到计数归零
//
// 常见误用：在 goroutine 内部调用 Add，导致 Wait 提前返回。

func demoWaitGroup() {
	var wg sync.WaitGroup
	const workers = 5

	for i := 1; i <= workers; i++ {
		wg.Add(1) // 必须在 go 之前
		go func(id int) {
			defer wg.Done()
			d := time.Duration(20+rand.Intn(30)) * time.Millisecond
			time.Sleep(d)
			fmt.Printf("  worker %d 完成（用时 %v）\n", id, d)
		}(i) // 传值，避免循环变量捕获问题
	}

	wg.Wait() // 阻塞直到全部 Done
	fmt.Println("  全部 worker 已完成")
}

// ============================================
// 5. sync.Mutex / sync.RWMutex
// ============================================
//
// Mutex 互斥锁：同一时刻只允许一个 goroutine 持有。
// RWMutex 读写锁：多个读可并发，写独占。
//
// 经验：能用 channel 解决就用 channel；要保护一段共享数据时再上锁。
//      读多写少且临界区有一定开销的场景，RWMutex 收益最明显。

type counter struct {
	mu sync.Mutex
	n  int
}

func (c *counter) inc() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.n++
}

func (c *counter) value() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.n
}

type cache struct {
	mu   sync.RWMutex
	data map[string]string
}

func (c *cache) get(k string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.data[k]
	return v, ok
}

func (c *cache) set(k, v string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[k] = v
}

func demoMutex() {
	// Mutex：1000 个 goroutine 各自 +1，结果应稳定为 1000
	var wg sync.WaitGroup
	cnt := &counter{}
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cnt.inc()
		}()
	}
	wg.Wait()
	fmt.Printf("  Mutex 保护下计数 = %d（期望 1000）\n", cnt.value())

	// RWMutex：写少读多
	cc := &cache{data: map[string]string{"hello": "world"}}
	cc.set("go", "fast")
	for _, k := range []string{"hello", "go", "missing"} {
		if v, ok := cc.get(k); ok {
			fmt.Printf("  cache[%s] = %s\n", k, v)
		} else {
			fmt.Printf("  cache[%s] 未命中\n", k)
		}
	}
}

// ============================================
// 6. sync.Once
// ============================================
//
// sync.Once.Do(f) 保证 f 在程序生命周期内最多被执行一次，且对调用者可见。
// 典型场景：单例、惰性初始化、只需做一次的资源准备。

var (
	onceConfig sync.Once
	configVal  map[string]string
)

func loadConfig() map[string]string {
	onceConfig.Do(func() {
		fmt.Println("  loadConfig 真正执行了一次")
		configVal = map[string]string{"env": "prod", "version": "1.0"}
	})
	return configVal
}

func demoOnce() {
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c := loadConfig()
			_ = c["env"]
		}()
	}
	wg.Wait()
	fmt.Printf("  最终配置：%v\n", loadConfig())
}

// ============================================
// 7. sync/atomic
// ============================================
//
// atomic 包提供"无锁"原子操作（CAS、Add、Load、Store、Swap 等）。
// 适合简单的计数 / 状态标志位场景；比 Mutex 更轻量。
// 注意：原子只保证单一操作的原子性，不能替代 Mutex 保护多步骤的复合操作。

func demoAtomic() {
	var counter int64
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			atomic.AddInt64(&counter, 1)
		}()
	}
	wg.Wait()
	fmt.Printf("  atomic 计数 = %d\n", atomic.LoadInt64(&counter))

	// CompareAndSwap：实现一次性"赢家通吃"
	var winner int32
	var contestWG sync.WaitGroup
	for i := 1; i <= 5; i++ {
		contestWG.Add(1)
		go func(id int32) {
			defer contestWG.Done()
			// 谁先把 winner 从 0 改到 id，谁就赢
			atomic.CompareAndSwapInt32(&winner, 0, id)
		}(int32(i))
	}
	contestWG.Wait()
	fmt.Printf("  CAS 抢占赢家 = %d\n", atomic.LoadInt32(&winner))
}

// ============================================
// 8. context：取消、超时、传值
// ============================================
//
// context.Context 用来在 goroutine 之间传递截止时间、取消信号和请求作用域的数据。
// 建议：所有阻塞 / 可能长时间运行的函数，第一个参数都接 ctx context.Context。
//
// 三种常见派生方式：
//   WithCancel       手动 cancel
//   WithTimeout      到时间自动 cancel
//   WithDeadline     到某个时刻自动 cancel
//   WithValue        传递请求级数据（不要塞配置/可选参数！）

func worker(ctx context.Context, id int) error {
	for {
		select {
		case <-ctx.Done():
			// ctx.Err() == context.Canceled 或 DeadlineExceeded
			return ctx.Err()
		case <-time.After(15 * time.Millisecond):
			fmt.Printf("  worker %d 正在干活...\n", id)
		}
	}
}

func demoContextCancel() {
	// 1. 手动取消
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		err := worker(ctx, 1)
		fmt.Printf("  worker 1 退出：%v\n", err)
	}()
	time.Sleep(50 * time.Millisecond)
	cancel() // 通知 worker 退出
	time.Sleep(10 * time.Millisecond)

	// 2. 带超时的 context（请求级最常用）
	ctx2, cancel2 := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel2() // 即便已超时，也应主动 cancel 释放资源
	err := worker(ctx2, 2)
	fmt.Printf("  worker 2 退出：%v\n", err)

	// 3. WithValue 传值（注意：仅用于请求作用域的"元数据"）
	type reqIDKey struct{}
	ctx3 := context.WithValue(context.Background(), reqIDKey{}, "req-001")
	id, _ := ctx3.Value(reqIDKey{}).(string)
	fmt.Printf("  请求 ID = %s\n", id)
}

// ============================================
// 9. 常见并发模式 (patterns)
// ============================================

// 9.1 Worker Pool：固定数量的工作协程从同一个任务通道消费
func workerPool(tasks <-chan int, results chan<- int, workers int) {
	var wg sync.WaitGroup
	for w := 1; w <= workers; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for t := range tasks {
				// 模拟处理
				time.Sleep(5 * time.Millisecond)
				results <- t * t
			}
		}(w)
	}
	go func() {
		wg.Wait()
		close(results)
	}()
}

// 9.2 Pipeline 阶段：generate -> double -> filter
func gen(nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, n := range nums {
			out <- n
		}
	}()
	return out
}

func double(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			out <- n * 2
		}
	}()
	return out
}

func evenOnly(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			if n%2 == 0 {
				out <- n
			}
		}
	}()
	return out
}

// 9.3 Fan-out / Fan-in：多个 worker 并行处理同一来源，再合并结果
func fanInInts(channels ...<-chan int) <-chan int {
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

func demoPatterns() {
	// (A) Worker Pool
	tasks := make(chan int, 10)
	results := make(chan int, 10)
	workerPool(tasks, results, 3)
	go func() {
		for i := 1; i <= 6; i++ {
			tasks <- i
		}
		close(tasks)
	}()
	var collected []int
	for r := range results {
		collected = append(collected, r)
	}
	fmt.Printf("  Worker Pool 结果（顺序不定，共 %d 个）：%v\n", len(collected), collected)

	// (B) Pipeline
	pipeOut := evenOnly(double(gen(1, 2, 3, 4, 5)))
	var pipeRes []int
	for v := range pipeOut {
		pipeRes = append(pipeRes, v)
	}
	fmt.Printf("  Pipeline 结果（×2 后取偶数）：%v\n", pipeRes)

	// (C) Fan-out + Fan-in
	src := gen(1, 2, 3, 4, 5, 6, 7, 8)
	// 把 src 拆给两个 double，做"扇出"
	d1 := double(src)
	d2 := double(src) // 注意：两个消费者共享同一个 in，会均分（不可靠的演示，仅说明原理）
	merged := fanInInts(d1, d2)
	count := 0
	for range merged {
		count++
	}
	fmt.Printf("  Fan-out/Fan-in 收到 %d 个元素\n", count)

	// (D) 用 channel 做信号量（限制并发数）
	sem := make(chan struct{}, 2) // 同时最多 2 个 goroutine 进入临界区
	var wg sync.WaitGroup
	for i := 1; i <= 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			sem <- struct{}{} // 获取许可
			defer func() { <-sem }()
			time.Sleep(20 * time.Millisecond)
			fmt.Printf("  semaphore 任务 %d 完成\n", id)
		}(i)
	}
	wg.Wait()
}

// ============================================
// 10. 常见陷阱 (pitfalls)
// ============================================

// 10.1 协程泄漏：发送方阻塞却没人接收，goroutine 永远挂着
func leakyReceiver() {
	// 错误示范（用注释保留思路；这里不真的让它泄漏）
	//   ch := make(chan int)
	//   go func() { ch <- 1 }() // 没有接收者 -> 永久阻塞
	//
	// 正确做法：用带 ctx 的发送，或者保证接收方一定存在
	ch := make(chan int, 1) // 缓冲容量 1 让发送方可以立即返回
	go func() { ch <- 1 }()
	v := <-ch
	fmt.Printf("  避免泄漏：通过缓冲让发送方可退出，读到 %d\n", v)
}

// 10.2 循环变量捕获：Go 1.22 之前必须显式复制
func loopCapturePitfall() {
	fns := make([]func() int, 0, 3)
	for i := 1; i <= 3; i++ {
		i := i // 关键：每次迭代为闭包创建独立的副本（Go 1.22+ 已经默认这样做）
		fns = append(fns, func() int { return i })
	}
	for _, f := range fns {
		fmt.Printf("  捕获到的 i = %d\n", f())
	}
}

// 10.3 死锁：所有 goroutine 都在等彼此 -> runtime 报 "all goroutines are asleep"
//      演示中我们 *不会* 真的触发死锁，而是用 select+default 模拟成 "差点死锁"
func deadlockNearMiss() {
	ch := make(chan int) // 无缓冲
	select {
	case ch <- 1:
		fmt.Println("  发送成功（不会到达，因为没有接收者）")
	default:
		fmt.Println("  没人接收 -> 走 default，避免阻塞（在生产代码里要警惕这种情况）")
	}
}

// 10.4 错误地不带锁地共享 map：会触发 fatal "concurrent map writes"
//      演示中用 sync.Mutex 修正
func sharedMapSafe() {
	var mu sync.Mutex
	m := make(map[int]int)
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(k int) {
			defer wg.Done()
			mu.Lock()
			m[k] = k * k
			mu.Unlock()
		}(i)
	}
	wg.Wait()
	fmt.Printf("  并发写入 map（加锁后）大小 = %d\n", len(m))
}

// 10.5 错误用法：在 goroutine 内才 wg.Add(1)，Wait 可能提前返回
//      演示中用正确写法对比
func waitGroupCorrect() {
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1) // ✅ 在 go 之前
		go func(id int) {
			defer wg.Done()
			_ = id
		}(i)
	}
	wg.Wait()
	fmt.Println("  WaitGroup 正确用法：Add 在 go 之前调用")
}

// 10.6 errgroup 风格的错误收集（手写简化版）
//      实际项目推荐 "golang.org/x/sync/errgroup"
type errGroup struct {
	wg   sync.WaitGroup
	once sync.Once
	err  error
}

func (g *errGroup) Go(f func() error) {
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		if err := f(); err != nil {
			g.once.Do(func() { g.err = err })
		}
	}()
}

func (g *errGroup) Wait() error {
	g.wg.Wait()
	return g.err
}

func demoPitfalls() {
	leakyReceiver()
	loopCapturePitfall()
	deadlockNearMiss()
	sharedMapSafe()
	waitGroupCorrect()

	// errgroup 演示：多个 goroutine 同时跑，最先失败的那个错误被保留下来，
	// 其余错误被 sync.Once 忽略；最终 Wait() 一次性返回。
	var eg errGroup
	eg.Go(func() error { time.Sleep(5 * time.Millisecond); return nil })
	eg.Go(func() error { time.Sleep(15 * time.Millisecond); return errors.New("最先失败的任务") })
	eg.Go(func() error { time.Sleep(30 * time.Millisecond); return errors.New("更晚失败的任务（被忽略）") })
	if err := eg.Wait(); err != nil {
		fmt.Printf("  errGroup 保留的错误：%v\n", err)
	}
}

// ============================================
// 知识点小抄 (Cheat Sheet)
// ============================================
//
// 1. 启动协程： go f(args...)
// 2. 同步等待： sync.WaitGroup{Add, Done, Wait}
// 3. 通信：     ch := make(chan T) / make(chan T, N)，<-ch 读，ch<-v 写
// 4. 关闭：     close(ch) 只由发送方调用；v, ok := <-ch 检测关闭
// 5. 多路：     select { case ...: ; default: }
// 6. 超时：     time.After / context.WithTimeout
// 7. 取消：     context.WithCancel；尊重 <-ctx.Done()
// 8. 互斥：     sync.Mutex / sync.RWMutex；defer Unlock 几乎必备
// 9. 原子：     sync/atomic（适合简单计数 / 状态位）
// 10. 单次：    sync.Once.Do(fn)
//
// 调试与排错：
//   - 跑测试时加 -race 检测数据竞争： go test -race ./...
//   - runtime.NumGoroutine() 观察协程数变化，怀疑协程泄漏时有用
//   - 死锁运行时通常会打印 "fatal error: all goroutines are asleep - deadlock!"

// ============================================
// 练习题 (Exercises)
// ============================================
//
// 1. 实现一个固定大小的 worker pool，可以通过 context 中途取消，
//    取消后正在执行的任务允许跑完，但不再领新任务。
//
// 2. 写一个 Rate Limiter：每秒最多允许 N 次操作，超出的调用阻塞等待。
//    （提示：time.Ticker 或令牌桶 + 带缓冲 channel）
//
// 3. 编写一个 Pub/Sub：一个发布者通过 channel 广播消息给 M 个订阅者，
//    且任何一个订阅者堵塞都不能影响其他订阅者。
//
// 4. 用 sync.Cond 实现"有界缓冲"：生产者-消费者模型，缓冲满时生产者等待，
//    空时消费者等待。
//
// 5. 给定一个 URL 列表，并发抓取每个 URL，最多保留 3 个并发，
//    任一个出错则取消其余请求并返回首个错误。
//    （提示：context + semaphore + errgroup）
