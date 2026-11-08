package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Env      string        `yaml:"env"`
	Server   ServerConfig  `yaml:"server"`
	Storage  StorageConfig `yaml:"storage"`
	Counter  CounterConfig `yaml:"counter"`
	URL      URLConfig     `yaml:"url"`
	Postgres DBConfig      `yaml:"postgres"`
	Redis    RedisConfig   `yaml:"redis"`
}

type ServerConfig struct {
	Port        int           `yaml:"port"`
	Timeout     time.Duration `yaml:"timeout"`
	IdleTimeout time.Duration `yaml:"idle_timeout"`
}

type StorageConfig struct {
	Type string `yaml:"type"`
}

type CounterConfig struct {
	Type string `yaml:"type"`
}

type URLConfig struct {
	DefaultTTLSec uint64 `yaml:"default_ttl_sec"`
	MaxTTLSec     uint64 `yaml:"max_ttl_sec"`
}

type DBConfig struct {
	Host                  string        `yaml:"host"`
	Port                  int           `yaml:"port"`
	User                  string        `yaml:"user"`
	Password              string        `yaml:"password"`
	DBName                string        `yaml:"db_name"`
	MaxOpenConnections    int           `yaml:"max_open_connections"`
	MaxIdleConnections    int           `yaml:"max_idle_connections"`
	MaxConnectionLifeTime time.Duration `yaml:"max_connection_lifetime"`
}

type PostgresConfig = DBConfig

type RedisConfig struct {
	Host         string        `yaml:"host"`
	Port         int           `yaml:"port"`
	Password     string        `yaml:"password"`
	DB           int           `yaml:"db"`
	DialTimeout  time.Duration `yaml:"dial_timeout"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
	PoolSize     int           `yaml:"pool_size"`
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yml"
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %q: %w", configPath, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config yaml: %w", err)
	}

	overrideWithEnv(&cfg)

	return &cfg, nil
}

func overrideWithEnv(cfg *Config) {
	if val := os.Getenv("ENV"); val != "" {
		cfg.Env = val
	}

	if val := os.Getenv("SERVER_PORT"); val != "" {
		if port, err := strconv.Atoi(val); err == nil {
			cfg.Server.Port = port
		}
	}

	if val := os.Getenv("STORAGE_TYPE"); val != "" {
		cfg.Storage.Type = val
	}

	if val := os.Getenv("COUNTER_TYPE"); val != "" {
		cfg.Counter.Type = val
	}

	if val := os.Getenv("POSTGRES_HOST"); val != "" {
		cfg.Postgres.Host = val
	}

	if val := os.Getenv("POSTGRES_PORT"); val != "" {
		if port, err := strconv.Atoi(val); err == nil {
			cfg.Postgres.Port = port
		}
	}

	if val := os.Getenv("POSTGRES_USER"); val != "" {
		cfg.Postgres.User = val
	}

	if val := os.Getenv("POSTGRES_PASSWORD"); val != "" {
		cfg.Postgres.Password = val
	}

	if val := os.Getenv("POSTGRES_DB"); val != "" {
		cfg.Postgres.DBName = val
	}

	if val := os.Getenv("POSTGRES_MAX_OPEN_CONNS"); val != "" {
		if n, err := strconv.Atoi(val); err == nil {
			cfg.Postgres.MaxOpenConnections = n
		}
	}

	if val := os.Getenv("POSTGRES_MAX_IDLE_CONNS"); val != "" {
		if n, err := strconv.Atoi(val); err == nil {
			cfg.Postgres.MaxIdleConnections = n
		}
	}

	if val := os.Getenv("POSTGRES_MAX_CONN_LIFETIME"); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			cfg.Postgres.MaxConnectionLifeTime = d
		}
	}

	if val := os.Getenv("REDIS_HOST"); val != "" {
		cfg.Redis.Host = val
	}

	if val := os.Getenv("REDIS_PORT"); val != "" {
		if port, err := strconv.Atoi(val); err == nil {
			cfg.Redis.Port = port
		}
	}

	if val := os.Getenv("REDIS_PASSWORD"); val != "" {
		cfg.Redis.Password = val
	}

	if val := os.Getenv("REDIS_DB"); val != "" {
		if db, err := strconv.Atoi(val); err == nil {
			cfg.Redis.DB = db
		}
	}
}
