package output

import (
	"fmt"
	"reflect"
	"strings"
)

type textFormatter struct{}

func (f *textFormatter) Format(v any) ([]byte, error) {
	return []byte(formatValue(v, 0)), nil
}

func formatValue(v any, indent int) string {
	if v == nil {
		return ""
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return ""
		}
		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.Struct:
		return formatStruct(rv, indent)
	case reflect.Slice, reflect.Array:
		return formatSlice(rv, indent)
	case reflect.Map:
		return formatMap(rv, indent)
	default:
		return fmt.Sprintf("%v", rv.Interface())
	}
}

func formatStruct(rv reflect.Value, indent int) string {
	rt := rv.Type()
	var sb strings.Builder
	prefix := strings.Repeat("  ", indent)

	for i := 0; i < rv.NumField(); i++ {
		field := rt.Field(i)
		if !field.IsExported() {
			continue
		}

		value := rv.Field(i)
		if value.IsZero() {
			continue
		}

		name := field.Tag.Get("json")
		if name == "" {
			name = field.Name
		}
		if idx := strings.Index(name, ","); idx >= 0 {
			name = name[:idx]
		}
		if name == "-" {
			continue
		}

		switch value.Kind() {
		case reflect.Struct, reflect.Slice, reflect.Array, reflect.Map:
			sb.WriteString(fmt.Sprintf("%s%s:\n", prefix, name))
			sb.WriteString(formatValue(value.Interface(), indent+1))
		default:
			sb.WriteString(fmt.Sprintf("%s%s: %v\n", prefix, name, value.Interface()))
		}
	}
	return sb.String()
}

func formatSlice(rv reflect.Value, indent int) string {
	var sb strings.Builder
	prefix := strings.Repeat("  ", indent)

	for i := 0; i < rv.Len(); i++ {
		elem := rv.Index(i)
		if elem.Kind() == reflect.Struct || (elem.Kind() == reflect.Ptr && elem.Elem().Kind() == reflect.Struct) {
			sb.WriteString(fmt.Sprintf("%s- \n", prefix))
			sb.WriteString(formatValue(elem.Interface(), indent+1))
		} else {
			sb.WriteString(fmt.Sprintf("%s- %v\n", prefix, elem.Interface()))
		}
	}
	return sb.String()
}

func formatMap(rv reflect.Value, indent int) string {
	var sb strings.Builder
	prefix := strings.Repeat("  ", indent)

	for _, key := range rv.MapKeys() {
		value := rv.MapIndex(key)
		sb.WriteString(fmt.Sprintf("%s%v: %v\n", prefix, key.Interface(), value.Interface()))
	}
	return sb.String()
}
