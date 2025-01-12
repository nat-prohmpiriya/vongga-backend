package utils

import (
	"encoding/json"
	"fmt"
	"time"
)

// Logger provides structured logging functionality
type Logger struct {
	FunctionName string
	ServiceName  string
	Environment  string
}

// NewLogger creates a new Logger instance
func NewLogger(functionName string) *Logger {
	return &Logger{
		FunctionName: functionName,
	}
}

// LogInput logs the input parameters of a function
func (l *Logger) LogInput(params ...interface{}) {
	logEntry := map[string]interface{}{
		"level":       "INPUT",
		"timestamp":   time.Now().Format(time.RFC3339),
		"service":     l.ServiceName,
		"environment": l.Environment,
		"function":    l.FunctionName,
	}

	paramsFindMany := append(make([]interface{}, 0), params...)
	logEntry["params"] = paramsFindMany

	jsonBytes, _ := json.Marshal(logEntry)
	fmt.Printf("%s\n", string(jsonBytes))
}

func (l *Logger) LogInfo(params ...interface{}) {
	logEntry := map[string]interface{}{
		"level":       "INFO",
		"timestamp":   time.Now().Format(time.RFC3339),
		"service":     l.ServiceName,
		"environment": l.Environment,
		"function":    l.FunctionName,
	}

	// ยังคง loop เพื่อเก็บ params
	paramsFindMany := append(make([]interface{}, 0), params...)
	logEntry["params"] = paramsFindMany

	jsonBytes, _ := json.Marshal(logEntry)
	fmt.Printf("%s\n", string(jsonBytes))
}

func (l *Logger) LogWarning(err error, params ...interface{}) {
	logEntry := map[string]interface{}{
		"level":       "WARNING",
		"timestamp":   time.Now().Format(time.RFC3339),
		"service":     l.ServiceName,
		"environment": l.Environment,
		"function":    l.FunctionName,
		"error":       err.Error(),
	}

	paramsFindMany := append(make([]interface{}, 0), params...)
	logEntry["params"] = paramsFindMany

	jsonBytes, _ := json.Marshal(logEntry)
	fmt.Printf("%s\n", string(jsonBytes))
}

func (l *Logger) LogOutput(output interface{}, err error) {
	logEntry := map[string]interface{}{
		"level":       "OUTPUT",
		"timestamp":   time.Now().Format(time.RFC3339),
		"service":     l.ServiceName,
		"environment": l.Environment,
		"function":    l.FunctionName,
	}

	if err != nil {
		logEntry["error"] = err.Error()
		logEntry["status"] = "error"
	} else if output != nil {
		logEntry["output"] = output
		logEntry["status"] = "success"
	}

	jsonBytes, _ := json.Marshal(logEntry)
	fmt.Printf("%s\n", string(jsonBytes))
}
