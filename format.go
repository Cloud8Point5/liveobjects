package liveobjects

import (
	"fmt"
	"reflect"
)

func FormatLiveData(value any) any {
	if value == nil {
		return value
	}
	vType := reflect.TypeOf(value)
	valOf := reflect.ValueOf(value)
	if !valOf.IsValid() {
		return nil
	}
	switch vType.Kind() {
	case reflect.Interface:
		fallthrough
	case reflect.Pointer:
		if valOf.IsNil() {
			return nil
		}
		return FormatLiveData(valOf.Elem().Interface())

	// maps and structs
	case reflect.Map:
		if valOf.IsNil() {
			return nil
		}
		newMap := make(map[string]any)
		for _, k := range valOf.MapKeys() {
			newMap[fmt.Sprint(k.Interface())] = FormatLiveData(valOf.MapIndex(k).Interface())
		}
		return newMap
	case reflect.Struct:
		// structs are converted to maps

		newMap := make(map[string]any)

		for i := 0; i < valOf.NumField(); i++ {
			field := vType.Field(i)
			jsonName, exists := field.Tag.Lookup("codec")
			if !exists {
				jsonName = field.Name
			}
			if jsonName == "-" {
				continue
			}
			if field.IsExported() {
				fieldVal := valOf.Field(i).Interface()
				formatted := FormatLiveData(fieldVal)
				newMap[jsonName] = formatted
			}
		}
		return newMap

	// arrays and slices
	case reflect.Slice:
		if valOf.IsNil() {
			return nil
		}
		fallthrough
	case reflect.Array:
		l := valOf.Len()
		newSlice := make([]any, valOf.Len())
		for i := 0; i < l; i++ {
			newSlice[i] = FormatLiveData(valOf.Index(i).Interface())
		}
		return newSlice

	// non-serializable types
	case reflect.Invalid:
		return nil
	case reflect.Chan:
		return nil
	case reflect.Func:
		return nil

	// primitives
	default:
		// Bool, Int, Int8, Int16, Int32, Int64,
		// Uint, Uint8, Uint16, Uint32, Uint64,
		// Uintptr, Float32, Float64, String
		// UnsafePointer, Complex64, Complex128 (All 3 ignored by Codec)
		return value
	}
}