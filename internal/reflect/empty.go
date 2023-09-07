package reflect

import (
	"fmt"
	"reflect"
)

func CheckIfOneStrcutFieldIsEmpty(s interface{}) string {
	v := reflect.ValueOf(s)
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field, typeOfField := v.Field(i), t.Field(i)
		if field.IsZero() {
			return fmt.Sprintf("field %s is empty", typeOfField.Name)
		}
	}
	return ""
}
