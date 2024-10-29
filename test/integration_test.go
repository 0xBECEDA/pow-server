package test

import (
	"bytes"
	"context"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
	"time"
	"world-of-wisdom/internal/client"
	"world-of-wisdom/internal/pow"
	"world-of-wisdom/internal/quotes"
	"world-of-wisdom/internal/server"
	"world-of-wisdom/internal/storage"
)

func TestClientAndServer(t *testing.T) {
	// Start server
	const (
		challengeTTL = 2 * time.Second
		writeTimeout = 5 * time.Second
		readTimeout  = 5 * time.Second
		port         = 8080
		connsLimit   = 1
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		time.Sleep(1 * time.Second)
	}()

	db := storage.NewStorage(ctx, challengeTTL)

	srv := server.NewServer(
		&server.Config{
			Port:             port,
			ConnectionsLimit: connsLimit,
			WriteTimeout:     writeTimeout,
			ReadTimeout:      readTimeout,
			ChallengeTTL:     challengeTTL,
			MinWorkers:       1,
			MaxWorkers:       1,
		},
		pow.NewChallengeService(db),
		quotes.NewService())

	go func() {
		assert.NoError(t, srv.Listen(ctx))
	}()

	var buf bytes.Buffer
	log.SetOutput(&buf)

	// Start client and try to get quote
	client := client.New(&client.Config{
		Hostname:     "localhost",
		Resource:     "example",
		Port:         port,
		WriteTimeout: writeTimeout,
		ReadTimeout:  readTimeout,
	})

	assert.NoError(t, client.Dial())
	assert.NoError(t, client.Run())
	assert.Contains(t, buf.String(), "got quote: '") // check that really got a quote
}
