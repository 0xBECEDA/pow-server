package main

import (
	"log"
	"world-of-wisdom/configs"
	clientcfg "world-of-wisdom/configs/client"

	"world-of-wisdom/internal/client"
)

const cfgPath = "../configs/client/config.yaml"

func main() {
	cfg := &clientcfg.Config{}
	if err := configs.LoadConfig(cfgPath, cfg); err != nil {
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
