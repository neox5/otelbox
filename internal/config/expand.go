package config

import (
	"reflect"
	"regexp"
	"strings"
)

// iteratorPattern matches {iterator_name} placeholders in strings
var iteratorPattern = regexp.MustCompile(`\{([a-zA-Z_][a-zA-Z0-9_]*)\}`)

// findIteratorsInStruct finds all iterator references in a struct's string fields.
// Returns unique iterator names found in {name} patterns.
func findIteratorsInStruct(v interface{}) []string {
	found := make(map[string]bool)
	val := reflect.ValueOf(v)

	walkStructFields(val, func(field reflect.Value) {
		if field.Kind() == reflect.String {
			for _, name := range extractIteratorNames(field.String()) {
				found[name] = true
			}
		}
	})

	// Convert map to sorted slice for deterministic order
	result := make([]string, 0, len(found))
	for name := range found {
		result = append(result, name)
	}
	return result
}

// extractIteratorNames extracts iterator names from {name} patterns in a string.
func extractIteratorNames(s string) []string {
	matches := iteratorPattern.FindAllStringSubmatch(s, -1)
	if len(matches) == 0 {
		return nil
	}

	names := make([]string, len(matches))
	for i, match := range matches {
		names[i] = match[1] // Capture group 1 contains the iterator name
	}
	return names
}

// substituteIterators replaces {iterator_name} patterns with actual values.
// Modifies string fields in-place using reflection.
func substituteIterators(v interface{}, values map[string]string) {
	val := reflect.ValueOf(v)

	walkStructFields(val, func(field reflect.Value) {
		if field.Kind() == reflect.String && field.CanSet() {
			s := field.String()
			// Replace all {name} patterns with corresponding values
			for name, value := range values {
				placeholder := "{" + name + "}"
				s = strings.ReplaceAll(s, placeholder, value)
			}
			field.SetString(s)
		}
	})
}

// walkStructFields recursively walks struct fields, calling fn for each field.
// Handles pointers, nested structs, but stops at certain types (time.Duration, etc.).
func walkStructFields(val reflect.Value, fn func(reflect.Value)) {
	// Dereference pointers
	for val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return
		}
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		fn(val)
		return
	}

	// Walk struct fields
	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// Skip unexported fields
		if !fieldType.IsExported() {
			continue
		}

		// Recursively walk nested structs and pointers
		walkStructFields(field, fn)
	}
}
