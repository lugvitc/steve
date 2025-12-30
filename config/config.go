package config

import (
	"encoding/json"
	"os"

	"go.mau.fi/whatsmeow/types"
)

type Config struct {
	Sudo              []string `json:"sudo"`
	GithubWebhookPort int      `json:"github_webhook_port"`
	DeliveryJID       string   `json:"delivery_jid"`
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
	if c.DeliveryJID != "" {
		_, err := types.ParseJID(c.DeliveryJID)
		if err != nil {
			return err
		}
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

func GetGithubWebhookPort() int {
	port := c.GithubWebhookPort
	if port == 0 {
		port = 8080
	}
	return port
}
