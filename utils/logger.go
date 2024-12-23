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
	fmt.Printf("\n#### %s OUTPUT #### ", l.FunctionName)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else if output != nil {
		jsonBytes, _ := json.Marshal(output)
		fmt.Printf("%s\n", string(jsonBytes))
	}
}
