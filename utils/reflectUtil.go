package utils

import (
	"log"
	"reflect"
)

// 将结构体转换为map
func StructToMap(s interface{}) map[string]interface{} {
	// 创建一个空map
	result := make(map[string]interface{})

	// 使用反射获取结构体的值
	val := reflect.ValueOf(s)

	// 遍历结构体的字段
	for i := 0; i < val.NumField(); i++ {
		// 获取字段的类型信息
		field := val.Type().Field(i)
		fieldName := field.Name
		fieldValue := val.Field(i)
		fieldKind := val.Field(i).Kind() // 字段的类型（如指针、字符串、整数等）

		// 如果是指针则将值取出放到map中
		if fieldKind == reflect.Ptr {
			log.Println("发现指针", fieldName)
			if !fieldValue.IsNil() {
				result[fieldName] = fieldValue.Elem().Interface()
			}
		} else {
			result[fieldName] = fieldValue.Interface()
		}
	}

	return result
}
