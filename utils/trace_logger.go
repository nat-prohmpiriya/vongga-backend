package utils

import (
	"fmt"

	"go.opentelemetry.io/otel/trace"
)

type TraceLogger struct {
	span trace.Span
}

func NewTraceLogger(span trace.Span) *TraceLogger {
	return &TraceLogger{span: span}
}

func (t *TraceLogger) Input(data interface{}) {
	t.span.AddEvent(fmt.Sprintf("input # %s", ToJSONString(data)))
}

func (t *TraceLogger) Output(data interface{}, err error) {
	if err == nil {
		t.span.AddEvent(fmt.Sprintf("output # %s", ToJSONString(data)))
		return
	}
	mappedErr := map[string]interface{}{
		"message": data,
		"error":   err.Error(),
	}
	t.span.RecordError(err)
	t.span.AddEvent(fmt.Sprintf("output Error # %s", ToJSONString(mappedErr)))
}

func (t *TraceLogger) Warn(data interface{}) {
	t.span.AddEvent(fmt.Sprintf("warn  # %s", ToJSONString(data)))
}

func (t *TraceLogger) Info(data interface{}) {
	t.span.AddEvent(fmt.Sprintf("info # %s", ToJSONString(data)))
}

func (t *TraceLogger) Error(err error) {
	t.span.RecordError(err)
	t.span.AddEvent(fmt.Sprintf("error %s", err.Error()))
}
