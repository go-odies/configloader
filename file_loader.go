package configloader

import "fmt"

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
