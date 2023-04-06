package service

import (
	"reflect"

	"github.com/godbus/dbus/v5"
)

// fills given struct with values from dbus dict payload
func MapDbusDictToStruct[T interface{}](payload map[string]dbus.Variant, structType T) T {
	// unwrap result struct
	elem := reflect.ValueOf(structType).Elem()
	elemType := elem.Type()

	// iterate over struct fields
	for i := 0; i < elem.NumField(); i++ {
		// get and check field value
		fieldValue := elem.Field(i)
		if !fieldValue.IsValid() || !fieldValue.CanSet() {
			continue
		}
		// get field definition
		field := elemType.Field(i)
		// check field exists in map
		if val, ok := payload[field.Name]; ok {
			// extract value
			unwrapped := val.Value()
			// set value if assignable
			if reflect.TypeOf(unwrapped).AssignableTo(field.Type) {
				fieldValue.Set(reflect.ValueOf(unwrapped))
			}
		}
	}
	return structType
}
