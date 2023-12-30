// Business logic for device enrollment and attestation. This package
// implements high level handing of incoming evidence, scheduling analysis jobs
// and finalizing the result of them into an appraisal.
package workflow

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"reflect"
	"strconv"
	"time"

	acert "github.com/google/go-attestation/attributecert"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/klauspost/compress/zstd"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/immune-gmbh/agent/v3/pkg/typevisit"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/api"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/appraisal"
	auth "github.com/immune-gmbh/guard/apisrv/v2/pkg/authentication"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/binarly"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/blob"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/check"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/check/baseline"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/check/policy"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/configuration"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/database"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/device"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/event"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/evidence"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/inteltsc"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/organization"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/report"
	tel "github.com/immune-gmbh/guard/apisrv/v2/pkg/telemetry"
)

var (
	// The user's device quota is full
	ErrQuotaExceeded = errors.New("quota exceeded")
	// The evidence is signed by and unknown or invalid key
	ErrNoAttestationKey = errors.New("no aik")
	// The device to be attested is marked as retired
	ErrDeviceRetired = errors.New("device retired")
	// A referenced blob could not be found
	ErrBlobMissing = errors.New("blob-missing")
)

var (
	DisableIntelTSC = false
	DisableBinarly  = false
)

var BlobStoreVisitor *typevisit.TypeVisitorTree

func init() {
	// construct a type visitor tree for re-use
	tvt, err := typevisit.New(&api.FirmwareProperties{}, api.HashBlob{}, "blobstore")
	if err != nil {
		panic(err)
	}
	BlobStoreVisitor = tvt
}

// Verify incoming evidence, deconstruct and deduplicate sections of it and
// schedule background jobs handing longer running analysis tasks. May return
// an appraisal if all analysis results were already cached, otherwise returns
// nil.
func Attest(ctx context.Context, pool *pgxpool.Pool, store *blob.Storage, ev *api.Evidence, hashBlobs map[string]multipart.File, aikName *api.Name, serviceName string, now time.Time, jobTimeout time.Duration) (*api.Appraisal, error) {
	ctx, span := tel.Start(ctx, "workflow.Attest")
	defer span.End()

	// 1st database transaction: fetch device. retrieveDevice() is read only so
	// tx should never fail b/c of serialization issues.
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
	if err != nil {
		tel.Log(ctx).WithError(err).Error("start 1st db tx")
		return nil, err
	}
	defer tx.Rollback(ctx)

	// get device and correct AIK from db, check organization's quota and whether
	// the device is retired
	dev, bline, pol, err := retrieveDevice(ctx, tx, aikName)
	if err != nil {
		return nil, err
	}

	// commit transaction
	err = tx.Commit(ctx)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("commit 1st tx")
		return nil, err
	}

	// preflight evidence verification and flash image dedup
	values, blobs, flash, err := validateEvidence(ctx, pool, store, dev, ev, hashBlobs, now)
	if err != nil {
		return nil, err
	}

	var appr *api.Appraisal
	// 2nd database transaction: schedule analysis and optionally appraise evidence
	// we must use repeatable read isolation to prevent race conditions in concurrently
	// running appraiseOne calls
	err = database.QueryIsolatedRetry(ctx, pool, func(ctx context.Context, tx pgx.Tx) error {
		//XXX: we need to temporarily make SMBIOS abailable again as inline data for scheduleTasks TSC stuff to work
		//XXX: remove it again after scheduleTasks to make sure it is not stored in the database when it has been uploaded to S3
		//XXX: also make sure that this workaround is not used if SMBIOS is no longer stored in S3
		//XXX: in general all operations working on the data should be running in the same processing stage as the checks do, so we can have a proper pipeline
		if v, ok := blobs[hex.EncodeToString([]byte(values.SMBIOS.Sha256))]; ok {
			values.SMBIOS.Data = v
		}

		// schedule background jobs, return refs and cached results, if any
		binRep, binRef, tscData, tscCerts, tscRef, err := scheduleTasks(ctx, tx, flash, values, now)
		if err != nil {
			return err
		}

		// XXX see above XXX comments
		// remove data again to make sure it is not stored in the database
		values.SMBIOS.Data = nil

		evrow, ready, err := persistEvidence(ctx, tx, dev, values, bline, pol, flash, binRef, tscRef, now, jobTimeout)
		if err != nil {
			return err
		}
		span.AddEvent("evidence store", trace.WithAttributes(attribute.String("guard.evidence.id", evrow.Id)))

		// use read-back row as values to build subject. this is
		// a) consistent with how Appraise() works
		// b) makes sure all values pass through FromRow(..) before being consumed
		//    by the checks engine - giving us a single place to do some processing
		//    like unpacking blobs
		values, err = evidence.FromRow(evrow)
		if err != nil {
			tel.Log(ctx).WithError(err).Error("parse evidence values")
			return err
		}

		if ready {
			// build subject
			subj, err := check.NewSubject(ctx, values, bline, pol,
				check.WithBinarly{Report: binRep},
				check.WithIntelTSC{Data: tscData, Certificates: tscCerts},
				check.WithBlobs{Blobs: blobs})
			if err != nil {
				return err
			}

			// create an appraisal from evidence
			appr, err = appraiseOne(ctx, tx, store, evrow, subj, serviceName, now)
			if err != nil {
				return err
			}
			span.AddEvent("appraisal", trace.WithAttributes(attribute.String("guard.appraisal.id", appr.Id)))
		}

		return nil
	})

	return appr, err
}

var errNotReady = errors.New("not ready")

// Called by queue after each job. Checks for all in progress appraisals
// whether depended jobs are done and finishes appraisals where that's the
// case.
func Appraise(ctx context.Context, pool *pgxpool.Pool, store *blob.Storage, ref string, serviceName string, now time.Time, jobTimeout time.Duration) error {
	ctx, span := tel.Start(ctx, "workflow.Appraise")
	defer span.End()

	span.SetAttributes(
		attribute.String("job.timeout", jobTimeout.String()),
		attribute.String("job.reference", ref))

	maxAge := now.Add(-1 * jobTimeout)
	candidates, err := evidence.ByReference(ctx, pool, ref)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("fetch affected evidence")
		return err
	}

	for _, ev := range candidates {
		// begin tx
		// we must use repeatable read isolation to prevent race conditions in concurrently
		// running appraiseOne calls
		err = database.QueryIsolatedRetry(ctx, pool, func(ctx context.Context, tx pgx.Tx) error {
			// all tasks done?
			ready, err := evidence.IsReadyAndLatest(ctx, tx, &ev, maxAge)
			if err != nil || !ready {
				if err != nil {
					tel.Log(ctx).WithError(err).Error("check evidence readiness")
				}
				tx.Rollback(ctx)
				return nil
			}

			// create appraisal
			_, err = appraiseOne(ctx, tx, store, &ev, nil /* maybeSubj */, serviceName, now)
			if err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			return nil
		}
	}

	return nil
}

// Finalizes a single in progress attestation
// this function, when used concurrently, needs a transaction with nonrepeatable read protection
// b/c the device re-read might produce dirty appraisal references
func appraiseOne(ctx context.Context, tx pgx.Tx, store *blob.Storage, ev *evidence.Row, maybeSubj *check.Subject, serviceName string, now time.Time) (*api.Appraisal, error) { // fetch the device again. It may have changed since uploading the image
	ctx, span := tel.Start(ctx, "workflow.appraiseOne")
	defer span.End()

	dev, _, _, err := retrieveDevice(ctx, tx, &ev.SignedBy)
	if err != nil {
		return nil, err
	}
	// fetch analysis results
	checkResult, rep, err := runAnalysis(ctx, tx, store, ev, maybeSubj, dev, now)
	if err != nil {
		return nil, err
	}

	// create and persist appraisal structure
	newAppraisalId, err := appraisal.Create(ctx, tx, ev, rep, checkResult, dev.AIK, dev.Id, now, auth.SubjectAgent)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("create appraisal")
		return nil, err
	}

	// just fetch the latest appraisals instead of re-fetching the whole device
	// runAnalysis may update the baseline, but it is not used in subsequent code
	// as the device is just converted to an API structure that does not include the baseline
	// thus it is okay to re-use the old device data
	appraisals, err := appraisal.GetLatestAppraisalsByDeviceId(ctx, tx, dev.Id, false, 2)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("fetch new appraisal set")
		return nil, err
	}

	// sanity
	if len(appraisals) == 0 || appraisals[0].Id != strconv.FormatInt(newAppraisalId, 10) {
		tel.Log(ctx).Error("new appraisal not found")
		return nil, database.ErrNotFound
	}

	// send event
	// first construct a minimal api struct conveying the necessary device fields
	devapi := api.Device{Id: fmt.Sprint(dev.Id), Name: dev.Name}

	nextappr := appraisals[0]
	var prevappr *api.Appraisal
	if len(appraisals) > 1 {
		prevappr = appraisals[1]
	}

	organizationExternal, err := organization.GetExternalById(ctx, tx, dev.OrganizationId)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("get org external id")
		return nil, err
	}

	_, err = event.NewAppraisal(ctx, tx, serviceName, *organizationExternal, &devapi, prevappr, nextappr, now)
	if err != nil {
		tel.Log(ctx).WithError(err).WithField("previd", prevappr.Id).WithField("nextid", nextappr.Id).Error("send event")
		return nil, err
	}
	return nextappr, nil
}

// Fetches the device to be attested from the database and does preflight quota
// and integrity checks on it.
func retrieveDevice(ctx context.Context, tx pgx.Tx, aikName *api.Name) (*device.DevAikRow, *baseline.Values, *policy.Values, error) {
	ctx, span := tel.Start(ctx, "workflow.retrieveDevice")
	defer span.End()

	// fetch device for key
	dev, err := device.GetByFingerprint(ctx, tx, aikName)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("fetch device by fpr")
		return nil, nil, nil, err
	}
	span.SetAttributes(attribute.Int64("guard.device.id", dev.Id))
	span.SetAttributes(attribute.String("guard.device.name", dev.Name))
	span.SetAttributes(attribute.Int64("guard.organization.id", dev.OrganizationId))

	// check device state
	if dev.Retired {
		tel.Log(ctx).Error("device retired")
		return nil, nil, nil, ErrDeviceRetired
	}

	// check quota
	current, allowed, err := organization.Quota(ctx, tx, dev.OrganizationId)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("fetch quota")
		return nil, nil, nil, err
	}
	if current > allowed {
		tel.Log(ctx).
			WithFields(log.Fields{"current": current, "allowed": allowed}).
			Info("over quota")
		return nil, nil, nil, ErrQuotaExceeded
	}

	// check that the device has an attestation key enrolled
	if dev.AIK == nil {
		tel.Log(ctx).Error("no attestation key")
		return nil, nil, nil, ErrNoAttestationKey
	}

	// parse baseline
	bline, err := baseline.FromRow(dev.Baseline)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("parse baseline")
		return nil, nil, nil, err
	}

	// parse policy
	pol, err := dev.GetPolicy()
	if err != nil {
		tel.Log(ctx).WithError(err).Error("parse policy")
		return nil, nil, nil, err
	}

	return dev, bline, pol, nil
}

func scheduleTasks(ctx context.Context, tx pgx.Tx, flash *blob.Row, values *evidence.Values, now time.Time) (*binarly.Report, string, *inteltsc.Data, []*acert.AttributeCertificate, string, error) {
	ctx, span := tel.Start(ctx, "workflow.scheduleTasks")
	defer span.End()

	// schedule binarly analysis -> get ref
	var (
		err    error
		binRef string
		bin    *binarly.Report
	)
	if !DisableBinarly && flash != nil {
		bin, err = binarly.Analyse(ctx, tx, flash, now)
		if err != nil && err != binarly.ErrInProgress {
			tel.Log(ctx).WithError(err).Error("schedule binarly analysis")
			return nil, "", nil, nil, "", err
		}
		if err == nil || err == binarly.ErrInProgress {
			binRef = binarly.Reference(flash)
		}
	}

	// schedule inteltsc analysis -> get ref
	var (
		tscRef string
		data   *inteltsc.Data
		certs  []*acert.AttributeCertificate
	)
	if !DisableIntelTSC {
		vendor, serialno, err := values.PlatformSerial()
		if err == nil {
			data, certs, err = inteltsc.Schedule(ctx, tx, vendor, serialno, now)
			if err != nil && err != inteltsc.ErrInProgress {
				tel.Log(ctx).WithError(err).Error("schedule intel tsc retrival")
				return nil, "", nil, nil, "", err
			}
			if err == nil || err == inteltsc.ErrInProgress {
				tscRef = inteltsc.Reference(vendor, serialno)
			}
		}
	}

	return bin, binRef, data, certs, tscRef, nil
}

func buildSubject(ctx context.Context, tx pgx.Tx, store *blob.Storage, ev *evidence.Row, allowMissing bool) (*check.Subject, error) {
	ctx, span := tel.Start(ctx, "workflow.buildSubject")
	defer span.End()

	// decode misc evidence values
	values, err := evidence.FromRow(ev)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("parse evidence values")
		return nil, err
	}

	// decode baseline
	bline, err := baseline.FromRow(ev.Baseline)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("parse baseline")
		return nil, err
	}

	// decode policy
	pol, err := policy.Parse(ev.Policy)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("parse policy")
		return nil, err
	}

	isMissing := func(err error) bool {
		return errors.Is(err, binarly.ErrInProgress) ||
			errors.Is(err, database.ErrNotFound) ||
			errors.Is(err, inteltsc.ErrInProgress)
	}

	// fetch binarly analysis
	var bin *binarly.Report
	if ev.BinarlyReference != nil {
		bin, err = binarly.Fetch(ctx, tx, *ev.BinarlyReference)
		if err != nil {
			if !isMissing(err) || !allowMissing {
				tel.Log(ctx).WithError(err).Error("fetch binarly result")
				return nil, err
			}
		}
	}
	// fetch inteltsc analysis
	var (
		data  *inteltsc.Data
		certs []*acert.AttributeCertificate
	)
	if ev.IntelTSCReference != nil {
		data, certs, err = inteltsc.Fetch(ctx, tx, *ev.IntelTSCReference)
		if err != nil {
			if !isMissing(err) || !allowMissing {
				tel.Log(ctx).WithError(err).Error("fetch tsc result")
				return nil, err
			}
		}
	}

	// download linked blobs
	uniqueDownloads := make(map[string]bool)
	var batch []*blob.DownloadJob
	for _, digest := range values.EarlyLaunchDrivers {
		if len(digest) != 32 {
			tel.Log(ctx).WithField("digest", fmt.Sprintf("%x", digest)).Error("invalid blob digest")
			continue
		}

		if _, ok := uniqueDownloads[hex.EncodeToString(digest)]; !ok {
			job := blob.DownloadJob{Namespace: blob.WindowsExecutable}
			copy(job.Digest[:], []byte(digest))
			batch = append(batch, &job)
			uniqueDownloads[hex.EncodeToString(digest)] = true
		}
	}
	for _, digest := range values.AntiMalwareProcesses {
		if len(digest) != 32 {
			tel.Log(ctx).WithField("digest", fmt.Sprintf("%x", digest)).Error("invalid blob digest")
			continue
		}

		if _, ok := uniqueDownloads[hex.EncodeToString(digest)]; !ok {
			job := blob.DownloadJob{Namespace: blob.WindowsExecutable}
			copy(job.Digest[:], []byte(digest))
			batch = append(batch, &job)
			uniqueDownloads[hex.EncodeToString(digest)] = true
		}
	}
	for _, digest := range values.BootApps {
		if len(digest) != 32 {
			tel.Log(ctx).WithField("digest", fmt.Sprintf("%x", digest)).Error("invalid blob digest")
			continue
		}

		if _, ok := uniqueDownloads[hex.EncodeToString(digest)]; !ok {
			job := blob.DownloadJob{Namespace: blob.UEFIApp}
			copy(job.Digest[:], []byte(digest))
			batch = append(batch, &job)
			uniqueDownloads[hex.EncodeToString(digest)] = true
		}
	}
	for _, hb := range values.TPM2EventLogs {
		// this blob was stored inline
		if len(hb.Data) > 0 {
			continue
		}
		if len(hb.Sha256) != 32 {
			tel.Log(ctx).WithField("digest", fmt.Sprintf("%x", hb.Sha256)).Error("invalid blob digest")
			continue
		}

		if _, ok := uniqueDownloads[hex.EncodeToString(hb.Sha256)]; !ok {
			job := blob.DownloadJob{Namespace: blob.Eventlog}
			copy(job.Digest[:], []byte(hb.Sha256))
			batch = append(batch, &job)
			uniqueDownloads[hex.EncodeToString(hb.Sha256)] = true
		}
	}
	for _, hb := range values.ACPI.Blobs {
		// this blob was stored inline
		if len(hb.Data) > 0 {
			continue
		}
		if len(hb.Sha256) != 32 {
			tel.Log(ctx).WithField("digest", fmt.Sprintf("%x", hb.Sha256)).Error("invalid blob digest")
			continue
		}

		if _, ok := uniqueDownloads[hex.EncodeToString(hb.Sha256)]; !ok {
			job := blob.DownloadJob{Namespace: blob.ACPI}
			copy(job.Digest[:], []byte(hb.Sha256))
			batch = append(batch, &job)
			uniqueDownloads[hex.EncodeToString(hb.Sha256)] = true
		}
	}
	// some blobs are inline and it is the case then the data member is already set
	if len(values.SMBIOS.Data) == 0 && len(values.SMBIOS.Sha256) == 32 {
		if _, ok := uniqueDownloads[hex.EncodeToString(values.SMBIOS.Sha256)]; !ok {
			job := blob.DownloadJob{Namespace: blob.SMBIOS}
			copy(job.Digest[:], []byte(values.SMBIOS.Sha256))
			batch = append(batch, &job)
			uniqueDownloads[hex.EncodeToString(values.SMBIOS.Sha256)] = true
		}
	}
	// some blobs are inline and it is the case then the data member is already set
	if len(values.TXTPublicSpace.Data) == 0 && len(values.TXTPublicSpace.Sha256) == 32 {
		if _, ok := uniqueDownloads[hex.EncodeToString(values.TXTPublicSpace.Sha256)]; !ok {
			job := blob.DownloadJob{Namespace: blob.TXT}
			copy(job.Digest[:], []byte(values.TXTPublicSpace.Sha256))
			batch = append(batch, &job)
			uniqueDownloads[hex.EncodeToString(values.TXTPublicSpace.Sha256)] = true
		}
	}

	err = store.DownloadBatch(ctx, tx, batch)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("batch")
		return nil, err
	}
	var blobs = make(map[string][]byte)
	for _, job := range batch {
		if job.Error != nil {
			tel.Log(ctx).WithField("digest", fmt.Sprintf("%x", job.Digest)).WithError(job.Error).Error("download")
			return nil, job.Error
		}
		blobs[fmt.Sprintf("%x", job.Digest[:])] = job.Blob
	}

	// build subject
	subj, err := check.NewSubject(ctx, values, bline, pol,
		check.WithBinarly{Report: bin},
		check.WithIntelTSC{Data: data, Certificates: certs},
		check.WithBlobs{Blobs: blobs})
	if err != nil {
		tel.Log(ctx).WithError(err).Error("subject")
		return nil, err
	}

	return subj, nil
}

func runAnalysis(ctx context.Context, tx pgx.Tx, store *blob.Storage, row *evidence.Row, subj *check.Subject, dev *device.DevAikRow, now time.Time) (*check.Result, *api.Report, error) {
	ctx, span := tel.Start(ctx, "workflow.runAnalysis")
	defer span.End()

	var err error

	// get results
	if subj == nil {
		subj, err = buildSubject(ctx, tx, store, row, false)
		if err != nil {
			return nil, nil, err
		}
	}

	// run checks
	res, err := check.Run(ctx, subj)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("run check()")
		return nil, nil, err
	}

	// update device baseline
	if subj.BaselineModified {
		err = device.Patch(ctx, tx, dev.Id, dev.OrganizationId, nil, subj.Baseline, nil, now, auth.SubjectAgent)
		if err != nil {
			tel.Log(ctx).WithError(err).Error("update baseline")
			return nil, nil, err
		}
	}

	// compile report (deprecated)
	rep, err := report.Compile(ctx, subj.Values)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("compile report")
		return nil, nil, err
	}

	return res, rep, nil
}

// handleHashBlob puts OOB transferred, compressed blobs back into the HashBlob struct and if a blob should be uploaded to S3
// then it uncompresses the data, verifies the hash and returns a blob.UploadJob.
// when an upload job is created, the ZData and Data members are set to nil.
// if no upload job is created, Data is left untouched and ZData will contain the possibly plugged-back OOB transferred blob.
// this means that ZData and Data would be stored in the database inside the evidence json.
func handleHashBlob(blobsIn map[string]multipart.File, hb *api.HashBlob, opts typevisit.FieldOpts, decoder *zstd.Decoder) (*blob.UploadJob, error) {
	// when ZData is empty we must have a multipart blob or the data member must be set, otherwise it is an error
	// this ensures that clients get visible errors and tests fail when out-of-band transfer is broken client-side
	// it however is complicated with zero length files in which case we must implement a differentiation between nil and an array with no elements
	oob := false
	if len(hb.ZData) == 0 && len(hb.Sha256) == 32 {
		key := hex.EncodeToString(hb.Sha256)
		if val, ok := blobsIn[key]; ok {
			buf, err := io.ReadAll(val)
			if err != nil {
				return nil, err
			}
			hb.ZData = buf
			oob = true
		} else {
			return nil, ErrBlobMissing
		}
	} else if len(hb.Data) == 0 {
		// just ignore broken stuff (hash length mismatch) and empty blobs
		return nil, nil
	}

	// skip s3 upload if no blobstore options are set
	if len(opts) == 0 {
		// check hash in case we had an OOB transfer and also nil Data b/c it would be overriden anyway when ZData is decompressed
		if oob {
			uncompressed, err := decoder.DecodeAll(hb.ZData, make([]byte, 0, len(hb.ZData)))
			if err != nil {
				return nil, err
			}

			// validate that the uncompressed contents match the hash to be sure that
			// this is the one that was included in the quote
			measuredHash := sha256.Sum256(uncompressed)
			if !bytes.Equal(measuredHash[:], hb.Sha256) {
				return nil, ErrBlobMissing
			}

			hb.Data = nil
		}

		return nil, nil
	}

	// decompress blobs here before uploading to s3
	var buf []byte
	if len(hb.ZData) > 0 {
		var err error
		buf, err = decoder.DecodeAll(hb.ZData, make([]byte, 0, len(hb.ZData)))
		if err != nil {
			return nil, err
		}

		// validate that the uncompressed contents match the hash to be sure that
		// this is the one that was included in the quote
		measuredHash := sha256.Sum256(buf)
		if !bytes.Equal(measuredHash[:], hb.Sha256) {
			return nil, ErrBlobMissing
		}
	} else {
		// hash data that was transferred uncompressed inline (which is the case for blobs < 1024 bytes)
		// and store the hash within the evidence structure as this is the
		// reference that is required to download the deduplicated data later
		// (hash needs no verification because the full data was part of the quote)
		h := sha256.Sum256(hb.Data)
		hb.Sha256 = h[:]
		buf = hb.Data
	}

	// finally create an upload job with the uncompressed data
	up := blob.UploadJob{
		Contents:  buf,
		Namespace: string(opts),
	}

	// if we don't nil Data & ZData here it will be stored in DB and S3
	hb.ZData = nil
	hb.Data = nil

	return &up, nil
}

/*
XXX
The way this function works with is problematic in any way. There are some blobs that are deduplicated whose SHA256 references
are stored in evidence.Values and the hashes are used for OOB transfer and the blobs are maybe stored in S3. But the bios flash
works completely different, it has a different mechanism to be passed along and it's SHA256 hash is not stored in evidence.Values.
This means we can not work on evidence.Values to plug back OOB transfered blobs into the evidence.Values. We must work on the raw evidence
to do this. This is a problem because when the evidence is loaded from the DB it is evidence.Values which lacks the tags used for S3 storage.
So either we move the tags to evidence.Values and must do two passes over the evidence, one to load the blobs from OOB over the raw evidence and one to store the blobs in S3 over evidence.Values
or we must maintain the struct tags in two places. This is a mess.
*/
func validateEvidence(ctx context.Context, pool *pgxpool.Pool, store *blob.Storage, dev *device.DevAikRow, ev *api.Evidence, hashBlobs map[string]multipart.File, now time.Time) (*evidence.Values, map[string][]byte, *blob.Row, error) {
	ctx, span := tel.Start(ctx, "workflow.validateEvidence")
	defer span.End()

	// verify evidence signatrue with AIK
	// verify evidence matches configuration
	err := evidence.Validate(ctx, ev, &configuration.DefaultConfiguration, dev.Fingerprint, &dev.AIK.Public)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("validate evidence")
		return nil, nil, nil, err
	}

	// decompress and upload blobs
	var batch []*blob.UploadJob

	// flash special legacy case:
	// the data field that is supposed to be uncompressed really is compressed here, but we have no hash
	// of the uncompressed data so we can't just stick it into the zdata field. instead we just manually
	// upload it to S3. newer agents transfer a proper hashblob with hash and zdata, obsoleting this conditional block
	legacyFlash := ev.Firmware.Flash.Error == "" && len(ev.Firmware.Flash.Data) > 0 && len(ev.Firmware.Flash.Sha256) == 0
	if legacyFlash {
		up := blob.UploadJob{
			CompressedContents: ev.Firmware.Flash.Data,
			Namespace:          blob.BIOS,
		}
		batch = append(batch, &up)
	}

	// plug back out-of-band transferred hashblobs into evidence and upload tagged ones to S3 after we have validated it
	// the validation does the expensive JCS transform and hashing so we must make sure that we
	// verify that the hash of our out-of-band data matches the hashblob hash value in the evidence that was verified as part of the quoted extradata
	//XXX: we could become a lot more efficient if we passed along stream readers to the S3 and validate the hash only when we read the data during upload
	//XXX: we could even skip parsing the multipart data in router.doAttest and require that the first part is always the evidence and the remaining parts be the referenced hashes in a sorted order
	//XXX: we use an map here to only create one upload job for a given hash; the evidence structure could reference the same OOB blob multiple times (f.e. boot apps) and due to the snowflaking it would end up being uploaded multiple times); this could be optimized by storing the upload jobs inside the map and passing it directly to s3 upload
	decoder, err := zstd.NewReader(nil)
	if err != nil {
		return nil, nil, nil, err
	}
	defer decoder.Close()
	uniqueUploads := make(map[string]bool)
	BlobStoreVisitor.Visit(&ev.Firmware, func(v reflect.Value, opts typevisit.FieldOpts) {
		// sanitize opts, we only accept certain values
		if opts != blob.WindowsExecutable && opts != blob.UEFIApp && opts != blob.BIOS && opts != blob.Eventlog && opts != blob.SMBIOS && opts != blob.TXT && opts != blob.ACPI {
			opts = ""
		}

		// we need a special treatment for maps because there are no pointers to map elements
		// and that would prevent us from being able to plug-back the values
		if v.Kind() == reflect.Map {
			mi := v.MapRange()
			for mi.Next() {
				hb := mi.Value().Interface().(api.HashBlob)
				if up, err := handleHashBlob(hashBlobs, &hb, opts, decoder); err != nil {
					tel.Log(ctx).WithError(err).Error("read hashblob")
				} else {
					if up != nil {
						hexHash := hex.EncodeToString(hb.Sha256)
						if _, ok := uniqueUploads[hexHash]; !ok {
							batch = append(batch, up)
							uniqueUploads[hexHash] = true
						}
					}

					// plug the updated hashblob back into the map
					v.SetMapIndex(mi.Key(), reflect.ValueOf(hb))
				}
			}
		} else {
			hb := v.Addr().Interface().(*api.HashBlob)
			if up, err := handleHashBlob(hashBlobs, hb, opts, decoder); err != nil {
				tel.Log(ctx).WithError(err).Error("read hashblob")
			} else if up != nil {
				hexHash := hex.EncodeToString(hb.Sha256)
				if _, ok := uniqueUploads[hexHash]; !ok {
					batch = append(batch, up)
					uniqueUploads[hexHash] = true
				}
			}
		}
	})

	// we must wrap the evidence after in-line stored blobs that were OOB transferred are
	// plugged back to ensure those blobs make it into the database
	values, err := evidence.WrapInsecure(ev)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("wrap evidence")
		return nil, nil, nil, err
	}

	store.UploadBatch(ctx, pool, batch, now)
	blobs := make(map[string][]byte)

	// special case for in-line flash
	var flashDigest string
	var flash *blob.Row
	if !legacyFlash && len(ev.Firmware.Flash.Sha256) > 0 {
		flashDigest = hex.EncodeToString([]byte(ev.Firmware.Flash.Sha256))
	}
	for i, b := range batch {
		if b.Error != nil {
			tel.Log(ctx).WithError(b.Error).WithField("index", i).Error("upload blob")
			return nil, nil, nil, b.Error
		}
		blobs[b.Row.Digest] = b.Contents
		if b.Row.Digest == flashDigest {
			flash = b.Row
		}
	}
	if legacyFlash && len(blobs) > 0 {
		flash = batch[0].Row
	}

	return values, blobs, flash, nil
}

func persistEvidence(ctx context.Context, tx pgx.Tx, device *device.DevAikRow, values *evidence.Values, bline *baseline.Values, pol *policy.Values, flash *blob.Row, binRef string, tscRef string, now time.Time, jobTimeout time.Duration) (*evidence.Row, bool, error) {
	maxAge := now.Add(-1 * jobTimeout)
	ctx, span := tel.Start(ctx, "workflow.persistEvidence")
	defer span.End()

	var digest string
	if flash != nil {
		digest = flash.Digest
	}
	// persist evidence data and try to finalize attestation.
	// keep the tx open to prevent another actor from finalizing.
	row, err := evidence.Persist(ctx, tx, values, bline, pol, device, digest, binRef, tscRef, now)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("persist evidence")
		return nil, false, err
	}
	ready, err := evidence.IsReadyAndLatest(ctx, tx, row, maxAge)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("check evidence readiness")
		return nil, false, err
	}

	return row, ready, nil
}
