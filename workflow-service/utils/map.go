package utils

import (
	"fmt"
	"reflect"
)

// ToMap(&u1, "json")
func ToMap(in interface{}, tagName string) (map[string]interface{}, error) {
	out := make(map[string]interface{})

	v := reflect.ValueOf(in)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("ToMap only accepts struct or struct pointer; got %T", v)
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		fi := t.Field(i)
		if tagValue := fi.Tag.Get(tagName); tagValue != "" {
			out[tagValue] = v.Field(i).Interface()
		}
	}
	return out, nil
}

//ToMap2(&u1, "json")
func ToMap2(in interface{}, tag string) (map[string]interface{}, error) {

	// The current function only receives struct types
	v := reflect.ValueOf(in)
	if v.Kind() == reflect.Ptr { // Structure Pointer
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("ToMap only accepts struct or struct pointer; got %T", v)
	}

	out := make(map[string]interface{})
	queue := make([]interface{}, 0, 1)
	queue = append(queue, in)

	for len(queue) > 0 {
		v := reflect.ValueOf(queue[0])
		if v.Kind() == reflect.Ptr { // Structure Pointer
			v = v.Elem()
		}
		queue = queue[1:]
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			vi := v.Field(i)
			if vi.Kind() == reflect.Ptr { // Embedded Pointer
				vi = vi.Elem()
				if vi.Kind() == reflect.Struct { // Structures
					queue = append(queue, vi.Interface())
				} else {
					ti := t.Field(i)
					if tagValue := ti.Tag.Get(tag); tagValue != "" {
						// Save to map
						out[tagValue] = vi.Interface()
					}
				}
				break
			}
			if vi.Kind() == reflect.Struct { // Embedded Structs
				queue = append(queue, vi.Interface())
				break
			}
			// General Fields
			ti := t.Field(i)
			if tagValue := ti.Tag.Get(tag); tagValue != "" {
				// Save to map
				out[tagValue] = vi.Interface()
			}
		}
	}
	return out, nil
}