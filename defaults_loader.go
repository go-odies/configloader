package configloader

import "reflect"

type DefaultsLoader[T any] struct{}

func NewDefaultsLoader[T any]() *DefaultsLoader[T] {
	return &DefaultsLoader[T]{}
}

func (d *DefaultsLoader[T]) Load(cfg *T) error {
	return loadConfigDefaults(cfg)
}

func loadConfigDefaults[T any](config *T) error {
	if config == nil {
		return nil
	}

	value := reflect.ValueOf(config)
	if value.Kind() != reflect.Pointer || value.IsNil() {
		return nil
	}

	return applyStructDefaults(value.Elem())
}

func applyStructDefaults(value reflect.Value) error {
	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return nil
		}
		value = value.Elem()
	}
	if value.Kind() != reflect.Struct {
		return nil
	}

	for i := 0; i < value.NumField(); i++ {
		field := value.Type().Field(i)
		if field.PkgPath != "" {
			continue
		}

		fieldValue := value.Field(i)

		underlying := fieldValue
		if underlying.Kind() == reflect.Pointer && !underlying.IsNil() {
			underlying = underlying.Elem()
		}
		if underlying.Kind() == reflect.Struct {
			if err := applyStructDefaults(fieldValue); err != nil {
				return err
			}
			continue
		}

		defaultValue, ok := field.Tag.Lookup("default")
		if !ok || !fieldValue.IsZero() {
			continue
		}
		if err := setFieldValue(fieldValue, defaultValue); err != nil {
			return err
		}
	}

	return nil
}
