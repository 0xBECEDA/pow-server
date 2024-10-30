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
	defaultConnections  = 100
	defaultChallengeTTL = 20

	portEnv             = "PORT"
	writeTimeoutEnv     = "WRITE_TIMEOUT"
	readTimeoutEnv      = "READ_TIMEOUT"
	challengeTTLEnv     = "CHALLENGE_TTL"
	connectionsLimitEnv = "CONNECTIONS_LIMIT"
	maxWorkersEnv       = "MAX_NUM_WORKERS"
)

var (
	ErrEmptyPort = errors.New("empty server port")
)

type Config struct {
	Port             uint64
	ConnectionsLimit uint64
	MinWorkers       uint64
	MaxWorkers       uint64
	WriteTimeout     time.Duration
	ReadTimeout      time.Duration
	ChallengeTTL     time.Duration
}

func (c *Config) Load() error {
	port := os.Getenv(portEnv)
	if port == "" {
		return ErrEmptyPort
	}

	portInt, err := strconv.ParseUint(port, 10, 64)
	if err != nil {
		return err
	}

	c.Port = portInt

	writeDeadline := os.Getenv(writeTimeoutEnv)
	if writeDeadline == "" {
		c.WriteTimeout = defaultWriteTimeout * time.Second
	} else {
		dur, err := strconv.Atoi(writeDeadline)
		if err != nil {
			return err
		}
		c.WriteTimeout = time.Duration(dur) * time.Second
	}

	readDeadline := os.Getenv(readTimeoutEnv)
	if readDeadline == "" {
		c.ReadTimeout = defaultReadTimeout * time.Second
	} else {
		dur, err := strconv.Atoi(readDeadline)
		if err != nil {
			return err
		}
		c.ReadTimeout = time.Duration(dur) * time.Second
	}

	keyTTL := os.Getenv(challengeTTLEnv)
	if keyTTL == "" {
		c.ChallengeTTL = defaultChallengeTTL * time.Second
	} else {
		dur, err := strconv.Atoi(keyTTL)
		if err != nil {
			return err
		}
		c.ChallengeTTL = time.Duration(dur) * time.Second
	}

	connLimit := os.Getenv(connectionsLimitEnv)
	if connLimit == "" {
		c.ConnectionsLimit = defaultConnections
	} else {
		limit, err := strconv.ParseUint(connLimit, 10, 64)
		if err != nil {
			return err
		}
		c.ConnectionsLimit = limit
	}

	maxWorkers := os.Getenv(maxWorkersEnv)
	if maxWorkers == "" {
		c.MaxWorkers = c.ConnectionsLimit
	} else {
		workers, err := strconv.ParseUint(maxWorkers, 10, 64)
		if err != nil {
			return err
		}
		c.MaxWorkers = workers
	}
	c.MinWorkers = c.MaxWorkers / 2

	if c.MaxWorkers == 1 {
		c.MinWorkers = c.MaxWorkers
	}
	return nil
}
