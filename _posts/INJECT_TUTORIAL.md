---
layout: post
title: "Inject 依赖注入原理深度教程"
date:   2024-11-23
tags: [go]
comments: true
author: xiaodp
toc: true
---

本文由浅入深详细介绍golang自动注入工具inject

# Inject 依赖注入原理深度教程

## 📚 目录

1. [第一部分：基础概念](#第一部分基础概念)
2. [第二部分：Go 反射基础](#第二部分go-反射基础)
3. [第三部分：Inject 库入门](#第三部分inject-库入门)
4. [第四部分：深入源码](#第四部分深入源码)
5. [第五部分：实战演练](#第五部分实战演练)
6. [第六部分：常见问题](#第六部分常见问题)

---

## 第一部分：基础概念

### 1.1 什么是依赖注入？

**依赖注入（Dependency Injection, DI）** 是一种设计模式，用于解耦代码，提高可测试性和可维护性。

#### 传统方式（紧耦合）

```go
// ❌ 不好的方式：在内部创建依赖
type UserService struct {
    db *sql.DB
}

func NewUserService() *UserService {
    db, _ := sql.Open("mysql", "user:pass@/dbname")
    return &UserService{db: db}
}
```

**问题：**
- 无法替换数据库实现（比如测试时用 mock）
- 无法控制数据库的创建时机
- 代码耦合度高

#### 依赖注入方式（松耦合）

```go
// ✅ 好的方式：从外部注入依赖
type UserService struct {
    db *sql.DB
}

func NewUserService(db *sql.DB) *UserService {
    return &UserService{db: db}
}

// 使用时
db, _ := sql.Open("mysql", "user:pass@/dbname")
service := NewUserService(db)
```

**优势：**
- 可以轻松替换依赖（测试时注入 mock）
- 依赖的创建和对象的创建分离
- 代码更灵活、可测试

### 1.2 手动依赖注入的问题

当依赖关系复杂时，手动注入会变得繁琐：

```go
// 假设有这些依赖关系：
// UserService -> DB, Logger, Cache
// OrderService -> DB, Logger, UserService
// PaymentService -> DB, Logger, OrderService

func SetupServices() {
    db := NewDB()
    logger := NewLogger()
    cache := NewCache()
    
    userService := NewUserService(db, logger, cache)
    orderService := NewOrderService(db, logger, userService)
    paymentService := NewPaymentService(db, logger, orderService)
    
    // 需要手动管理所有依赖关系，容易出错
}
```

**问题：**
- 需要手动管理依赖顺序
- 容易遗漏依赖
- 代码冗长且容易出错

### 1.3 自动依赖注入的解决方案

**Inject 库的作用：** 自动扫描结构体字段，识别依赖关系，自动创建和注入依赖。

```go
type UserService struct {
    DB     *sql.DB    `inject:""`
    Logger *Logger    `inject:""`
    Cache  *Cache     `inject:""`
}

// 只需要提供基础对象，其他依赖自动注入
inject.Provide(&inject.Object{Value: db})
inject.Provide(&inject.Object{Value: logger})
inject.Populate(&userService)  // 自动注入所有依赖
```

---

## 第二部分：Go 反射基础

Inject 库基于 Go 的 `reflect` 包实现，理解反射是掌握 inject 的关键。

### 2.1 反射的核心概念

**反射（Reflection）** 允许程序在运行时检查类型信息、修改变量值。

#### 核心类型

```go
import "reflect"

// reflect.Type  - 类型信息
// reflect.Value - 值信息
```

### 2.2 获取类型信息

```go
type User struct {
    Name string
    Age  int
}

func main() {
    u := &User{Name: "Alice", Age: 30}
    
    // 获取类型
    t := reflect.TypeOf(u)        // *User
    fmt.Println(t)                 // *main.User
    fmt.Println(t.Kind())          // ptr (指针类型)
    fmt.Println(t.Elem())          // User (指向的类型)
    fmt.Println(t.Elem().Kind())   // struct (结构体类型)
    
    // 获取值
    v := reflect.ValueOf(u)        // 获取值
    fmt.Println(v.Kind())          // ptr
    fmt.Println(v.Elem())          // 解引用，得到 User 的值
}
```

### 2.3 遍历结构体字段

```go
type Service struct {
    DB     *sql.DB
    Logger *Logger
    Cache  *Cache
}

func main() {
    s := &Service{}
    t := reflect.TypeOf(s).Elem()  // 获取 Service 类型（不是指针）
    v := reflect.ValueOf(s).Elem()  // 获取 Service 的值（不是指针）
    
    // 遍历所有字段
    for i := 0; i < t.NumField(); i++ {
        field := t.Field(i)           // 字段类型信息
        fieldValue := v.Field(i)      // 字段值
        
        fmt.Printf("字段名: %s\n", field.Name)
        fmt.Printf("字段类型: %s\n", field.Type)
        fmt.Printf("字段标签: %s\n", field.Tag)
        fmt.Printf("是否可设置: %v\n", fieldValue.CanSet())
    }
}
```

### 2.4 读取结构体标签（Struct Tag）

```go
type Service struct {
    DB     *sql.DB `inject:""`
    Logger *Logger `inject:"private"`
    Cache  *Cache  `inject:"my_cache"`
}

func main() {
    t := reflect.TypeOf(&Service{}).Elem()
    
    for i := 0; i < t.NumField(); i++ {
        field := t.Field(i)
        tag := field.Tag.Get("inject")  // 获取 inject 标签的值
        
        fmt.Printf("字段 %s 的 inject 标签: %s\n", field.Name, tag)
        // 输出:
        // 字段 DB 的 inject 标签: 
        // 字段 Logger 的 inject 标签: private
        // 字段 Cache 的 inject 标签: my_cache
    }
}
```

### 2.5 设置字段值

```go
type Service struct {
    DB *sql.DB
}

func main() {
    s := &Service{}
    v := reflect.ValueOf(s).Elem()
    
    // 创建新的 DB 实例
    db := &sql.DB{}  // 实际使用时需要正确初始化
    
    // 设置字段值
    fieldValue := v.FieldByName("DB")
    if fieldValue.CanSet() {
        fieldValue.Set(reflect.ValueOf(db))
    }
    
    fmt.Println(s.DB == db)  // true
}
```

### 2.6 创建新实例

```go
type User struct {
    Name string
}

func main() {
    // 获取类型
    t := reflect.TypeOf((*User)(nil)).Elem()  // User 类型
    
    // 创建新实例
    newValue := reflect.New(t)  // *User
    
    // 设置字段值
    newValue.Elem().FieldByName("Name").SetString("Alice")
    
    // 转换为实际类型
    user := newValue.Interface().(*User)
    fmt.Println(user.Name)  // Alice
}
```

### 2.7 类型匹配检查

```go
func main() {
    var db *sql.DB
    var logger *Logger
    
    dbType := reflect.TypeOf(db)      // *sql.DB
    loggerType := reflect.TypeOf(logger)  // *Logger
    
    // 检查类型是否可赋值
    fieldType := reflect.TypeOf((*sql.DB)(nil)).Elem()
    fmt.Println(dbType.AssignableTo(reflect.PtrTo(fieldType)))  // true
    
    // 检查接口实现
    var writer io.Writer
    writerType := reflect.TypeOf(&writer).Elem()
    fmt.Println(loggerType.Implements(writerType))  // 取决于 Logger 是否实现 io.Writer
}
```

### 2.8 实战练习：手写一个简单的注入器

```go
package main

import (
    "fmt"
    "reflect"
)

// 简单的对象图
type Graph struct {
    objects map[reflect.Type]interface{}
}

func NewGraph() *Graph {
    return &Graph{
        objects: make(map[reflect.Type]interface{}),
    }
}

// 提供对象
func (g *Graph) Provide(obj interface{}) {
    t := reflect.TypeOf(obj)
    g.objects[t] = obj
}

// 填充依赖
func (g *Graph) Populate(target interface{}) error {
    targetValue := reflect.ValueOf(target).Elem()
    targetType := reflect.TypeOf(target).Elem()
    
    for i := 0; i < targetType.NumField(); i++ {
        field := targetType.Field(i)
        fieldValue := targetValue.Field(i)
        
        // 检查是否有 inject 标签
        tag := field.Tag.Get("inject")
        if tag == "" {
            continue
        }
        
        // 查找匹配的对象
        fieldType := field.Type
        if obj, found := g.objects[fieldType]; found {
            fieldValue.Set(reflect.ValueOf(obj))
        } else {
            return fmt.Errorf("找不到类型 %s 的对象", fieldType)
        }
    }
    
    return nil
}

// 使用示例
type DB struct{}

type Service struct {
    DB *DB `inject:""`
}

func main() {
    graph := NewGraph()
    db := &DB{}
    graph.Provide(db)
    
    service := &Service{}
    graph.Populate(service)
    
    fmt.Println(service.DB == db)  // true
}
```

---

## 第三部分：Inject 库入门

### 3.1 三种注入方式详解

#### 方式 1: `inject:""` - 单例注入（最常用）

```go
type UserService struct {
    DB     *sql.DB    `inject:""`
    Logger *Logger    `inject:""`
}

func main() {
    var g inject.Graph
    
    // 提供基础对象
    db := &sql.DB{}
    logger := &Logger{}
    
    g.Provide(&inject.Object{Value: db})
    g.Provide(&inject.Object{Value: logger})
    
    // 创建服务（依赖会自动注入）
    service := &UserService{}
    g.Provide(&inject.Object{Value: service})
    
    // 填充所有依赖
    g.Populate()
    
    // 现在 service.DB 和 service.Logger 已经被自动注入
    fmt.Println(service.DB == db)      // true
    fmt.Println(service.Logger == logger)  // true
}
```

**特点：**
- 如果图中已有该类型的对象，会复用（单例）
- 如果没有，会自动创建新实例

#### 方式 2: `inject:"private"` - 私有实例

```go
type Service struct {
    Logger *Logger `inject:"private"`
}

func main() {
    var g inject.Graph
    
    service1 := &Service{}
    service2 := &Service{}
    
    g.Provide(&inject.Object{Value: service1})
    g.Provide(&inject.Object{Value: service2})
    g.Populate()
    
    // 每个 Service 都有自己独立的 Logger 实例
    fmt.Println(service1.Logger != service2.Logger)  // true
}
```

**特点：**
- 每个对象都获得独立的实例
- 不会加入全局对象图（其他对象无法使用）

#### 方式 3: `inject:"name"` - 命名依赖

```go
type Service struct {
    DB *sql.DB `inject:"main_db"`
}

func main() {
    var g inject.Graph
    
    // 提供命名对象
    mainDB := &sql.DB{}
    g.Provide(&inject.Object{
        Value: mainDB,
        Name:  "main_db",
    })
    
    service := &Service{}
    g.Provide(&inject.Object{Value: service})
    g.Populate()
    
    fmt.Println(service.DB == mainDB)  // true
}
```

**特点：**
- 通过名称精确匹配
- 适用于同一类型有多个实例的场景

### 3.2 完整示例：多层依赖

```go
package main

import (
    "fmt"
    "github.com/facebookgo/inject"
)

// 定义依赖关系
type DB struct {
    Name string
}

type Logger struct {
    Level string
}

type Cache struct {
    Size int
}

type UserService struct {
    DB     *DB     `inject:""`
    Logger *Logger `inject:""`
}

type OrderService struct {
    DB          *DB          `inject:""`
    Logger      *Logger      `inject:""`
    UserService *UserService `inject:""`
}

type PaymentService struct {
    DB          *DB          `inject:""`
    Logger      *Logger      `inject:""`
    OrderService *OrderService `inject:""`
}

func main() {
    var g inject.Graph
    
    // 1. 提供基础对象
    db := &DB{Name: "production_db"}
    logger := &Logger{Level: "info"}
    
    g.Provide(&inject.Object{Value: db})
    g.Provide(&inject.Object{Value: logger})
    
    // 2. 提供服务（依赖会自动注入）
    userService := &UserService{}
    orderService := &OrderService{}
    paymentService := &PaymentService{}
    
    g.Provide(&inject.Object{Value: userService})
    g.Provide(&inject.Object{Value: orderService})
    g.Provide(&inject.Object{Value: paymentService})
    
    // 3. 填充所有依赖
    if err := g.Populate(); err != nil {
        panic(err)
    }
    
    // 4. 验证依赖注入成功
    fmt.Println("UserService.DB:", userService.DB.Name)           // production_db
    fmt.Println("OrderService.UserService:", orderService.UserService == userService)  // true
    fmt.Println("PaymentService.OrderService:", paymentService.OrderService == orderService)  // true
    
    // 所有服务共享同一个 DB 和 Logger（单例）
    fmt.Println(userService.DB == orderService.DB)     // true
    fmt.Println(userService.Logger == orderService.Logger)  // true
}
```

### 3.3 快捷方法：Populate

```go
// 等价于上面的代码
func main() {
    db := &DB{Name: "production_db"}
    logger := &Logger{Level: "info"}
    
    userService := &UserService{}
    orderService := &OrderService{}
    
    // 一行代码完成所有操作
    if err := inject.Populate(db, logger, userService, orderService); err != nil {
        panic(err)
    }
}
```

---

## 第四部分：深入源码

### 4.1 Graph 结构体

```go
type Graph struct {
    Logger      Logger // 可选的日志记录器
    unnamed     []*Object  // 未命名的对象列表
    unnamedType map[reflect.Type]bool  // 类型到对象的映射（用于去重）
    named       map[string]*Object     // 命名对象映射
}
```

**设计思路：**
- `unnamed`: 存储所有未命名的对象（通过类型匹配）
- `named`: 存储所有命名对象（通过名称匹配）
- `unnamedType`: 快速检查某个类型是否已存在

### 4.2 Object 结构体

```go
type Object struct {
    Value        interface{}           // 对象的值
    Name         string                // 可选名称
    Complete     bool                  // 是否已完成注入
    Fields       map[string]*Object    // 被注入的字段及其对应的对象
    reflectType  reflect.Type          // 反射类型（缓存）
    reflectValue reflect.Value         // 反射值（缓存）
    private      bool                  // 是否为私有实例
    created      bool                  // 是否由 inject 创建
    embedded     bool                  // 是否为嵌入结构体
}
```

### 4.3 Provide 方法详解

```go
func (g *Graph) Provide(objects ...*Object) error {
    for _, o := range objects {
        // 1. 缓存反射信息
        o.reflectType = reflect.TypeOf(o.Value)
        o.reflectValue = reflect.ValueOf(o.Value)
        
        // 2. 验证：必须是结构体指针
        if o.Name == "" {
            if !isStructPtr(o.reflectType) {
                return fmt.Errorf("expected pointer to struct")
            }
            
            // 3. 检查类型是否已存在（防止重复）
            if g.unnamedType[o.reflectType] {
                return fmt.Errorf("duplicate type")
            }
            g.unnamedType[o.reflectType] = true
            g.unnamed = append(g.unnamed, o)
        } else {
            // 4. 命名对象存储到 named map
            if g.named[o.Name] != nil {
                return fmt.Errorf("duplicate name")
            }
            g.named[o.Name] = o
        }
    }
    return nil
}
```

**关键点：**
1. 验证对象类型（必须是结构体指针）
2. 防止重复（类型或名称）
3. 分别存储到 `unnamed` 或 `named`

### 4.4 Populate 方法详解

```go
func (g *Graph) Populate() error {
    // 第一轮：处理命名对象
    for _, o := range g.named {
        if o.Complete {
            continue
        }
        if err := g.populateExplicit(o); err != nil {
            return err
        }
    }
    
    // 第二轮：处理未命名对象（动态扩展）
    i := 0
    for {
        if i == len(g.unnamed) {
            break
        }
        o := g.unnamed[i]
        i++
        
        if o.Complete {
            continue
        }
        if err := g.populateExplicit(o); err != nil {
            return err
        }
    }
    
    // 第三轮：处理接口注入（需要先创建所有具体类型）
    for _, o := range g.unnamed {
        if err := g.populateUnnamedInterface(o); err != nil {
            return err
        }
    }
    
    for _, o := range g.named {
        if err := g.populateUnnamedInterface(o); err != nil {
            return err
        }
    }
    
    return nil
}
```

**为什么分三轮？**
1. **第一轮**：处理命名对象（优先级高）
2. **第二轮**：处理未命名对象，可能创建新对象（动态扩展 `unnamed` 列表）
3. **第三轮**：处理接口注入（需要先有所有具体类型）

### 4.5 populateExplicit 方法详解

这是核心方法，负责填充结构体字段：

```go
func (g *Graph) populateExplicit(o *Object) error {
    // 遍历结构体的所有字段
    for i := 0; i < o.reflectValue.Elem().NumField(); i++ {
        field := o.reflectValue.Elem().Field(i)      // 字段值
        fieldType := field.Type()                    // 字段类型
        fieldTag := o.reflectType.Elem().Field(i).Tag  // 字段标签
        
        // 1. 解析 inject 标签
        tag, err := parseTag(string(fieldTag))
        if err != nil {
            return err
        }
        
        // 2. 跳过没有 inject 标签的字段
        if tag == nil {
            continue
        }
        
        // 3. 检查字段是否可设置（必须是导出字段）
        if !field.CanSet() {
            return fmt.Errorf("unexported field")
        }
        
        // 4. 如果字段已有值，跳过
        if !isNilOrZero(field, fieldType) {
            continue
        }
        
        // 5. 处理命名注入
        if tag.Name != "" {
            existing := g.named[tag.Name]
            if existing == nil {
                return fmt.Errorf("named object not found")
            }
            field.Set(reflect.ValueOf(existing.Value))
            continue
        }
        
        // 6. 处理接口注入（延迟到第二轮）
        if fieldType.Kind() == reflect.Interface {
            continue
        }
        
        // 7. 处理指针类型注入
        if !isStructPtr(fieldType) {
            return fmt.Errorf("unsupported field type")
        }
        
        // 8. 查找现有对象（单例模式）
        if !tag.Private {
            for _, existing := range g.unnamed {
                if existing.private {
                    continue
                }
                if existing.reflectType.AssignableTo(fieldType) {
                    field.Set(reflect.ValueOf(existing.Value))
                    continue
                }
            }
        }
        
        // 9. 创建新对象
        newValue := reflect.New(fieldType.Elem())
        newObject := &Object{
            Value:   newValue.Interface(),
            private: tag.Private,
            created: true,
        }
        
        // 10. 将新对象加入图
        g.Provide(newObject)
        
        // 11. 赋值给字段
        field.Set(newValue)
    }
    
    o.Complete = true
    return nil
}
```

**关键步骤：**
1. 解析标签 → 2. 检查字段 → 3. 查找依赖 → 4. 创建或复用 → 5. 赋值

### 4.6 标签解析

```go
func parseTag(t string) (*tag, error) {
    found, value, err := structtag.Extract("inject", t)
    if !found {
        return nil, nil  // 没有 inject 标签
    }
    
    if value == "" {
        return &tag{}, nil  // inject:""
    }
    if value == "private" {
        return &tag{Private: true}, nil
    }
    if value == "inline" {
        return &tag{Inline: true}, nil
    }
    return &tag{Name: value}, nil  // inject:"name"
}
```

### 4.7 类型匹配逻辑

```go
// 检查类型是否可赋值
if existing.reflectType.AssignableTo(fieldType) {
    // 可以赋值，使用现有对象
    field.Set(reflect.ValueOf(existing.Value))
}

// AssignableTo 检查：
// - 类型完全匹配
// - 接口实现关系
// - 指针类型匹配
```

---

## 第五部分：实战演练

### 5.1 项目中的实际使用

基于你的代码库，我们来看一个完整的例子：

```go
// 1. 初始化组件（提供基础对象）
func (svr *InteractionSvr) initComponent() {
    db.Init(svr.config.DBConf[env.GetCID()]...)
    redis.Init(svr.config.RedisConfig[env.GetCID()]...)
    
    // 提供命名对象到注入图
    inject.Provide(&inject.Object{
        Value: db.GetDB("common_mysql"),
        Name:  "common_mysql",
    })
    inject.Provide(&inject.Object{
        Value: db.GetDB("vote_mysql"),
        Name:  "vote_mysql",
    })
    // ... 更多数据库和 Redis 连接
}

// 2. 定义服务（使用 inject 标签）
type ReplyService struct {
    CommonDB    *sql.DB `inject:"common_mysql"`
    CommentDB   *sql.DB `inject:"comment_mysql"`
    CommonRedis *redis.Client `inject:"common_redis"`
}

// 3. 填充依赖
func (svr *InteractionSvr) InitProcessor() {
    replyService := &ReplyService{}
    inject.Populate(replyService)
    
    // 现在 replyService 的所有字段都已自动注入
}
```

### 5.2 完整示例：构建一个微服务

```go
package main

import (
    "fmt"
    "github.com/facebookgo/inject"
)

// ========== 基础设施层 ==========
type Database struct {
    Name string
}

type Redis struct {
    Addr string
}

type Logger struct {
    Level string
}

// ========== 数据访问层 ==========
type UserDAO struct {
    DB     *Database `inject:""`
    Logger *Logger   `inject:""`
}

func (d *UserDAO) FindUser(id int) {
    fmt.Printf("UserDAO.FindUser: 使用数据库 %s, 日志级别 %s\n", 
        d.DB.Name, d.Logger.Level)
}

type OrderDAO struct {
    DB     *Database `inject:""`
    Logger *Logger   `inject:""`
}

// ========== 业务逻辑层 ==========
type UserService struct {
    UserDAO *UserDAO `inject:""`
    Logger  *Logger  `inject:""`
    Cache   *Redis   `inject:""`
}

func (s *UserService) GetUser(id int) {
    fmt.Printf("UserService.GetUser: 缓存地址 %s\n", s.Cache.Addr)
    s.UserDAO.FindUser(id)
}

type OrderService struct {
    OrderDAO   *OrderDAO   `inject:""`
    UserService *UserService `inject:""`
    Logger     *Logger     `inject:""`
}

// ========== 控制器层 ==========
type UserController struct {
    UserService *UserService `inject:""`
}

func (c *UserController) HandleRequest() {
    c.UserService.GetUser(123)
}

// ========== 主程序 ==========
func main() {
    var g inject.Graph
    
    // 1. 提供基础设施
    db := &Database{Name: "production_db"}
    redis := &Redis{Addr: "localhost:6379"}
    logger := &Logger{Level: "info"}
    
    g.Provide(&inject.Object{Value: db})
    g.Provide(&inject.Object{Value: redis})
    g.Provide(&inject.Object{Value: logger})
    
    // 2. 提供业务对象（依赖会自动注入）
    userDAO := &UserDAO{}
    orderDAO := &OrderDAO{}
    userService := &UserService{}
    orderService := &OrderService{}
    userController := &UserController{}
    
    g.Provide(&inject.Object{Value: userDAO})
    g.Provide(&inject.Object{Value: orderDAO})
    g.Provide(&inject.Object{Value: userService})
    g.Provide(&inject.Object{Value: orderService})
    g.Provide(&inject.Object{Value: userController})
    
    // 3. 填充所有依赖
    if err := g.Populate(); err != nil {
        panic(err)
    }
    
    // 4. 使用
    userController.HandleRequest()
    
    // 验证：所有对象共享同一个基础设施（单例）
    fmt.Println("\n=== 验证单例 ===")
    fmt.Println("UserDAO.DB == OrderDAO.DB:", userDAO.DB == orderDAO.DB)  // true
    fmt.Println("UserService.Logger == OrderService.Logger:", 
        userService.Logger == orderService.Logger)  // true
}
```

### 5.3 处理循环依赖

**问题：** 如果 A 依赖 B，B 依赖 A，会发生什么？

```go
type A struct {
    B *B `inject:""`
}

type B struct {
    A *A `inject:""`
}

func main() {
    var g inject.Graph
    
    a := &A{}
    b := &B{}
    
    g.Provide(&inject.Object{Value: a})
    g.Provide(&inject.Object{Value: b})
    
    // 这会成功！因为 inject 会先创建对象，再填充依赖
    g.Populate()
    
    fmt.Println(a.B == b)  // true
    fmt.Println(b.A == a)  // true
}
```

**原理：** Inject 会先创建所有对象（字段为 nil），然后再填充依赖，所以可以处理循环依赖。

### 5.4 接口注入

```go
type Writer interface {
    Write(string)
}

type FileWriter struct {
    Path string
}

func (f *FileWriter) Write(s string) {
    fmt.Printf("写入文件 %s: %s\n", f.Path, s)
}

type Service struct {
    Writer Writer `inject:""`  // 注入接口
}

func main() {
    var g inject.Graph
    
    // 提供具体实现
    fileWriter := &FileWriter{Path: "/tmp/log.txt"}
    g.Provide(&inject.Object{Value: fileWriter})
    
    service := &Service{}
    g.Provide(&inject.Object{Value: service})
    
    g.Populate()
    
    service.Writer.Write("hello")  // 写入文件 /tmp/log.txt: hello
}
```

**注意：** 接口注入在第二轮处理，确保所有具体类型都已创建。

---

## 第六部分：常见问题

### Q1: 为什么必须是结构体指针？

**A:** 因为需要通过反射修改字段值，只有指针才能修改指向的结构体。

```go
// ❌ 错误
type Service struct {
    DB sql.DB  // 值类型，无法注入
}

// ✅ 正确
type Service struct {
    DB *sql.DB  // 指针类型，可以注入
}
```

### Q2: 私有字段为什么不能注入？

**A:** Go 的反射机制限制：只有导出字段（首字母大写）才能通过反射设置。

```go
type Service struct {
    db *sql.DB `inject:""`  // ❌ 小写，无法注入
    DB *sql.DB `inject:""`  // ✅ 大写，可以注入
}
```

### Q3: 如何调试依赖注入问题？

**A:** 使用 Graph 的 Logger：

```go
type MyLogger struct{}

func (l *MyLogger) Debugf(format string, v ...interface{}) {
    fmt.Printf("[DEBUG] "+format+"\n", v...)
}

var g inject.Graph
g.Logger = &MyLogger{}
// 现在会输出详细的注入过程
```

### Q4: 性能影响大吗？

**A:** 
- **启动时：** 反射有开销，但通常只在初始化时执行一次
- **运行时：** 没有影响，所有依赖在启动时已注入完成
- **建议：** 对于性能敏感的场景，考虑使用代码生成工具（如 wire）

### Q5: 如何处理可选依赖？

**A:** Inject 不支持可选依赖，如果找不到依赖会报错。可以这样处理：

```go
type Service struct {
    OptionalDB *sql.DB `inject:"optional_db"`  // 如果不存在会报错
}

// 解决方案：提供 nil 值或默认实现
g.Provide(&inject.Object{
    Value: (*sql.DB)(nil),
    Name:  "optional_db",
})
```

### Q6: 如何测试？

**A:** 在测试中提供 mock 对象：

```go
func TestUserService(t *testing.T) {
    var g inject.Graph
    
    // 提供 mock 对象
    mockDB := &MockDB{}
    mockLogger := &MockLogger{}
    
    g.Provide(&inject.Object{Value: mockDB})
    g.Provide(&inject.Object{Value: mockLogger})
    
    service := &UserService{}
    g.Provide(&inject.Object{Value: service})
    g.Populate()
    
    // 现在 service 使用的是 mock 对象
    // 可以进行测试
}
```

---

## 总结

### 核心原理

1. **反射扫描：** 通过 `reflect` 包扫描结构体字段和标签
2. **对象图管理：** 维护一个对象图，存储所有可用的对象
3. **类型匹配：** 通过类型或名称匹配依赖
4. **自动创建：** 如果找不到依赖，自动创建新实例
5. **递归注入：** 递归处理依赖的依赖

### 关键要点

- ✅ 使用结构体指针
- ✅ 字段必须可导出
- ✅ 理解三种注入方式
- ✅ 注意依赖顺序（命名依赖需要先 Provide）
- ✅ 接口注入在第二轮处理

### 进阶学习

1. 阅读 `github.com/facebookgo/inject` 源码
2. 尝试实现一个简化版的 inject
3. 了解其他 DI 框架（wire, fx, dig）
4. 学习设计模式：依赖注入、控制反转

---

**祝你学习愉快！** 🎉

