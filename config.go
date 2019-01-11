package hudorbot

import (
	"fmt"
	"strings"
	"time"
)

// RedisConfig is required because hudor
// heavily used Redis
type RedisConfig struct {
	DB       int    `mapstructure:"db"`
	Port     int    `mapstructure:"port"`
	Host     string `mapstructure:"hostname"`
	Password string `mapstructure:"password"`
}

// Addr return redis "hostname:port" address
func (rCfg *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", rCfg.Host, rCfg.Port)
}

// ExpiresConfig is optional
//
// * State expiry: user state (pv state) and defualt expiry is set to 6 hours
//
// * Warn expiry: after an usual user adding a non whitelisted bot to they group
// 	 they will get a warning and default expiry is set to 7 days
type ExpiresConfig struct {
	State time.Duration `mapstructure:"state"`
	Warn  time.Duration `mapstructure:"Warn"`
}

// HudorConfig contains whole hudor configurations
type HudorConfig struct {
	TelegramToken string        `mapstructure:"telegramToken"`
	Redis         RedisConfig   `mapstructure:"redis"`
	Expiry        ExpiresConfig `mapstructure:"expiry"`
}

// Clean sanitize and fill defaults for missing configurations
func (cfg *HudorConfig) Clean() {
	cfg.TelegramToken = strings.TrimSpace(cfg.TelegramToken)
	cfg.Redis.Host = strings.TrimSpace(cfg.Redis.Host)

	// ------------ redis defaults ------------
	if len(cfg.Redis.Host) == 0 {
		cfg.Redis.Host = "localhost"
	}
	if cfg.Redis.Port == 0 {
		cfg.Redis.Port = 6379
	}

	// ------------ expiry defaults ------------
	if cfg.Expiry.State == 0 {
		cfg.Expiry.State = 6 * time.Hour
	}
	if cfg.Expiry.Warn == 0 {
		cfg.Expiry.Warn = 7 * 24 * time.Hour
	}
}
