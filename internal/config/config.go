package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Config struct {
	path       string
	HourlyRate float64 `json:"hourly_rate"`
	Currency   string  `json:"currency"`
	UpdatedAt  string  `json:"updated_at,omitempty"`
}

func Load() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}
	c := &Config{path: path, Currency: "R$"}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return c, nil
	}
	if err != nil {
		return nil, err
	}
	return c, json.Unmarshal(data, c)
}

func (c *Config) Save() error {
	c.UpdatedAt = time.Now().Format("2006-01-02 15:04")
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(c.path, data, 0644)
}

func (c *Config) HasRate() bool {
	return c.HourlyRate > 0
}

// Earn calculates earnings for a given duration.
func (c *Config) Earn(d float64) string {
	if !c.HasRate() {
		return ""
	}
	return fmt.Sprintf("%s %.2f", c.Currency, d*c.HourlyRate)
}

func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".hourly")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}
