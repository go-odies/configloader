package configloader

import (
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type ConfigFileType string

const (
	ConfigFileTypeYAML    ConfigFileType = "yaml"
	ConfigFileTypeJSON    ConfigFileType = "json"
	ConfigFileTypeInvalid ConfigFileType = ""
)

func Load[T any](opts ...Option) (*T, error) {
	options := &Options{
		FilePath:     "config.yaml",
		EnvPrefix:    "",
		EnvSeparator: "__",
	}

	for _, opt := range opts {
		opt(options)
	}

	var config T
	err := loadConfigFromFile(options.FilePath, &config)
	if err != nil {
		return nil, err
	}

	err = loadConfigFromEnv(options.EnvPrefix, options.EnvSeparator, &config)
	if err != nil {
		return nil, err
	}

	err = loadConfigFromArgs(&config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func loadConfigFromFile[T any](filePath string, config *T) error {
	fileType, err := getConfigFileType(filePath)
	if err != nil {
		return err
	}

	fileContents, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read config file %q: %w", filePath, err)
	}

	switch fileType {
	case ConfigFileTypeYAML:
		if err := yaml.Unmarshal(fileContents, config); err != nil {
			return fmt.Errorf("unmarshal YAML config %q: %w", filePath, err)
		}
	case ConfigFileTypeJSON:
		if err := json.Unmarshal(fileContents, config); err != nil {
			return fmt.Errorf("unmarshal JSON config %q: %w", filePath, err)
		}
	default:
		return fmt.Errorf("unsupported config file type: %s", fileType)
	}

	return nil
}
