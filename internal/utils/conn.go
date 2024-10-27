package utils

import (
	"io"
	"log"
	"net"
	"time"
	"world-of-wisdom/internal/message"
)

func WriteConn(msg message.Message, conn net.Conn, deadline time.Duration) error {
	marshaledMsg, err := msg.Marshal()
	if err != nil {
		return err
	}

	if err := conn.SetWriteDeadline(time.Now().Add(deadline)); err != nil {
		log.Printf("failed set write deadline %v", err)
	}

	_, err = conn.Write(marshaledMsg)
	if err != nil {
		return err
	}
	return nil
}

func ReadConn(conn net.Conn, deadline time.Duration) ([]byte, error) {
	buffer := make([]byte, 1024)
	var data []byte

	if err := conn.SetReadDeadline(time.Now().Add(deadline)); err != nil {
		log.Printf("failed set read deadline %v", err)
	}

	for {
		n, err := conn.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		data = append(data, buffer[:n]...)
		if n < len(buffer) {
			break
		}
	}
	return data, nil
}
