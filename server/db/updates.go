package db

import (
	"context"
	"database/sql/driver"
	"fmt"
	"reflect"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

// Updates behaves like GORM's Updates(map), but applies serializers declared on
// model fields before map values are bound. GORM only does this automatically
// for struct updates.
func Updates(query *gorm.DB, values map[string]interface{}) *gorm.DB {
	statement := &gorm.Statement{DB: query}
	if query.Statement.Model == nil {
		query.AddError(fmt.Errorf("serialize updates: model is required"))
		return query
	}
	if err := statement.Parse(query.Statement.Model); err != nil {
		query.AddError(err)
		return query
	}

	serialized := make(map[string]interface{}, len(values))
	for name, value := range values {
		field := statement.Schema.LookUpField(name)
		_, isExpression := value.(clause.Expression)
		_, isSubquery := value.(*gorm.DB)
		if field == nil || field.Serializer == nil || isExpression || isSubquery {
			serialized[name] = value
			continue
		}
		serialized[name] = serializedFieldValue{
			context:     query.Statement.Context,
			field:       field,
			destination: reflect.ValueOf(query.Statement.Model),
			value:       value,
		}
	}

	return query.Updates(serialized)
}

type serializedFieldValue struct {
	context     context.Context
	field       *schema.Field
	destination reflect.Value
	value       interface{}
}

func (value serializedFieldValue) Value() (driver.Value, error) {
	serialized, err := value.field.Serializer.Value(
		value.context,
		value.field,
		value.destination,
		value.value,
	)
	if err != nil {
		return nil, err
	}
	if driver.IsValue(serialized) {
		return serialized, nil
	}
	if valuer, ok := serialized.(driver.Valuer); ok {
		return valuer.Value()
	}
	return nil, fmt.Errorf("serialize %s: unsupported database value %T", value.field.Name, serialized)
}
