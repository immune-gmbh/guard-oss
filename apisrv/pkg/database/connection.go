package database

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"regexp"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	tel "github.com/immune-gmbh/guard/apisrv/v2/pkg/telemetry"
)

var (
	// supposed to match:
	// --
	// -- Blahblahblah
	// --
	queryNameComment = regexp.MustCompile(`^[[:space:]]*--[[:space:]]+--[[:space:]]+([[:print:]]+)[[:space:]]+--[[:space:]]+`)

	// database connection metrics
	idleConnectionGauge  *prometheus.GaugeFunc
	totalConnectionGauge *prometheus.GaugeFunc
	maxConnectionGauge   *prometheus.GaugeFunc
)

func Connect(host string, port uint16, user string, password string, db string, pemCert string, enableTracing bool, enableLogging bool, poolSize int) (*pgxpool.Pool, error) {
	// TLS certificate
	var tlsConfig *tls.Config
	if pemCert != "" {
		pool := x509.NewCertPool()

		if pool.AppendCertsFromPEM([]byte(pemCert)) {
			tlsConfig = &tls.Config{
				RootCAs:    pool,
				ServerName: host,
			}
		} else {
			return nil, fmt.Errorf("cannot parse database TLS certificate")
		}
	}

	// Database connection
	connOpts, err := pgxpool.ParseConfig("")
	if err != nil {
		return nil, err
	}
	connOpts.ConnConfig.Host = host
	connOpts.ConnConfig.Port = port
	connOpts.ConnConfig.User = user
	connOpts.ConnConfig.Password = password
	connOpts.ConnConfig.Database = db
	connOpts.ConnConfig.TLSConfig = tlsConfig
	connOpts.LazyConnect = true
	connOpts.MaxConns = int32(poolSize)
	connOpts.HealthCheckPeriod = time.Second * 30
	connOpts.MaxConnLifetime = time.Minute

	if enableTracing {
		connOpts.ConnConfig.Logger = pgxTracer{
			attribs: []attribute.KeyValue{
				attribute.String("db.system", "postgresql"),
				attribute.String("db.user", user),
				attribute.String("db.connection_string",
					fmt.Sprintf("postgres://%s@%s:%d/%s?ssl=%t", user, host, port, db, tlsConfig != nil)),
			},
		}
	} else if enableLogging {
		connOpts.ConnConfig.Logger = pgxLogger{}
	}

	pool, err := pgxpool.ConnectConfig(context.Background(), connOpts)
	if err != nil {
		return nil, err
	}

	// Metrics
	opts := prometheus.GaugeOpts{
		Name: "database_connections",
		Help: "Gauge of idle, aquire and total PostgreSQL connections",
	}
	if idleConnectionGauge == nil {
		o := opts
		o.ConstLabels = prometheus.Labels{"state": "idle"}
		g := prometheus.NewGaugeFunc(o, func() float64 { return float64(pool.Stat().IdleConns()) })
		prometheus.MustRegister(g)
		idleConnectionGauge = &g
	}
	if totalConnectionGauge == nil {
		o := opts
		o.ConstLabels = prometheus.Labels{"state": "total"}
		g := prometheus.NewGaugeFunc(o, func() float64 { return float64(pool.Stat().TotalConns()) })
		prometheus.MustRegister(g)
		totalConnectionGauge = &g
	}
	if maxConnectionGauge == nil {
		o := opts
		o.ConstLabels = prometheus.Labels{"state": "max"}
		g := prometheus.NewGaugeFunc(o, func() float64 { return float64(pool.Stat().MaxConns()) })
		prometheus.MustRegister(g)
		maxConnectionGauge = &g
	}

	return pool, nil
}

type pgxLogger struct{}

func (m pgxLogger) Log(ctx context.Context, level pgx.LogLevel, msg string, data map[string]interface{}) {
	if sql, ok := data["sql"]; ok {
		log.Infof("%s", sql)
	}
}

type pgxTracer struct {
	attribs []attribute.KeyValue
}

func (m pgxTracer) Log(ctx context.Context, level pgx.LogLevel, msg string, data map[string]interface{}) {
	if !trace.SpanFromContext(ctx).IsRecording() {
		return
	}

	if t, ok := data["time"]; ok {
		if t, ok := t.(time.Duration); ok {
			endTime := time.Now().UTC()
			startTime := endTime.Add(-1 * t)

			if tracer := tel.Tracer(); tracer != nil {
				_, span := (*tracer).Start(ctx, "Query",
					trace.WithTimestamp(startTime),
					trace.WithSpanKind(trace.SpanKindClient),
					trace.WithAttributes(m.attribs...))

				attribs := []attribute.KeyValue{}
				if ct, ok := data["commandTag"]; ok {
					if ct, ok := ct.(pgconn.CommandTag); ok {
						attribs = append(attribs, attribute.String("db.operation", ct.String()))
					}
				}
				if sql, ok := data["sql"]; ok {
					if sql, ok := sql.(string); ok {
						submatches := queryNameComment.FindStringSubmatch(sql)
						if len(submatches) == 2 {
							attribs = append(attribs, attribute.String("db.statement_shorthand", submatches[1]))
						}
						attribs = append(attribs, attribute.String("db.statement", sql))
					}
				}

				span.SetAttributes(attribs...)
				span.End(trace.WithTimestamp(endTime))
			}
		}
	}
}

// Ping is originally from pgxpool package
func Ping(ctx context.Context, p *pgxpool.Pool) error {
	c, err := p.Acquire(ctx)
	if err != nil {
		return err
	}
	defer c.Release()

	readOnly := ""
	err = c.QueryRow(ctx, "SHOW default_transaction_read_only;").Scan(&readOnly)
	if err != nil {
		return err
	} else if readOnly != "off" {
		return fmt.Errorf("default_transaction_read_only (%v): %w", readOnly, ErrInvalidTransactionMode)
	}
	return nil
}
