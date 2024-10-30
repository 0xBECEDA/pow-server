package server

import (
	"context"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type WorkerService interface {
	Start(ctx context.Context, conns chan net.Conn, workerFn WorkerFn)
}

type WorkerFn func(ctx context.Context, conn chan net.Conn, stop chan struct{})

type workerManager struct {
	minWorkers  uint64
	maxWorkers  uint64
	workerCount atomic.Int64

	wg sync.WaitGroup
}

func NewWorkerManager(minWorkers, maxWorkers uint64) WorkerService {
	return &workerManager{
		minWorkers: minWorkers,
		maxWorkers: maxWorkers,
	}
}

func (w *workerManager) Start(ctx context.Context, conns chan net.Conn, workerFn WorkerFn) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	stopChan := make(chan struct{}, w.maxWorkers)
	for i := uint64(0); i < w.minWorkers; i++ {
		w.runWorker(ctx, conns, stopChan, workerFn)
	}

	for {
		select {
		case <-ctx.Done():
			w.wg.Wait()
			return
		case <-ticker.C:
			queueLength := len(conns)
			workerCount := int(w.workerCount.Load())

			if queueLength > 0 && workerCount < int(w.maxWorkers) {
				w.runWorker(ctx, conns, stopChan, workerFn)
			}

			if queueLength == 0 && workerCount > int(w.minWorkers) {
				stopChan <- struct{}{} // empty value to close random worker
			}
		}
	}
}

func (w *workerManager) runWorker(ctx context.Context, conns chan net.Conn, stopChan chan struct{}, workerFn WorkerFn) {
	w.wg.Add(1)
	w.workerCount.Add(1)

	go func() {
		defer func() {
			w.wg.Done()
			w.workerCount.Add(-1)
		}()
		workerFn(ctx, conns, stopChan)
	}()
}
