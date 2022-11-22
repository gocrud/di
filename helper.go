package di

import (
	"reflect"
	"strings"
)

func getValueKindStr(filed reflect.Value) string {
	kindStr := filed.Type().String()
	kindStr = strings.TrimLeft(kindStr, "*")
	return kindStr
}

func getTypeKindStr(filed reflect.Type) string {
	kindStr := filed.String()
	kindStr = strings.TrimLeft(kindStr, "*")
	return kindStr
}
