package web

import (
	"compress/gzip"
	"context"
	"crypto"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	prommux "github.com/albertogviana/prometheus-middleware"
	"github.com/google/jsonapi"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/otel/attribute"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/blob"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/key"
	tel "github.com/immune-gmbh/guard/apisrv/v2/pkg/telemetry"
)

type EnableErrorOptionType string

var EnableErrorOption = EnableErrorOptionType("enable_errors")

var (
	errEncoding = errors.New("content encoding")
)

func contentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Trace(fmt.Sprintf("%s %s", r.Method, r.URL.Path))
		next.ServeHTTP(w, r)
	})
}

func debugErrors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rr := r.WithContext(context.WithValue(r.Context(), EnableErrorOption, true))

		next.ServeHTTP(w, rr)
	})
}

func catchPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		armed := true

		defer func() {
			if err := recover(); err != nil || armed {
				// send a response
				ctx, span := tel.Start(r.Context(), "PANIC")
				defer span.End()
				span.SetAttributes(attribute.Bool("panic", true))
				tel.Log(ctx).WithField("panicData", err).WithField("stack", string(debug.Stack())).Error("catch panic")
				span.End()
				tel.Flush(ctx)

				// respond over HTTP
				status, errobjs := buildError(ctx, errInternalServerError)
				w.WriteHeader(status)
				jsonapi.MarshalErrors(w, errobjs)
				if f, ok := w.(http.Flusher); ok {
					f.Flush()
				}

				// rethrow to hand back all resources (fixes blocking db connection pool f.e.)
				panic(fmt.Errorf("emergency shutdown due to panic: %v", err))
			}
		}()

		next.ServeHTTP(w, r)
		armed = false
	})
}

func decompressRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Header.Get("Content-Encoding") {
		case "":
			break

		case "gzip":
			gzrd, err := gzip.NewReader(r.Body)
			if err != nil {
				tel.Log(r.Context()).WithError(err).Error("inflate")
				marshalError(w, r, err)
				return
			}
			r.Body = gzrd

		default:
			tel.Log(r.Context()).WithError(errEncoding).Error("inflate")
			marshalError(w, r, errEncoding)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func NewRouter(ctx context.Context, pool *pgxpool.Pool, store *blob.Storage, pubCa crypto.PublicKey, privCa crypto.PrivateKey, ks *key.Set, ping_ch chan bool, serviceName string, baseURL string, webAppBaseURL string, corsOrigins []string, jobTimeout time.Duration) (http.Handler, error) {
	router, err := New(ctx, pool, store, pubCa, privCa, ks, ping_ch, serviceName, baseURL, webAppBaseURL, jobTimeout)
	if err != nil {
		return nil, err
	}

	root := mux.NewRouter()
	root.Use(logRequest)
	root.Use(prommux.NewPrometheusMiddleware(prommux.Opts{}).InstrumentHandlerDuration)
	root.Use(otelmux.Middleware(serviceName))
	root.Use(contentType)
	root.Use(debugErrors)
	root.Use(catchPanic)
	root.Use(handlers.CompressHandler)
	root.Use(decompressRequest)

	router.Populate(root)

	root.Handle("/v2/metrics", promhttp.Handler())
	corsRoot := cors.New(cors.Options{
		AllowCredentials: true,
		AllowedOrigins:   corsOrigins,
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE"},
		AllowedHeaders: []string{
			"authorization", "content-type", "if-modified-since",
			"x-b3-traceid", "x-b3-parentspanid", "x-b3-spanid", "x-b3-sampled",
			"traceparent", "tracestate",
		},
	}).Handler(root)

	return corsRoot, nil
}

func Serve(ctx context.Context, listenAddr string, serviceName string, baseURL string, corsOrigins []string, webAppBaseURL string, jobTimeout time.Duration, pubCa crypto.PublicKey, privCa crypto.PrivateKey, ks *key.Set, ping_ch chan bool, pool *pgxpool.Pool, store *blob.Storage) (func(), error) {
	handler, err := NewRouter(ctx, pool, store, pubCa, privCa, ks, ping_ch, serviceName, baseURL, webAppBaseURL, corsOrigins, jobTimeout)
	if err != nil {
		return func() {}, err
	}

	srv := http.Server{
		Addr:    listenAddr,
		Handler: handler,
	}
	go func() {
		log.Infof("Listen on %s\n", listenAddr)
		srv.ListenAndServe()
	}()

	cancel := func() {
		err := srv.Shutdown(context.Background())
		if err != nil {
			srv.Close()
		}
	}

	return cancel, nil
}
