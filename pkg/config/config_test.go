package config

import (
	"testing"

	config_env "github.com/EvolutionAPI/evolution-go/pkg/config/env"
)

func TestLoadDefaultsConnectOnStartupToTrue(t *testing.T) {
	t.Setenv(config_env.POSTGRES_USERS_DB, "postgresql://postgres:postgres@localhost:5432/evogo_users?sslmode=disable")
	t.Setenv(config_env.DATABASE_SAVE_MESSAGES, "false")
	t.Setenv(config_env.GLOBAL_API_KEY, "test-key")
	t.Setenv(config_env.CONNECT_ON_STARTUP, "")

	cfg := Load()

	if !cfg.ConnectOnStartup {
		t.Fatal("expected CONNECT_ON_STARTUP to default to true")
	}
}

func TestLoadAllowsConnectOnStartupToBeDisabledExplicitly(t *testing.T) {
	t.Setenv(config_env.POSTGRES_USERS_DB, "postgresql://postgres:postgres@localhost:5432/evogo_users?sslmode=disable")
	t.Setenv(config_env.DATABASE_SAVE_MESSAGES, "false")
	t.Setenv(config_env.GLOBAL_API_KEY, "test-key")
	t.Setenv(config_env.CONNECT_ON_STARTUP, "false")

	cfg := Load()

	if cfg.ConnectOnStartup {
		t.Fatal("expected explicit CONNECT_ON_STARTUP=false to be respected")
	}
}
