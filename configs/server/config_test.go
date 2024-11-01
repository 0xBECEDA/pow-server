package server

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"strconv"
	"testing"
	"time"
	"world-of-wisdom/configs"
)

func TestLoadYamlConfig(t *testing.T) {
	expectedConfig := &Config{
		Port:             8080,
		ConnectionsLimit: 100,
		MinWorkers:       10,
		MaxWorkers:       20,
		Complexity:       5,
		WriteTimeout:     10 * time.Second,
		ReadTimeout:      10 * time.Second,
		ChallengeTTL:     20 * time.Second,
	}

	actualConfig := &Config{}
	assert.NoError(t, configs.LoadConfig(DefaultFilePath, actualConfig))
	assert.Equal(t, expectedConfig, actualConfig)
}

func TestOverwriteConfigWithEnvs(t *testing.T) {
	const (
		port             = 4567
		writeTimeout     = 3
		readTimeout      = 9
		challengeTTL     = 10
		complexity       = 10
		maxWorkers       = 30
		connectionsLimit = 5
	)

	os.Setenv("PORT", strconv.Itoa(port))
	os.Setenv("WRITE_TIMEOUT", "3s")
	os.Setenv("READ_TIMEOUT", "9s")
	os.Setenv("CHALLENGE_TTL", "10s")
	os.Setenv("COMPLEXITY", fmt.Sprintf("%v", complexity))
	os.Setenv("MAX_NUM_WORKERS", strconv.Itoa(maxWorkers))
	os.Setenv("CONNECTIONS_LIMIT", strconv.Itoa(connectionsLimit))

	expectedConfig := &Config{
		Port:             port,
		WriteTimeout:     writeTimeout * time.Second,
		ReadTimeout:      readTimeout * time.Second,
		ChallengeTTL:     challengeTTL * time.Second,
		Complexity:       complexity,
		MaxWorkers:       maxWorkers,
		MinWorkers:       maxWorkers / 2,
		ConnectionsLimit: connectionsLimit,
	}

	actualConfig := &Config{}
	assert.NoError(t, configs.LoadConfig(DefaultFilePath, actualConfig))
	assert.Equal(t, expectedConfig, actualConfig)
}
