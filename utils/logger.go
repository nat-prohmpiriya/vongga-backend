package utils

import (
	"encoding/json"
	"fmt"
	"time"
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
	fmt.Printf("\n[%s] ### %s INPUT ### ", time.Now().Format(time.RFC3339), l.FunctionName)
	for _, param := range params {
		jsonBytes, _ := json.Marshal(param)
		fmt.Printf("%s\n", string(jsonBytes))
	}
}

func (l *Logger) LogInfo(params ...interface{}) {
	fmt.Printf("\n[%s] ### %s INFO ### ", time.Now().Format(time.RFC3339), l.FunctionName)
	for _, param := range params {
		jsonBytes, _ := json.Marshal(param)
		fmt.Printf("%s\n", string(jsonBytes))
	}
}

func (l *Logger) LogWarning(params ...interface{}) (err error) {
	fmt.Printf("\n[%s] ### %s WARNING ### ", time.Now().Format(time.RFC3339), l.FunctionName)
	for _, param := range params {
		jsonBytes, _ := json.Marshal(param)
		fmt.Printf("%s\n", string(jsonBytes))
	}
	return nil
}

// LogOutput logs the output of a function
func (l *Logger) LogOutput(output interface{}, err error) {
	if err != nil {
		// กรณีเกิด error เพิ่ม ERROR ในชื่อ
		fmt.Printf("\n[%s] ### %s ERROR OUTPUT ### ", time.Now().Format(time.RFC3339), l.FunctionName)
		fmt.Printf("Error: %v\n", err)
	} else if output != nil {
		// กรณีปกติ
		fmt.Printf("\n[%s] ### %s OUTPUT ### ", time.Now().Format(time.RFC3339), l.FunctionName)
		jsonBytes, _ := json.Marshal(output)
		fmt.Printf("%s\n", string(jsonBytes))
	}
}
