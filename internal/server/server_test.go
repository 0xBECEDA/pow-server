package server

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/vmihailenco/msgpack/v5"
	"math"
	"net"
	"testing"
	"time"
	"world-of-wisdom/internal/message"
	"world-of-wisdom/internal/pow"
	"world-of-wisdom/internal/quotes"
	"world-of-wisdom/internal/storage"
	"world-of-wisdom/internal/utils"
)

func TestServer(t *testing.T) {
	const (
		challengeTTL = 5 * time.Second
		writeTimeout = defaultWriteTimeout * time.Second
		readTimeout  = defaultReadTimeout * time.Second
		port         = 8080
		connsLimit   = 1
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		time.Sleep(1 * time.Second)
	}()

	db := storage.NewStorage(ctx, challengeTTL)

	// run server
	srv := NewServer(
		&Config{
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
	assert.Equal(t, ErrConnectionsLimitExceeded.Error(), resp.Data)

	// 2. invalid message type -> fail
	req = message.NewMessage(100, "")
	assert.NoError(t, utils.WriteConn(*req, conn, writeTimeout))

	resp = readResp(t, conn, readTimeout)
	assert.Equal(t, message.ErrResp, resp.Type)
	assert.Equal(t, ErrUnknownRequest.Error(), resp.Data)

	// 3. request challenge -> send unresolved challenge  -> fail
	req = message.NewMessage(message.ChallengeReq, "")
	assert.NoError(t, utils.WriteConn(*req, conn, writeTimeout))

	resp = readResp(t, conn, readTimeout)
	resp.Type = message.QuoteReq

	err := utils.WriteConn(*resp, conn, writeTimeout)
	assert.NoError(t, err)

	resp = readResp(t, conn, readTimeout)
	assert.Equal(t, message.ErrResp, resp.Type)
	assert.Equal(t, ErrChallengeUnsolved.Error(), resp.Data)

	// 4. request challenge -> send resolved challenge, but too late -> fail
	solved := solveChallenge(t, conn, challengeTTL, readTimeout, writeTimeout)
	err = utils.WriteConn(*solved, conn, writeTimeout)
	assert.NoError(t, err)

	resp = readResp(t, conn, readTimeout)
	assert.Equal(t, message.ErrResp, resp.Type)
	assert.Equal(t, ErrFailedToGetNonce.Error(), resp.Data)

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
	assert.Equal(t, ErrFailedToGetNonce.Error(), resp.Data)

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
