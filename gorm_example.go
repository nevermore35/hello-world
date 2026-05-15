// GORM 使用示例
// 一个适合 Go 语言初学者的 GORM（Go ORM）入门教程
//
// 运行方式：
//   # 初始化模块
//   go mod init example && go get gorm.io/gorm gorm.io/driver/sqlite
//   # 运行代码
//   go run gorm_example.go
//
//   # 只看某个主题
//   go run gorm_example.go -topic crud
//   go run gorm_example.go -topic associations
//   go run gorm_example.go -topic query
//
// 覆盖的主题：
//   model       模型定义、字段标签、约定
//   connect     数据库连接、驱动
//   migrate     自动迁移、索引
//   crud        Create / Read / Update / Delete
//   query       条件查询、排序、分页、聚合
//   associations 一对一、一对多、多对多、预加载
//   transaction 事务、嵌套事务、SavePoint
//   advanced    原生 SQL、Scopes、批量操作、软删除
//   pitfalls    常见陷阱：零值更新、N+1 查询、连接池配置
//
// @author 初学者学习项目

package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

// ============================================
// 命令行参数
// ============================================

var topicFlag = flag.String("topic", "", "运行指定主题（model/connect/migrate/crud/query/associations/transaction/advanced/pitfalls），留空运行全部")

// ============================================
// 模型定义
// ============================================

// User 用户模型
type User struct {
	ID        uint           `gorm:"primaryKey;autoIncrement"`
	Name      string         `gorm:"type:varchar(100);not null;index;comment:用户名"`
	Email     string         `gorm:"type:varchar(100);uniqueIndex;not null;comment:邮箱"`
	Age       uint8          `gorm:"default:18;comment:年龄"`
	Active    bool           `gorm:"default:true;comment:是否激活"`
	Bio       *string        `gorm:"type:text;comment:简介"`
	CreatedAt time.Time      `gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime;comment:更新时间"`
	DeletedAt gorm.DeletedAt `gorm:"index;comment:软删除时间"`

	// 关联
	Profile Profile   `gorm:"foreignKey:UserID"`
	Posts   []Post    `gorm:"foreignKey:AuthorID"`
	Groups  []Group   `gorm:"many2many:user_groups;"`
	Orders  []Order   `gorm:"foreignKey:UserID"`
}

// Profile 用户资料（一对一）
type Profile struct {
	ID        uint   `gorm:"primaryKey"`
	UserID    uint   `gorm:"uniqueIndex;not null;comment:用户ID"`
	Avatar    string `gorm:"type:varchar(255);comment:头像"`
	Phone     string `gorm:"type:varchar(20);comment:手机号"`
	Address   string `gorm:"type:varchar(200);comment:地址"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Post 文章（一对多，属于 User）
type Post struct {
	ID        uint           `gorm:"primaryKey"`
	Title     string         `gorm:"type:varchar(200);not null;index;comment:标题"`
	Content   string         `gorm:"type:text;comment:内容"`
	AuthorID  uint           `gorm:"not null;index;comment:作者ID"`
	Author    User           `gorm:"foreignKey:AuthorID"`
	Tags      []Tag          `gorm:"many2many:post_tags;"`
	Comments  []Comment      `gorm:"foreignKey:PostID"`
	ViewCount uint           `gorm:"default:0;comment:浏览量"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// Comment 评论（一对多，属于 Post 和 User）
type Comment struct {
	ID        uint   `gorm:"primaryKey"`
	Content   string `gorm:"type:text;not null;comment:评论内容"`
	PostID    uint   `gorm:"not null;index;comment:文章ID"`
	Post      Post   `gorm:"foreignKey:PostID"`
	UserID    uint   `gorm:"not null;index;comment:用户ID"`
	User      User   `gorm:"foreignKey:UserID"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Tag 标签（多对多，与 Post 关联）
type Tag struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"type:varchar(50);uniqueIndex;not null;comment:标签名"`
	Posts     []Post `gorm:"many2many:post_tags;"`
	CreatedAt time.Time
}

// Group 用户组（多对多，与 User 关联）
type Group struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"type:varchar(50);uniqueIndex;not null;comment:组名"`
	Users     []User `gorm:"many2many:user_groups;"`
	CreatedAt time.Time
}

// Order 订单（一对多，属于 User，演示复合查询）
type Order struct {
	ID        uint      `gorm:"primaryKey"`
	OrderNo   string    `gorm:"type:varchar(32);uniqueIndex;not null;comment:订单号"`
	UserID    uint      `gorm:"not null;index;comment:用户ID"`
	User      User      `gorm:"foreignKey:UserID"`
	Amount    float64   `gorm:"type:decimal(10,2);not null;comment:金额"`
	Status    string    `gorm:"type:varchar(20);default:pending;comment:状态"`
	PaidAt    *time.Time `gorm:"comment:支付时间"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ============================================
// 数据库连接
// ============================================

// connectDB 连接数据库
func connectDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("gorm_demo.db"), &gorm.Config{
		// 日志级别：Silent / Error / Warn / Info
		Logger: logger.Default.LogMode(logger.Info),
		// 禁用默认事务（提升性能）
		SkipDefaultTransaction: false,
		// 命名策略
		// NamingStrategy: schema.NamingStrategy{
		// 	TablePrefix:   "t_",  // 表名前缀
		// 	SingularTable: true,  // 使用单数表名
		// },
	})
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	return db
}

// ============================================
// 主题：模型定义 (model)
// ============================================

func runModelDemo(db *gorm.DB) {
	printSection("模型定义 (Model Definition)")

	printSubSection("常用字段标签")
	fmt.Println(`
	gorm.Model（内嵌结构体）包含：
	  ID        uint ` + "`gorm:\"primaryKey\"`" + `
	  CreatedAt time.Time
	  UpdatedAt time.Time
	  DeletedAt gorm.DeletedAt ` + "`gorm:\"index\"`" + `

	常用标签说明：
	  primaryKey        主键
	  autoIncrement     自增
	  uniqueIndex       唯一索引
	  index             创建索引
	  not null          非空
	  default:xxx       默认值
	  type:varchar(100) 字段类型
	  comment:xxx       注释
	  size:255          字段大小
	  autoCreateTime    创建时自动设置时间
	  autoUpdateTime    更新时自动设置时间
	  <-:false          只读（禁止写入）
	  ->:false          只写（禁止读取）
	  -                 忽略此字段`)

	printSubSection("约定优于配置")
	fmt.Println(`
	GORM 默认约定：
	  - 表名 = 结构体名（蛇形复数），如 User → users
	  - 主键默认字段名 ID，类型为整型
	  - CreatedAt / UpdatedAt 自动管理时间戳
	  - DeletedAt 实现软删除（值不为 NULL 即为已删除）
	  - 外键默认为 关联结构体名 + ID，如 UserID`)
}

// ============================================
// 主题：数据库连接 (connect)
// ============================================

func runConnectDemo(db *gorm.DB) {
	printSection("数据库连接 (Database Connection)")

	fmt.Println(`
	支持的数据驱动：
	  - SQLite:    gorm.io/driver/sqlite
	  - MySQL:     gorm.io/driver/mysql
	  - PostgreSQL: gorm.io/driver/postgres
	  - SQL Server: gorm.io/driver/sqlserver

	连接示例：
	  // SQLite
	  db, _ := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})

	  // MySQL
	  dsn := "user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True"
	  db, _ := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	  // PostgreSQL
	  dsn := "host=localhost user=test password=test dbname=test port=9920"
	  db, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	连接池配置：
	  sqlDB, _ := db.DB()
	  sqlDB.SetMaxIdleConns(10)           // 最大空闲连接数
	  sqlDB.SetMaxOpenConns(100)          // 最大打开连接数
	  sqlDB.SetConnMaxLifetime(time.Hour) // 连接最大存活时间`)

	// 检查数据库连接是否可用（仅示例，需要真实数据库）
	_ = db
}

// ============================================
// 主题：自动迁移 (migrate)
// ============================================

func runMigrateDemo(db *gorm.DB) {
	printSection("自动迁移 (Auto Migration)")

	fmt.Println("执行 AutoMigrate...")

	// 自动迁移：创建表、添加字段、创建索引（不会删除已有字段）
	err := db.AutoMigrate(
		&User{},
		&Profile{},
		&Post{},
		&Comment{},
		&Tag{},
		&Group{},
		&Order{},
	)
	if err != nil {
		log.Fatalf("迁移失败: %v", err)
	}

	fmt.Println("✓ 表结构迁移完成")

	printSubSection("迁移能力与限制")
	fmt.Println(`
	AutoMigrate 会：
	  ✓ 创建不存在的表
	  ✓ 添加不存在的字段
	  ✓ 创建不存在的索引
	  ✗ 不会删除已有字段
	  ✗ 不会修改字段类型
	  ✗ 不会删除已有索引

	手动迁移：
	  db.Migrator().CreateTable(&User{})       // 创建表
	  db.Migrator().DropTable(&User{})          // 删除表
	  db.Migrator().AddColumn(&User{}, "Bio")   // 添加列
	  db.Migrator().DropColumn(&User{}, "Bio")  // 删除列
	  db.Migrator().CreateIndex(&User{}, "IdxAge") // 创建索引
	  db.Migrator().HasTable(&User{})           // 检查表是否存在`)
}

// ============================================
// 主题：CRUD 操作 (crud)
// ============================================

func runCRUDDemo(db *gorm.DB) {
	printSection("CRUD 基本操作")

	// ---- Create ----
	printSubSection("1. Create（创建）")

	// 单条创建
	user := User{Name: "张三", Email: "zhangsan@example.com", Age: 25}
	result := db.Create(&user)
	fmt.Printf("创建用户: ID=%d, RowsAffected=%d, Error=%v\n", user.ID, result.RowsAffected, result.Error)

	// 批量创建
	users := []User{
		{Name: "李四", Email: "lisi@example.com", Age: 30},
		{Name: "王五", Email: "wangwu@example.com", Age: 28},
		{Name: "赵六", Email: "zhaoliu@example.com", Age: 22},
	}
	db.Create(&users)
	fmt.Printf("批量创建 %d 个用户\n", len(users))

	// 指定字段创建
	db.Select("Name", "Email").Create(&User{Name: "只插入名字和邮箱", Age: 99})

	// 排除字段创建
	db.Omit("Age").Create(&User{Name: "不插入Age", Email: "noage@example.com", Age: 99})

	// ---- Read ----
	printSubSection("2. Read（查询）")

	// 根据主键查询
	var firstUser User
	db.First(&firstUser, user.ID)
	fmt.Printf("First: ID=%d, Name=%s, Email=%s\n", firstUser.ID, firstUser.Name, firstUser.Email)

	// 根据条件查询
	var userByEmail User
	db.Where("email = ?", "lisi@example.com").First(&userByEmail)
	fmt.Printf("按邮箱查询: %s\n", userByEmail.Name)

	// First vs Find vs Take 的区别
	fmt.Println(`
	First / Last / Take / Find 的区别：
	  First  按主键升序返回第一条，无结果返回 ErrRecordNotFound
	  Last   按主键降序返回第一条
	  Take   不排序返回第一条
	  Find   返回所有匹配记录，无结果返回空切片（不报错）`)

	// 获取所有用户
	var allUsers []User
	db.Find(&allUsers)
	fmt.Printf("用户总数: %d\n", len(allUsers))

	// ---- Update ----
	printSubSection("3. Update（更新）")

	// 更新单个字段
	db.Model(&user).Update("Age", 26)
	fmt.Printf("更新年龄: %d\n", user.Age)

	// 更新多个字段（使用 struct）
	db.Model(&user).Updates(User{Name: "张三（已更新）", Age: 27})

	// 更新多个字段（使用 map，可更新零值）
	db.Model(&user).Updates(map[string]interface{}{
		"Age":   28,
		"Active": false, // map 可以更新零值，struct 不行
	})
	fmt.Println("使用 map 更新后（包括零值 Active=false）")

	// 批量更新
	db.Model(&User{}).Where("age < ?", 25).Update("Active", false)

	// 表达式更新
	db.Model(&user).UpdateColumn("age", gorm.Expr("age + ?", 1))
	fmt.Printf("表达式更新年龄+1: %d\n", user.Age)

	// ---- Delete ----
	printSubSection("4. Delete（删除）")

	// 软删除（设置了 DeletedAt 字段自动启用）
	db.Delete(&User{}, users[2].ID) // 删除王五（软删除）
	fmt.Println("软删除完成（DeletedAt 被设置为当前时间）")

	// 查询时自动过滤已软删除的记录
	var remainingUsers []User
	db.Find(&remainingUsers)
	fmt.Printf("查询到的用户数（已过滤软删除）: %d\n", len(remainingUsers))

	// 查询包含软删除的记录
	var allWithDeleted []User
	db.Unscoped().Find(&allWithDeleted)
	fmt.Printf("包含软删除的用户数: %d\n", len(allWithDeleted))

	// 物理删除
	db.Unscoped().Delete(&User{}, users[2].ID)
	fmt.Println("物理删除完成（记录从数据库彻底移除）")
}

// ============================================
// 主题：查询进阶 (query)
// ============================================

func runQueryDemo(db *gorm.DB) {
	printSection("查询进阶 (Advanced Query)")

	// ---- 条件查询 ----
	printSubSection("1. 条件查询")

	var users []User

	// 多种条件方式
	db.Where("name LIKE ?", "%张%").Find(&users)
	fmt.Printf("LIKE 查询: 找到 %d 个用户\n", len(users))

	// struct 条件（零值字段不会被使用）
	db.Where(&User{Age: 25, Active: true}).Find(&users)

	// map 条件（零值也会被使用）
	db.Where(map[string]interface{}{"Age": 25, "Active": false}).Find(&users)

	// IN 查询
	db.Where("age IN ?", []int{25, 28, 30}).Find(&users)

	// OR 查询
	db.Where("name = ? OR name = ?", "张三", "李四").Find(&users)

	// NOT 查询
	db.Not("age = ?", 25).Find(&users)

	// ---- 排序、分页、聚合 ----
	printSubSection("2. 排序、分页、聚合")

	// 排序
	db.Order("age desc, name asc").Find(&users)

	// Limit & Offset（分页）
	var page1 []User
	db.Limit(2).Offset(0).Find(&page1) // 第 1 页
	fmt.Printf("第 1 页: %d 条\n", len(page1))

	var page2 []User
	db.Limit(2).Offset(2).Find(&page2) // 第 2 页
	fmt.Printf("第 2 页: %d 条\n", len(page2))

	// 聚合函数
	var count int64
	var maxAge int
	db.Model(&User{}).Count(&count)
	db.Model(&User{}).Select("MAX(age)").Scan(&maxAge)
	fmt.Printf("用户总数: %d, 最大年龄: %d\n", count, maxAge)

	// Group By & Having
	type AgeStat struct {
		Age   uint8
		Count int64
	}
	var stats []AgeStat
	db.Model(&User{}).Select("age, COUNT(*) as count").
		Group("age").Having("COUNT(*) > ?", 0).
		Find(&stats)
	for _, s := range stats {
		fmt.Printf("  年龄 %d: %d 人\n", s.Age, s.Count)
	}

	// ---- 子查询 ----
	printSubSection("3. 子查询")

	var subUsers []User
	db.Where("age > (?)", db.Model(&User{}).Select("AVG(age)")).Find(&subUsers)
	fmt.Printf("年龄大于平均值的用户数: %d\n", len(subUsers))

	// ---- 选择特定字段 ----
	printSubSection("4. 选择特定字段")

	var names []string
	db.Model(&User{}).Pluck("name", &names)
	fmt.Printf("所有用户名: %v\n", names)

	type NameAndEmail struct {
		Name  string
		Email string
	}
	var ne []NameAndEmail
	db.Model(&User{}).Select("name", "email").Find(&ne)
	for _, v := range ne {
		fmt.Printf("  %s <%s>\n", v.Name, v.Email)
	}
}

// ============================================
// 主题：关联关系 (associations)
// ============================================

func runAssociationsDemo(db *gorm.DB) {
	printSection("关联关系 (Associations)")

	// 获取第一个用户
	var user User
	db.First(&user)

	// ---- 一对一 (Has One) ----
	printSubSection("1. 一对一 (Has One) — User → Profile")

	profile := Profile{
		UserID:  user.ID,
		Avatar:  "/avatars/default.png",
		Phone:   "13800138000",
		Address: "北京市朝阳区",
	}
	db.Create(&profile)
	fmt.Printf("为用户 %s 创建了 Profile\n", user.Name)

	// 查询时预加载
	var userWithProfile User
	db.Preload("Profile").First(&userWithProfile, user.ID)
	fmt.Printf("  Avatar: %s, Phone: %s\n", userWithProfile.Profile.Avatar, userWithProfile.Profile.Phone)

	// ---- 一对多 (Has Many) ----
	printSubSection("2. 一对多 (Has Many) — User → Posts")

	posts := []Post{
		{Title: "Go 入门教程", Content: "Go 语言基础...", AuthorID: user.ID},
		{Title: "GORM 使用指南", Content: "GORM 是一个强大的 ORM...", AuthorID: user.ID},
		{Title: "微服务架构实践", Content: "微服务设计模式...", AuthorID: user.ID},
	}
	db.Create(&posts)
	fmt.Printf("为用户 %s 创建了 %d 篇文章\n", user.Name, len(posts))

	// 预加载关联
	var userWithPosts User
	db.Preload("Posts").First(&userWithPosts, user.ID)
	fmt.Printf("  文章列表:\n")
	for _, p := range userWithPosts.Posts {
		fmt.Printf("    - %s\n", p.Title)
	}

	// ---- 多对多 (Many to Many) ----
	printSubSection("3. 多对多 (Many to Many) — Posts ↔ Tags")

	tags := []Tag{
		{Name: "Go"},
		{Name: "ORM"},
		{Name: "数据库"},
	}
	db.Create(&tags)

	// 为第一篇文章添加标签
	var firstPost Post
	db.First(&firstPost, posts[0].ID)
	db.Model(&firstPost).Association("Tags").Append([]Tag{tags[0], tags[1]})
	fmt.Printf("为文章「%s」添加了标签: Go, ORM\n", firstPost.Title)

	// 预加载多对多
	var postWithTags Post
	db.Preload("Tags").First(&postWithTags, firstPost.ID)
	fmt.Printf("  标签: ")
	for _, t := range postWithTags.Tags {
		fmt.Printf("%s ", t.Name)
	}
	fmt.Println()

	// ---- 预加载策略 ----
	printSubSection("4. 预加载策略")

	fmt.Println(`
	Preload 方式：
	  db.Preload("Profile").Find(&users)                   // 基础预加载
	  db.Preload("Posts.Comments").Find(&users)            // 嵌套预加载
	  db.Preload("Posts", "view_count > ?", 10).Find(&users) // 条件预加载
	  db.Preload(clause.Associations).Find(&users)          // 预加载全部关联

	Joins 方式（适合一对一，生成一条 SQL）：
	  db.Joins("Profile").Find(&users)

	延迟加载（按需查询，注意 N+1 问题）：
	  db.Model(&user).Association("Posts").Find(&user.Posts)`)

	// ---- 关联操作 ----
	printSubSection("5. 关联增删改")

	// 替换关联
	db.Model(&firstPost).Association("Tags").Replace([]Tag{tags[0], tags[2]})
	fmt.Println("替换标签为: Go, 数据库")

	// 删除关联
	db.Model(&firstPost).Association("Tags").Delete(tags[0])
	fmt.Println("删除了标签: Go")

	// 清空关联
	// db.Model(&firstPost).Association("Tags").Clear()

	// 关联计数
	count := db.Model(&firstPost).Association("Tags").Count()
	fmt.Printf("当前关联标签数: %d\n", count)
}

// ============================================
// 主题：事务 (transaction)
// ============================================

func runTransactionDemo(db *gorm.DB) {
	printSection("事务 (Transaction)")

	// ---- 自动事务 ----
	printSubSection("1. 自动事务（推荐）")

	err := db.Transaction(func(tx *gorm.DB) error {
		// 在事务中创建订单
		order := Order{
			OrderNo: fmt.Sprintf("ORD-%d", time.Now().Unix()),
			UserID:  1,
			Amount:  99.99,
			Status:  "pending",
		}
		if err := tx.Create(&order).Error; err != nil {
			return err // 返回 error 自动回滚
		}

		// 更新用户状态
		if err := tx.Model(&User{}).Where("id = ?", 1).
			Update("active", true).Error; err != nil {
			return err
		}

		fmt.Printf("事务提交成功: 订单 %s, 金额 %.2f\n", order.OrderNo, order.Amount)
		return nil // 返回 nil 自动提交
	})
	if err != nil {
		fmt.Printf("事务失败（自动回滚）: %v\n", err)
	}

	// ---- 手动事务 ----
	printSubSection("2. 手动事务")

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			fmt.Println("发生 panic，事务回滚")
		}
	}()

	order := Order{
		OrderNo: fmt.Sprintf("ORD-%d", time.Now().Unix()+1),
		UserID:  1,
		Amount:  199.99,
		Status:  "pending",
	}
	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		fmt.Printf("创建订单失败，回滚: %v\n", err)
		return
	}

	tx.Commit()
	fmt.Printf("手动事务提交: 订单 %s\n", order.OrderNo)

	// ---- 嵌套事务 / SavePoint ----
	printSubSection("3. 嵌套事务（SavePoint）")

	db.Transaction(func(tx *gorm.DB) error {
		tx.Create(&Order{
			OrderNo: fmt.Sprintf("ORD-%d", time.Now().Unix()+2),
			UserID:  1,
			Amount:  50.00,
		})

		// 嵌套事务
		err := tx.Transaction(func(tx2 *gorm.DB) error {
			tx2.Create(&Order{
				OrderNo: fmt.Sprintf("ORD-%d", time.Now().Unix()+3),
				UserID:  1,
				Amount:  30.00,
			})
			return errors.New("模拟嵌套事务失败")
		})

		if err != nil {
			fmt.Printf("嵌套事务失败（外层事务继续）: %v\n", err)
		}
		return nil
	})
}

// ============================================
// ============================================


// ============================================
// 主题：高级特性 (advanced)
// ============================================

func runAdvancedDemo(db *gorm.DB) {
	printSection("高级特性 (Advanced Features)")

	// ---- 原生 SQL ----
	printSubSection("1. 原生 SQL")

	var name string
	db.Raw("SELECT name FROM users WHERE id = ?", 1).Scan(&name)
	fmt.Printf("原生查询: name=%s\n", name)

	// 执行原生写操作
	db.Exec("UPDATE users SET age = age + 1 WHERE id = ?", 1)

	// ---- Scopes ----
	printSubSection("2. Scopes（查询作用域）")

	// 定义 scope
	activeUsers := func(db *gorm.DB) *gorm.DB {
		return db.Where("active = ?", true)
	}
	adultUsers := func(db *gorm.DB) *gorm.DB {
		return db.Where("age >= ?", 18)
	}
	paginate := func(page, pageSize int) func(db *gorm.DB) *gorm.DB {
		return func(db *gorm.DB) *gorm.DB {
			return db.Offset((page - 1) * pageSize).Limit(pageSize)
		}
	}

	var activeAdults []User
	db.Scopes(activeUsers, adultUsers, paginate(1, 10)).Find(&activeAdults)
	fmt.Printf("激活的成年用户数: %d\n", len(activeAdults))

	// ---- 批量操作 ----
	printSubSection("3. 批量操作")

	// 批量插入
	bulkUsers := make([]User, 3)
	for i := 0; i < 3; i++ {
		bulkUsers[i] = User{
			Name:  fmt.Sprintf("批量用户%d", i+1),
			Email: fmt.Sprintf("bulk%d@example.com", i+1),
			Age:   20,
		}
	}
	db.CreateInBatches(bulkUsers, 100) // 每批 100 条
	fmt.Printf("批量插入 %d 个用户\n", len(bulkUsers))

	// 批量更新
	db.Where("name LIKE ?", "批量%").Updates(User{Active: true})

	// ---- 软删除详解 ----
	printSubSection("4. 软删除详解")

	fmt.Println(`
	特点：
	  - DeletedAt 不为 NULL 即为已删除
	  - 所有查询自动过滤软删除记录
	  - 使用 Unscoped() 可以查询已删除记录

	操作：
	  db.Delete(&user)             // 软删除
	  db.Unscoped().Delete(&user)  // 物理删除
	  db.Unscoped().Find(&users)   // 查询包含已删除
	  db.Unscoped().Where("id = ?", id).Find(&user) // 查找已删除记录`)

	// ---- 错误处理 ----
	printSubSection("5. 错误处理")

	fmt.Println(`
	常用错误判断：
	  errors.Is(result.Error, gorm.ErrRecordNotFound)  // 记录未找到
	  errors.Is(result.Error, gorm.ErrDuplicatedKey)    // 唯一键冲突
	  errors.Is(result.Error, gorm.ErrForeignKeyViolated) // 外键冲突

	最佳实践：
	  1. 始终检查 result.Error
	  2. 使用 First 而非 Find 获取单条记录（Find 不报 RecordNotFound）
	  3. 避免忽略 RowsAffected = 0 的更新/删除`)
}

// ============================================
// 主题：常见陷阱 (pitfalls)
// ============================================

func runPitfallsDemo(db *gorm.DB) {
	printSection("常见陷阱 (Common Pitfalls)")

	fmt.Println(`
	1. 零值更新问题
	   使用 struct 更新时，零值（0, "", false）不会更新到数据库。
	   解决方案：使用 map 或 Select 指定字段。

	   ❌ db.Model(&user).Updates(User{Age: 0, Active: false})  // Age 和 Active 不会更新
	   ✓ db.Model(&user).Updates(map[string]interface{}{"Age": 0, "Active": false})
	   ✓ db.Model(&user).Select("Age", "Active").Updates(User{Age: 0, Active: false})

	2. N+1 查询问题
	   在循环中加载关联会导致大量查询。

	   ❌ for _, user := range users {
	          db.Model(&user).Association("Posts").Find(&user.Posts)
	      }
	   ✓ db.Preload("Posts").Find(&users)

	3. 连接池耗尽
	   忘记配置连接池或配置不当会导致连接泄漏。

	   ✓ sqlDB, _ := db.DB()
	     sqlDB.SetMaxIdleConns(10)
	     sqlDB.SetMaxOpenConns(100)
	     sqlDB.SetConnMaxLifetime(time.Hour)

	4. 事务忘记提交/回滚
	   手动事务必须显式 Commit 或 Rollback。

	   ✓ 使用 db.Transaction() 闭包，自动处理提交/回滚

	5. 循环变量捕获陷阱
	   在循环中启动 goroutine 时捕获循环变量。

	   ❌ for _, u := range users {
	          go func() { db.Create(&u) }()  // 所有 goroutine 用同一个 u
	      }
	   ✓ for _, u := range users {
	          u := u  // 创建副本
	          go func() { db.Create(&u) }()
	      }

	6. 默认事务开销
	   对于只读查询，可以跳过默认事务以提升性能。

	   ✓ db.Session(&gorm.Session{SkipDefaultTransaction: true}).Find(&users)

	7. 表名约定混淆
	   默认表名为复数形式（User → users），可以通过配置修改。

	   ✓ db.Config.NamingStrategy = schema.NamingStrategy{SingularTable: true}
	`)
}

// ============================================
// 主函数
// ============================================

func main() {
	flag.Parse()

	fmt.Println("==================================")
	fmt.Println("    GORM 使用示例教程 v1.0")
	fmt.Println("==================================")
	fmt.Println()

	// 连接数据库
	db := connectDB()

	// 获取底层 sql.DB 并配置连接池
	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
	defer sqlDB.Close()

	// 自动迁移（始终执行，确保表结构正确）
	db.AutoMigrate(
		&User{},
		&Profile{},
		&Post{},
		&Comment{},
		&Tag{},
		&Group{},
		&Order{},
	)

	// 主题路由
	topic := *topicFlag
	runTopic := func(name string, fn func(*gorm.DB)) {
		if topic == "" || topic == name {
			fn(db)
		}
	}

	// 运行所有或指定主题
	allTopics := []struct {
		name string
		fn   func(*gorm.DB)
	}{
		{"model", runModelDemo},
		{"connect", runConnectDemo},
		{"migrate", runMigrateDemo},
		{"crud", runCRUDDemo},
		{"query", runQueryDemo},
		{"associations", runAssociationsDemo},
		{"transaction", runTransactionDemo},
		
		{"advanced", runAdvancedDemo},
		{"pitfalls", runPitfallsDemo},
	}

	found := topic == ""
	for _, t := range allTopics {
		if topic == "" || topic == t.name {
			t.fn(db)
			found = true
		}
	}

	if !found {
		fmt.Printf("未知主题: %s\n", topic)
		fmt.Println("可用主题: model, connect, migrate, crud, query, associations, transaction, advanced, pitfalls")
	} else {
		fmt.Println()
		fmt.Println("==================================")
		fmt.Println("        全部主题完成！")
		fmt.Println("==================================")
	}
}

// ============================================
// 辅助函数
// ============================================

func printSection(title string) {
	fmt.Println()
	fmt.Println("--------------------------------------------------")
	fmt.Printf("  %s\n", title)
	fmt.Println("--------------------------------------------------")
}

func printSubSection(title string) {
	fmt.Printf("\n  ▶ %s\n\n", title)
}

// ============================================
// 代码要点总结
// ============================================
//
// GORM 核心概念：
//
// 1. 模型定义
//    - 使用 struct + tag 定义数据库模型
//    - 遵循约定优于配置原则
//    - 内嵌 gorm.Model 快速获得 ID/CreatedAt/UpdatedAt/DeletedAt
//
// 2. CRUD
//    - Create: db.Create(&record)
//    - Read:   db.First(&record, id) / db.Find(&records)
//    - Update: db.Model(&record).Update("field", val) / db.Updates(map/struct)
//    - Delete: db.Delete(&record)  // 默认软删除
//
// 3. 查询
//    - Where / Not / Or / Order / Limit / Offset
//    - Select 指定字段 / Pluck 提取单列
//    - Group / Having 聚合
//    - Joins / Preload 关联加载
//    - Scopes 抽取可复用的查询逻辑
//
// 4. 关联
//    - Has One: 一对一
//    - Has Many: 一对多
//    - Belongs To: 属于
//    - Many To Many: 多对多（需要中间表）
//    - Preload / Joins 预加载
//    - Association API 进行关联操作
//
// 5. 事务
//    - db.Transaction(func(tx *gorm.DB) error { ... }) 自动事务
//    - db.Begin() / tx.Commit() / tx.Rollback() 手动事务
//    - 嵌套事务使用 SavePoint
//
// 6. 高级特性
//    - Raw / Exec 原生 SQL
//    - Scopes 查询作用域
//    - CreateInBatches 批量插入
//    - Migrator API 精细控制表结构
//
// 与标准库 database/sql 的对比：
//
// 特性          | database/sql     | GORM
// -------------|-----------------|------------------
// 查询方式      | 手写 SQL        | 链式 API + SQL
// 结果映射      | rows.Scan 手动  | 自动映射到 struct
// 迁移         | 手动写 SQL      | AutoMigrate
// 关联         | 手写 JOIN       | Preload / Association
// 事务         | tx.Begin 手动   | Transaction 闭包
//
// 适用场景：
//   - database/sql：简单查询、高性能需求、精细 SQL 控制
//   - GORM：快速开发、复杂关联、自动迁移、减少样板代码

// ============================================
// 练习题
// ============================================
//
// 练习 1: 扩展用户模型
//   问题: 为 User 增加一个 Address 结构体（内嵌 value 类型），如何处理？
//   提示: 使用 serializer:json 标签或自定义 Scan/Value
//
// 练习 2: 实现复杂搜索
//   问题: 如何实现一个支持多条件组合的动态搜索接口？
//   提示: 使用 Scopes 和条件链式调用
//
// 练习 3: 实现乐观锁
//   问题: 如何使用 GORM 实现乐观锁防止并发更新冲突？
//
// 练习 4: 改造 Todo API
//   问题: 将 todo_api.go 中的数据存储改为 GORM + SQLite
//   提示: 定义 Todo struct 的 GORM 模型，替换内存 map
//
// 练习 5: 改造博客系统
//   问题: 将 blog_system.go 中的文章和用户存储改为 GORM
//   提示: 使用 GORM 的关联特性处理文章-用户-评论关系
