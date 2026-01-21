package config

import (
	"fmt"
	"os"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/maker2413/term-idle/internal/ssh"
)

// Config represents the application configuration
type Config struct {
	SSH      *ssh.Config   `yaml:"ssh" koanf:"ssh"`
	Game     *GameConfig   `yaml:"game" koanf:"game"`
	Database *DBConfig     `yaml:"database" koanf:"database"`
	Server   *ServerConfig `yaml:"server" koanf:"server"`
	Logging  *LogConfig    `yaml:"logging" koanf:"logging"`
}

// GameConfig holds game-specific settings
type GameConfig struct {
	SaveInterval      string `yaml:"save_interval" koanf:"save_interval"`
	ProductionTick    string `yaml:"production_tick" koanf:"production_tick"`
	MaxPlayers        int    `yaml:"max_players" koanf:"max_players"`
	OfflineProduction bool   `yaml:"offline_production" koanf:"offline_production"`
}

// DBConfig holds database configuration
type DBConfig struct {
	Path     string `yaml:"path" koanf:"path"`
	MaxConns int    `yaml:"max_connections" koanf:"max_connections"`
	Timeout  string `yaml:"timeout" koanf:"timeout"`
}

// ServerConfig holds HTTP API server configuration
type ServerConfig struct {
	Port         string `yaml:"port" koanf:"port"`
	Host         string `yaml:"host" koanf:"host"`
	ReadTimeout  string `yaml:"read_timeout" koanf:"read_timeout"`
	WriteTimeout string `yaml:"write_timeout" koanf:"write_timeout"`
}

// LogConfig holds logging configuration
type LogConfig struct {
	Level  string `yaml:"level" koanf:"level"`
	Format string `yaml:"format" koanf:"format"`
	File   string `yaml:"file" koanf:"file"`
}

// LoadConfig loads configuration from file and environment variables
func LoadConfig(configPath string) (*Config, error) {
	k := koanf.New(".")

	// Load default configuration
	if err := loadDefaults(k); err != nil {
		return nil, fmt.Errorf("failed to load defaults: %w", err)
	}

	// Load configuration file if it exists
	if configPath != "" {
		if _, err := os.Stat(configPath); err == nil {
			if err := k.Load(file.Provider(configPath), yaml.Parser()); err != nil {
				return nil, fmt.Errorf("failed to load config file: %w", err)
			}
		}
	}

	// Load environment variables with prefix
	if err := k.Load(env.Provider("TERMIDLE_", ".", func(s string) string {
		// Convert TERMIDLE_SSH_PORT to ssh.port
		return s
	}), nil); err != nil {
		return nil, fmt.Errorf("failed to load env vars: %w", err)
	}

	// Unmarshal configuration
	var config Config
	if err := k.Unmarshal("", &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

// loadDefaults sets default configuration values
func loadDefaults(k *koanf.Koanf) error {
	defaults := map[string]interface{}{
		"ssh.port":                 2222,
		"ssh.host_key_file":        "./ssh_host_key",
		"ssh.max_sessions":         100,
		"game.save_interval":       "30s",
		"game.production_tick":     "1s",
		"game.max_players":         1000,
		"game.offline_production":  true,
		"database.path":            "./term_idle.db",
		"database.max_connections": 10,
		"database.timeout":         "30s",
		"server.port":              "8080",
		"server.host":              "0.0.0.0",
		"server.read_timeout":      "30s",
		"server.write_timeout":     "30s",
		"logging.level":            "info",
		"logging.format":           "text",
		"logging.file":             "",
	}

	for key, value := range defaults {
		if err := k.Set(key, value); err != nil {
			return fmt.Errorf("failed to set default %s: %w", key, err)
		}
	}

	return nil
}
