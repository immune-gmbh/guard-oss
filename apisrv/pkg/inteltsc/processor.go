package inteltsc

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/queue"
	tel "github.com/immune-gmbh/guard/apisrv/v2/pkg/telemetry"
)

const (
	jobType      = "IntelTSC-v1"
	pollInterval = time.Second * 10
)

var (
	defaultBackoff = queue.Exponetial{
		Min: 5 * time.Second,
		Max: 30 * time.Minute,
	}

	ErrIllegalConfig = errors.New("illegal config")
)

type Processor struct {
	pool   *pgxpool.Pool
	sites  map[string]Site
	client *http.Client
}

func NewProcessor(ctx context.Context, pool *pgxpool.Pool, sites []Site) (*Processor, error) {
	var client = &http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}
	sitesmap := make(map[string]Site, len(sites))
	for _, site := range sites {
		if _, ok := sitesmap[site.Vendor]; ok {
			tel.Log(ctx).WithField("vendor", site.Vendor).Error("dup site")
			return nil, ErrIllegalConfig
		}
		sitesmap[site.Vendor] = site
	}

	return &Processor{pool: pool, sites: sitesmap, client: client}, nil
}

type jobArgs struct {
	Vendor string `json:"vendor"`
	Serial string `json:"serial"`
}

func (p *Processor) Type() string {
	return jobType
}

func (p *Processor) Run(ctx context.Context, job *queue.Job) {
	// parse args
	var args jobArgs
	err := job.Arguments(&args)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("parse args")
		job.Failed()
		return
	}

	row, err := getRow(ctx, p.pool, job.Row.Reference)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("fetch row")
		job.Retry(defaultBackoff)
		return
	}

	if row.Data != nil {
		tel.Log(ctx).Error("data already present")
	}

	site, siteok := p.sites[args.Vendor]
	if siteok {
		// get a oauth access token
		sess, err := openSession(ctx, &site, p.client)
		if err != nil {
			tel.Log(ctx).WithError(err).Error("open session")
			job.Retry(defaultBackoff)
			return
		}

		// fetch file links for serial
		links, err := getFiles(ctx, &site, p.client, sess, args.Serial)
		if err != nil && err != ErrNotFound {
			tel.Log(ctx).WithError(err).Error("fetch links")
			job.Retry(defaultBackoff)
			return
		}

		// download files
		if links != nil && links.LinkZip != "" {
			data, certs, err := downloadFile(ctx, p.client, links)
			if err != nil {
				tel.Log(ctx).WithError(err).Error("download file")
				job.Retry(defaultBackoff)
				return
			}

			row.Data = &data
			row.Certificates = make([]string, len(certs))
			for i, c := range certs {
				row.Certificates[i] = base64.StdEncoding.EncodeToString(c.Raw)
			}
		}
	}

	now := time.Now()
	row.FinishedAt = &now

	err = updateRow(ctx, p.pool, row)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("update row")
		job.Retry(defaultBackoff)
	}
}
