package configloader

type ConfigFileType string

const (
	ConfigFileTypeYAML    ConfigFileType = "yaml"
	ConfigFileTypeJSON    ConfigFileType = "json"
	ConfigFileTypeInvalid ConfigFileType = ""
)

func Load[T any](opts ...Option) (*T, error) {
	options := &Options{
		FilePath:     "",
		EnvPrefix:    "",
		EnvSeparator: "__",
	}

	for _, opt := range opts {
		opt(options)
	}

	loaders := make([]ConfigLoader[T], 0, 3)
	if options.FilePath != "" {
		loaders = append(loaders, NewFileLoader[T](options.FilePath))
	}
	loaders = append(loaders,
		NewEnvLoader[T](options.EnvPrefix, options.EnvSeparator),
		NewArgsLoader[T](),
	)

	return LoadCustom(loaders...)
}
