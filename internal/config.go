package internal

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/spf13/viper"
)

type LogConfig struct {
	OutputFile string `yaml:"outputFile"`
}

type GrpcConfig struct {
	Address string `yaml:"address"`
	Port    int    `yaml:"port"`
}

type OrgConfig struct {
	BaseURL          string        `yaml:"baseURL"`
	APIClientTimeout time.Duration `yaml:"apiClientTimeout"`
}

type ServiceConfig struct {
	Organization OrgConfig `yaml:"organization"`
}

type AppConfig struct {
	Port int `yaml:"port"`

	IdleTimeout          time.Duration `yaml:"idleTimeout"`
	ReadTimeout          time.Duration `yaml:"readTimeout"`
	WriteTimeout         time.Duration `yaml:"writeTimeout"`
	GracefulTimeout      time.Duration `yaml:"gracefulTimeout"`
	DefaultClientTimeout time.Duration `yaml:"defaultClientTimeout"`
}

type RedisConfig struct {
	Address      string `yaml:"address"`
	Password     string `yaml:"password"`
	Database     int    `yaml:"database"`
	Protocol     int    `yaml:"protocol"`
	PoolSize     int    `yaml:"poolSize"`
	MaxRetries   int    `yaml:"maxRetries"`
	DialTimeout  int    `yaml:"dialTimeout"`
	ReadTimeout  int    `yaml:"readTimeout"`
	MinIdleConns int    `yaml:"minIdleConns"`
	WriteTimeout int    `yaml:"writeTimeout"`
}

type Config struct {
	Log         LogConfig     `yml:"log"`
	Redis       RedisConfig   `yml:"redis"`
	Service     ServiceConfig `yml:"service"`
	Application AppConfig     `yml:"application"`
}

var (
	config     *Config
	configOnce sync.Once
	configErr  error
)

func loadConfigs(env string) (*Config, error) {
	configOnce.Do(func() {
		configFilePath := fmt.Sprintf("config-%s.yaml", env)
		viper.SetConfigName(configFilePath)
		viper.AddConfigPath("./configs")
		viper.SetConfigType("yaml")

		// AutomaticEnv check for an environment variable any time a viper.Get request is made.

		// Rules: viper checks for an environment variable w/ a name matching the key uppercased and prefixed with the EnvPrefix if set.
		viper.AutomaticEnv()
		viper.SetEnvPrefix("XRF_SE") // will be uppercased automatically
		// this is useful, e.g., want to use . in Get() calls, but environmental variables are to use _ delimiters (e.g., app.port -> APP_PORT)
		viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

		// Read the config file
		err := viper.ReadInConfig()
		if err != nil {
			configErr = fmt.Errorf("failed to read config file: %w :: env=%s", err, env)
			return
		}

		appConfig := Config{}
		err = viper.Unmarshal(&appConfig)
		if err != nil {
			configErr = fmt.Errorf("failed to unmarshal config file: %w :: env=%s", err, env)
			return
		}

		config = &appConfig
	})

	// Important: Check the global error variable *after* once.Do.
	if configErr != nil {
		return nil, configErr // Return the stored error
	}
	return config, nil
}

func GetConfig(env string) (*Config, error) {
	return loadConfigs(env)
}
