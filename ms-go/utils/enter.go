package utils

import "strings"

func Slice_find(slice []any, val any) {
	switch val.(type) {
	case int:
	case string:
	case bool:
		isBaseType()
	case []any:
		isReferenceType()
	}

}

func isBaseType()      {}
func isReferenceType() {}

func SubStringLast(str string, substr string) string {
	index := strings.Index(str, substr)
	if index < 0 {
		return ""
	} else {
		return str[index+len(substr):]
	}
}
