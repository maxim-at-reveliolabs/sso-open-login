package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	SSHHost    string
	SSHUser    string
	SSHKey     string
	Username   string
	Password   string
	TOTPSecret string
	Camoufox   string
}

func Load() (Config, error) {
	cfg := Config{
		SSHHost:    strings.TrimSpace(os.Getenv("SSH_HOST")),
		SSHUser:    strings.TrimSpace(os.Getenv("SSH_USER")),
		SSHKey:     os.ExpandEnv(strings.TrimSpace(os.Getenv("SSH_KEY"))),
		Username:   strings.TrimSpace(os.Getenv("OKTA_USERNAME")),
		Password:   strings.TrimSpace(os.Getenv("OKTA_PASSWORD")),
		TOTPSecret: strings.TrimSpace(os.Getenv("OKTA_TOTP_SECRET")),
		Camoufox:   os.ExpandEnv(strings.TrimSpace(os.Getenv("CAMOUFOX"))),
	}
	switch {
	case cfg.SSHHost == "":
		return Config{}, fmt.Errorf("SSH_HOST is empty")
	case cfg.SSHUser == "":
		return Config{}, fmt.Errorf("SSH_USER is empty")
	case cfg.SSHKey == "":
		return Config{}, fmt.Errorf("SSH_KEY is empty")
	case cfg.Username == "":
		return Config{}, fmt.Errorf("OKTA_USERNAME is empty")
	case cfg.Password == "":
		return Config{}, fmt.Errorf("OKTA_PASSWORD is empty")
	case cfg.TOTPSecret == "":
		return Config{}, fmt.Errorf("OKTA_TOTP_SECRET is empty")
	case cfg.Camoufox == "":
		return Config{}, fmt.Errorf("CAMOUFOX is empty")
	}
	return cfg, nil
}
