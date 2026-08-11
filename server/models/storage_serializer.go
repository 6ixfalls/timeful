package models

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"gorm.io/gorm/schema"
)

// storageJSONSerializer preserves fields deliberately hidden from API JSON.
// Calendar credentials and reminder task IDs must be stored even though their
// json tags are "-" when responses are sent to clients.
type storageJSONSerializer struct{}

func init() {
	schema.RegisterSerializer("storage_json", storageJSONSerializer{})
}

func (storageJSONSerializer) Scan(ctx context.Context, field *schema.Field, dst reflect.Value, dbValue interface{}) error {
	fieldValue := reflect.New(field.FieldType)
	if dbValue != nil {
		var data []byte
		switch value := dbValue.(type) {
		case []byte:
			data = value
		case string:
			data = []byte(value)
		default:
			return fmt.Errorf("storage JSON scan %T", dbValue)
		}
		if len(data) != 0 && string(data) != "null" {
			if err := unmarshalStorageJSON(data, fieldValue.Elem()); err != nil {
				return err
			}
		}
	}
	field.ReflectValueOf(ctx, dst).Set(fieldValue.Elem())
	return nil
}

func (storageJSONSerializer) Value(_ context.Context, _ *schema.Field, _ reflect.Value, fieldValue interface{}) (interface{}, error) {
	return marshalStorageJSON(fieldValue)
}

func marshalStorageJSON(value interface{}) ([]byte, error) {
	return json.Marshal(storageJSONValue(reflect.ValueOf(value)))
}

func unmarshalStorageJSON(data []byte, value reflect.Value) error {
	if !value.CanSet() {
		return fmt.Errorf("storage JSON destination cannot be set")
	}
	if value.Kind() == reflect.Pointer {
		if string(data) == "null" {
			value.SetZero()
			return nil
		}
		value.Set(reflect.New(value.Type().Elem()))
		return unmarshalStorageJSON(data, value.Elem())
	}
	if value.CanAddr() {
		if unmarshaler, ok := value.Addr().Interface().(json.Unmarshaler); ok {
			return unmarshaler.UnmarshalJSON(data)
		}
	}
	switch value.Kind() {
	case reflect.Struct:
		fields := make(map[string]json.RawMessage)
		if err := json.Unmarshal(data, &fields); err != nil {
			return err
		}
		typeOfValue := value.Type()
		for index := 0; index < value.NumField(); index++ {
			field := typeOfValue.Field(index)
			if field.PkgPath != "" {
				continue
			}
			name, _, skip := storageFieldTag(field)
			if skip {
				continue
			}
			if raw, ok := fields[name]; ok {
				if err := unmarshalStorageJSON(raw, value.Field(index)); err != nil {
					return err
				}
			}
		}
		return nil
	case reflect.Map:
		fields := make(map[string]json.RawMessage)
		if err := json.Unmarshal(data, &fields); err != nil {
			return err
		}
		value.Set(reflect.MakeMapWithSize(value.Type(), len(fields)))
		for key, raw := range fields {
			element := reflect.New(value.Type().Elem()).Elem()
			if err := unmarshalStorageJSON(raw, element); err != nil {
				return err
			}
			value.SetMapIndex(reflect.ValueOf(key).Convert(value.Type().Key()), element)
		}
		return nil
	case reflect.Slice:
		var fields []json.RawMessage
		if err := json.Unmarshal(data, &fields); err != nil {
			return err
		}
		value.Set(reflect.MakeSlice(value.Type(), len(fields), len(fields)))
		for index, raw := range fields {
			if err := unmarshalStorageJSON(raw, value.Index(index)); err != nil {
				return err
			}
		}
		return nil
	default:
		return json.Unmarshal(data, value.Addr().Interface())
	}
}

func storageJSONValue(value reflect.Value) interface{} {
	if !value.IsValid() {
		return nil
	}
	if value.Kind() == reflect.Interface || value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return nil
		}
		return storageJSONValue(value.Elem())
	}
	if value.CanInterface() {
		if marshaler, ok := value.Interface().(json.Marshaler); ok {
			data, err := marshaler.MarshalJSON()
			if err == nil {
				var decoded interface{}
				if json.Unmarshal(data, &decoded) == nil {
					return decoded
				}
			}
		}
	}
	switch value.Kind() {
	case reflect.Struct:
		result := make(map[string]interface{})
		typeOfValue := value.Type()
		for index := 0; index < value.NumField(); index++ {
			field := typeOfValue.Field(index)
			if field.PkgPath != "" {
				continue
			}
			name, omitEmpty, skip := storageFieldTag(field)
			if skip || (omitEmpty && value.Field(index).IsZero()) {
				continue
			}
			result[name] = storageJSONValue(value.Field(index))
		}
		return result
	case reflect.Map:
		result := make(map[string]interface{}, value.Len())
		iter := value.MapRange()
		for iter.Next() {
			result[fmt.Sprint(iter.Key().Interface())] = storageJSONValue(iter.Value())
		}
		return result
	case reflect.Slice, reflect.Array:
		result := make([]interface{}, value.Len())
		for index := 0; index < value.Len(); index++ {
			result[index] = storageJSONValue(value.Index(index))
		}
		return result
	default:
		return value.Interface()
	}
}

// storageFieldTag keeps PostgreSQL JSONB compatible with the field names and
// omission behavior previously used by MongoDB. BSON tags are plain metadata;
// reading them does not require the MongoDB driver.
func storageFieldTag(field reflect.StructField) (name string, omitEmpty, skip bool) {
	tag := field.Tag.Get("bson")
	if tag == "" {
		tag = field.Tag.Get("json")
	}
	parts := strings.Split(tag, ",")
	name = parts[0]
	if name == "-" {
		return "", false, true
	}
	for _, option := range parts[1:] {
		if option == "omitempty" {
			omitEmpty = true
		}
	}
	if name == "" {
		name = strings.ToLower(field.Name[:1]) + field.Name[1:]
	}
	return name, omitEmpty, false
}
