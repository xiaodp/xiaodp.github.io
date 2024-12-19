---
layout: post
title: "Golang并发模式之Pipeline"
date:   2024-12-14
tags: [go]
comments: true
author: xiaodp
toc: true
---

本文介绍Golang Pipeline并发模式
<!-- more -->

## 引子

先考虑一个简单的问题。一个任务包含4个步骤，需要依次执行：

![image.png](https://raw.githubusercontent.com/xiaodp/xiaodp.github.io/master/2024-12-14-concurreny_in_golang_pipeline/image.png)

![image.png](https://raw.githubusercontent.com/xiaodp/xiaodp.github.io/master/2024-12-14-concurreny_in_golang_pipeline/image%201.png)

输入input，依次经过step1，step2，step3，step4后，处理完毕，输出output

```yaml
var input 
for step in {step1~4}
    output = step(input)
    input = output
```

现在需要处理多个input, 最直接的方法如下：

```yaml
var outputs
for input in {input1, input2, input3}
	for step in {step1~4}
	    output = step(input)
	    input = output
  outputs.add(output)
```

执行过程如下图：

![image.png](https://raw.githubusercontent.com/xiaodp/xiaodp.github.io/master/2024-12-14-concurreny_in_golang_pipeline/image%202.png)

当然，你会想到使用并发来优化性能。如

从input维度，每个input开启一个协程来处理，伪代码如下：

```go
var outputs
for input in {input1, input2, input3}
  input := input
	go func(){
	  	for step in {step1~4}
	    output = step(input)
	    input = output
      outputs.add(output) 	
	}()

```

或者，从step维度，在每个step，多个协程处理各个input。注意，后续步骤需要等待前序步骤完成后才能执行，因此，每个step需要等到所有处理input的协程执行完毕。（这里可以使用waitgroup），伪代码如下：

```yaml
var inputs = {input1, input2, input3} 
var outputs
var wg sync.WaitGroup{}
for step in {step1~4}
  for i, input in {input1, input2, input3} 
    i := i
    input := input
	  go func(){
	    wg.Add(1)
	    defer wg.Done()
	    outputs[i] = step(input)
		}()
	wg.Wait() // 这里需要等待所有input已经完成该step的处理
	inputs = outputs
```

很棒，以上并发方式都能够提升效率。但现在需要增减步骤，或者调整步骤顺序，那么上述代码将变得不好维护。

## Pipeline模式介绍

现在我们来介绍另一种方式：Pipeline模式

![image.png](https://raw.githubusercontent.com/xiaodp/xiaodp.github.io/master/2024-12-14-concurreny_in_golang_pipeline/image%203.png)

在Go语言中，管道（Pipeline）是一种编程模式，用于将数据处理分成多个独立的阶段，每个阶段通过channel进行通信。每个阶段通常在一个独立的goroutine中运行，数据从一个阶段流向下一个阶段，形成一个数据处理流水线。所谓“阶段”，可以理解为组成pipeline的单元，负责输入→ 处理 → 输出。

这种模式的主要优点包括：

1. **并发处理**：每个阶段可以在独立的goroutine中并发执行，提高处理效率。
2. **模块化**：每个阶段可以独立开发、测试和维护，增强代码的可读性和可维护性。
3. **灵活性**：可以轻松地组合、修改和重用不同的阶段，适应不同的需求。

现在，我们通过一个简单的demo来加深理解。

首先，我们需要一些数据输入：

```go
	generator := func(
		done <-chan interface{},
		integers ...int,
	) <-chan int {
		intStream := make(chan int)
		go func() {
			defer close(intStream)
			for _, i := range integers {
				select {
				case <-done:
					return
				case intStream <- i:
				}
			}
		}()
		return intStream
	}
```

这个函数简单地启动了一个 goroutine，它将可变参数integers列表放入channel，而函数本身只返回该channel。在这个通道上生成的值将作为后续阶段的输入。

此外，还在函数中传递了一个 `done` 通道，用于优雅地退出（等待当前处理结束后再退出）生成过程，这种模式也被称为 [Poison Pill Pattern](https://java-design-patterns.com/patterns/poison-pill/)。

注意，这里将通道存放的数据类型定义为struct{}的原因，涉及一个小知识点，如下

> **`chan struct{}`**
> 
> - **用途**:
>     - 用于仅仅传递信号，而不携带任何数据。
>     - `struct{}` 是 Go 中最小的类型，零大小（zero-size type），不占用内存。
>     - 常用于同步操作或信号通知，比如关闭某个 goroutine、广播事件等。
> - **内存开销**:
>     - 几乎为零，因为 `struct{}` 不占用内存。
>     - 性能更优，尤其是在高并发场景下。

接下来使用相同的方式定义两个处理函数add、multiply，即定义不同的**阶段**

```go
	add := func(
		done <-chan **struct**{},
		intStream <-chan int,
		additive int,
	) <-chan int {
		addedStream := make(chan int)
		go func() {
			defer close(addedStream)
			for i := range intStream {
				select {
				case <-done:
					return
				case addedStream <- i + additive:
				}
			}
		}()
		return addedStream
	}
```

```go
	multiply := func(
		done <-chan **struct**{},
		intStream <-chan int,
		multiplier int,
	) <-chan int {
		multipliedStream := make(chan int)
		go func() {
			defer close(multipliedStream)
			for i := range intStream {
				select {
				case <-done:
					return
				case multipliedStream <- i * multiplier:
				}
			}
		}()
		return multipliedStream
	}
```

这里的函数特征：

1. 使用chan作为输入参数，从chan中取出值，做add/multipy后放入新的结果通道中，并返回结果通道
2. 使用done chan，结合for select，实现优雅退出。在一些资料中，这也被称为防止协程泄漏的模式（Concurrency in Go）
3. 使用协程，实现并发

接下来，就可以任意组合阶段，实现简单的数值处理的pipeline

```go
done := make(chan **struct**{})
defer close(done)
intStream := gen(done, 1, 2, 3, 4)
pipeline := multiply(done, add(done, multiply(done, intStream, 2), 1), 2)
for v := range pipeline {
	fmt.Println(v)
}
```

- 完整代码
  
    ```go
    
    package main
    
    import "fmt"
    
    func gen(done <-chan struct{}, integers ...int) chan int {
    	out := make(chan int)
    	go func() {
    		defer close(out) // 如果没有close，会发生色什么
    		for _, integer := range integers {
    			select {
    			case <-done:
    				return
    			case out <- integer:
    			}
    		}
    	}()
    	return out
    }
    
    func add(done <-chan struct{}, in <-chan int, additive int) chan int {
    	out := make(chan int)
    	go func() {
    		defer close(out)
    		for integer := range in {
    			select {
    			case <-done:
    				return
    			case out <- integer + additive:
    			}
    		}
    	}()
    	return out
    }
    
    func mul(done <-chan struct{}, in <-chan int, multiplier int) chan int {
    	out := make(chan int)
    	go func() {
    		defer close(out)
    		for integer := range in {
    			select {
    			case <-done:
    				return
    			case out <- integer * multiplier:
    			}
    		}
    	}()
    	return out
    }
    
    func main() {
    	done := make(chan struct{})
    	defer close(done)
    	got := mul(done, add(done, gen(done, 1, 2, 3, 4), 5), 6)
    	for i := range got {
    		fmt.Println(i)
    	}
    }
    
    ```
    

## 更加通用的Pipeline库

todo：代码地址

现在我们来设计一个更加实用的Pipeline库。这个库需要包含如下要素

1. Stage。Stage结构接受一个某种type的参数，经过处理后，返回一个或多个相同type的结果
2. 参数类型在使用时指定，因此在库中需要抽象为interface{}
3. 初始化时，能够指定Stage的顺序。
4. 针对某一个阶段而言，可能需要处理多个输入。因此需要能够设置Stage处理函数的并发数
5. 整个Pipeline需要有启动和停止机制

### Stage

先说Stage

```go
type Stage interface {
	Process(input interface{}) ([]interface{}, error)
}
```

Stage为一个抽象的接口。`Process()`接受一个输入参数，处理完成后返回多个结果。（这是常见的，比如将一段文本拆分为多行）。Stage具备处理数据的能力，具体怎么处理，由其实现来决定。

我们将Process方法的参数进一步抽象，更加明确了我们的参数语义：

```go

type Message interface{}

type Stage interface {
	Process(input Message) ([]Message, error)
}
```

为什么不是`Process(inputs []Message ) ([]Message, error)`处理多个输入、得到多个输出？这里保持最小化单元，多个输入可以交由多个协程来处理。因此我们需要封装一个管理本阶段并发执行的中间层——`StageWorker`

```go
type Opt struct {
	Parallel int
}

type StageWorker struct {
	wg    sync.WaitGroup
	stage Stage

	input  chan Message
	output chan Message

	parallel int
}

func NewStageWorker(stage Stage, input chan Message, output chan Message, opt *Opt) *StageWorker {
	return &StageWorker{
		stage:    stage,
		input:    input,
		output:   output,
		parallel: opt.Parallel,
	}
}
```

`StageWorker` 负责启动本阶段需要执行的处理，并管理并发数

```go
func (s *StageWorker) Start() error {
	if s.input == nil || s.output == nil {
		return fmt.Errorf("not initialized")
	}
	for i := 0; i < s.parallel; i++ {
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			fmt.Println("start to work")
			for message := range s.input {
				results, err := s.stage.Process(message)
				if err != nil {
					log.Println("process error", err)
					return err
				}
				for _, result := range results {
					s.output <- result
				}
			}
		}()
	}
	return nil
}

func (s *StageWorker) WaitStop() {
	s.wg.Wait()
}

```

### Pileline

接下来然后是Pipeline

```go
type Pipeline interface {
  // 启动前向pipeline中添加Stage，并且将stage按照添加顺序组成一条pipeline。这里定义了一个选项参数，用于指定Stage的并发数
	AddStage(stage Stage, opt *Opt)
	// 启动pipeline
	Start() error 
	// 停止pipeline。这里仍然需要等待正在被处理的message处理完毕后才能停止。 
	Stop() error
	// 传入待处理的数据
	Input() chan<- Message
	// 输出pipeline处理结果
	Output() <-chan Message
}
```

注意这里Input、Output返回的参数分别为只写channel、只读channel。主要是为了明确：对于一个pipeline，外部只能往Input()中填入待处理的数据，同理，外部只能从Output()中读取已处理的数据。

接下来实现一个具体的pipeline。

```go

type ConcurrentPipeline struct {
	workers []*StageWorker
}

func NewConcurrentPipeline() *ConcurrentPipeline {
	return new(ConcurrentPipeline)
}

func (c *ConcurrentPipeline) AddStage(stage Stage, opt *Opt) {
	if opt == nil || opt.Parallel == 0 {
		opt = &Opt{Parallel: 1}
	}

	var input, output chan Message
	output = make(chan Message, 10)
	if len(c.workers) == 0 {
		input = make(chan Message, 10)
	} else {
		input = c.workers[len(c.workers)-1].output
	}

	worker := NewStageWorker(stage, input, output, opt)
	c.workers = append(c.workers, worker)

}

func (c *ConcurrentPipeline) Start() error {
	for _, worker := range c.workers {
		err := worker.Start()
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *ConcurrentPipeline) Stop() {
	for _, worker := range c.workers {
		close(worker.input)
		worker.WaitStop()
	}
	close(c.workers[len(c.workers)-1].output)
}

func (c *ConcurrentPipeline) Input() chan<- Message {
	if len(c.workers) == 0 {
		return nil
	}
	return c.workers[0].input
}

func (c *ConcurrentPipeline) Output() <-chan Message {
	if len(c.workers) == 0 {
		return nil
	}
	return c.workers[len(c.workers)-1].output
}
```

如果调用`p.Stop()`, 必须等到所有stage处理完，这并优雅。我们引入context，利用context的Done来提前结束（这里仍然需要等待本阶段正在处理的数据处理结束）

```go
func (s *StageWorker) Start(ctx context.Context) error {
	if s.input == nil || s.output == nil {
		return fmt.Errorf("not initialized")
	}
	for i := 0; i < s.parallel; i++ {
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			for message := range s.input {
				select {
				case <-ctx.Done():
					log.Println("process canceled")
					return
				default:
					results, err := s.stage.Process(message)
					if err != nil {
						log.Println("process error", err)
						return
					}
					for _, result := range results {
						s.output <- result
					}
				}
			}
		}()
	}
	return nil
}
```

```go
func (c *ConcurrentPipeline) Start(ctx context.Context) error {
	for _, worker := range c.workers {
		err := worker.Start(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}
```

现在，我们可以写一个demo

```go

type MultiplyTenSlow struct{}

func (m MultiplyTenSlow) Process(result mypipeline.Message) ([]mypipeline.Message, error) {
	time.Sleep(1 * time.Second)
	number := result.(int)
	return []mypipeline.Message{number * 10, number * 10}, nil
}

type MultiplyHundredSlow struct{}

func (m MultiplyHundredSlow) Process(result mypipeline.Message) ([]mypipeline.Message, error) {
	time.Sleep(1 * time.Second)
	number := result.(int)
	return []mypipeline.Message{number * 100, number * 100}, nil
}

type DivideThreeSlow struct{}

func (m DivideThreeSlow) Process(result mypipeline.Message) ([]mypipeline.Message, error) {
	time.Sleep(1 * time.Second)
	number := result.(int)
	return []mypipeline.Message{number / 3}, nil
}

func main() {

	p := mypipeline.NewConcurrentPipeline()
	p.AddStage(MultiplyHundredSlow{}, &mypipeline.Opt{Parallel: 2})
	p.AddStage(MultiplyTenSlow{}, &mypipeline.Opt{Parallel: 2})
	p.AddStage(DivideThreeSlow{}, &mypipeline.Opt{Parallel: 2})

	ctx := context.Background()
	ctx, _ = context.WithTimeout(ctx, 2*time.Second)
	if err := p.Start(ctx); err != nil {
		log.Println(err)
	}

	for i := 1; i <= 3; i++ {
		p.Input() <- i
	}

	go func() {
	  // 1
		for number := range p.Output() {
			fmt.Println(number)
		}
	}()

	p.Stop()
	// 这里是为了让1处有足够的时间读取并打印数据
	time.Sleep(time.Second * 1)
}
```

```go
666
666
333
333
333
666
666
333
1000
1000
1000
1000
```

整个过程如下

![image.png](https://raw.githubusercontent.com/xiaodp/xiaodp.github.io/master/2024-12-14-concurreny_in_golang_pipeline/image%204.png)

思考：

在ConcurrentPipeline的Stop方法中，我们先close(worker.input)，再worker.WaitStop()，最后close(c.workers[len(c.workers)-1].output)。

即关闭每一个stage的input channel，顺道关闭最后一个output channel
close(worker.input)是否必须的，如果不这么做，会发生什么？

close(c.workers[len(c.workers)-1].output)是否必须的，如果不这么做，会发生什么？

```go
func (c *ConcurrentPipeline) Stop() {
	for _, worker := range c.workers {
		close(worker.input)
		worker.WaitStop()
	}
	close(c.workers[len(c.workers)-1].output)
}
```

### 参考文献

[https://ketansingh.me/posts/pipeline-pattern-in-go-part-2/](https://ketansingh.me/posts/pipeline-pattern-in-go-part-2/)

### 完整代码
https://github.com/xiaodp/xiaodp.github.io/master/src/golang