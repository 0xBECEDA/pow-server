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

const maxIterations = 10000000

type Client struct {
	resource     string
	writeTimeout time.Duration
	readTimeout  time.Duration
	conn         net.Conn
}

func New(config *Config) *Client {
	return &Client{
		resource:     config.Resource,
		writeTimeout: config.WriteTimeout,
		readTimeout:  config.ReadTimeout,
	}
}

func (c *Client) Dial(host string, port uint64) error {
	addr := fmt.Sprintf("%v:%v", host, port)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}

	c.conn = conn
	return nil
}

func (c *Client) Run() error {
	req := message.NewMessage(message.ChallengeReq, "")
	if err := utils.WriteConn(*req, c.conn, c.writeTimeout); err != nil {
		return err
	}

	resp, err := utils.ReadConn(c.conn, c.readTimeout)
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
		}, c.conn, c.writeTimeout); err != nil {
		return err
	}

	respQuote, err := utils.ReadConn(c.conn, c.readTimeout)
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
