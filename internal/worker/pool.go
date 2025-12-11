package worker

import (
	"sync"
)

type Pool struct {
	workers   int
	workQueue chan func()
	wg        sync.WaitGroup
}

func NewPool(workers int) *Pool {
	if workers <= 0 {
		workers = 4
	}

	p := &Pool{
		workers:   workers,
		workQueue: make(chan func(), workers*2),
	}

	// Start worker goroutines
	for i := 0; i < workers; i++ {
		p.wg.Add(1)
		go p.worker()
	}

	return p
}

func (p *Pool) worker() {
	defer p.wg.Done()
	for work := range p.workQueue {
		if work != nil {
			work()
		}
	}
}

func (p *Pool) Submit(work func()) {
	p.workQueue <- work
}

func (p *Pool) Wait() {
	close(p.workQueue)
	p.wg.Wait()
}
