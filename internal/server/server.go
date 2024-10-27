package server

import (
	"context"
	"net"
	"sync"
	"time"
	"world-of-wisdom/internal/pow"
	"world-of-wisdom/internal/quotes"
)

type Server struct {
	powService   pow.Service
	quoteService quotes.Service

	writeDeadline time.Duration
	readDeadline  time.Duration
	wg            sync.WaitGroup
	sem           chan struct{}
}

func NewServer(
	hr pow.Service,
	quoteService quotes.Service,
	cfg *Config) *Server {
	return &Server{
		powService:    hr,
		quoteService:  quoteService,
		writeDeadline: cfg.WriteTimeout,
		readDeadline:  cfg.ReadTimeout,
		sem:           make(chan struct{}, cfg.ConnectionsLimit),
	}
}

func (s *Server) Listen(ctx context.Context, address string) error {
	l, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	defer func() {
		s.wg.Wait()
		l.Close()
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case s.sem <- struct{}{}:
			clientConn, err := l.Accept()
			if err != nil {
				<-s.sem
				continue
			}

			s.wg.Add(1)
			go s.handle(ctx, clientConn)
		}
	}
}
