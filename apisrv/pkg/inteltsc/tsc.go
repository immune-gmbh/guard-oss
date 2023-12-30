// Client for the Intel Transparent Supply Chain API
package inteltsc

import (
	"context"
	"crypto/x509"
	"embed"
	"encoding/base64"
	"errors"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/beevik/etree"
	"github.com/georgysavva/scany/pgxscan"
	acert "github.com/google/go-attestation/attributecert"
	"github.com/jackc/pgx/v4"
	dsig "github.com/russellhaering/goxmldsig"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/database"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/queue"
	tel "github.com/immune-gmbh/guard/apisrv/v2/pkg/telemetry"
)

var (
	//go:embed certs
	certificateFiles embed.FS
	certificates     map[string]*x509.Certificate
	certificateStore dsig.MemoryX509CertificateStore

	ErrInProgress    = errors.New("in progress")
	ErrUnknownVendor = errors.New("unknown vendor")
	ErrInvalidXML    = errors.New("invalid xml")
)

func init() {
	certificates = make(map[string]*x509.Certificate)
	pool := x509.NewCertPool()
	files, err := certificateFiles.ReadDir("certs")
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		if !file.IsDir() {
			buf, err := certificateFiles.ReadFile(path.Join("certs", file.Name()))
			if err != nil {
				panic(err)
			}
			cert, err := x509.ParseCertificate(buf)
			if err != nil {
				panic(err)
			}
			pool.AddCert(cert)

			if cert.Subject.String() != cert.Issuer.String() {
				if strings.Contains(cert.Subject.String(), "LENOVO_DCG") {
					if _, ok := certificates["Lenovo"]; !ok {
						certificates["Lenovo"] = cert
					} else {
						panic("duplicate Lenovo TSC certificate")
					}
				} else {
					panic("unknown TSC leaf certificate")
				}
			}
		}
	}
	opts := x509.VerifyOptions{Roots: pool}
	for _, cert := range certificates {
		_, err := cert.Verify(opts)
		if err != nil {
			panic(err)
		}
		certificateStore.Roots = append(certificateStore.Roots, cert)
	}
}

type Data struct {
	rawxml string
}

func ParseData(col string) (*Data, error) {
	var data Data

	data.rawxml = col
	return &data, nil
}

func (data *Data) Verify() (*etree.Document, error) {
	unverified := etree.NewDocument()
	err := unverified.ReadFromString(data.rawxml)
	if err != nil {
		return nil, err
	}

	// the dsig library expects certificates to be embedded with the signature
	vendor := unverified.Root().FindElement("/DirectPlatformData/Header/Manufacturer")
	if vendor == nil {
		return nil, ErrInvalidXML
	}
	sigcert, ok := certificates[vendor.Text()]
	if !ok {
		return nil, ErrUnknownVendor
	}
	for _, e := range unverified.Root().FindElements("/DirectPlatformData/Signature/KeyInfo/X509Data") {
		ne := e.CreateElement("X509Certificate")
		ne.SetText(base64.StdEncoding.EncodeToString(sigcert.Raw))
	}

	// Construct a signing context with one or more roots of trust.
	valctx := dsig.NewDefaultValidationContext(&certificateStore)
	validated, err := valctx.Validate(unverified.Root())
	if err != nil {
		return nil, err
	}

	gooddoc := etree.NewDocument()
	gooddoc.SetRoot(validated)
	return gooddoc, nil
}

func Reference(vendor, serial string) string {
	return fmt.Sprintf("tsc/1(%s,%s)", vendor, serial)
}

func Fetch(ctx context.Context, q pgxscan.Querier, ref string) (*Data, []*acert.AttributeCertificate, error) {
	row, err := getRow(ctx, q, ref)
	if err != nil {
		return nil, nil, err
	}

	if row.FinishedAt != nil {
		if row.Data != nil {
			return decode(row)
		} else {
			// no results
			return nil, nil, nil
		}
	} else {
		// report in progress
		return nil, nil, ErrInProgress
	}
}

func Schedule(ctx context.Context, tx pgx.Tx, vendor, serial string, now time.Time) (*Data, []*acert.AttributeCertificate, error) {
	ctx, span := tel.Start(ctx, "inteltsc.Schedule")
	defer span.End()

	ref := Reference(vendor, serial)
	row, err := getRow(ctx, tx, ref)
	switch err {
	case database.ErrNotFound:
		// no report -> schedule background job
		row := &Row{
			Reference: ref,
			CreatedAt: now,
		}
		err = insertRow(ctx, tx, row)
		if err != nil {
			return nil, nil, err
		}
		args := jobArgs{
			Vendor: vendor,
			Serial: serial,
		}
		jrow, err := queue.Enqueue(ctx, tx, jobType, ref, args, now, now)
		if err != nil {
			return nil, nil, err
		}
		tel.Log(ctx).WithFields(map[string]interface{}{"job": jrow.Id, "ref": jrow.Reference}).
			Info("report scheduled")
		return nil, nil, ErrInProgress
	case nil:
		if row.FinishedAt != nil {
			tel.Log(ctx).WithField("ref", row.Reference).Info("report exists")
			if row.Data != nil {
				return decode(row)
			} else {
				// no results
				return nil, nil, nil
			}
		} else {
			// report in progress
			tel.Log(ctx).WithField("ref", row.Reference).Info("report in progress")
			return nil, nil, ErrInProgress
		}
	default:
		return nil, nil, err
	}
}

func decode(row *Row) (*Data, []*acert.AttributeCertificate, error) {
	// results exist
	data, err := ParseData(*row.Data)
	if err != nil {
		return nil, nil, err
	}
	certs := make([]*acert.AttributeCertificate, len(row.Certificates))
	for i, c := range row.Certificates {
		der, err := base64.StdEncoding.DecodeString(c)
		if err != nil {
			return nil, nil, err
		}
		certs[i], err = acert.ParseAttributeCertificate(der)
		if err != nil {
			return nil, nil, err
		}
	}
	return data, certs, nil
}
