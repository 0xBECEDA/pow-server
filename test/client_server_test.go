package test

import (
	"bytes"
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go/modules/compose"
	"github.com/testcontainers/testcontainers-go/wait"
	"log"
	"strconv"
	"testing"
	"time"
	"world-of-wisdom/internal/client"
)

func TestClientAndServer(t *testing.T) {
	// Start server
	const (
		challengeTTL = 7
		writeTimeout = 10 * time.Second
		readTimeout  = 10 * time.Second
		port         = 8080
		connsLimit   = 1
	)

	c, err := compose.NewDockerCompose("../docker/docker-compose.yml")
	assert.NoError(t, err)

	err = c.WithEnv(map[string]string{
		"CHALLENGE_TTL":     strconv.Itoa(challengeTTL),
		"CONNECTIONS_LIMIT": strconv.Itoa(connsLimit + 1), // add extra conn for Ryuk (https://golang.testcontainers.org/features/garbage_collector/)
	}).
		WaitForService("server", wait.ForListeningPort("8080/tcp").
			WithStartupTimeout(5*time.Second)).
		Up(context.TODO(), compose.RunServices("server"))
	assert.NoError(t, err)

	defer func() {
		assert.NoError(t, c.Down(context.TODO()))
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
