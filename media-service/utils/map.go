package utils

import "reflect"

func BuildTagValueMap[T any](data T) (map[string]interface{}, error) {
	t := reflect.TypeOf(data)
	result := make(map[string]interface{})
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("json")
		if tag == "" {
			continue
		}
		value := reflect.ValueOf(data).Field(i).Interface()
		result[tag] = value
	}

	return result, nil
}
