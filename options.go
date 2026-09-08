package configloader


type Options struct {
	FilePath     string
	EnvPrefix    string
	EnvSeparator string
}

type Option func(*Options)

func WithFilePath(filePath string) Option {
	return func(o *Options) {
		o.FilePath = filePath
	}
}

func WithEnvPrefix(envPrefix string) Option {
	return func(o *Options) {
		o.EnvPrefix = envPrefix
	}
}

func WithEnvSeparator(envSeparator string) Option {
	return func(o *Options) {
		o.EnvSeparator = envSeparator
	}
}

