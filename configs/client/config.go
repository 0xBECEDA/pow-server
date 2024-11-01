package client

import (
	"errors"
	"time"
)

var (
	ErrEmptyPort = errors.New("empty client port")
	ErrEmptyHost = errors.New("empty client host")
)

const DefaultFilePath = "config.yaml"

type Config struct {
	Hostname     string        `yaml:"hostname" envconfig:"HOSTNAME"`
	Resource     string        `yaml:"resource" envconfig:"RESOURCE"`
	Port         uint64        `yaml:"port" envconfig:"PORT"`
	WriteTimeout time.Duration `yaml:"write_timeout" envconfig:"WRITE_TIMEOUT"`
	ReadTimeout  time.Duration `yaml:"read_timeout" envconfig:"READ_TIMEOUT"`
}

func (c *Config) Validate() error {
	if c.Hostname == "" {
		return ErrEmptyHost
	}

	if c.Port == 0 {
		return ErrEmptyPort
	}
	return nil
}
