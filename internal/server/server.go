package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"
	"world-of-wisdom/internal/pow"
	"world-of-wisdom/internal/quotes"
)

type Server struct {
	cfg          *Config
	powService   pow.Service
	quoteService quotes.Service

	wg          sync.WaitGroup
	workerCount atomic.Int64
	sem         chan struct{}
}

func NewServer(cfg *Config, challengeServ pow.Service, quoteServ quotes.Service) *Server {
	return &Server{
		cfg:          cfg,
		powService:   challengeServ,
		quoteService: quoteServ,
		sem:          make(chan struct{}, cfg.ConnectionsLimit),
	}
}

func (s *Server) worker(ctx context.Context, conns chan net.Conn) {
	defer func() {
		s.wg.Done()
		s.workerCount.Add(-1)
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case conn, ok := <-conns:
			if !ok || conn == nil {
				return
			}
			s.handleConn(ctx, conn)
		}
	}
}

func (s *Server) workerManager(ctx context.Context, conns chan net.Conn) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for i := uint64(0); i < s.cfg.MinWorkers; i++ {
		s.wg.Add(1)
		go s.worker(ctx, conns)
	}

	for {
		select {
		case <-ctx.Done():
			s.wg.Wait()
			return
		case <-ticker.C:
			queueLength := len(conns)
			workerCount := int(s.workerCount.Load())

			if queueLength > 0 && workerCount < int(s.cfg.MaxWorkers) {
				s.wg.Add(1)
				s.workerCount.Add(1)
				go s.worker(ctx, conns)
			}

			if queueLength == 0 && workerCount > int(s.cfg.MinWorkers) {
				conns <- nil // empty value to close worker
			}
		}
	}
}

func (s *Server) Listen(ctx context.Context) error {
	addr := fmt.Sprintf(":%v", s.cfg.Port)
	log.Printf("starting tcp server on addr %v", addr)

	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	defer func() {
		l.Close()
		s.wg.Wait()
	}()

	connsQueue := make(chan net.Conn, s.cfg.ConnectionsLimit)
	go s.workerManager(ctx, connsQueue)

	go func() {
		for {
			conn, err := l.Accept()
			select {
			case <-ctx.Done():
				return
			default:
				if err != nil {
					log.Printf("error accepting client connection: %v", err)
					continue
				}
			}

			select {
			case s.sem <- struct{}{}:
				connsQueue <- conn
			default:
				s.handleConnRefused(conn)
			}
		}
	}()

	<-ctx.Done()
	return nil
}
