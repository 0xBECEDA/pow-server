package main

import (
	"context"
	"log"
	"world-of-wisdom/internal/pow"
	"world-of-wisdom/internal/quotes"
	"world-of-wisdom/internal/server"
	"world-of-wisdom/internal/storage"
)

func main() {
	cfg := &server.Config{}
	if err := cfg.Load(); err != nil {
		log.Fatalf("failed start server %s", err.Error())
	}

	ctx := context.Background()
	db := storage.NewStorage(ctx, cfg.ChallengeTTL)

	tcpServer := server.NewServer(
		cfg,
		pow.NewChallengeService(db),
		quotes.NewService(),
		server.NewWorkerManager(cfg.MinWorkers, cfg.MaxWorkers),
	)

	if err := tcpServer.Listen(ctx); err != nil {
		log.Fatal(err)
	}
}
