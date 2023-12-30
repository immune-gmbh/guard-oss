package telemetry

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func setupLog(log bool) {
	var level logrus.Level
	if log {
		level = logrus.TraceLevel
	} else {
		level = logrus.ErrorLevel
	}

	// converts logs to span events
	logrus.AddHook(&logrusToOtelHook{})
	// limits logs to stderr
	logrus.SetFormatter(&logrusFilterFormatter{
		Level: level,
		Inner: &logrus.TextFormatter{},
	})
	// run all logs through the pipeline
	logrus.SetLevel(logrus.TraceLevel)
}

// attaches logs as events to the running span
type logrusToOtelHook struct{}

func (*logrusToOtelHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
		logrus.DebugLevel,
		logrus.TraceLevel,
	}
}

func (*logrusToOtelHook) Fire(e *logrus.Entry) error {
	// skip entries logging started trace operations
	if _, ok := e.Data[startOperationKey]; ok {
		return nil
	}

	span := trace.SpanFromContext(e.Context)
	attrs := []attribute.KeyValue{}

	if e.Level <= logrus.ErrorLevel {
		span.SetStatus(codes.Error, e.Message)
	}

	if e.Caller != nil {
		attrs = append(attrs, attribute.String("code.filepath", strings.TrimPrefix(e.Caller.File, sourcePrefix)))
		attrs = append(attrs, attribute.Int("code.lineno", e.Caller.Line))

		if e.Caller.Function != "" {
			attrs = append(attrs, attribute.String("code.function", e.Caller.Function))
		}
	}

	var name string
	if err, ok := e.Data[logrus.ErrorKey]; ok {
		name = "exception"
		if err, ok := err.(error); ok {
			attrs = append(attrs, attribute.String("exception.type", err.Error()))
		} else {
			attrs = append(attrs, attribute.String("exception.type", fmt.Sprintf("%#v", err)))
		}
		attrs = append(attrs, attribute.String("exception.message", e.Message))
	} else {
		name = "log"
		attrs = append(attrs, attribute.String("log.message", e.Message))
	}

	for k, v := range e.Data {
		if k == logrus.ErrorKey {
			continue
		}
		attrs = append(attrs, attribute.KeyValue{
			Key:   attribute.Key(fmt.Sprintf("debug.%s", k)),
			Value: interfaceToAttribute(v),
		})
	}

	//span.AddEventWithTimestamp(e.Context, e.Time, name, attrs...)
	span.AddEvent(name, trace.WithAttributes(attrs...))
	return nil
}

// filters entries by level and forwards them to another formatter
type logrusFilterFormatter struct {
	Inner logrus.Formatter
	Level logrus.Level
}

func (f *logrusFilterFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	if entry.Level <= f.Level {
		return f.Inner.Format(entry)
	} else {
		return nil, nil
	}
}

func interfaceToAttribute(value interface{}) attribute.Value {
	switch value.(type) {
	case string:
		return attribute.StringValue(value.(string))
	// XXX: more types
	default:
		return attribute.StringValue(fmt.Sprintf("%#v", value))
	}
}
