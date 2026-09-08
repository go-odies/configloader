package configloader

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

	return LoadCustom(
		NewFileLoader[T](options.FilePath),
		NewEnvLoader[T](options.EnvPrefix, options.EnvSeparator),
		NewArgsLoader[T](),
	)
}
