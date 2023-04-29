package utils

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
