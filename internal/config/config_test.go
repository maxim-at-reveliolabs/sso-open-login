package config

import (
	"strings"
	"testing"
)

func setValidEnv(t *testing.T) {
	t.Helper()
	t.Setenv("SSH_HOST", "dev.box")
	t.Setenv("SSH_USER", "maxim")
	t.Setenv("SSH_KEY", "$HOME/code/revelio-labs/.ssh/maxim.pem")
	t.Setenv("OKTA_USERNAME", "user@example.com")
	t.Setenv("OKTA_PASSWORD", "x")
	t.Setenv("OKTA_TOTP_SECRET", "MFRGGZDFMZTWQ2LK")
	t.Setenv("CAMOUFOX", "/home/maxim/.cache/camoufox/camoufox-bin")
}

func Test_Load_rejects_empty_required_env(t *testing.T) {
	for _, key := range []string{"SSH_HOST", "SSH_USER", "SSH_KEY", "OKTA_USERNAME", "OKTA_PASSWORD", "OKTA_TOTP_SECRET", "CAMOUFOX"} {
		t.Run(key, func(t *testing.T) {
			setValidEnv(t)
			t.Setenv(key, "")
			_, err := Load()
			if err == nil || !strings.Contains(err.Error(), key) {
				t.Fatalf("Load() error = %v, want empty %s", err, key)
			}
		})
	}
}

func Test_Load_reads_all_fields_from_env(t *testing.T) {
	setValidEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.SSHHost != "dev.box" {
		t.Fatalf("SSHHost = %q", cfg.SSHHost)
	}
	if cfg.SSHUser != "maxim" {
		t.Fatalf("SSHUser = %q", cfg.SSHUser)
	}
	if cfg.SSHKey == "" || cfg.SSHKey[0] == '$' {
		t.Fatalf("SSHKey not expanded: %q", cfg.SSHKey)
	}
	if cfg.Username != "user@example.com" {
		t.Fatalf("Username = %q", cfg.Username)
	}
	if cfg.Password != "x" {
		t.Fatal("Password was not loaded")
	}
	if cfg.TOTPSecret != "MFRGGZDFMZTWQ2LK" {
		t.Fatal("TOTPSecret was not loaded")
	}
	if cfg.Camoufox != "/home/maxim/.cache/camoufox/camoufox-bin" {
		t.Fatalf("Camoufox = %q", cfg.Camoufox)
	}
}
