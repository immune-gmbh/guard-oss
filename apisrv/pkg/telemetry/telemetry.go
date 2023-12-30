package telemetry

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"path"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

const (
	defaultCodeRepositoryUrl = "https://github.com/immune-gmbh/guard/apisrv/v2"
	startOperationKey        = "immune.start-operation"
)

var (
	defaultTracer *trace.Tracer = nil
	sourcePrefix  string

	// used by web.marshalError
	Version string = "unknown"

	TraceConfigErr   = errors.New("duplicate trace config")
	LogConfigErr     = errors.New("duplicate log config")
	UnknownOptionErr = errors.New("unknown option")
)

func init() {
	_, this, _, ok := runtime.Caller(0)
	if ok {
		// assumes we're in v2/pkg/telemetry/telemetry.go
		sourcePrefix = path.Clean(path.Join(path.Dir(this), "..", "..", ".."))
	}
}

type WithOpenTelemtry struct {
	Endpoint  string
	Token     string
	Lightstep bool
	TLS       bool
}

type WithJaeger struct {
	Endpoint string
}

type WithLog struct {
	Stdout bool
}

func Setup(service string, tag string, opts ...interface{}) (func(), error) {
	var (
		initLog   bool
		initTrace bool

		spanExporter sdktrace.SpanExporter
	)

	for _, opt := range opts {
		switch opt := opt.(type) {
		case WithOpenTelemtry:
			if initTrace {
				return nil, TraceConfigErr
			}
			exp, err := setupTraceOpenTelemetry(opt.Endpoint, opt.Token, opt.TLS)
			if err != nil {
				return nil, err
			}
			spanExporter = exp
			initTrace = true

		case WithJaeger:
			if initTrace {
				return nil, TraceConfigErr
			}
			exp, err := setupTraceJaeger(opt.Endpoint)
			if err != nil {
				return nil, err
			}
			spanExporter = exp
			initTrace = true

		case WithLog:
			if initLog {
				return nil, LogConfigErr
			}
			setupLog(opt.Stdout)

		default:
			return nil, UnknownOptionErr
		}
	}

	if initTrace {
		return setupTrace(spanExporter, service, defaultCodeRepositoryUrl, tag)
	} else {
		return func() {}, nil
	}
}

type WithAttributes struct {
	Attributes map[string]string
}

type WithOption struct {
	Option trace.SpanStartOption
}

func Start(ctx context.Context, name string, args ...interface{}) (context.Context, trace.Span) {
	var (
		attrs  map[string]string       = map[string]string{}
		opts   []trace.SpanStartOption = []trace.SpanStartOption{}
		fields logrus.Fields           = logrus.Fields{}
	)

	// log the operation start
	for _, arg := range args {
		switch arg := arg.(type) {
		case WithAttributes:
			for k, v := range arg.Attributes {
				fields[k] = v
				attrs[k] = v
			}
		case WithOption:
			opts = append(opts, arg.Option)
		}
	}
	fields[startOperationKey] = "true"
	logrus.StandardLogger().WithContext(ctx).WithFields(fields).Trace(name)

	// start trace with caller info
	tracer := defaultTracer
	if tracer != nil {
		ctx, span := (*tracer).Start(ctx, name, opts...)
		if _, file, line, ok := runtime.Caller(1); ok {
			span.SetAttributes(attribute.String("code.filepath", strings.TrimPrefix(file, sourcePrefix)))
			span.SetAttributes(attribute.Int("code.lineno", line))
			span.SetAttributes(attribute.String("trace.id", span.SpanContext().TraceID().String()))
		}
		for k, v := range attrs {
			span.SetAttributes(attribute.String(k, v))
		}

		return ctx, span
	} else {
		return trace.NewNoopTracerProvider().Tracer("").Start(ctx, name)
	}
}

func Log(ctx context.Context) *logrus.Entry {
	logger := logrus.StandardLogger().WithContext(ctx)
	if _, file, line, ok := runtime.Caller(1); ok {
		logger = logger.WithField("file", strings.TrimPrefix(file, sourcePrefix)).WithField("line", line)
	}
	return logger
}

func SpanContext(ctx context.Context) (string, error) {
	buf := new(bytes.Buffer)
	spanctx := trace.SpanContextFromContext(ctx)
	err := json.NewEncoder(buf).Encode(spanctx)

	return buf.String(), err
}

func Tracer() *trace.Tracer {
	return defaultTracer
}

func Flush(ctx context.Context) {
	if traceProvider != nil {
		traceProvider.ForceFlush(ctx)
	}
}
