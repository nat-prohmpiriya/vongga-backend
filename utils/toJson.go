package utils

import (
	"encoding/json"
	"fmt"
)

// ToJSONString แปลงข้อมูลเป็น JSON string
func ToJSONString(v interface{}) string {
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	return string(jsonBytes)
}

func ToJSONIndent(v interface{}) []byte {
	jsonBytes, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return []byte(fmt.Sprintf("error: %v", err))
	}
	return jsonBytes
}

func ToJSONBytes(v interface{}) []byte {
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return []byte(fmt.Sprintf("error: %v", err))
	}
	return jsonBytes
}

// StructToMap แปลง struct เป็น map[string]interface{}
func StructToMap(v interface{}) map[string]interface{} {
	// แปลงเป็น JSON ก่อน
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return map[string]interface{}{"error": err.Error()}
	}

	// แปลง JSON กลับเป็น map
	var result map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		return map[string]interface{}{"error": err.Error()}
	}

	return result
}
