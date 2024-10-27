package main

import (
	"context"
	"fmt"
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
		pow.NewChallengeService(db),
		quotes.NewService(),
		cfg)

	addr := fmt.Sprintf(":%v", cfg.Port)
	log.Printf("starting tcp server on addr %v", addr)

	if err := tcpServer.Listen(ctx, addr); err != nil {
		log.Fatal(err)
	}
}
