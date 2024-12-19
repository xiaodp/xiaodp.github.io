

package pipeline

import (
	"context"
	"fmt"
	"log"
	"sync"
)

type Message interface{}

type Stage interface {
	Process(message Message) ([]Message, error)
}

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

func (s *StageWorker) WaitStop() {
	s.wg.Wait()
}
