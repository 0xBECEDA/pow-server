package client

import (
	"fmt"
	"github.com/vmihailenco/msgpack/v5"
	"log"
	"net"
	"time"
	"world-of-wisdom/internal/message"
	"world-of-wisdom/internal/pow"
	"world-of-wisdom/internal/utils"
)

const (
	maxIterations = 10000000
	maxRetries    = 5
)

type Client struct {
	cfg  *Config
	conn net.Conn
}

func New(config *Config) *Client {
	return &Client{
		cfg: config,
	}
}

func (c *Client) Dial() error {
	var (
		err  error
		conn net.Conn
	)

	addr := fmt.Sprintf("%v:%v", c.cfg.Hostname, c.cfg.Port)

	for i := 0; i < maxRetries; i++ {
		conn, err = net.Dial("tcp", addr)
		if err == nil {
			c.conn = conn
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	return err
}

func (c *Client) Run() error {
	req := message.NewMessage(message.ChallengeReq, "")
	if err := utils.WriteConn(*req, c.conn, c.cfg.WriteTimeout); err != nil {
		return err
	}

	resp, err := utils.ReadConn(c.conn, c.cfg.ReadTimeout)
	if err != nil {
		return err
	}

	return c.handleChallenge(resp)
}

func (c *Client) handleChallenge(resp []byte) error {
	challengeResp, err := message.Unmarshal(resp)
	if err != nil {
		return err
	}

	challenge, err := pow.Unmarshal([]byte(challengeResp.Data))
	if err != nil {
		return err
	}

	if err := challenge.Compute(maxIterations); err != nil {
		return err
	}

	challengeBytes, _ := msgpack.Marshal(challenge)
	if err := utils.WriteConn(
		message.Message{
			Type: message.QuoteReq,
			Data: string(challengeBytes),
		}, c.conn, c.cfg.WriteTimeout); err != nil {
		return err
	}

	respQuote, err := utils.ReadConn(c.conn, c.cfg.ReadTimeout)
	if err != nil {
		return err
	}

	quoteResponseMessage, err := message.Unmarshal(respQuote)
	if err != nil {
		return err
	}

	log.Printf("got quote: '%s'(c)", quoteResponseMessage.Data)
	return nil
}
