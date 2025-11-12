package utils

import "strconv"

func StringToUint(s string) uint {
	val, _ := strconv.ParseUint(s, 10, 32)
	return uint(val) // fungsi ini bertujuan mengubah string ke uint
}
func InterfaceToUint(val interface{}) uint {
	switch v := val.(type) { // fungsi ini mengubah interface ke uintt
	case float64:
		return uint(v)
	case string:
		return StringToUint(v)
	case int:
		return uint(v)
	case uint:
		return v
	default:
		return 0
	}
}
