package config

import (
	"fmt"
	"reflect"
	"strings"

	cfgtypes "github.com/jedi4ever/addt/config"
)

// resolveField navigates the GlobalConfig struct using yaml tags to find the field
// corresponding to a dotted key like "firewall.enabled" or "docker.dind.enable".
// When allocate is true, nil pointer-to-struct intermediates are allocated (for Set).
// Returns the final reflect.Value and whether the field was found.
func resolveField(cfg *cfgtypes.GlobalConfig, key string, allocate bool) (reflect.Value, bool) {
	segments := strings.Split(key, ".")
	v := reflect.ValueOf(cfg).Elem()

	for i, seg := range segments {
		field, ok := fieldByYamlTag(v, seg)
		if !ok {
			return reflect.Value{}, false
		}

		// If this is a pointer to a struct and not the final segment, dereference it
		if field.Kind() == reflect.Ptr && field.Type().Elem().Kind() == reflect.Struct && i < len(segments)-1 {
			if field.IsNil() {
				if !allocate {
					// For read operations, return the nil pointer field itself
					// so the caller knows it's unset
					if i == len(segments)-2 {
						// We're one level up from the leaf — look for the leaf in the type
						elemType := field.Type().Elem()
						for j := 0; j < elemType.NumField(); j++ {
							tag := elemType.Field(j).Tag.Get("yaml")
							tag = strings.Split(tag, ",")[0]
							if tag == segments[i+1] {
								return reflect.Value{}, true // field exists but parent is nil
							}
						}
					}
					return reflect.Value{}, true // field path exists but parent is nil
				}
				field.Set(reflect.New(field.Type().Elem()))
			}
			v = field.Elem()
		} else {
			// Final segment or non-pointer field
			return field, true
		}
	}

	return reflect.Value{}, false
}

// fieldByYamlTag scans the struct fields of v for one whose yaml tag matches tag.
func fieldByYamlTag(v reflect.Value, tag string) (reflect.Value, bool) {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return reflect.Value{}, false
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return reflect.Value{}, false
	}

	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		ft := t.Field(i)
		yamlTag := ft.Tag.Get("yaml")
		if yamlTag == "" {
			continue
		}
		// Parse "name,omitempty" to get just the name
		name := strings.Split(yamlTag, ",")[0]
		if name == tag {
			return v.Field(i), true
		}
	}
	return reflect.Value{}, false
}

// reflectGetValue retrieves a config value using reflection
func reflectGetValue(cfg *cfgtypes.GlobalConfig, key string) string {
	field, ok := resolveField(cfg, key, false)
	if !ok || !field.IsValid() {
		return ""
	}
	return formatField(field)
}

// reflectSetValue sets a config value using reflection
func reflectSetValue(cfg *cfgtypes.GlobalConfig, key, value string) {
	field, ok := resolveField(cfg, key, true)
	if !ok || !field.IsValid() {
		return
	}
	kd := keyDefMap[key]
	if kd == nil {
		return
	}
	setField(field, value, kd.Type)
}

// reflectUnsetValue clears a config value using reflection
func reflectUnsetValue(cfg *cfgtypes.GlobalConfig, key string) {
	field, ok := resolveField(cfg, key, false)
	if !ok || !field.IsValid() {
		return
	}
	unsetField(field)
}

// formatField converts a reflect.Value to its string representation
func formatField(field reflect.Value) string {
	if !field.IsValid() {
		return ""
	}

	switch field.Kind() {
	case reflect.Ptr:
		if field.IsNil() {
			return ""
		}
		elem := field.Elem()
		switch elem.Kind() {
		case reflect.Bool:
			return fmt.Sprintf("%v", elem.Bool())
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return fmt.Sprintf("%d", elem.Int())
		case reflect.String:
			return elem.String()
		}
	case reflect.String:
		return field.String()
	case reflect.Bool:
		return fmt.Sprintf("%v", field.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", field.Int())
	case reflect.Slice:
		if field.Type().Elem().Kind() == reflect.String {
			parts := make([]string, field.Len())
			for i := 0; i < field.Len(); i++ {
				parts[i] = field.Index(i).String()
			}
			return strings.Join(parts, ",")
		}
	}
	return ""
}

// setField sets a reflect.Value from a string, using the type hint from the key definition
func setField(field reflect.Value, value string, typeName string) {
	switch field.Kind() {
	case reflect.Ptr:
		elemType := field.Type().Elem()
		switch elemType.Kind() {
		case reflect.Bool:
			b := value == "true"
			field.Set(reflect.ValueOf(&b))
		case reflect.Int:
			var i int
			fmt.Sscanf(value, "%d", &i)
			field.Set(reflect.ValueOf(&i))
		case reflect.String:
			field.Set(reflect.ValueOf(&value))
		}
	case reflect.String:
		field.SetString(value)
	case reflect.Bool:
		field.SetBool(value == "true")
	case reflect.Int, reflect.Int64:
		var i int64
		fmt.Sscanf(value, "%d", &i)
		field.SetInt(i)
	case reflect.Slice:
		if field.Type().Elem().Kind() == reflect.String {
			if value == "" {
				field.Set(reflect.Zero(field.Type()))
			} else {
				parts := strings.Split(value, ",")
				if typeName == "string_list" {
					// Trim spaces for string_list types (ports.expose behavior)
					for i := range parts {
						parts[i] = strings.TrimSpace(parts[i])
					}
				}
				field.Set(reflect.ValueOf(parts))
			}
		}
	}
}

// unsetField clears a reflect.Value to its zero value
func unsetField(field reflect.Value) {
	switch field.Kind() {
	case reflect.Ptr:
		field.Set(reflect.Zero(field.Type()))
	case reflect.String:
		field.SetString("")
	case reflect.Bool:
		field.SetBool(false)
	case reflect.Int, reflect.Int64:
		field.SetInt(0)
	case reflect.Slice:
		field.Set(reflect.Zero(field.Type()))
	}
}
