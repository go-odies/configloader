package configloader

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type ArgsLoader[T any] struct{}

func NewArgsLoader[T any]() *ArgsLoader[T] {
	return &ArgsLoader[T]{}
}

func (a *ArgsLoader[T]) Load(cfg *T) error {
	return loadConfigFromArgs(cfg)
}

func loadConfigFromArgs[T any](config *T) error {
	if config == nil {
		return nil
	}

	value := reflect.ValueOf(config)
	if value.Kind() != reflect.Pointer || value.IsNil() {
		return nil
	}
	value = value.Elem()
	if value.Kind() != reflect.Struct {
		return nil
	}

	for _, arg := range os.Args[1:] {
		if !strings.HasPrefix(arg, "--") {
			continue
		}
		flag := strings.TrimPrefix(arg, "--")
		if flag == "" {
			continue
		}

		parts := strings.SplitN(flag, "=", 2)
		name := parts[0]
		valueText := ""
		if len(parts) == 2 {
			valueText = parts[1]
		}

		if valueText == "" {
			continue
		}

		path := strings.Split(name, ".")
		if err := applyArgValue(value, path, valueText); err != nil {
			return err
		}
	}

	return nil
}

func applyArgValue(target reflect.Value, path []string, rawValue string) error {
	if !target.IsValid() {
		return nil
	}

	if target.Kind() == reflect.Pointer {
		if target.IsNil() {
			target.Set(reflect.New(target.Type().Elem()))
		}
		target = target.Elem()
	}

	if target.Kind() != reflect.Struct {
		return nil
	}

	segment := normalizeConfigKey(path[0])
	for i := 0; i < target.NumField(); i++ {
		field := target.Type().Field(i)
		if field.PkgPath != "" {
			continue
		}

		if normalizeConfigKey(fieldName(field)) != segment && normalizeConfigKey(field.Name) != segment {
			continue
		}

		fieldValue := target.Field(i)
		if len(path) == 1 {
			return setArgFieldValue(fieldValue, rawValue)
		}

		if fieldValue.Kind() == reflect.Pointer {
			if fieldValue.IsNil() {
				fieldValue.Set(reflect.New(fieldValue.Type().Elem()))
			}
			fieldValue = fieldValue.Elem()
		}
		if fieldValue.Kind() == reflect.Struct {
			return applyArgValue(fieldValue, path[1:], rawValue)
		}
		return nil
	}

	return nil
}

func parseArgValue(field reflect.Value, rawValue string) error {
	if !field.CanSet() {
		return nil
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(rawValue)
		return nil
	case reflect.Bool:
		parsed, err := strconv.ParseBool(rawValue)
		if err != nil {
			return fmt.Errorf("parse bool %q: %w", rawValue, err)
		}
		field.SetBool(parsed)
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		parsed, err := strconv.ParseInt(rawValue, 10, 64)
		if err != nil {
			return fmt.Errorf("parse int %q: %w", rawValue, err)
		}
		field.SetInt(parsed)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		parsed, err := strconv.ParseUint(rawValue, 10, 64)
		if err != nil {
			return fmt.Errorf("parse uint %q: %w", rawValue, err)
		}
		field.SetUint(parsed)
		return nil
	case reflect.Float32, reflect.Float64:
		parsed, err := strconv.ParseFloat(rawValue, field.Type().Bits())
		if err != nil {
			return fmt.Errorf("parse float %q: %w", rawValue, err)
		}
		field.SetFloat(parsed)
		return nil
	default:
		return nil
	}
}

func setArgFieldValue(field reflect.Value, rawValue string) error {
	return parseArgValue(field, rawValue)
}
