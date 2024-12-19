package pipeline

import "context"

type Pipeline interface {
	AddStage(stage Stage, opt *Opt)
	Start() error
	Stop() error
	Output() <-chan Message
	Input() chan<- Message
}

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

func (c *ConcurrentPipeline) Start(ctx context.Context) error {
	for _, worker := range c.workers {
		err := worker.Start(ctx)
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
