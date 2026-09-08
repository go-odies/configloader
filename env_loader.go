package configloader

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

type EnvLoader[T any] struct {
	envPrefix    string
	envSeparator string
}

func NewEnvLoader[T any](envPrefix string, envSeparator string) *EnvLoader[T] {
	return &EnvLoader[T]{
		envPrefix:    envPrefix,
		envSeparator: envSeparator,
	}
}

func (e *EnvLoader[T]) Load(cfg *T) error {
	return loadConfigFromEnv(e.envPrefix, e.envSeparator, cfg)
}

func loadConfigFromEnv[T any](envPrefix string, envSeparator string, config *T) error {
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

	for _, env := range os.Environ() {
		keyValue := strings.SplitN(env, "=", 2)
		if len(keyValue) != 2 {
			continue
		}

		key, rawValue := keyValue[0], keyValue[1]
		path, ok := envKeyToPath(key, envPrefix, envSeparator)
		if !ok {
			continue
		}
		if len(path) == 0 {
			continue
		}
		if err := applyEnvValue(value, path, rawValue); err != nil {
			return err
		}
	}

	return nil
}

func envKeyToPath(envKey string, envPrefix string, envSeparator string) ([]string, bool) {
	if envKey == "" {
		return nil, false
	}

	remainder := envKey
	if envPrefix != "" {
		trimmedPrefix := strings.TrimSuffix(envPrefix, "_")
		trimmedPrefix = strings.TrimSuffix(trimmedPrefix, envSeparator)
		if trimmedPrefix == "" {
			return nil, false
		}
		if !strings.HasPrefix(envKey, trimmedPrefix) {
			return nil, false
		}
		remainder = strings.TrimPrefix(envKey, trimmedPrefix)
		remainder = strings.TrimPrefix(remainder, "_")
	}

	remainder = strings.TrimPrefix(remainder, "_")
	if remainder == "" {
		return nil, false
	}

	if envSeparator != "" {
		parts := strings.Split(remainder, envSeparator)
		filtered := make([]string, 0, len(parts))
		for _, part := range parts {
			part = strings.Trim(part, "_")
			if part != "" {
				filtered = append(filtered, part)
			}
		}
		if len(filtered) > 0 {
			return filtered, true
		}
	}

	return []string{remainder}, true
}

func applyEnvValue(target reflect.Value, path []string, rawValue string) error {
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
			return setFieldValue(fieldValue, rawValue)
		}

		if fieldValue.Kind() == reflect.Pointer {
			if fieldValue.IsNil() {
				fieldValue.Set(reflect.New(fieldValue.Type().Elem()))
			}
			fieldValue = fieldValue.Elem()
		}
		if fieldValue.Kind() == reflect.Struct {
			return applyEnvValue(fieldValue, path[1:], rawValue)
		}
		return nil
	}

	return nil
}

func setFieldValue(field reflect.Value, rawValue string) error {
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

func normalizeConfigKey(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Trim(value, "_")
	value = strings.ReplaceAll(value, "-", "_")
	value = strings.ReplaceAll(value, ".", "_")

	var b strings.Builder
	b.Grow(len(value))
	for i, r := range value {
		if i > 0 {
			prev := rune(value[i-1])
			next := rune(0)
			if i+1 < len(value) {
				next = rune(value[i+1])
			}

			if unicode.IsUpper(r) && ((unicode.IsLower(prev) || unicode.IsDigit(prev)) || (unicode.IsUpper(prev) && unicode.IsLower(next))) {
				b.WriteByte('_')
			}
		}
		if r == ' ' {
			b.WriteByte('_')
			continue
		}
		b.WriteRune(unicode.ToLower(r))
	}

	result := strings.Trim(b.String(), "_")
	result = strings.ReplaceAll(result, "__", "_")
	return result
}

func fieldName(field reflect.StructField) string {
	for _, tagName := range []string{"json", "yaml", "env"} {
		if tag, ok := field.Tag.Lookup(tagName); ok && tag != "-" {
			name := strings.Split(tag, ",")[0]
			if name != "" {
				return name
			}
		}
	}
	return field.Name
}
