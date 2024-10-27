package server

import (
	"errors"
	"os"
	"strconv"
	"time"
)

const (
	defaultReadTimeout  = 10
	defaultWriteTimeout = 10
	defaultKeyTTL       = 20
	defaultConnections  = 10
)

var (
	ErrEmptyPort = errors.New("empty server port")
)

type Config struct {
	Port             string
	ConnectionsLimit uint64
	WriteTimeout     time.Duration
	ReadTimeout      time.Duration
	KeyTTL           time.Duration
}

func (c *Config) Load() error {
	port := os.Getenv("PORT")
	if port == "" {
		return ErrEmptyPort
	}
	c.Port = port

	writeDeadline := os.Getenv("WRITE_TIMEOUT")
	if writeDeadline == "" {
		c.WriteTimeout = defaultWriteTimeout * time.Second
	} else {
		dur, err := strconv.Atoi(writeDeadline)
		if err != nil {
			return err
		}
		c.WriteTimeout = time.Duration(dur) * time.Second
	}

	readDeadline := os.Getenv("READ_TIMEOUT")
	if readDeadline == "" {
		c.ReadTimeout = defaultReadTimeout * time.Second
	} else {
		dur, err := strconv.Atoi(readDeadline)
		if err != nil {
			return err
		}
		c.ReadTimeout = time.Duration(dur) * time.Second
	}

	keyTTL := os.Getenv("KEY_TTL")
	if keyTTL == "" {
		c.KeyTTL = defaultKeyTTL * time.Second
	} else {
		dur, err := strconv.Atoi(keyTTL)
		if err != nil {
			return err
		}
		c.KeyTTL = time.Duration(dur) * time.Second
	}

	connLimit := os.Getenv("CONNECTIONS_LIMIT")
	if connLimit == "" {
		c.ConnectionsLimit = defaultConnections
	} else {
		limit, err := strconv.ParseUint(connLimit, 10, 64)
		if err != nil {
			return err
		}
		c.ConnectionsLimit = limit
	}

	return nil
}
