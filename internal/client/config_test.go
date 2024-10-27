package client

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
				portEnv:         "8080",
				hostnameEnv:     "localhost",
				resourceEnv:     "example",
				writeTimeoutEnv: "15",
				readTimeoutEnv:  "20",
			},
			expectedConfig: Config{
				Hostname:     "localhost",
				Resource:     "example",
				Port:         8080,
				WriteTimeout: 15 * time.Second,
				ReadTimeout:  20 * time.Second,
			},
			expectedError: nil,
		},
		{
			name: "No hostname environment variable",
			envVars: map[string]string{
				portEnv: "8080",
			},
			expectedConfig: Config{},
			expectedError:  ErrEmptyHost,
		},
		{
			name: "No port environment variable",
			envVars: map[string]string{
				hostnameEnv: "localhost",
			},
			expectedConfig: Config{Hostname: "localhost"},
			expectedError:  ErrEmptyPort,
		},
		{
			name: "Invalid port",
			envVars: map[string]string{
				portEnv:     "invalid_port",
				hostnameEnv: "localhost",
			},
			expectedConfig: Config{Hostname: "localhost"},
			expectedError:  strconv.ErrSyntax,
		},
		{
			name: "Default write and read timeout variables",
			envVars: map[string]string{
				portEnv:     "8080",
				hostnameEnv: "localhost",
			},
			expectedConfig: Config{
				Hostname:     "localhost",
				Port:         8080,
				WriteTimeout: defaultWriteTimeout * time.Second,
				ReadTimeout:  defaultReadTimeout * time.Second,
			},
			expectedError: nil,
		},
		{
			name: "Invalid write timeout",
			envVars: map[string]string{
				portEnv:         "8080",
				hostnameEnv:     "localhost",
				writeTimeoutEnv: "invalid_timeout",
			},
			expectedConfig: Config{Port: 8080, Hostname: "localhost"},
			expectedError:  strconv.ErrSyntax,
		},
		{
			name: "Invalid read timeout",
			envVars: map[string]string{
				portEnv:        "8080",
				hostnameEnv:    "localhost",
				readTimeoutEnv: "invalid_timeout",
			},
			expectedConfig: Config{
				Port:         8080,
				Hostname:     "localhost",
				WriteTimeout: defaultWriteTimeout * time.Second,
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
