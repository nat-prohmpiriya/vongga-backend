package utils

import (
	"encoding/json"
	"fmt"
)

// Logger provides structured logging functionality
type Logger struct {
	FunctionName string
}

// NewLogger creates a new Logger instance
func NewLogger(functionName string) *Logger {
	return &Logger{
		FunctionName: functionName,
	}
}

// LogInput logs the input parameters of a function
func (l *Logger) LogInput(params ...interface{}) {
	fmt.Printf("\n#### %s INPUT #### ", l.FunctionName)
	for _, param := range params {
		jsonBytes, _ := json.Marshal(param)
		fmt.Printf("%s\n", string(jsonBytes))
	}
}

// LogOutput logs the output of a function
func (l *Logger) LogOutput(output interface{}, err error) {
	if err != nil {
		// กรณีเกิด error เพิ่ม ERROR ในชื่อ
		fmt.Printf("\n#### %s ERROR OUTPUT #### ", l.FunctionName)
		fmt.Printf("Error: %v\n", err)
	} else if output != nil {
		// กรณีปกติ
		fmt.Printf("\n#### %s OUTPUT #### ", l.FunctionName)
		jsonBytes, _ := json.Marshal(output)
		fmt.Printf("%s\n", string(jsonBytes))
	}
}
