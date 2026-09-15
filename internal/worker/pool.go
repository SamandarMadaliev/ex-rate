package worker

import (
	"context"
	"errors"
	"log"
	"sync"
)

var ErrPoolStopped = errors.New("worker pool is stopped")

type Job func(ctx context.Context)

type Pool struct {
	jobs chan Job
	wg   sync.WaitGroup

	mu      sync.RWMutex
	stopped bool
}

func NewPool(count, bufferSize int) *Pool {
	p := &Pool{
		jobs: make(chan Job, bufferSize),
	}

	p.wg.Add(count)
	for i := 0; i < count; i++ {
		go p.worker()
	}

	return p
}

func (p *Pool) worker() {
	defer p.wg.Done()

	for job := range p.jobs {
		runJob(job)
	}
}

func runJob(job Job) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("worker: job panicked: %v", r)
		}
	}()

	job(context.Background())
}

func (p *Pool) Submit(job Job) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.stopped {
		return ErrPoolStopped
	}

	p.jobs <- job
	return nil
}

func (p *Pool) Stop() {
	p.mu.Lock()
	if p.stopped {
		p.mu.Unlock()
		return
	}
	p.stopped = true
	close(p.jobs)
	p.mu.Unlock()

	p.wg.Wait()
}
