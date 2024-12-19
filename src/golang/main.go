package main

type MultiplyTenSlow struct{}

func (m MultiplyTenSlow) Process(result pipeline.Message) ([]pipeline.Message, error) {
	time.Sleep(1 * time.Second)
	number := result.(int)
	return []pipeline.Message{number * 10, number * 10}, nil
}

type MultiplyHundredSlow struct{}

func (m MultiplyHundredSlow) Process(result pipeline.Message) ([]pipeline.Message, error) {
	time.Sleep(1 * time.Second)
	number := result.(int)
	return []pipeline.Message{number * 100, number * 100}, nil
}

type DivideThreeSlow struct{}

func (m DivideThreeSlow) Process(result pipeline.Message) ([]pipeline.Message, error) {
	time.Sleep(1 * time.Second)
	number := result.(int)
	return []pipeline.Message{number / 3}, nil
}

func main() {

	p := pipeline.NewConcurrentPipeline()
	p.AddStage(MultiplyHundredSlow{}, &pipeline.Opt{Parallel: 4})
	p.AddStage(MultiplyTenSlow{}, &pipeline.Opt{Parallel: 4})
	p.AddStage(DivideThreeSlow{}, &pipeline.Opt{Parallel: 4})

	ctx := context.Background()
	ctx, _ = context.WithTimeout(ctx, 10*time.Second)
	if err := p.Start(ctx); err != nil {
		log.Println(err)
	}

	for i := 1; i <= 3; i++ {
		p.Input() <- i
	}

	go func() {
		for number := range p.Output() {
			fmt.Println(number)
		}
	}()

	p.Stop()
	time.Sleep(time.Second * 1)
}

```