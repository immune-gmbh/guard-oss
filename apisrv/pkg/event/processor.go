package event

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptrace"
	"time"

	ce "github.com/cloudevents/sdk-go/v2"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/collector/semconv/v1.9.0"
	"go.opentelemetry.io/contrib/instrumentation/net/http/httptrace/otelhttptrace"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	auth "github.com/immune-gmbh/guard/apisrv/v2/pkg/authentication"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/key"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/queue"
	tel "github.com/immune-gmbh/guard/apisrv/v2/pkg/telemetry"
)

const (
	jobType = "Event-v1"
)

var (
	UnknownOptionErr       = errors.New("unknown queue option")
	InvalidOptionErr       = errors.New("invalid queue option")
	EndpointUnreachableErr = errors.New("endpoint unreachable")
	defaultBackoff         = queue.Exponetial{
		Min: 5 * time.Second,
		Max: 30 * time.Minute,
	}

	// event metrics
	eventCounter *prometheus.CounterVec
	eventLatency *prometheus.HistogramVec
)

func init() {
	eventCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "events_outgoing_total",
			Help: "Counter of all outgoing cloudevents, partitioned by type, target and status",
		},
		[]string{"result", "endpoint", "type"})
	eventLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "events_outgoing_latency",
			Help: "Histogram of in flight cloudevent latency, partitioned by type, target and status",
		},
		[]string{"result", "endpoint", "type"})

	prometheus.DefaultRegisterer.MustRegister(eventCounter)
	prometheus.DefaultRegisterer.MustRegister(eventLatency)
}

type Processor struct {
	endpoint    string
	serviceName string
	kid         string
	privateKey  *ecdsa.PrivateKey
	client      *http.Client
}

type WithCredentials struct {
	PrivateKey *ecdsa.PrivateKey
	Kid        string
}

type WithClient struct {
	Client *http.Client
}

func NewProcessor(ctx context.Context, endpoint string, serviceName string, opts ...interface{}) (*Processor, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	kid, err := key.ComputeKid(&privateKey.PublicKey)
	if err != nil {
		return nil, err
	}
	var client *http.Client = &http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}

	for _, opt := range opts {
		switch opt := opt.(type) {
		case WithClient:
			if opt.Client == nil {
				tel.Log(ctx).Error("creds cannot be nil")
				return nil, InvalidOptionErr
			}
			client = opt.Client

		case WithCredentials:
			if opt.PrivateKey == nil || opt.Kid == "" {
				tel.Log(ctx).Error("creds cannot be nil")
				return nil, InvalidOptionErr
			}
			privateKey = opt.PrivateKey
			kid = opt.Kid

		default:
			tel.Log(ctx).WithField("opt", opt).Error("unknown processor option")
			return nil, UnknownOptionErr
		}
	}

	proc := Processor{
		serviceName: serviceName,
		kid:         kid,
		privateKey:  privateKey,
		client:      client,
		endpoint:    endpoint,
	}

	return &proc, nil
}

func (p *Processor) Type() string {
	return jobType
}

func (p *Processor) Run(ctx context.Context, job *queue.Job) {
	ctx, span := tel.Start(ctx, "Send Event")
	defer span.End()

	ctx = httptrace.WithClientTrace(ctx, otelhttptrace.NewClientTrace(ctx))
	span.SetAttributes(attribute.String(semconv.AttributeMessagingURL, p.endpoint))

	var event ce.Event
	err := job.Arguments(&event)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("parse event")
		job.Failed()
		return
	}
	span.SetAttributes(attribute.String(semconv.AttributeMessagingMessageID, event.ID()))

	var sub *string
	if cesub := event.Subject(); cesub != "" && cesub != "ext-1" {
		sub = &cesub
	}
	token, err := auth.IssueServiceCredential(p.serviceName, sub, time.Now(), p.kid, p.privateKey)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("issue token")
		job.Failed()
		return
	}

	client, err := ce.NewClientHTTP(
		cehttp.WithClient(*p.client),
		cehttp.WithTarget(p.endpoint),
		cehttp.WithMethod(http.MethodPost),
		cehttp.WithHeader("Authorization", fmt.Sprintf("Bearer %s", token)),
	)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("create client")
		job.Failed()
		return
	}

	begin := time.Now()
	result := client.Send(ctx, event)
	if err != nil {
		tel.Log(ctx).WithError(result).Error("send")
	}
	ret := "unknown"
	if ce.IsACK(result) {
		ret = "ack"
		job.Done()
	} else if ce.IsNACK(result) {
		ret = "nack"
		job.Retry(defaultBackoff)
		span.SetStatus(codes.Error, ret)
	} else if ce.IsUndelivered(result) {
		ret = "undelivered"
		job.Retry(defaultBackoff)
		span.SetStatus(codes.Error, ret)
	} else {
		ret = "unknown"
		job.Retry(defaultBackoff)
		span.SetStatus(codes.Error, ret)
	}

	eventCounter.WithLabelValues(ret, p.endpoint, event.Type()).Add(1)
	eventLatency.WithLabelValues(ret, p.endpoint, event.Type()).Observe(float64(time.Since(begin)) / float64(time.Second))
	tel.Log(ctx).WithField("result", ret).Info("send event")
}

func (p *Processor) PingReceiver() error {
	token, err := auth.IssueServiceCredential(p.serviceName, nil, time.Now(), p.kid, p.privateKey)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodHead, p.endpoint, bytes.NewBuffer(nil))
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return EndpointUnreachableErr
	} else {
		return nil
	}
}
