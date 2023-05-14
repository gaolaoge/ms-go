package utils

import (
	"strings"
	"unicode"
	"unsafe"
)

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

func IsASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > unicode.MaxASCII {
			return false
		}
	}
	return true
}

func StringToBytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(
		&struct {
			string
			Cap int
		}{
			s,
			len(s),
		}))
}
