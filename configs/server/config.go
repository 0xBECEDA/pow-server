package server

import (
	"errors"
	"time"
)

var (
	ErrEmptyPort             = errors.New("empty server port")
	ErrEmptyConnectionsLimit = errors.New("empty server connections limit parameter")
	ErrNoTimeouts            = errors.New("no timeout set")
	ErrNoComplexity          = errors.New("no complexity set")
)

const DefaultFilePath = "config.yaml"

type Config struct {
	Port             uint64 `yaml:"port" envconfig:"PORT"`
	ConnectionsLimit uint64 `yaml:"connections_limit" envconfig:"CONNECTIONS_LIMIT"`
	MinWorkers       uint64
	MaxWorkers       uint64        `yaml:"max_workers" envconfig:"MAX_NUM_WORKERS"`
	Complexity       int           `yaml:"complexity" envconfig:"COMPLEXITY"`
	WriteTimeout     time.Duration `yaml:"write_timeout" envconfig:"WRITE_TIMEOUT"`
	ReadTimeout      time.Duration `yaml:"read_timeout" envconfig:"READ_TIMEOUT"`
	ChallengeTTL     time.Duration `yaml:"challenge_ttl" envconfig:"CHALLENGE_TTL"`
}

func (c *Config) Validate() error {
	if c.Port == 0 {
		return ErrEmptyPort
	}

	if c.ConnectionsLimit == 0 {
		return ErrEmptyConnectionsLimit
	}

	if c.MaxWorkers == 0 {
		c.MaxWorkers = c.ConnectionsLimit
	}

	c.MinWorkers = c.MaxWorkers
	if c.MaxWorkers > 1 {
		c.MinWorkers = c.MaxWorkers / 2
	}

	if c.WriteTimeout == 0 || c.ReadTimeout == 0 {
		return ErrNoTimeouts
	}

	if c.Complexity == 0 {
		return ErrNoComplexity
	}
	return nil
}
