package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Server    ServerConfig
	Cassandra CassandraConfig
	Redis     RedisConfig
	Kafka     KafkaConfig
	JWT       JWTConfig
}

type ServerConfig struct {
	Port string
}

type CassandraConfig struct {
	Hosts    []string
	Keyspace string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type KafkaConfig struct {
	Brokers []string
	Topic   string
}

type JWTConfig struct {
	AccessSecret       string
	RefreshSecret      string
	AccessExpiry       time.Duration
	RefreshExpiry      time.Duration
	Issuer             string
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
		},
		Cassandra: CassandraConfig{
			Hosts:    strings.Split(getEnv("CASSANDRA_HOSTS", "127.0.0.1"), ","),
			Keyspace: getEnv("CASSANDRA_KEYSPACE", "image_platform"),
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "127.0.0.1:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
		},
		Kafka: KafkaConfig{
			Brokers: strings.Split(getEnv("KAFKA_BROKERS", "127.0.0.1:9092"), ","),
			Topic:   getEnv("KAFKA_TOPIC", "image-events"),
		},
		JWT: JWTConfig{
			AccessSecret:  getEnv("JWT_ACCESS_SECRET", "access-secret-change-me"),
			RefreshSecret: getEnv("JWT_REFRESH_SECRET", "refresh-secret-change-me"),
			AccessExpiry:  time.Duration(getEnvInt("JWT_ACCESS_EXPIRY_MIN", 15)) * time.Minute,
			RefreshExpiry: time.Duration(getEnvInt("JWT_REFRESH_EXPIRY_HOURS", 168)) * time.Hour,
			Issuer:        getEnv("JWT_ISSUER", "image-platform"),
		},
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return fallback
}
