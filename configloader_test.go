package configloader

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

type stubLoader struct {
	apply func(cfg *testConfig)
	err   error
}

type testConfig struct {
	Host string
	Port int
}

func (s *stubLoader) Load(cfg *testConfig) error {
	if s.err != nil {
		return s.err
	}
	if s.apply != nil {
		s.apply(cfg)
	}
	return nil
}

var (
	_ ConfigLoader[testConfig] = (*FileLoader[testConfig])(nil)
	_ ConfigLoader[testConfig] = (*EnvLoader[testConfig])(nil)
	_ ConfigLoader[testConfig] = (*ArgsLoader[testConfig])(nil)
	_ ConfigLoader[testConfig] = (*stubLoader)(nil)
)

func TestLoadCustomAppliesLoadersInOrder(t *testing.T) {
	first := &stubLoader{apply: func(cfg *testConfig) {
		cfg.Host = "from-first"
		cfg.Port = 1111
	}}
	second := &stubLoader{apply: func(cfg *testConfig) {
		cfg.Host = "from-second"
	}}

	cfg, err := LoadCustom(first, second)
	if err != nil {
		t.Fatalf("load custom: %v", err)
	}

	if cfg.Host != "from-second" {
		t.Fatalf("expected later loader to win, got host %q", cfg.Host)
	}
	if cfg.Port != 1111 {
		t.Fatalf("expected port set by first loader to survive, got %d", cfg.Port)
	}
}

func TestLoadCustomPropagatesLoaderError(t *testing.T) {
	wantErr := errors.New("boom")
	applied := false

	failing := &stubLoader{err: wantErr}
	after := &stubLoader{apply: func(cfg *testConfig) { applied = true }}

	cfg, err := LoadCustom(failing, after)

	if !errors.Is(err, wantErr) {
		t.Fatalf("expected error %v, got %v", wantErr, err)
	}
	if cfg != nil {
		t.Fatalf("expected nil config on error, got %+v", cfg)
	}
	if applied {
		t.Fatal("expected loaders after the failing one to be skipped")
	}
}

func TestLoadCustomWithNoLoaders(t *testing.T) {
	cfg, err := LoadCustom[testConfig]()
	if err != nil {
		t.Fatalf("load custom: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected a non-nil zero-value config")
	}
	if cfg.Host != "" || cfg.Port != 0 {
		t.Fatalf("expected zero-value config, got %+v", cfg)
	}
}

func TestLoadCustomWithSubsetOfBuiltinLoaders(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")
	if err := os.WriteFile(configPath, []byte(`{"host":"from-file","port":8080}`), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	t.Setenv("APP_HOST", "from-env")

	oldArgs := os.Args
	os.Args = []string{"app", "--host=from-args"}
	t.Cleanup(func() { os.Args = oldArgs })

	// Deliberately omit the args loader so CLI flags are not applied,
	// proving that callers can compose any subset of loaders they need.
	cfg, err := LoadCustom[testConfig](
		NewFileLoader[testConfig](configPath),
		NewEnvLoader[testConfig]("APP", "__"),
	)
	if err != nil {
		t.Fatalf("load custom: %v", err)
	}

	if cfg.Host != "from-env" {
		t.Fatalf("expected env loader to override file loader, got host %q", cfg.Host)
	}
	if cfg.Port != 8080 {
		t.Fatalf("expected file loader value to survive, got port %d", cfg.Port)
	}
}

func TestNewFileLoaderLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte(`{"host":"localhost","port":8080}`), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	var cfg testConfig
	if err := NewFileLoader[testConfig](path).Load(&cfg); err != nil {
		t.Fatalf("load file config: %v", err)
	}

	if cfg.Host != "localhost" || cfg.Port != 8080 {
		t.Fatalf("unexpected config %+v", cfg)
	}
}

func TestNewEnvLoaderLoad(t *testing.T) {
	t.Setenv("APP_HOST", "db.internal")

	cfg := testConfig{Host: "default"}
	if err := NewEnvLoader[testConfig]("APP", "__").Load(&cfg); err != nil {
		t.Fatalf("load env config: %v", err)
	}

	if cfg.Host != "db.internal" {
		t.Fatalf("expected host %q, got %q", "db.internal", cfg.Host)
	}
}

func TestNewArgsLoaderLoad(t *testing.T) {
	oldArgs := os.Args
	os.Args = []string{"app", "--host=0.0.0.0", "--port=9090"}
	t.Cleanup(func() { os.Args = oldArgs })

	var cfg testConfig
	if err := NewArgsLoader[testConfig]().Load(&cfg); err != nil {
		t.Fatalf("load args config: %v", err)
	}

	if cfg.Host != "0.0.0.0" || cfg.Port != 9090 {
		t.Fatalf("unexpected config %+v", cfg)
	}
}
