package server

import (
	"context"
	"crypto/tls"
	"github.com/stretchr/testify/assert"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestWorkerManager(t *testing.T) {
	const (
		minWorkers = 1
		maxWorkers = 3
	)

	workersCnt := atomic.Int64{}
	workerFn := func(ctx context.Context, conn chan net.Conn, stop chan struct{}) {
		workersCnt.Add(1)
		for {
			select {
			case <-conn:
			case <-stop:
				workersCnt.Add(-1)
				return
			case <-ctx.Done():
				workersCnt.Add(-1)
				return
			}
		}
	}

	conns := make(chan net.Conn, maxWorkers*3)

	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}

	manager := NewWorkerManager(minWorkers, maxWorkers)

	wg.Add(1)
	go func() {
		manager.Start(ctx, conns, workerFn)
		wg.Done()
	}()

	// 1. no connections in channel, keep minimal amount of workers
	time.Sleep(time.Second * 2)
	assert.Equal(t, int64(minWorkers), workersCnt.Load())

	// 2. increase load to manager till max
	connsCtx, cancel2 := context.WithCancel(context.Background())
	go func() {
		for {
			select {
			case <-connsCtx.Done():
			case conns <- &tls.Conn{}:
			}
		}
	}()

	var maxWorkersAchieved bool
	for i := 0; i < 10; i++ {
		if workersCnt.Load() == maxWorkers {
			maxWorkersAchieved = true
			break
		}
		time.Sleep(1 * time.Second)
	}
	cancel2()
	assert.True(t, maxWorkersAchieved)

	// 3. no new connections -> num of workers must be decreased
	var workersAmountDecreased bool
	for i := 0; i < 10; i++ {
		if workersCnt.Load() < maxWorkers {
			workersAmountDecreased = true
			break
		}

		time.Sleep(1 * time.Second)
	}

	assert.True(t, workersAmountDecreased)

	cancel()
	wg.Wait()
}
