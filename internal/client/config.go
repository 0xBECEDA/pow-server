package client

import (
	"errors"
	"os"
	"strconv"
	"time"
)

const (
	defaultReadTimeout  = 10
	defaultWriteTimeout = 10
)

var (
	ErrEmptyPort = errors.New("empty client port")
	ErrEmptyHost = errors.New("empty client host")
)

type Config struct {
	Hostname     string
	Resource     string
	Port         uint64
	WriteTimeout time.Duration
	ReadTimeout  time.Duration
}

func (c *Config) Load() error {
	host := os.Getenv("HOSTNAME")
	if host == "" {
		return ErrEmptyHost
	}
	c.Hostname = host

	port := os.Getenv("PORT")
	if port == "" {
		return ErrEmptyPort
	}
	portInt, err := strconv.ParseUint(port, 10, 64)
	if err != nil {
		return err
	}

	c.Port = portInt
	c.Resource = os.Getenv("RESOURCE")

	writeDeadline := os.Getenv("WRITE_DEADLINE")
	if writeDeadline == "" {
		c.WriteTimeout = defaultWriteTimeout * time.Second
	} else {
		dur, err := strconv.Atoi(writeDeadline)
		if err != nil {
			return err
		}
		c.WriteTimeout = time.Duration(dur) * time.Second
	}

	readDeadline := os.Getenv("READ_DEADLINE")
	if readDeadline == "" {
		c.ReadTimeout = defaultReadTimeout * time.Second
	} else {
		dur, err := strconv.Atoi(readDeadline)
		if err != nil {
			return err
		}
		c.ReadTimeout = time.Duration(dur) * time.Second
	}
	return nil
}
