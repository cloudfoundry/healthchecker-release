package config

import (
	"fmt"
	"io/ioutil"
	"time"

	"gopkg.in/yaml.v2"
)

type HealthCheckEndpoint struct {
	Socket   string `yaml:"socket"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Path     string `yaml:"path"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

var defaultConfig = Config{
	HealthCheckPollInterval: 10 * time.Second,
	HealthCheckTimeout:      5 * time.Second,
	StartupDelayBuffer:      5 * time.Second,
	LogLevel:                "info",
}

func DefaultConfig() (*Config, error) {
	c := defaultConfig
	return &c, nil
}

type Config struct {
	ComponentName              string              `yaml:"component_name,omitempty"`
	FailureCounterFile         string              `yaml:"failure_counter_file,omitempty"`
	HealthCheckEndpoint        HealthCheckEndpoint `yaml:"healthcheck_endpoint,omitempty"`
	HealthCheckPollInterval    time.Duration       `yaml:"healthcheck_poll_interval,omitempty"`
	HealthCheckTimeout         time.Duration       `yaml:"healthcheck_timeout,omitempty"`
	StartResponseDelayInterval time.Duration       `yaml:"start_response_delay_interval,omitempty"`
	StartupDelayBuffer         time.Duration       `yaml:"startup_delay_buffer,omitempty"`
	LogLevel                   string              `yaml:"log_level,omitempty"`
}

func LoadConfig(configFile string) (*Config, error) {
	c, err := DefaultConfig()
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("Could not read config file: %s, err: %s", configFile, err.Error())
	}

	err = c.Initialize(b)
	if err != nil {
		return nil, fmt.Errorf("Could not unmarshal config file: %s, err: %s", configFile, err.Error())
	}

	err = c.Validate()
	if err != nil {
		return nil, fmt.Errorf("Failed to validate config file: %s, err: %s", configFile, err.Error())
	}

	// c.ApplyDefaults()

	return c, nil
}

// func (c *Config) ApplyDefaults() {
// 	if c.HealthCheckPollInterval == 0 {
// 		c.HealthCheckPollInterval = defaultConfig.HealthCheckPollInterval
// 	}

// 	if c.HealthCheckTimeout == 0 {
// 		c.HealthCheckTimeout = defaultConfig.HealthCheckTimeout
// 	}

// 	if c.StartupDelayBuffer == 0 {
// 		c.StartupDelayBuffer = defaultConfig.StartupDelayBuffer
// 	}

// 	if c.LogLevel == "" {
// 		c.LogLevel = defaultConfig.LogLevel
// 	}
// }

func (c *Config) Validate() error {
	if c.ComponentName == "" {
		return fmt.Errorf("Missing component_name")
	}

	if c.HealthCheckEndpoint.Socket == "" {
		if c.HealthCheckEndpoint.Host == "" {
			return fmt.Errorf("Missing healthcheck endpoint host or socket")
		}
		if c.HealthCheckEndpoint.Port == 0 {
			return fmt.Errorf("Missing healthcheck endpoint port or socket")
		}
	} else {
		if c.HealthCheckEndpoint.Host != "" {
			return fmt.Errorf("Cannot specify both healthcheck endpoint host and socket")
		}
		if c.HealthCheckEndpoint.Port != 0 {
			return fmt.Errorf("Cannot specify both healthcheck endpoint port and socket")
		}
	}

	if c.FailureCounterFile == "" {
		return fmt.Errorf("Missing failure_counter_file")
	}
	return nil
}

func (c *Config) Initialize(configYAML []byte) error {
	return yaml.Unmarshal(configYAML, &c)
}
