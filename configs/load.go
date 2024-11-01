package configs

import (
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v3"
	"os"
)

type Config interface {
	Validate() error
}

func LoadConfig(fileName string, cfg Config) error {
	f, err := os.Open(fileName)
	if err != nil {
		return fmt.Errorf("unable to open configuration file (%s): %w", fileName, err)
	}

	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(cfg)
	if err != nil {
		return fmt.Errorf("unable to decode configuration file (%s): %w", fileName, err)
	}

	err = envconfig.Process("", cfg)
	if err != nil {
		return fmt.Errorf("can't load config from environment variables: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	return nil
}
