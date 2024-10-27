package server

import (
	"context"
	"errors"
	"github.com/vmihailenco/msgpack/v5"
	"log"
	"net"
	"world-of-wisdom/internal/message"
	"world-of-wisdom/internal/pow"
	"world-of-wisdom/internal/utils"
)

var (
	ErrEmptyMessage      = errors.New("empty message")
	ErrFailedToMarshal   = errors.New("error marshaling timestamp")
	ErrFailedToGetRand   = errors.New("error get rand from cache")
	ErrChallengeUnsolved = errors.New("challenge is not solved")
	ErrUnknownRequest    = errors.New("unknown request received")
)

func (s *Server) handle(ctx context.Context, conn net.Conn) {
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("error closing client connection: %s", err.Error())
		}

		<-s.sem
		s.wg.Done()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			req, err := utils.ReadConn(conn, s.readDeadline)
			if err != nil {
				log.Printf("error reading request: %s", err.Error())
				return
			}

			if len(req) == 0 {
				continue
			}

			response, err := s.processClientRequest(req)
			if err != nil {
				log.Printf("error processing request: %s", err.Error())
				continue
			}

			if response != nil {
				if err = utils.WriteConn(*response, conn, s.writeDeadline); err != nil {
					log.Printf("error sending tcp message: %s", err.Error())
				}
			}
		}
	}
}

func (s *Server) processClientRequest(clientRequest []byte) (*message.Message, error) {
	parsedRequest, err := message.Unmarshal(clientRequest)
	if err != nil {
		return nil, err
	}

	switch parsedRequest.Type {
	case message.ChallengeReq:
		return s.challengeHandler(parsedRequest)
	case message.QuoteReq:
		return s.quoteHandler(parsedRequest)
	default:
		return nil, ErrUnknownRequest
	}
}

func (s *Server) challengeHandler(req *message.Message) (*message.Message, error) {
	if req == nil {
		return nil, ErrEmptyMessage
	}

	hash := pow.NewHashcash(5, req.Data)

	log.Printf("adding hash %++v", hash)

	s.powService.Add(hash.GetNonce())
	marshaledStamp, err := msgpack.Marshal(hash)
	if err != nil {
		return nil, ErrFailedToMarshal
	}

	return message.NewMessage(message.ChallengeResp, string(marshaledStamp)), nil
}

func (s *Server) quoteHandler(parsedRequest *message.Message) (*message.Message, error) {
	hash, err := pow.Unmarshal([]byte(parsedRequest.Data))
	if err != nil {
		return nil, err
	}

	randNum := hash.GetNonce()

	if ok := s.powService.Exists(randNum); !ok {
		return nil, ErrFailedToGetRand
	}

	if !hash.Check() {
		return nil, ErrChallengeUnsolved
	}

	responseMessage := message.NewMessage(message.QuoteResp, s.quoteService.RandomQuote().Text())
	s.powService.Delete(randNum)

	return responseMessage, nil
}
