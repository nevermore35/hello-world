// URL Checker
// 一个适合 Go 语言初学者的并发 URL 检查器
// 学习目标：Go 并发编程、goroutine、channel、WaitGroup、HTTP 请求
//
// 运行方式：
//   go run url_checker.go
//   go run url_checker.go -urls https://google.com,https://github.com,https://golang.org
//   go run url_checker.go -timeout 10
//
// @author 初学者学习项目

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// 常量定义
// ============================================

const (
	// DefaultTimeout 默认超时时间（秒）
	DefaultTimeout = 5
	// DefaultWorkers 默认工作协程数
	DefaultWorkers = 5
	// UserAgent 用户代理
	UserAgent = "Go-URL-Checker/1.0"
)

// 全局配置
// ============================================

var (
	// timeout 超时时间
	timeout = DefaultTimeout * time.Second
	// maxWorkers 最大并发数
	maxWorkers = DefaultWorkers
)

// 结构体定义
// ============================================

// URLResult URL 检查结果
// 每个 URL 的检查结果存储在这里
type URLResult struct {
	URL       string    // URL 地址
	Status   string    // 状态：OK, ERROR, TIMEOUT
	StatusCode int     // HTTP 状态码
	Latency   int64    // 响应延迟（毫秒）
	Timestamp time.Time // 检查时间
	Error     string   // 错误信息（如果有）
}

// URLChecker URL 检查器
// 包含所有 URLs 的检查结果
type URLChecker struct {
	Results []URLResult // 检查结果列表
	mu      sync.Mutex  // 互斥锁（保护 Results）
	
	// 统计信息
	TotalURLs    int32 // 总 URL 数
	CheckedURLs int32 // 已检查数
	SuccessNum  int32 // 成功数
	FailedNum   int32 // 失败数
}

// NewURLChecker 创建新的 URL 检查器
func NewURLChecker() *URLChecker {
	return &URLChecker{
		Results: make([]URLResult, 0),
	}
}

// AddResult 添加结果（线程安全）
func (uc *URLChecker) AddResult(result URLResult) {
	uc.mu.Lock()
	uc.Results = append(uc.Results, result)
	uc.mu.Unlock()
}

// Job 工作单元
// 用于在协程间传递任务
type Job struct {
	URL   string
	Queue chan URLResult // 结果队列
}

// 函数定义
// ============================================

// main 主函数
func main() {
	fmt.Println("==================================")
	fmt.Println("      Go 并发 URL 检查器 v1.0")
	fmt.Println("==================================")
	fmt.Println()

	// 1. 解析命令行参数
	urls := parseURLs()
	if len(urls) == 0 {
		// 使用默认 URLs
		urls = []string{
			"https://google.com",
			"https://github.com",
			"https://golang.org",
			"https://stackoverflow.com",
			"https://www.baidu.com",
		}
		fmt.Println("使用默认 URLs:")
		for _, url := range urls {
			fmt.Printf("  - %s\n", url)
		}
	}

	// 2. 创建检查器
	checker := NewURLChecker()
	atomic.AddInt32(&checker.TotalURLs, int32(len(urls)))

	// 3. 创建结果队列
	resultQueue := make(chan URLResult, len(urls))

	// 4. 设置信号处理（优雅退出）
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt)

	// 5. 创建任务队列
	jobQueue := make(chan Job, len(urls))

	fmt.Println()
	fmt.Printf("开始检查 %d 个 URL...\n", len(urls))
	fmt.Printf("最大并发数: %d\n", maxWorkers)
	fmt.Printf("超时时间: %v\n", timeout)
	fmt.Println("==================================")
	fmt.Println()

	// 6. 启动工作协程（Worker Goroutines）
	// 这是一个核心的并发模式
	// 多个协程同时处理不同的 URL
	var wg sync.WaitGroup
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go worker(i+1, jobQueue, &wg)
	}

	// 7. 添加任务到队列
	go func() {
		for _, url := range urls {
			job := Job{
				URL:   url,
				Queue: resultQueue,
			}
			jobQueue <- job
		}
		close(jobQueue) // 关闭任务队列
	}()

	// 8. 收集结果
	go func() {
		for result := range resultQueue {
			checker.ProcessResult(result)
		}
	}()

	// 9. 等待所有任务完成
	wg.Wait()
	close(resultQueue) // 关闭结果队列

	// 10. 输出结果
	fmt.Println("==================================")
	fmt.Println()
	printResults(checker)

	// 11. 保存结果到文件
	saveResults(checker)
}

// worker 工作协程
// 从任务队列获取任务并执行
//
// 参数说明：
//   - id: 协程 ID
//   - jobQueue: 任务队列
//   - wg: WaitGroup
func worker(id int, jobQueue <-chan Job, wg *sync.WaitGroup) {
	defer wg.Done()

	fmt.Printf("[Worker %d] 启动\n", id)

	for job := range jobQueue {
		// 检查是否被中断
		select {
		case <-stopChan:
			fmt.Printf("[Worker %d] 收到停止信号\n", id)
			return
		default:
			// 继续处理
		}

		fmt.Printf("[Worker %d] 处理: %s\n", id, job.URL)
		
		// 检查 URL
		result := checkURL(job.URL)
		
		// 发送结果
		job.Queue <- result
	}

	fmt.Printf("[Worker %d] 退出\n", id)
}

// checkURL 检查单个 URL
// 发送 HTTP 请求并获取状态
//
// 参数：
//   url string - 要检查的 URL
//
// 返回值：
//   URLResult - 检查结果
func checkURL(url string) URLResult {
	startTime := time.Now()
	
	result := URLResult{
		URL:       url,
		Timestamp: startTime,
	}

	// 创建 HTTP 客户端
	client := &http.Client{
		Timeout: timeout,
	}

	// 创建请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		result.Status = "ERROR"
		result.Error = err.Error()
		return result
	}

	// 设置请求头
	req.Header.Set("User-Agent", UserAgent)

	// 发送请求
	resp, err := client.Do(req)
	latency := time.Since(startTime).Milliseconds()
	result.Latency = latency

	if err != nil {
		// 检查是否是超时错误
		if strings.Contains(err.Error(), "timeout") {
			result.Status = "TIMEOUT"
		} else {
			result.Status = "ERROR"
		}
		result.Error = err.Error()
		return result
	}

	// 延迟关闭响应体
	defer resp.Body.Close()

	// 获取状态码
	result.StatusCode = resp.StatusCode

	// 判断是否成功
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		result.Status = "OK"
	} else if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		result.Status = "REDIRECT"
	} else if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		result.Status = "CLIENT_ERROR"
	} else if resp.StatusCode >= 500 {
		result.Status = "SERVER_ERROR"
	} else {
		result.Status = "UNKNOWN"
	}

	return result
}

// ProcessResult 处理结果
func (uc *URLChecker) ProcessResult(result URLResult) {
	uc.mu.Lock()
	uc.Results = append(uc.Results, result)
	uc.mu.Unlock()

	// 更新统计
	atomic.AddInt32(&uc.CheckedURLs, 1)
	if result.Status == "OK" {
		atomic.AddInt32(&uc.SuccessNum, 1)
	} else {
		atomic.AddInt32(&uc.FailedNum, 1)
	}
}

// parseURLs 解析命令行参数中的 URLs
func parseURLs() []string {
	var urls []string

	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-urls", "-u":
			if i+1 < len(args) {
				urls = strings.Split(args[i+1], ",")
				i++
			}
		case "-timeout", "-t":
			if i+1 < len(args) {
				var seconds int
				fmt.Sscanf(args[i+1], "%d", &seconds)
				timeout = time.Duration(seconds) * time.Second
				i++
			}
		case "-workers", "-w":
			if i+1 < len(args) {
				fmt.Sscanf(args[i+1], "%d", &maxWorkers)
				i++
			}
		case "-help", "-h":
			printUsage()
		}
	}

	return urls
}

// printResults 打印结果
func printResults(checker *URLChecker) {
	// 按延迟排序
	// 这里简单处理，实际可以用 sort.Slice

	fmt.Println("📊 检查结果:")
	fmt.Println()
	
	// 统计信息
	total := atomic.LoadInt32(&checker.TotalURLs)
	checked := atomic.LoadInt32(&checker.CheckedURLs)
	success := atomic.LoadInt32(&checker.SuccessNum)
	failed := atomic.LoadInt32(&checker.FailedNum)

	fmt.Printf("总 URL 数: %d\n", total)
	fmt.Printf("已检查: %d\n", checked)
	fmt.Printf("成功: %d\n", success)
	fmt.Printf("失败: %d\n", failed)
	fmt.Printf("成功率: %.1f%%\n", float64(success)/float64(total)*100)
	fmt.Println()

	// 详细结果
	for _, result := range checker.Results {
		var icon string
		switch result.Status {
		case "OK":
			icon = "✅"
		case "REDIRECT":
			icon = "🔄"
		case "CLIENT_ERROR", "SERVER_ERROR":
			icon = "❌"
		case "TIMEOUT":
			icon = "⏱️"
		case "ERROR":
			icon = "⚠️"
		default:
			icon = "❓"
		}

		fmt.Printf("%s %s\n", icon, result.URL)
		fmt.Printf("   状态: %s (%d)\n", result.Status, result.StatusCode)
		fmt.Printf("   延迟: %dms\n", result.Latency)
		if result.Error != "" {
			fmt.Printf("   错误: %s\n", result.Error)
		}
		fmt.Println()
	}
}

// saveResults 保存结果到文件
func saveResults(checker *URLChecker) {
	// 保存为 JSON
	data, err := json.MarshalIndent(checker.Results, "", "  ")
	if err != nil {
		fmt.Printf("保存结果失败: %v\n", err)
		return
	}

	filename := fmt.Sprintf("url_check_results_%d.json", time.Now().Unix())
	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		fmt.Printf("保存结果失败: %v\n", err)
		return
	}

	fmt.Printf("结果已保存到: %s\n", filename)
}

// printUsage 打印使用说明
func printUsage() {
	fmt.Println("用法: url_checker [选项]")
	fmt.Println()
	fmt.Println("选项:")
	fmt.Println("  -urls, -u <URL列表>   要检查的 URL（逗号分隔）")
	fmt.Println("  -timeout, -t <秒>    超时时间（默认 5）")
	fmt.Println("  -workers, -w <数字>  并发工作协程数（默认 5）")
	fmt.Println("  -help, -h           显示此帮助信息")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  url_checker -urls https://google.com,https://github.com")
	fmt.Println("  url_checker -timeout 10 -workers 10")
	os.Exit(0)
}

// 代码要点总结
// ============================================
//
// Go 并发核心概念：
//
// 1. Goroutine（协程）
//    - 轻量级的执行单元
//    - 比线程更轻量，可以轻松创建数万个
//    - 使用 go 关键字启动
//
// 2. Channel（通道）
//    - 用于协程间通信
//    - 可以发送和接收值
//    - 阻塞直到有接收者
//
// 3. WaitGroup
//    - 用于等待一组协程完成
//    - Add() 添加计数
//    - Done() 减少计数
//    - Wait() 阻塞直到计数为 0
//
// 4. Mutex（互斥锁）
//    - 保护共享资源
//    - Lock() 加锁
//    - Unlock() 解锁
//
// 5. atomic 包
//    - 原子操作
//    - 用于计数器等场景
//    - 比锁更高效
//
// 6. select 语句
//    - 多通道选择
//    - 类似于 switch，但用于通道
//    - 常用于超时处理
//
// 并发模式：
//
// 1. Worker Pool（工作池）
//    - 预先创建固定数量的工作协程
//    - 从任务队列获取任务
//    - 提高资源利用率
//
// 2. Pipeline（管道）
//    - 数据从协程流向协程
//    - 每级处理一部分
//    - 解耦处理步骤
//
// 3. Fan-out / Fan-in
//    - 一个任务分发给多个协程
//    - 结果汇总

// 练习题
// ============================================
//
// 练习 1: 添加重试功能
//   问题: 失败的 URL 自动重试 3 次？
//   提示: 使用循环和计数器
//
// 练习 2: 添加限流功能
//   问题: 限制每秒请求数？
//   提示: 使用 time.Ticker
//
// 练习 3: 添加进度条
//   问题: 显示检查进度？
//   提示: 使用 atomic.LoadInt32
//
// 练习 4: 添加导出 CSV
//   问题: 导出为 CSV 格式？
//   提示: 使用 encoding/csv 包