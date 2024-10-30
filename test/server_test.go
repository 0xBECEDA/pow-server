package test

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go/modules/compose"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/vmihailenco/msgpack/v5"
	"math"
	"net"
	"strconv"
	"testing"
	"time"
	"world-of-wisdom/internal/message"
	"world-of-wisdom/internal/pow"
	"world-of-wisdom/internal/server"
	"world-of-wisdom/internal/utils"
)

func TestServer(t *testing.T) {
	const (
		challengeTTL = 5
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
		"COMPLEXITY":        "2",
	}).
		WaitForService("server", wait.ForListeningPort("8080/tcp").
			WithStartupTimeout(5*time.Second)).
		Up(context.TODO(), compose.RunServices("server"))
	assert.NoError(t, err)

	defer func() {
		assert.NoError(t, c.Down(context.TODO()))
	}()

	// connect to server
	conn := connect(t, port)
	defer conn.Close()

	conn2 := connect(t, port)
	defer conn2.Close()

	// 1. request challenge -> conn accepted, but refused
	req := message.NewMessage(message.ChallengeReq, "")
	assert.NoError(t, utils.WriteConn(*req, conn2, writeTimeout))

	resp := readResp(t, conn2, readTimeout)
	assert.Equal(t, message.ErrResp, resp.Type)
	assert.Equal(t, server.ErrConnectionsLimitExceeded.Error(), resp.Data)

	// 2. invalid message type -> fail
	req = message.NewMessage(100, "")
	assert.NoError(t, utils.WriteConn(*req, conn, writeTimeout))

	resp = readResp(t, conn, readTimeout)
	assert.Equal(t, message.ErrResp, resp.Type)
	assert.Equal(t, server.ErrUnknownRequest.Error(), resp.Data)

	// 3. request challenge -> send unresolved challenge  -> fail
	req = message.NewMessage(message.ChallengeReq, "")
	assert.NoError(t, utils.WriteConn(*req, conn, writeTimeout))

	resp = readResp(t, conn, readTimeout)
	resp.Type = message.QuoteReq

	err = utils.WriteConn(*resp, conn, writeTimeout)
	assert.NoError(t, err)

	resp = readResp(t, conn, readTimeout)
	assert.Equal(t, message.ErrResp, resp.Type)
	assert.Equal(t, server.ErrChallengeUnsolved.Error(), resp.Data)

	// 4. request challenge -> send resolved challenge, but too late -> fail
	solved := solveChallenge(t, conn, challengeTTL*time.Second, readTimeout, writeTimeout)
	err = utils.WriteConn(*solved, conn, writeTimeout)
	assert.NoError(t, err)

	resp = readResp(t, conn, readTimeout)
	assert.Equal(t, message.ErrResp, resp.Type)
	assert.Equal(t, server.ErrFailedToGetNonce.Error(), resp.Data)

	// 5. request challenge -> send unknown resolved challenge in time -> fail
	req = message.NewMessage(message.ChallengeReq, "")
	assert.NoError(t, utils.WriteConn(*req, conn, writeTimeout))

	readResp(t, conn, readTimeout)

	unknownChallenge := pow.NewHashcash(5, "example")
	assert.NoError(t, unknownChallenge.Compute(math.MaxInt64))

	solvedUnknownChallenge, _ := msgpack.Marshal(unknownChallenge)
	assert.NoError(t, utils.WriteConn(
		message.Message{
			Type: message.QuoteReq,
			Data: string(solvedUnknownChallenge),
		}, conn, writeTimeout),
	)

	resp = readResp(t, conn, readTimeout)
	assert.Equal(t, message.ErrResp, resp.Type)
	assert.Equal(t, server.ErrFailedToGetNonce.Error(), resp.Data)

	// 6. request challenge -> send resolved challenge in time -> success
	solved = solveChallenge(t, conn, 0, readTimeout, writeTimeout)
	assert.NoError(t, utils.WriteConn(*solved, conn, writeTimeout))

	resp = readResp(t, conn, readTimeout)
	assert.Equal(t, message.QuoteResp, resp.Type)
	assert.NotEmpty(t, resp.Data)
}

func connect(t *testing.T, port int64) net.Conn {
	var (
		err  error
		conn net.Conn
	)
	for i := 0; i < 3; i++ {
		conn, err = net.Dial("tcp", fmt.Sprintf("localhost:%v", port))
		if err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	assert.NoError(t, err)
	return conn
}

func solveChallenge(
	t *testing.T, conn net.Conn,
	delay, readTimeout, writeTimeout time.Duration) *message.Message {
	req := message.NewMessage(message.ChallengeReq, "")
	assert.NoError(t, utils.WriteConn(*req, conn, writeTimeout))

	resp := readResp(t, conn, readTimeout)
	challenge, err := pow.Unmarshal([]byte(resp.Data))
	assert.NoError(t, err)
	assert.NoError(t, challenge.Compute(math.MaxInt64))

	time.Sleep(delay)

	solved, err := msgpack.Marshal(challenge)
	assert.NoError(t, err)
	return &message.Message{Type: message.QuoteReq, Data: string(solved)}
}

func readResp(
	t *testing.T,
	conn net.Conn,
	readTimeout time.Duration) *message.Message {

	resp, err := utils.ReadConn(conn, readTimeout)
	assert.NoError(t, err)

	dataResp, err := message.Unmarshal(resp)
	assert.NoError(t, err)

	return dataResp
}
