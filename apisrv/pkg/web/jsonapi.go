package web

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/google/jsonapi"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	auth "github.com/immune-gmbh/guard/apisrv/v2/pkg/authentication"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/database"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/evidence"
	tel "github.com/immune-gmbh/guard/apisrv/v2/pkg/telemetry"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/workflow"
)

// convert to JSON:API
func marshalWithoutIncluded(ctx context.Context, data interface{}) (int, []*jsonapi.Node, []*jsonapi.ErrorObject) {
	span := trace.SpanFromContext(ctx)

	pdoc, err := jsonapi.Marshal(data)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("serialize")
		return http.StatusInternalServerError, nil, nil
	}
	doc, ok := pdoc.(*jsonapi.ManyPayload)
	if !ok {
		span.SetStatus(codes.Error, "Failed to cast jsonapi.Payloader to multiple payload doc")
		return http.StatusInternalServerError, nil, nil
	}

	return http.StatusOK, doc.Data, nil
}

func buildMeta(ctx context.Context) map[string]interface{} {
	meta := make(map[string]interface{})

	meta["version"] = tel.Version
	if name, err := os.Hostname(); err == nil {
		meta["instance"] = name
	} else if name := os.Getenv("HOSTNAME"); name != "" {
		meta["instance"] = name
	}
	if span := trace.SpanFromContext(ctx); span.SpanContext().HasTraceID() {
		meta["trace"] = span.SpanContext().TraceID().String()
	}

	return meta
}

func marshalError(w http.ResponseWriter, r *http.Request, err error) {
	status, errobjs := buildError(r.Context(), err)

	w.WriteHeader(status)
	jsonapi.MarshalErrors(w, errobjs)
}

// map generic errors that only occur during json decodes in the scope of a client request
func mapJsonClientRequestErrors(err error) error {
	if errors.Is(err, io.ErrUnexpectedEOF) {
		return errRequestIncomplete
	}
	return err
}

func buildError(ctx context.Context, err error) (int, []*jsonapi.ErrorObject) {
	var errobj jsonapi.ErrorObject
	var jsonInvalidUnmarshalError *json.InvalidUnmarshalError
	var inputErr database.InputErr
	var pErr payloadErr

	status := http.StatusInternalServerError

	if err == nil {
		return http.StatusOK, nil
	}

	// title, status & detail
	if errors.As(err, &pErr) {
		errobj.Detail = pErr.Argument
		err = pErr.Err
	}

	if errors.Is(err, database.ErrConnection) { // take case that this error is caught before others, especially PayloadErr, b/c they sometimes wrap this one
		status = http.StatusServiceUnavailable
		errobj.Title = "Database unreachable"
	} else if errors.Is(err, database.ErrInvalidTransactionMode) { // take case that this error is caught before others, especially PayloadErr, b/c they sometimes wrap this one
		status = http.StatusServiceUnavailable
		errobj.Title = "Invalid database transaction mode"
	} else if errors.Is(err, database.ErrNotFound) { // take case that this error is caught before others, especially PayloadErr, b/c they sometimes wrap this one
		status = http.StatusNotFound
		errobj.Title = "Not found"
	} else if errors.Is(err, errEnrollment) {
		status = http.StatusBadRequest
		errobj.Title = "Enrollment format invalid"
	} else if errors.Is(err, errEvidence) {
		status = http.StatusBadRequest
		errobj.Title = "Evidence format invalid"
	} else if errors.Is(err, errDevicePolicy) {
		status = http.StatusBadRequest
		errobj.Title = "Device policy invalid"
	} else if errors.Is(err, errEncoding) {
		status = http.StatusUnsupportedMediaType
		errobj.Title = "Unsupported content encoding"
	} else if errors.Is(err, errQueryStringWrong) {
		status = http.StatusBadRequest
		errobj.Title = "Invalid query string"
	} else if errors.Is(err, errJsonApiContentType) {
		status = http.StatusBadRequest
		errobj.Title = "Wrong resource type"
	} else if errors.Is(err, errInvalidPatch) {
		status = http.StatusBadRequest
		errobj.Title = "Patch contents invalid"
	} else if errors.Is(err, errSerialize) {
		status = http.StatusInternalServerError
		errobj.Title = "Unable to serialize response"
	} else if errors.Is(err, workflow.ErrNoAttestationKey) {
		status = http.StatusNotFound
		errobj.Title = "No attestation key available"
		errobj.Detail = "Evidence was signed with an unknown key. Is the device enrolled?"
	} else if errors.Is(err, workflow.ErrDeviceRetired) {
		status = http.StatusBadRequest
		errobj.Title = "Device retired"
	} else if errors.As(err, &inputErr) {
		status = http.StatusBadRequest

		if inputErr.IsCheck() {
			errobj.Title = "Constraint violated"
		} else if inputErr.IsForeignKey() {
			errobj.Title = "Invalid reference"
		} else if inputErr.IsPrimaryKey() {
			errobj.Title = "Duplicate value"
		} else {
			errobj.Title = "Invalid value"
		}
		errobj.Detail = fmt.Sprintf("%s.%s", inputErr.Type(), inputErr.Field())
	} else if errors.Is(err, evidence.ErrQuote) {
		status = http.StatusBadRequest
		errobj.Title = "TPM 2.0 quote invalid"
	} else if errors.Is(err, evidence.ErrPayload) {
		status = http.StatusBadRequest
		errobj.Title = "EventLog manipulated"
	} else if errors.Is(err, evidence.ErrFormat) {
		status = http.StatusBadRequest
		errobj.Title = "Evidence structure invalid"
	} else if errors.Is(err, workflow.ErrQuotaExceeded) {
		status = http.StatusPaymentRequired
		errobj.Title = "User quota exceeded"
		errobj.Detail = "The maximum number of devices for your organisation has been reached. Please contact sales@immu.ne."
	} else if errors.As(err, &jsonInvalidUnmarshalError) {
		status = http.StatusBadRequest
		errobj.Title = "POST body invald"
		errobj.Detail += jsonInvalidUnmarshalError.Error()
	} else if errors.Is(err, jsonapi.ErrInvalidTime) || errors.Is(err, jsonapi.ErrInvalidISO8601) || errors.Is(err, jsonapi.ErrInvalidRFC3339) {
		status = http.StatusBadRequest
		errobj.Title = "Invalid timestamp format"
		errobj.Detail += err.Error()
	} else if errors.Is(err, jsonapi.ErrUnknownFieldNumberType) || errors.Is(err, strconv.ErrRange) || errors.Is(err, strconv.ErrSyntax) {
		status = http.StatusBadRequest
		errobj.Title = "Invalid number format"
		errobj.Detail += err.Error()
	} else if errors.Is(err, ErrAuthentication) {
		status = http.StatusUnauthorized
		errobj.Title = "Authentication failed"
		errobj.Detail += err.Error()
	} else if errors.Is(err, auth.ErrFormat) || errors.Is(err, auth.ErrExpiry) || errors.Is(err, auth.ErrSignature) || errors.Is(err, auth.ErrKey) || errors.Is(err, auth.ErrClaims) || errors.Is(err, auth.ErrInternal) {
		status = http.StatusUnauthorized
		errobj.Title = "Authentication failed"
		errobj.Detail += err.Error()
	} else if errors.Is(err, jsonapi.ErrInvalidType) {
		status = http.StatusBadRequest
		errobj.Title = "Invalid field format"
		errobj.Detail += err.Error()
	} else if errors.Is(err, errRequestIncomplete) {
		status = http.StatusBadRequest
		errobj.Title = "Request is incomplete"
	} else if errors.Is(err, context.Canceled) {
		status = HTTPStatusClientClosedRequest
		errobj.Title = "Client closed request"
	} else {
		status = http.StatusInternalServerError
		errobj.Title = "Internal error"
	}

	// code
	if b, ok := ctx.Value(EnableErrorOption).(bool); ok && b {
		errobj.Code = err.Error()
	}

	// meta
	meta := buildMeta(ctx)

	errobj.Meta = &meta
	errobj.Status = fmt.Sprintf("%d", status)

	return status, []*jsonapi.ErrorObject{&errobj}
}

func marshalResult(ctx context.Context, doc jsonapi.Payloader, w http.ResponseWriter) {
	ctx, span := tel.Start(ctx, "Send")
	defer span.End()

	// meta
	meta := buildMeta(ctx)

	switch pdoc := doc.(type) {
	case *jsonapi.OnePayload:
		if pdoc.Meta == nil {
			pdoc.Meta = (*jsonapi.Meta)(&meta)
		}
		(*pdoc.Meta)["trace"] = span.SpanContext().TraceID().String()
	case *jsonapi.ManyPayload:
		if pdoc.Meta == nil {
			pdoc.Meta = (*jsonapi.Meta)(&meta)
		}
		(*pdoc.Meta)["trace"] = span.SpanContext().TraceID().String()
	default:
	}

	// write out JSON
	err := json.NewEncoder(w).Encode(doc)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("serialize")
		status, errobjs := buildError(ctx, err)
		w.WriteHeader(status)
		jsonapi.MarshalErrors(w, errobjs)
	}
}

func parseInclude(str string, whitelist map[string]struct{}) (map[string]struct{}, bool) {
	var ret = make(map[string]struct{})
	for _, s := range strings.Split(str, ",") {
		if _, ok := whitelist[s]; ok {
			ret[s] = struct{}{}
		} else {
			return nil, false
		}
	}

	return ret, true
}

func parseFieldset(str string) *map[string]bool {
	if str == "" {
		return nil
	}
	ret := make(map[string]bool)
	for _, f := range strings.Split(str, ",") {
		ret[f] = true
	}
	return &ret
}

func filterFields(node *jsonapi.Node, fieldset *map[string]bool) {
	if node == nil || fieldset == nil {
		return
	}
	for key := range node.Attributes {
		if _, ok := (*fieldset)[key]; !ok {
			delete(node.Attributes, key)
		}
	}
	for key := range node.Relationships {
		if _, ok := (*fieldset)[key]; !ok {
			delete(node.Relationships, key)
		}
	}
}
