package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"world-of-wisdom/internal/pow"
	"world-of-wisdom/internal/quotes"
)

type Server struct {
	cfg           *Config
	powService    pow.Service
	quoteService  quotes.Service
	workerService WorkerService

	connectionLimit chan struct{}
}

func NewServer(cfg *Config, challengeServ pow.Service, quoteServ quotes.Service, workerService WorkerService) *Server {
	return &Server{
		cfg:             cfg,
		powService:      challengeServ,
		quoteService:    quoteServ,
		workerService:   workerService,
		connectionLimit: make(chan struct{}, cfg.ConnectionsLimit),
	}
}

func (s *Server) connWorker(
	ctx context.Context,
	conns chan net.Conn,
	stop chan struct{}) {
	for {
		select {
		case <-ctx.Done():
			return
		case conn, ok := <-conns:
			if !ok {
				return
			}
			s.handleConn(ctx, conn)
		case <-stop:
			return
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

	var wg sync.WaitGroup
	defer func() {
		l.Close()
		wg.Wait()
	}()

	connsQueue := make(chan net.Conn, s.cfg.ConnectionsLimit)
	listenerErr := make(chan error)

	wg.Add(1)
	go func() {
		s.workerService.Start(ctx, connsQueue, s.connWorker)
		wg.Done()
	}()

	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				if netErr, ok := err.(net.Error); ok {
					if netErr.Timeout() {
						log.Printf("accept timeout: %v", netErr)
						continue
					}
				}

				log.Printf("fatal error accepting client connection: %v", err)
				listenerErr <- err
				return
			}

			select {
			case <-ctx.Done():
				return
			case s.connectionLimit <- struct{}{}:
				connsQueue <- conn
			default:
				s.handleConnRefused(conn)
			}
		}
	}()

	select {
	case <-ctx.Done():
		return nil
	case err := <-listenerErr:
		return err
	}
}
