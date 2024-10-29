package main

import (
	"log"

	"world-of-wisdom/internal/client"
)

func main() {
	cfg := &client.Config{}
	if err := cfg.Load(); err != nil {
		log.Fatalf("failed to load client config: %s", err.Error())
	}

	tcpClient := client.New(cfg)

	if err := tcpClient.Dial(); err != nil {
		log.Fatalf("failed to dial: %s", err.Error())
	}

	if err := tcpClient.Run(); err != nil {
		log.Fatalf("failed to run client: %s", err.Error())
	}
}
