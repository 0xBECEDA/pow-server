package client

import (
	"github.com/stretchr/testify/assert"
	"os"
	"strconv"
	"testing"
	"time"
	"world-of-wisdom/configs"
)

func TestLoadYamlConfig(t *testing.T) {
	expectedConfig := &Config{
		Hostname:     "localhost",
		Resource:     "example.com",
		Port:         8080,
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}

	actualConfig := &Config{}
	assert.NoError(t, configs.LoadConfig(DefaultFilePath, actualConfig))
	assert.Equal(t, expectedConfig, actualConfig)
}

func TestOverwriteConfigWithEnvs(t *testing.T) {
	const (
		hostname     = "https://example.com"
		resource     = "blablabla"
		port         = 4567
		writeTimeout = 3
		readTimeout  = 9
	)

	os.Setenv("HOSTNAME", hostname)
	os.Setenv("RESOURCE", resource)
	os.Setenv("PORT", strconv.Itoa(port))
	os.Setenv("WRITE_TIMEOUT", "3s")
	os.Setenv("READ_TIMEOUT", "9s")

	expectedConfig := &Config{
		Hostname:     hostname,
		Resource:     resource,
		Port:         port,
		WriteTimeout: writeTimeout * time.Second,
		ReadTimeout:  readTimeout * time.Second,
	}

	actualConfig := &Config{}
	assert.NoError(t, configs.LoadConfig(DefaultFilePath, actualConfig))
	assert.Equal(t, expectedConfig, actualConfig)
}
