package configloader

import "testing"

func TestLoadConfigDefaults(t *testing.T) {
	cfg := struct {
		Host     string `json:"host" default:"localhost"`
		Port     int    `json:"port" default:"8080"`
		Enabled  bool   `json:"enabled" default:"true"`
		Database struct {
			Host string `json:"host" default:"db.internal"`
		} `json:"database"`
	}{}

	if err := loadConfigDefaults(&cfg); err != nil {
		t.Fatalf("load defaults: %v", err)
	}

	if cfg.Host != "localhost" {
		t.Fatalf("expected host %q, got %q", "localhost", cfg.Host)
	}
	if cfg.Port != 8080 {
		t.Fatalf("expected port %d, got %d", 8080, cfg.Port)
	}
	if !cfg.Enabled {
		t.Fatal("expected enabled default to be true")
	}
	if cfg.Database.Host != "db.internal" {
		t.Fatalf("expected database host %q, got %q", "db.internal", cfg.Database.Host)
	}
}

func TestLoadConfigDefaultsDoesNotOverrideExistingValue(t *testing.T) {
	cfg := struct {
		Host string `json:"host" default:"localhost"`
	}{
		Host: "already-set",
	}

	if err := loadConfigDefaults(&cfg); err != nil {
		t.Fatalf("load defaults: %v", err)
	}

	if cfg.Host != "already-set" {
		t.Fatalf("expected host to remain %q, got %q", "already-set", cfg.Host)
	}
}

func TestLoadIntegrationUsesDefaultsAsLowestPrecedence(t *testing.T) {
	type config struct {
		Host string `json:"host" default:"default-host"`
		Port int    `json:"port" default:"9000"`
	}

	t.Setenv("APP_HOST", "prod.example.com")

	cfg, err := Load[config](WithEnvPrefix("APP"), WithEnvSeparator("__"))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.Host != "prod.example.com" {
		t.Fatalf("expected env to override default, got %q", cfg.Host)
	}
	if cfg.Port != 9000 {
		t.Fatalf("expected default port %d, got %d", 9000, cfg.Port)
	}
}
