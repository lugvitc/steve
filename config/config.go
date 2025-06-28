package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Sudo []string `json:"sudo"`
}

var c Config

func LoadConfig() error {
	f, err := os.Open("config.json")
	if err != nil {
		return err
	}
	defer f.Close()
	if err := json.NewDecoder(f).Decode(&c); err != nil {
		return err
	}
	return nil
}

func GetConfig() *Config {
	return &c
}

func IsSudo(user string) bool {
	for _, sudo := range c.Sudo {
		if sudo == user {
			return true
		}
	}
	return false
}
