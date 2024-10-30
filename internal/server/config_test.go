package server

import (
	"github.com/stretchr/testify/assert"
	"os"
	"strconv"
	"testing"
	"time"
)

func TestConfigLoad(t *testing.T) {
	tests := []struct {
		name           string
		envVars        map[string]string
		expectedConfig Config
		expectedError  error
	}{
		{
			name: "Valid config",
			envVars: map[string]string{
				portEnv:             "8080",
				writeTimeoutEnv:     "15",
				readTimeoutEnv:      "20",
				connectionsLimitEnv: "80",
				challengeTTLEnv:     "5",
				maxWorkersEnv:       "4",
				complexityEnv:       "2",
			},
			expectedConfig: Config{
				Port:             8080,
				WriteTimeout:     15 * time.Second,
				ReadTimeout:      20 * time.Second,
				ConnectionsLimit: 80,
				ChallengeTTL:     5 * time.Second,
				MaxWorkers:       4,
				MinWorkers:       2,
				Complexity:       2,
			},
			expectedError: nil,
		},
		{
			name:           "No port environment variable",
			envVars:        map[string]string{},
			expectedConfig: Config{},
			expectedError:  ErrEmptyPort,
		},
		{
			name: "Invalid port",
			envVars: map[string]string{
				portEnv: "invalid_port",
			},
			expectedConfig: Config{},
			expectedError:  strconv.ErrSyntax,
		},
		{
			name: "Default values",
			envVars: map[string]string{
				portEnv: "8080",
			},
			expectedConfig: Config{
				Port:             8080,
				WriteTimeout:     defaultWriteTimeout * time.Second,
				ReadTimeout:      defaultReadTimeout * time.Second,
				ConnectionsLimit: defaultConnections,
				ChallengeTTL:     defaultChallengeTTL * time.Second,
				MaxWorkers:       defaultConnections,
				MinWorkers:       defaultConnections / 2,
				Complexity:       defaultComplexity,
			},
			expectedError: nil,
		},
		{
			name: "Invalid write timeout",
			envVars: map[string]string{
				portEnv:         "8080",
				writeTimeoutEnv: "invalid_timeout",
			},
			expectedConfig: Config{Port: 8080},
			expectedError:  strconv.ErrSyntax,
		},
		{
			name: "Invalid read timeout",
			envVars: map[string]string{
				portEnv:        "8080",
				readTimeoutEnv: "invalid_timeout",
			},
			expectedConfig: Config{
				Port:         8080,
				WriteTimeout: defaultWriteTimeout * time.Second,
			},
			expectedError: strconv.ErrSyntax,
		},
		{
			name: "Invalid challenge TTL",
			envVars: map[string]string{
				portEnv:         "8080",
				challengeTTLEnv: "invalid_ttl",
			},
			expectedConfig: Config{
				Port:         8080,
				WriteTimeout: defaultWriteTimeout * time.Second,
				ReadTimeout:  defaultReadTimeout * time.Second,
			},
			expectedError: strconv.ErrSyntax,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment variables
			for key, value := range tt.envVars {
				assert.NoError(t, os.Setenv(key, value))
			}
			defer func() {
				// Clean up environment variables after each test
				for key := range tt.envVars {
					assert.NoError(t, os.Unsetenv(key))
				}
			}()

			config := &Config{}
			err := config.Load()

			if tt.expectedError != nil {
				assert.Error(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedConfig, *config)
		})
	}
}
