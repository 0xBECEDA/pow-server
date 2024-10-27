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

	db := storage.NewStorage(cfg.KeyTTL)

	tcpServer := server.NewServer(
		pow.NewChallengeService(db),
		quotes.NewService(),
		cfg)

	addr := ":" + cfg.Port
	log.Printf("starting tcp server on addr %v", addr)

	ctx := context.Background()

	go db.Clean(ctx) // run cronjob to clean up unresolved challenges

	if err := tcpServer.Listen(ctx, addr); err != nil {
		log.Fatal(err)
	}
}
