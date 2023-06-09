package utils

import (
	"reflect"
)

func CheckReflect(x interface{}) reflect.Kind {
	return reflect.TypeOf(x).Kind()
}
