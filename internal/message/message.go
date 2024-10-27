package message

import (
	"github.com/vmihailenco/msgpack/v5"
)

type Type int

const (
	ChallengeReq Type = iota
	ChallengeResp
	QuoteReq
	QuoteResp
)

type Message struct {
	Type Type
	Data string
}

func (m *Message) Marshal() ([]byte, error) {
	return msgpack.Marshal(m)
}

func Unmarshal(req []byte) (*Message, error) {
	var msg Message
	err := msgpack.Unmarshal(req, &msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

func NewMessage(messageType Type, data string) *Message {
	return &Message{
		Type: messageType,
		Data: data,
	}
}
