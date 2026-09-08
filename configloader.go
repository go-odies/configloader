package configloader

type ConfigLoader[T any] interface {
	Load(cfg *T) error
}

func LoadCustom[T any](loaders ...ConfigLoader[T]) (*T, error) {
	var result T
	var err error
	for _, loader := range loaders {
		err = loader.Load(&result)
		if err != nil {
			return nil, err
		}
	}
	return &result, nil
}
