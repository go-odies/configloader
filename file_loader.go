package configloader

import (
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type FileLoader[T any] struct {
	filePath string
}

func NewFileLoader[T any](filePath string) *FileLoader[T] {
	return &FileLoader[T]{
		filePath: filePath,
	}
}

func (f *FileLoader[T]) Load(cfg *T) error {
	return loadConfigFromFile(f.filePath, cfg)
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

func getConfigFileType(filePath string) (ConfigFileType, error) {
	// Implement the logic to determine the file type based on the file extension.
	// For example, you can check if the filePath ends with ".yaml", ".json", etc., and return the corresponding file type.
	if len(filePath) >= 5 && filePath[len(filePath)-5:] == ".yaml" {
		return ConfigFileTypeYAML, nil
	} else if len(filePath) >= 5 && filePath[len(filePath)-5:] == ".json" {
		return ConfigFileTypeJSON, nil
	}

	return ConfigFileTypeInvalid, fmt.Errorf("unsupported config file type for file: %s", filePath)
}
