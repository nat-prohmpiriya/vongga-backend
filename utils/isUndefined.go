package utils

func IsUndefined(value interface{}) bool {
	switch v := value.(type) {
	case nil:
		return true
	case string:
		return v == "" || v == "undefined"
	case int:
		return v == 0
	case float64:
		return v == 0.0
	// เพิ่ม case อื่นๆ ตามต้องการ
	default:
		return false
	}
}
