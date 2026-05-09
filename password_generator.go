// Password Generator
// 一个适合 Go 语言初学者的密码生成器项目
// 学习目标：Go 基本语法、字符串处理、随机数、命令行参数
//
// 运行方式：
//   go run password_generator.go
//   go run password_generator.go -length 16 -numbers -symbols
//
// @author 初学者学习项目

package main

import (
	"bufio"
	"crypto/rand" // 加密安全的随机数生成器
	"fmt"
	"os"
	"strings"
	"time"
)

// 常量定义
// ============================================

const (
	// DefaultPasswordLength 默认密码长度
	DefaultPasswordLength = 12
	// MinPasswordLength 最小密码长度
	MinPasswordLength = 4
	// MaxPasswordLength 最大密码长度
	MaxPasswordLength = 64
)

// 全局变量
// ============================================

var (
	// lowercaseLetters 小写字母集
	lowercaseLetters = "abcdefghijklmnopqrstuvwxyz"
	// uppercaseLetters 大写字母集
	upperercaseLetters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	// numbers 数字集
	numbers = "0123456789"
	// symbols 符号集
	symbols = "!@#$%^&*()_+-=[]{}|;:,.<>?"
)

// 结构体定义
// ============================================

// PasswordOptions 密码生成选项
// 使用结构体来组织相关的配置参数
type PasswordOptions struct {
	Length   int  // 密码长度
	UseUpper bool // 是否包含大写字母
	UseLower bool // 是否包含小写字母
	UseNum   bool // 是否包含数字
	UseSymbol bool // 是否包含符号
}

// Config 配置结构体 (可选扩展)
// 用于存储程序运行时的配置
type Config struct {
	Version     string // 版本号
	Author      string // 作者
	Description string // 描述
}

// 函数定义
// ============================================

// main 程序入口
// 每个 Go 程序都必须有一个 main 函数
// 这是程序开始执行的地方
func main() {
	fmt.Println("==================================")
	fmt.Println("      Go 密码生成器 v1.0")
	fmt.Println("==================================")
	fmt.Println()

	// 1. 解析命令行参数
	options := parseArguments()

	// 2. 验证参数
	if !validateOptions(&options) {
		fmt.Println("参数验证失败，请检查输入的参数")
		os.Exit(1)
	}

	// 3. 生成密码
	password := generatePassword(options)

	fmt.Println("生成的密码:", password)
	fmt.Println()
	fmt.Println("密码强度分析:")
	analyzePasswordStrength(password, options)
}

// generatePassword 生成密码
// 根据提供的选项生成随机密码
//
// 参数：
//   options PasswordOptions - 密码生成选项
//
// 返回值：
//   string - 生成的密码
func generatePassword(options PasswordOptions) string {
	// 1. 创建字符池（所有可用字符的集合）
	charPool := buildCharPool(options)

	// 2. 处理空字符池的情况
	if len(charPool) == 0 {
		fmt.Println("警告: 没有选择任何字符类型，将使用小写字母")
		charPool = lowercaseLetters
	}

	fmt.Printf("INFO: 字符池大小 = %d\n", len(charPool))
	fmt.Printf("INFO: 目标密码长度 = %d\n", options.Length)

	// 3. 生成密码
	// 使用加密安全的随机数生成器
	password := make([]byte, options.Length)
	for i := 0; i < options.Length; i++ {
		// 生成随机索引
		randomIndex, err := rand.Int(rand.Reader, int64(len(charPool)))
		if err != nil {
			// 如果加密随机失败，回退到时间戳随机
			fmt.Println("WARNING: 使用时间戳随机数（建议生产环境使用 crypto/rand）")
			randomIndex = int64(time.Now().UnixNano() % int64(len(charPool)))
			time.Sleep(time.Nanosecond) // 避免重复时间戳
		}
		password[i] = charPool[randomIndex]
	}

	return string(password)
}

// buildCharacterPool 构建字符池
// 根据选项组合所有可用的字符
//
// 参数：
//   options PasswordOptions - 密码生成选项
//
// 返回值：
//   string - 组合后的字符池
func buildCharPool(options PasswordOptions) string {
	var charPool []string
	charPool = append(charPool, "")
	
	// 小写字母
	if options.UseLower {
		charPool[0] += lowercaseLetters
	}
	// 大写字母
	if options.UseUpper {
		charPool[0] += uppercaseLetters
	}
	// 数字
	if options.UseNum {
		charPool[0] += numbers
	}
	// 符号
	if options.UseSymbol {
		charPool[0] += symbols
	}

	return charPool[0]
}

// parseArguments 解析命令行参数
// 这是一个简单的命令行参数解析示例
//
// 返回值：
//   PasswordOptions - 解析后的选项
func parseArguments() PasswordOptions {
	options := PasswordOptions{
		Length:   DefaultPasswordLength,
		UseLower: true,
		UseUpper: true,
		UseNum:   true,
		UseSymbol: false,
	}

	// 获取命令行参数（排除程序名）
	args := os.Args[1:]

	// 遍历参数
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-length", "-l":
			// 处理长度参数
			if i+1 < len(args) {
				fmt.Sscanf(args[i+1], "%d", &options.Length)
				i++ // 跳过下一个参数（已经处理过了）
			}
		case "-numbers", "-n":
			options.UseNum = true
		case "-symbols", "-s":
			options.UseSymbol = true
		case "-help", "-h":
			printUsage()
		}
	}

	return options
}

// validateOptions 验证选项
// 确保选项值是有效的
//
// 参数：
//   options *PasswordOptions - 要验证的选项（指针）
//
// 返回值：
//   bool - 是否有效
func validateOptions(options *PasswordOptions) bool {
	// 验证长度
	if options.Length < MinPasswordLength {
		fmt.Printf("错误: 密码长度不能小于 %d\n", MinPasswordLength)
		return false
	}
	if options.Length > MaxPasswordLength {
		fmt.Printf("错误: 密码长度不能大于 %d\n", MaxPasswordLength)
		return false
	}

	// 至少选择一种字符类型
	if !options.UseLower && !options.UseUpper && !options.UseNum && !options.UseSymbol {
		fmt.Println("警告: 没有选择任何字符类型，将使用默认配置")
		options.UseLower = true
	}

	return true
}

// printUsage 打印使用说明
func printUsage() {
	fmt.Println("用法: password_generator [选项]")
	fmt.Println()
	fmt.Println("选项:")
	fmt.Println("  -length, -l <数字>  设置密码长度 (默认 12，最小 4，最大 64)")
	fmt.Println("  -numbers, -n      包含数字")
	fmt.Println("  -symbols, -s      包含符号")
	fmt.Println("  -help, -h         显示此帮助信息")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  password_generator -length 16 -numbers -symbols")
	fmt.Println()
	os.Exit(0)
}

// analyzePasswordStrength 分析密码强度
// 这是一个简单的密码强度分析函数
// 实际项目中应该使用更复杂的算法
func analyzePasswordStrength(password string, options PasswordOptions) {
	length := len(password)
	score := 0

	// 长度评分
	if length >= 8 {
		score += 1
	}
	if length >= 12 {
		score += 1
	}
	if length >= 16 {
		score += 1
	}

	// 字符类型评分
	types := 0
	if options.UseLower {
		types++
	}
	if options.UseUpper {
		types++
	}
	if options.UseNum {
		types++
	}
	if options.UseSymbol {
		types += 2 // 符号提供更高安全性
	}
	score += types

	// 输出强度等级
	fmt.Printf("  - 密码长度: %d\n", length)
	fmt.Printf("  - 字符类型数: %d\n", types)
	fmt.Printf("  - 强度评分: %d/7\n", score)

	switch {
	case score >= 6:
		fmt.Println("  - 强度等级: 非常强 🛡️🛡️🛡️")
	case score >= 4:
		fmt.Println("  - 强度等级: 强 🛡️🛡️")
	case score >= 2:
		fmt.Println("  - 强度等级: 中等 🛡️")
	default:
		fmt.Println("  - 强度等级: 弱 ⚠️")
	}

	fmt.Println()
	fmt.Println("💡 小贴士:")
	fmt.Println("  - 密码越长越安全")
	fmt.Println("  - 使用多种字符类型更安全")
	fmt.Println("  - 符号能显著提高密码强度")

	// 保持控制台不退出（按回车键退出）
	fmt.Println()
	fmt.Print("按回车键退出...")
	bufio.NewReader(os.Stdin).ReadString('\n')
}

// 附加练习题（供初学者练习）
// ============================================
//
// 练习 1: 添加更多字符集
//   问题: 如何添加更多字符，如中文、俄文等？
//   提示: 修改全局变量 charSets
//
// 练习 2: 添加排除功能
//   问题: 如何排除某些特定字符？
//   提示: 添加 ExcludeChars 字段到 PasswordOptions
//
// 练习 3: 添加密码存储功能
//   问题: 如何将生成的密码保存到文件？
//   提示: 使用 os 包或 ioutil 包
//
// 练习 4: 添加 GUI 界面
//   问题: 如何使用 Go 的 Web 框架创建 GUI？
//   提示: 看看项目 4（简易博客系统）

// 代码要点总结
// ============================================
//
// 1. Go 程序从 main() 函数开始执行
// 2. 使用 package 声明包名
// 3. 使用 import 导入依赖
// 4. 变量声明: var, :=, const
// 5. 结构体用于组织相关数据
// 6. 函数参数传递: 值传递 vs 指针传递
// 7. crypto/rand 生成加密安全的随机数
// 8. 条件语句: if, switch
// 9. 循环语句: for
// 10. 格式化输出: fmt.Printf