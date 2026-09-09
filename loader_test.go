package configloader

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadConfigFromFileJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	content := `{"host":"localhost","port":8080}`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write JSON config: %v", err)
	}

	var cfg struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	}

	if err := loadConfigFromFile(path, &cfg); err != nil {
		t.Fatalf("load JSON config: %v", err)
	}

	if cfg.Host != "localhost" {
		t.Fatalf("expected host %q, got %q", "localhost", cfg.Host)
	}
	if cfg.Port != 8080 {
		t.Fatalf("expected port %d, got %d", 8080, cfg.Port)
	}
}

func TestLoadConfigFromFileYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	content := "host: localhost\nport: 8080\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write YAML config: %v", err)
	}

	var cfg struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	}

	if err := loadConfigFromFile(path, &cfg); err != nil {
		t.Fatalf("load YAML config: %v", err)
	}

	if cfg.Host != "localhost" {
		t.Fatalf("expected host %q, got %q", "localhost", cfg.Host)
	}
	if cfg.Port != 8080 {
		t.Fatalf("expected port %d, got %d", 8080, cfg.Port)
	}
}

func TestLoadConfigFromFileUnsupportedType(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.txt")

	if err := os.WriteFile(path, []byte("host: localhost\n"), 0o600); err != nil {
		t.Fatalf("write unsupported config: %v", err)
	}

	var cfg struct{}
	err := loadConfigFromFile(path, &cfg)
	if err == nil {
		t.Fatal("expected an error for unsupported config type")
	}
	if !strings.Contains(err.Error(), "unsupported config file type") {
		t.Fatalf("expected unsupported type error, got %v", err)
	}
}

func TestLoadConfigFromEnvNestedOverride(t *testing.T) {
	t.Setenv("APP_HOST", "localhost")
	t.Setenv("APP_DATABASE__HOST", "db.internal")
	t.Setenv("APP_DATABASE__PORT", "5432")

	cfg := struct {
		Host     string `json:"host"`
		Database struct {
			Host string `json:"host"`
			Port int    `json:"port"`
		} `json:"database"`
	}{
		Host: "default-host",
	}

	if err := loadConfigFromEnv("APP", "__", &cfg); err != nil {
		t.Fatalf("load env config: %v", err)
	}

	if cfg.Host != "localhost" {
		t.Fatalf("expected host %q, got %q", "localhost", cfg.Host)
	}
	if cfg.Database.Host != "db.internal" {
		t.Fatalf("expected database host %q, got %q", "db.internal", cfg.Database.Host)
	}
	if cfg.Database.Port != 5432 {
		t.Fatalf("expected database port %d, got %d", 5432, cfg.Database.Port)
	}
}

func TestLoadConfigFromEnvUsesJSONTags(t *testing.T) {
	t.Setenv("PREFIX_MY_CONFIG", "enabled")
	t.Setenv("PREFIX_DATABASE__HOST", "db.internal")

	cfg := struct {
		MyConfig string `json:"my_config"`
		Database struct {
			Host string `json:"host"`
		} `json:"database"`
	}{
		MyConfig: "disabled",
	}

	if err := loadConfigFromEnv("PREFIX", "__", &cfg); err != nil {
		t.Fatalf("load tagged env config: %v", err)
	}

	if cfg.MyConfig != "enabled" {
		t.Fatalf("expected my_config %q, got %q", "enabled", cfg.MyConfig)
	}
	if cfg.Database.Host != "db.internal" {
		t.Fatalf("expected database host %q, got %q", "db.internal", cfg.Database.Host)
	}
}

func TestLoadConfigFromEnvUsesCamelCaseFields(t *testing.T) {
	t.Setenv("APP_API_KEY", "secret")
	t.Setenv("APP_DATABASE__HOST_NAME", "db.internal")

	cfg := struct {
		APIKey   string
		Database struct {
			HostName string
		}
	}{
		APIKey: "fallback",
	}

	if err := loadConfigFromEnv("APP", "__", &cfg); err != nil {
		t.Fatalf("load camelCase env config: %v", err)
	}

	if cfg.APIKey != "secret" {
		t.Fatalf("expected APIKey %q, got %q", "secret", cfg.APIKey)
	}
	if cfg.Database.HostName != "db.internal" {
		t.Fatalf("expected Database.HostName %q, got %q", "db.internal", cfg.Database.HostName)
	}
}

func TestLoadIntegration(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")

	if err := os.WriteFile(configPath, []byte(`{"host":"localhost","database":{"host":"default-db","port":5432}}`), 0o600); err != nil {
		t.Fatalf("write integration config: %v", err)
	}

	t.Setenv("APP_HOST", "prod.example.com")
	t.Setenv("APP_DATABASE__PORT", "6543")

	type config struct {
		Host     string `json:"host"`
		Database struct {
			Host string `json:"host"`
			Port int    `json:"port"`
		} `json:"database"`
	}

	cfg, err := Load[config](WithFilePath(configPath), WithEnvPrefix("APP"), WithEnvSeparator("__"))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.Host != "prod.example.com" {
		t.Fatalf("expected host %q, got %q", "prod.example.com", cfg.Host)
	}
	if cfg.Database.Host != "default-db" {
		t.Fatalf("expected database host %q, got %q", "default-db", cfg.Database.Host)
	}
	if cfg.Database.Port != 6543 {
		t.Fatalf("expected database port %d, got %d", 6543, cfg.Database.Port)
	}
}

func TestLoadWithoutFilePathSkipsFileLoader(t *testing.T) {
	oldArgs := os.Args
	os.Args = []string{"app"}
	t.Cleanup(func() { os.Args = oldArgs })

	t.Setenv("APP_HOST", "prod.example.com")

	type config struct {
		Host string `json:"host"`
	}

	cfg, err := Load[config](WithEnvPrefix("APP"), WithEnvSeparator("__"))
	if err != nil {
		t.Fatalf("load config without file path: %v", err)
	}

	if cfg.Host != "prod.example.com" {
		t.Fatalf("expected host %q, got %q", "prod.example.com", cfg.Host)
	}
}

func TestLoadConfigFromArgsOverridesEnvAndFile(t *testing.T) {
	oldArgs := os.Args
	os.Args = []string{"app", "--host=0.0.0.0", "--database.host=db.cli", "--database.port=7443"}
	t.Cleanup(func() { os.Args = oldArgs })

	t.Setenv("APP_HOST", "prod.example.com")
	t.Setenv("APP_DATABASE__PORT", "6543")

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")
	if err := os.WriteFile(configPath, []byte(`{"host":"localhost","database":{"host":"default-db","port":5432}}`), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	type config struct {
		Host     string `json:"host"`
		Database struct {
			Host string `json:"host"`
			Port int    `json:"port"`
		} `json:"database"`
	}

	cfg, err := Load[config](WithFilePath(configPath), WithEnvPrefix("APP"), WithEnvSeparator("__"))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.Host != "0.0.0.0" {
		t.Fatalf("expected host %q, got %q", "0.0.0.0", cfg.Host)
	}
	if cfg.Database.Host != "db.cli" {
		t.Fatalf("expected database host %q, got %q", "db.cli", cfg.Database.Host)
	}
	if cfg.Database.Port != 7443 {
		t.Fatalf("expected database port %d, got %d", 7443, cfg.Database.Port)
	}
}
