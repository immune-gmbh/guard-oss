package web

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/x509"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"mime/multipart"
	"net/http"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	cebind "github.com/cloudevents/sdk-go/v2/binding"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
	"github.com/google/jsonapi"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/api"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/appraisal"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/authentication"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/blob"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/change"
	policy2 "github.com/immune-gmbh/guard/apisrv/v2/pkg/check/policy"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/configuration"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/database"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/debugv1"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/device"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/event"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/filter"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/key"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/organization"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/tag"
	tel "github.com/immune-gmbh/guard/apisrv/v2/pkg/telemetry"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/workflow"
)

const HTTPStatusClientClosedRequest = 499 // NGINX introduced this custom code

//go:embed readiness.html.template
var readinessTemplateBlob string

var (
	deviceIncludeWhitelist = map[string]struct{}{"appraisals": {}, "changes": {}, "tags": {}}

	numDevices = 64
	numChanges = 64
	numTags    = 64

	// the maximum memory used for multipart form data during request parsing; the remaining data will be stored as tmp files
	maxMultiPartFormMem = int64(1024 * 1024 * 128)

	errInternalServerError = errors.New("internal server error")
	errQueryStringWrong    = errors.New("wrong query string")
	errJsonApiContentType  = errors.New("wrong jsonapi type")
	errInvalidPatch        = errors.New("invalid resource patch")
	errSerialize           = errors.New("cannot serialize response")
	ErrNotEcdsaKey         = errors.New("not a ecdsa private key")
	errRequestIncomplete   = errors.New("client request unexpected EOF")
	errDevicePolicy        = errors.New("invalid policy")
	errEvidence            = errors.New("invalid evidence")
	errEnrollment          = errors.New("invalid enrollment")

	readinessTemplate = template.Must(template.New("readiness.html").Funcs(template.FuncMap{
		"PKCS8": func(pub ecdsa.PublicKey) string {
			if pub.Equal(ecdsa.PublicKey{}) || pub.Y == nil || pub.X == nil {
				return "&lt;nil&gt;"
			}
			buf, err := x509.MarshalPKIXPublicKey(&pub)
			if err != nil {
				return err.Error()
			}
			return base64.StdEncoding.EncodeToString(buf)
		},
	}).Parse(readinessTemplateBlob))
)

type pointQuery struct {
	Id int64
}

type pointQueryStrId struct {
	Id string
}

type rangeQuery struct {
	Start  string
	Length int
}

type setQuery struct {
	Set []int64
}

type setQueryStr struct {
	Set []string
}

type textQuery struct {
	Fragment string
}

type payloadErr struct {
	Err      error
	Argument string
}

func newPayloadErr(arg string, err error) payloadErr {
	return payloadErr{
		Argument: arg,
		Err:      err,
	}
}

func (err payloadErr) Error() string {
	return fmt.Sprintf("%s: %s", err.Argument, err.Err.Error())
}

type Router struct {
	database                *pgxpool.Pool
	store                   *blob.Storage
	debug                   http.Handler
	keyCertificateAuth      *ecdsa.PrivateKey
	keyCertificateKid       string
	keyset                  *key.Set
	keyDiscoveryPingChannel chan bool
	serviceName             string
	baseURL                 string
	webAppBaseURL           string
	jobTimeout              time.Duration
}

func New(ctx context.Context, pool *pgxpool.Pool, store *blob.Storage, pubCa crypto.PublicKey, privCa crypto.PrivateKey, keyset *key.Set, ksPingCh chan bool, serviceName string, baseURL string, webAppBaseURL string, jobTimeout time.Duration) (*Router, error) {
	kid, err := key.ComputeKid(pubCa)
	if err != nil {
		return nil, err
	}

	eccPriv, ok := privCa.(*ecdsa.PrivateKey)
	if !ok {
		return nil, ErrNotEcdsaKey
	}

	// graphql debug API
	debug, err := debugv1.New(ctx, debugv1.WithDatabase{Pool: pool})
	if err != nil {
		return nil, err
	}

	c := Router{
		database:                pool,
		debug:                   debug,
		store:                   store,
		keyCertificateAuth:      eccPriv,
		keyCertificateKid:       kid,
		keyset:                  keyset,
		keyDiscoveryPingChannel: ksPingCh,
		serviceName:             serviceName,
		baseURL:                 baseURL,
		webAppBaseURL:           webAppBaseURL,
		jobTimeout:              jobTimeout,
	}
	return &c, nil
}

func (c *Router) crudAuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tel.Start(r.Context(), "crudAuthenticationMiddleware")
		defer span.End()

		// Bearer token
		// users are always authenticated within the context of an organization and
		// authenticatedUser represents that
		authenticatedUser, err := authenticateCrud(ctx, c.keyset, r)
		if err != nil {
			if errors.Is(err, authentication.ErrSignature) {
				tel.Log(ctx).WithError(err).Info("client login failed")
			} else {
				tel.Log(ctx).WithError(err).Error("auth")
			}

			marshalError(w, r, err)
			return
		}

		// validate if organization exists right here and store the internal id
		// into vars so the code that runs after authentication never has to
		// deal with the technical details of this authentication method
		conn, err := c.database.Acquire(ctx)
		if err != nil {
			tel.Log(ctx).WithError(err).Error("acquire db conn")
			marshalError(w, r, err)
			return
		}
		// defer here to be panic safe, but release conn immediately after use to free it before all handlers have run
		defer conn.Release()
		org, err := organization.Get(ctx, conn, authenticatedUser.OrganizationExternal)
		conn.Release()
		if org == nil || (err != nil && errors.Is(err, database.ErrNotFound)) {
			tel.Log(ctx).WithError(err).Error("no organization for external")
			marshalError(w, r, ErrAuthentication)
			return
		}
		if err != nil {
			tel.Log(ctx).WithError(err).Error("get organization")
			marshalError(w, r, err)
			return
		}

		if authenticatedUser.Actor == "" {
			authenticatedUser.Actor = "tag:immu.ne,2021:anonymous"
		}

		// XXX: Revokation check

		vars := mux.Vars(r)
		vars["org-id"] = strconv.FormatInt(org.Id, 10)
		vars["org-ext"] = authenticatedUser.OrganizationExternal // required for billing updates
		vars["actor"] = authenticatedUser.Actor
		span.End()
		next.ServeHTTP(w, r)
	})
}

func (c *Router) doConfiguration(w http.ResponseWriter, r *http.Request) {
	ctx, span := tel.Start(r.Context(), "Handler")
	defer span.End()

	if mtime, err := time.Parse(http.TimeFormat, r.Header.Get("If-Modified-Since")); err == nil {
		if configuration.DefaultConfigurationModTime.UTC().Before(mtime.UTC()) {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	pdoc, err := jsonapi.Marshal(&configuration.DefaultConfiguration)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("serialize")
		marshalError(w, r, newPayloadErr("configuration", err))
		return
	}

	// set the type
	doc, ok := pdoc.(*jsonapi.OnePayload)
	if !ok {
		tel.Log(ctx).WithError(errSerialize).Error("serialize")
		marshalError(w, r, newPayloadErr("configuration", errSerialize))
		return
	}
	doc.Data.Type = "configurations"

	// write out JSON
	marshalResult(ctx, doc, w)
}

func (c *Router) doInfo(w http.ResponseWriter, r *http.Request) {
	ctx, span := tel.Start(r.Context(), "Handler")
	defer span.End()
	span.SetAttributes(attribute.String("http.handler", "Handler"))

	info := api.Info{
		APIVersion: api.CurrentAPIVersion,
	}
	pdoc, err := jsonapi.Marshal(&info)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("serialize")
		marshalError(w, r, newPayloadErr("configuration", err))
		return
	}

	// set the type
	doc, ok := pdoc.(*jsonapi.OnePayload)
	if !ok {
		tel.Log(ctx).WithError(errSerialize).Error("serialize")
		marshalError(w, r, newPayloadErr("info", errSerialize))
		return
	}
	doc.Data.Type = "infos"

	// write out JSON
	marshalResult(ctx, doc, w)
}

func (c *Router) doHealthy(w http.ResponseWriter, r *http.Request) {
	ctx, span := tel.Start(r.Context(), "Handler")
	defer span.End()

	// ping authsrv event endpoint
	// verify token key against pubkey

	// ping key discovery goroutine
	timeout := time.After(time.Second * 6)
	select {
	case <-timeout:
		tel.Log(ctx).Error("ping key discovery goroutine")
		w.WriteHeader(http.StatusServiceUnavailable)

	case c.keyDiscoveryPingChannel <- true:
		// nop
	}
	w.WriteHeader(http.StatusOK)
}

type readinessTemplateDatabase struct {
	Hostname      string
	Role          string
	RoundTripTime time.Duration
}

type readinessTemplateStore struct {
	URL           string
	Bucket        string
	RoundTripTime time.Duration
}

type readinessTemplateValues struct {
	Service   string
	Timestamp time.Time
	AppURL    string
	APIURL    string

	Database      *readinessTemplateDatabase
	DatabaseError error
	Store         *readinessTemplateStore
	StoreError    error
	Keys          []key.Key
}

func (c *Router) doReady(w http.ResponseWriter, r *http.Request) {
	ctx, span := tel.Start(r.Context(), "Handler")
	defer span.End()

	values := readinessTemplateValues{
		Service:   c.serviceName,
		AppURL:    c.webAppBaseURL,
		APIURL:    c.baseURL,
		Timestamp: time.Now(),
		Keys:      c.keyset.Copy(),
	}
	status := http.StatusOK

	// wait for two parallel ping goroutines
	var wg sync.WaitGroup
	wg.Add(2)

	// ping database with timeout
	ctxDb, cancelDbPing := context.WithTimeout(ctx, time.Second*10)
	defer cancelDbPing()
	var dbrtt time.Duration
	go func() {
		defer wg.Done()
		dbts := time.Now()
		values.DatabaseError = database.Ping(ctxDb, c.database)
		dbrtt = time.Since(dbts)
	}()

	// ping storage with timeout
	ctxS3, cancelS3Ping := context.WithTimeout(ctx, time.Second*10)
	defer cancelS3Ping()
	var strtt time.Duration
	go func() {
		defer wg.Done()
		stts := time.Now()
		values.StoreError = c.store.Ping(ctxS3)
		strtt = time.Since(stts)
	}()

	// keyset
	if len(values.Keys) == 0 {
		tel.Log(ctx).Error("keyset empty")
		status = http.StatusServiceUnavailable
	}

	// wait for goroutines before finishing readyness evaluation
	wg.Wait()

	// process database errors
	if values.DatabaseError != nil {
		tel.Log(ctx).WithError(values.DatabaseError).Error("ping database")
		status = http.StatusServiceUnavailable
	} else {
		dbcfg := c.database.Config()
		values.Database = &readinessTemplateDatabase{
			Hostname:      dbcfg.ConnConfig.Host,
			Role:          dbcfg.ConnConfig.User,
			RoundTripTime: dbrtt,
		}
	}

	// process S3 errors
	if values.StoreError == nil {
		ep, err := c.store.Config.EndpointResolverWithOptions.ResolveEndpoint("s3", "")
		values.Store = &readinessTemplateStore{
			URL:           ep.URL,
			Bucket:        c.store.Bucket,
			RoundTripTime: strtt,
		}
		if err != nil {
			tel.Log(ctx).WithError(values.StoreError).Error("resolve endpoint")
			status = http.StatusServiceUnavailable
		}
	} else {
		tel.Log(ctx).WithError(values.StoreError).Error("ping store")
		status = http.StatusServiceUnavailable
	}

	w.Header().Set("content-type", "text/html")
	w.WriteHeader(status)
	err := readinessTemplate.Execute(w, values)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("execute template")
	}
}

func (c *Router) doEnroll(w http.ResponseWriter, r *http.Request) {
	ctx, span := tel.Start(r.Context(), "Handler")
	defer span.End()
	now := time.Now().UTC()

	w.Header().Set("Content-Type", jsonapi.MediaType)

	// Bearer token
	organizationExternal, err := authenticateEnroll(ctx, c.keyset, r)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("auth")
		marshalError(w, r, err)
		return
	}

	// parse enrollment structs
	enroll, err := parseEnrollmentBody(ctx, r.Body)
	if err != nil {
		marshalError(w, r, newPayloadErr("enrollment", err))
		return
	}

	// enroll device
	_, creds, err := workflow.Enroll(ctx, c.database, &enroll, c.keyCertificateAuth, c.keyCertificateKid, c.serviceName, organizationExternal, now)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("enroll")
		marshalError(w, r, err)
		return
	}
	// write out JSON
	ctx, sendSpan := tel.Start(r.Context(), "Send")
	defer sendSpan.End()
	err = jsonapi.MarshalPayload(w, creds)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("serialize json")
		marshalError(w, r, err)
	}
}

func (c *Router) doAttest(w http.ResponseWriter, r *http.Request) {
	ctx, span := tel.Start(r.Context(), "Handler")
	defer span.End()
	now := time.Now().UTC()

	w.Header().Set("Content-Type", jsonapi.MediaType)

	// Bearer token
	aikName, _, err := authenticateAttest(ctx, c.keyset, r)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("auth")
		marshalError(w, r, err)
		return
	}
	span.AddEvent("authenticated", trace.WithAttributes(attribute.String("guard.device.aikName", aikName.String())))

	// check if this is a legcy style json body or if we have fancy new multipart form and parse the evidence body accordingly
	var evidence *api.Evidence
	var blobs map[string]multipart.File
	ct := r.Header.Get("Content-Type")
	if ct == "application/json" {
		evidence, err = parseEvidenceBody(ctx, r.Body)
		if err != nil {
			marshalError(w, r, newPayloadErr("evidence", err))
			return
		}
	} else {
		tel.Log(ctx).Info("multipart json")
		// read multi-part request body now
		// this is potentially large and could involve tmp files on disk, so we do it after we know the smaller json parts could be parsed
		// if the content type is not multipart then this will throw an error and if there is any error we flat out reject as if the evidence was invalid
		err = r.ParseMultipartForm(maxMultiPartFormMem)
		if err != nil {
			tel.Log(ctx).WithError(err).Error("multipart")
			marshalError(w, r, errEvidence)
			return
		}
		defer r.MultipartForm.RemoveAll()

		// parse multipart evidence body
		evidence, blobs, err = parseEvidenceMultipart(ctx, r)
		if err != nil {
			marshalError(w, r, newPayloadErr("evidence", err))
			return
		}
	}

	// timeout for database operation
	ctx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	// start attestation process
	appr, err := workflow.Attest(ctx, c.database, c.store, evidence, blobs, aikName, c.serviceName, now, c.jobTimeout)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("start")
		marshalError(w, r, newPayloadErr("evidence", err))
		return
	}

	ctx, spanBuildReply := tel.Start(ctx, "build reply")
	defer spanBuildReply.End()

	// fetch device for key (required for device page link)
	devRow, err := device.GetByFingerprint(ctx, c.database, aikName)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("fetch device by fpr")
		marshalError(w, r, err)
		return
	}
	// org external also required
	organizationExternal, err := organization.GetExternalById(ctx, c.database, devRow.OrganizationId)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("get org external id")
		marshalError(w, r, err)
		return
	}

	// write out JSON
	if appr != nil {
		// set link to device page
		appr.SetLinkSelfWeb(device.FormatDevLinkSelfWeb(devRow.Id, *organizationExternal, c.webAppBaseURL))

		err = jsonapi.MarshalPayload(w, appr)
		if err != nil {
			tel.Log(ctx).WithError(err).Error("serialize0")
			marshalError(w, r, err)
		}
	} else {
		// write out minimal API device if we have no appraisal so client can access at least some information
		// currently the client will only access the link so it is sufficient if we pass that
		// note: before adding any extra stuff: we need a different type to return to the agent; the appraisal and device types used on the web don't cut it
		apiDev := api.Device{Id: fmt.Sprint(devRow.Id)}

		// set link to device page
		apiDev.SetLinkSelfWeb(device.FormatDevLinkSelfWeb(devRow.Id, *organizationExternal, c.webAppBaseURL))

		w.WriteHeader(http.StatusAccepted)
		err = jsonapi.MarshalPayload(w, &apiDev)
		if err != nil {
			tel.Log(ctx).WithError(err).Error("serialize1")
			marshalError(w, r, err)
		}
	}
}

func (c *Router) doEvents(w http.ResponseWriter, r *http.Request) {
	ctx, span := tel.Start(r.Context(), "Handler")
	defer span.End()

	// Bearer token
	svc, err := authenticateEvent(ctx, c.keyset, r)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("auth")
		marshalError(w, r, err)
		return
	}

	// parse CloudEvent
	msg := cehttp.NewMessageFromHttpRequest(r)
	ev, err := cebind.ToEvent(ctx, msg)
	if err != nil {
		tel.Log(ctx).WithError(err).WithField("request", *r).Error("parse")
		marshalError(w, r, err)
		return
	}

	// make sure token matches event source
	if ev.Source() != svc {
		tel.Log(ctx).WithField("event", ev).WithField("token", svc).Error("event source")
		marshalError(w, r, ErrAuthentication)
		return
	}

	// process event
	err = event.Receive(ctx, c.database, c.serviceName, ev, time.Now())
	if err != nil {
		tel.Log(ctx).WithError(err).WithField("event", ev).Error("event processing")
		marshalError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (c *Router) doListDevices(w http.ResponseWriter, r *http.Request) {
	ctx, span := tel.Start(r.Context(), "Handler")
	defer span.End()
	vars := mux.Vars(r)
	now := time.Now().UTC()
	fieldsDevices := r.URL.Query().Get("fields[devices]")
	filterTags := r.URL.Query().Get("filter[tags]")
	filterIssue := r.URL.Query().Get("filter[issue]")
	filterState := r.URL.Query().Get("filter[state]")
	iterStr := r.URL.Query().Get("i")

	w.Header().Set("Content-Type", jsonapi.MediaType)

	// convert org id to int
	orgId, err := strconv.ParseInt(vars["org-id"], 10, 64)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("error converting org id")
		marshalError(w, r, err)
		return
	}

	// start database transaction
	tx, err := c.database.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
	if err != nil {
		tel.Log(ctx).WithError(err).Error("start db tx")
		marshalError(w, r, err)
		return
	}
	defer tx.Rollback(ctx)

	// fetch devices
	state, doc, errobjs := doListDeviceImpl(ctx, tx, orgId, int64(numDevices), iterStr, fieldsDevices, c.baseURL, now, filterTags, filterIssue, filterState)
	if state != http.StatusOK {
		w.WriteHeader(state)
		jsonapi.MarshalErrors(w, errobjs)
		return
	}

	// end database transaction
	err = tx.Commit(ctx)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("commit tx")
		marshalError(w, r, err)
		return
	}

	// write out JSON
	marshalResult(ctx, doc, w)
}

func doListDeviceImpl(ctx context.Context, tx pgx.Tx, orgId, limit int64, start, devicesFieldsStr, baseURL string, now time.Time, filterTags, filterIssue, filterState string) (int, jsonapi.Payloader, []*jsonapi.ErrorObject) {
	span := trace.SpanFromContext(ctx)

	// parse fieldsets
	devicesFields := parseFieldset(devicesFieldsStr)

	// fetch device attributes. fills only indices of relationships
	var devrows []device.Row
	var next *string
	var err error

	// filtered device list with interator
	var firstDevice *int64
	if start != "" {
		if i, err := strconv.ParseInt(start, 10, 32); err == nil {
			firstDevice = &i
		}
	}
	var tagsArray []string
	if len(filterTags) > 0 {
		tagsArray = strings.Split(filterTags, ",")
	}
	devrows, err = filter.ListActiveDevicesFiltered(ctx, tx, orgId, limit, firstDevice, now, tagsArray, filterIssue, filterState)

	// error handling
	if err != nil {
		tel.Log(ctx).WithError(err).Error("fetch devices")
		status, errobjs := buildError(ctx, err)
		return status, nil, errobjs
	}

	// build next str
	if len(devrows) > 0 {
		str := strconv.FormatInt(devrows[len(devrows)-1].Id-1, 10)
		next = &str
	}

	// convert device rows to API structures w/o attached resources
	devs := make([]*api.Device, len(devrows))
	for i, row := range devrows {
		dev, err := row.ToApiStruct()
		if err != nil {
			tel.Log(ctx).WithError(err).Error("convert device row")
			status, errobjs := buildError(ctx, err)
			return status, nil, errobjs
		}
		devs[i] = dev
	}

	// fetch tags to include
	tagIds := make(map[string]*api.Tag)
	for i, row := range devrows {
		tags, err := tag.GetTagsByDeviceId(ctx, tx, row.Id)
		if err != nil {
			tel.Log(ctx).WithError(err).Error("fetch tags")
			status, errobjs := buildError(ctx, err)
			return status, nil, errobjs
		}
		if len(tags) > 0 {
			devs[i].Tags = make([]*api.Tag, len(tags))
			for j, t := range tags {
				tmp, err := tag.FromRow(&t)
				if err != nil {
					tel.Log(ctx).WithError(err).Error("conv tag")
					status, errobjs := buildError(ctx, err)
					return status, nil, errobjs
				}
				tagIds[t.Id] = tmp
				devs[i].Tags[j] = tmp
			}
		}
	}

	var included []*jsonapi.Node
	var tags []*api.Tag
	if len(tagIds) > 0 {
		for _, v := range tagIds {
			tags = append(tags, v)
		}

		status, nodes, errobjs := marshalWithoutIncluded(ctx, tags)
		if status != http.StatusOK {
			return status, nil, errobjs
		}

		included = append(included, nodes...)
	}

	// convert to JSON:API document
	var pdoc jsonapi.Payloader
	pdoc, err = jsonapi.Marshal(devs)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("serialize")
		status, errobjs := buildError(ctx, err)
		return status, nil, errobjs
	}

	// insert included, filter device fields and enforce many payload
	if doc, ok := pdoc.(*jsonapi.ManyPayload); ok {
		for i := range doc.Data {
			filterFields(doc.Data[i], devicesFields)
		}
		if next != nil {
			if doc.Links == nil {
				doc.Links = &jsonapi.Links{}
			}
			s := fmt.Sprintf(path.Join(baseURL, "devices?i=%s"), *next)
			if devicesFieldsStr != "" {
				s = fmt.Sprint(s, "&fields[devices]=", devicesFieldsStr)
			}
			if filterTags != "" {
				s = fmt.Sprint(s, "&filter[tags]=", filterTags)
			}
			if filterIssue != "" {
				s = fmt.Sprint(s, "&filter[issue]=", filterIssue)
			}
			if filterState != "" {
				s = fmt.Sprint(s, "&filter[state]=", filterState)
			}

			(*doc.Links)["next"] = s
		}
		doc.Included = included
		return http.StatusOK, doc, nil
	}

	span.SetStatus(codes.Error, "Failed to cast jsonapi.Payloader to many payload doc")
	status, errobjs := buildError(ctx, errSerialize)
	return status, nil, errobjs
}

func (c *Router) doGetDevice(w http.ResponseWriter, r *http.Request) {
	ctx, span := tel.Start(r.Context(), "Handler")
	defer span.End()
	vars := mux.Vars(r)
	now := time.Now().UTC()
	includeStr := r.URL.Query().Get("include")
	devicesFieldsStr := r.URL.Query().Get("fields[devices]")
	changesFieldsStr := r.URL.Query().Get("fields[changes]")
	appraisalsFieldsStr := r.URL.Query().Get("fields[appraisals]")

	w.Header().Set("Content-Type", jsonapi.MediaType)

	// convert org id to int
	orgId, err := strconv.ParseInt(vars["org-id"], 10, 64)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("error converting org id")
		marshalError(w, r, err)
		return
	}

	// convert dev id to int
	devId, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("error converting org id")
		marshalError(w, r, err)
		return
	}

	// start database transaction
	tx, err := c.database.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
	if err != nil {
		tel.Log(ctx).WithError(err).Error("start db tx")
		marshalError(w, r, err)
		return
	}
	defer tx.Rollback(ctx)

	// fetch device
	state, doc, errobjs := doGetDeviceImpl(ctx, tx, pointQuery{Id: devId}, orgId, includeStr, devicesFieldsStr, changesFieldsStr, appraisalsFieldsStr, c.baseURL, now)
	if state != http.StatusOK {
		w.WriteHeader(state)
		jsonapi.MarshalErrors(w, errobjs)
		return
	}

	// end database transaction
	err = tx.Commit(ctx)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("commit tx")
		marshalError(w, r, err)
		return
	}

	// write out JSON
	marshalResult(ctx, doc, w)
}

func doGetDeviceImpl(ctx context.Context, tx pgx.Tx, query interface{}, orgId int64, includeStr string, devicesFieldsStr string, changesFieldsStr string, appraisalsFieldsStr string, baseURL string, now time.Time) (int, jsonapi.Payloader, []*jsonapi.ErrorObject) {
	span := trace.SpanFromContext(ctx)

	// defaults, changed by include=* query string parameter
	includeTags := true
	includeAppraisals := true

	// parse include set
	if includeStr != "" {
		if includeSet, ok := parseInclude(includeStr, deviceIncludeWhitelist); !ok {
			status, errobjs := buildError(ctx, errQueryStringWrong)
			return status, nil, errobjs
		} else {
			_, includeTags = includeSet["tags"]
			_, includeAppraisals = includeSet["appraisals"]
		}
	}

	// parse fieldsets
	devicesFields := parseFieldset(devicesFieldsStr)
	appraisalsFields := parseFieldset(appraisalsFieldsStr)

	// fetch device attributes. fills only indices of relationships
	var devrows []device.Row
	var next *string
	var err error
	var singleton bool
	var minResultSize int

	switch query := query.(type) {
	case pointQuery:
		var devrow *device.Row
		singleton = true
		minResultSize = 1
		devrow, err = device.Get(ctx, tx, query.Id, orgId, now)
		if err == nil {
			devrows = []device.Row{*devrow}
		}

	case setQuery:
		// device set
		singleton = false
		minResultSize = len(query.Set)
		devrows, err = device.SetRow(ctx, tx, query.Set, orgId, now)

	case rangeQuery:
		// device list
		var i *string
		start := query.Start
		if start != "" {
			i = &start
		}
		singleton = false
		devrows, next, err = device.ListRow(ctx, tx, i, query.Length, orgId, now)

	default:
		tel.Log(ctx).WithField("query", fmt.Sprintf("%#v", query)).Error("fetch devices")
		err = errInternalServerError
	}

	// error handling
	if err != nil {
		tel.Log(ctx).WithError(err).Error("fetch devices")
		status, errobjs := buildError(ctx, err)
		return status, nil, errobjs
	}
	if len(devrows) < minResultSize {
		status, errobjs := buildError(ctx, database.ErrNotFound)
		return status, nil, errobjs
	}

	var included []*jsonapi.Node

	// convert device rows to API structures w/o attached resources
	devs := make([]*api.Device, len(devrows))
	for i, row := range devrows {
		dev, err := row.ToApiStruct()
		if err != nil {
			tel.Log(ctx).WithError(err).Error("convert device row")
			status, errobjs := buildError(ctx, err)
			return status, nil, errobjs
		}
		devs[i] = dev
	}

	// fetch only latest appraisal for each device
	if includeAppraisals {
		var appraisals []*api.Appraisal
		for _, dev := range devs {
			// we can safely ignore the IDs in dev.Appraisals b/c we are in a repeatable read isolated
			// transaction and it is thus safe to query the latest appraisal by device ID
			devId, err := strconv.ParseInt(dev.Id, 10, 64)
			if err != nil {
				// just panic here, enough is enough, if this is no int something is seriously broken
				// and we don't need to marshal any string conversion errors
				panic(err)
			}

			// include issues in query if requestes; this is an expensive operation,
			// so we don't do it by default
			includeIssues := false
			if appraisalsFields != nil {
				_, includeIssues = (*appraisalsFields)["issues"]
			}

			tmp, err := appraisal.GetLatestAppraisalsByDeviceId(ctx, tx, devId, includeIssues, 1)
			if err != nil {
				tel.Log(ctx).WithError(err).Error("fetch appraisals")
				status, errobjs := buildError(ctx, err)
				return status, nil, errobjs
			}
			dev.Appraisals = tmp
			appraisals = append(appraisals, tmp...)
		}

		status, nodes, errobjs := marshalWithoutIncluded(ctx, appraisals)
		if status != http.StatusOK {
			return status, nil, errobjs
		}
		for i := range nodes {
			filterFields(nodes[i], appraisalsFields)
		}

		included = append(included, nodes...)
	}

	// fetch tags to include
	tagIds := make(map[string]*api.Tag)
	var tags []*api.Tag
	if includeTags {
		for i, row := range devrows {
			tags, err := tag.GetTagsByDeviceId(ctx, tx, row.Id)
			if err != nil {
				tel.Log(ctx).WithError(err).Error("fetch tags")
				status, errobjs := buildError(ctx, err)
				return status, nil, errobjs
			}
			if len(tags) > 0 {
				devs[i].Tags = make([]*api.Tag, len(tags))
				for j, t := range tags {
					tmp, err := tag.FromRow(&t)
					if err != nil {
						tel.Log(ctx).WithError(err).Error("conv tag")
						status, errobjs := buildError(ctx, err)
						return status, nil, errobjs
					}
					tagIds[t.Id] = tmp
					devs[i].Tags[j] = tmp
				}
			}
		}

		if len(tagIds) > 0 {
			for _, v := range tagIds {
				tags = append(tags, v)
			}

			status, nodes, errobjs := marshalWithoutIncluded(ctx, tags)
			if status != http.StatusOK {
				return status, nil, errobjs
			}

			included = append(included, nodes...)
		}
	}

	// convert to JSON:API document
	var pdoc jsonapi.Payloader
	if singleton {
		pdoc, err = jsonapi.Marshal(devs[0])
	} else {
		pdoc, err = jsonapi.Marshal(devs)
	}
	if err != nil {
		tel.Log(ctx).WithError(err).Error("serialize")
		status, errobjs := buildError(ctx, err)
		return status, nil, errobjs
	}

	// insert included, filter device fields
	if doc, ok := pdoc.(*jsonapi.OnePayload); ok {
		filterFields(doc.Data, devicesFields)
		doc.Included = included
		return http.StatusOK, doc, nil
	} else if doc, ok := pdoc.(*jsonapi.ManyPayload); ok {
		for i := range doc.Data {
			filterFields(doc.Data[i], devicesFields)
		}
		if next != nil {
			if doc.Links == nil {
				doc.Links = &jsonapi.Links{}
			}
			(*doc.Links)["next"] = fmt.Sprintf(path.Join(baseURL, "devices?i=%s"), *next)
		}
		doc.Included = included
		return http.StatusOK, doc, nil
	} else {
		span.SetStatus(codes.Error, "Failed to cast jsonapi.Payloader to single payload doc")
		status, errobjs := buildError(ctx, errSerialize)
		return status, nil, errobjs
	}
}

func (c *Router) doListChangesOfDevice(w http.ResponseWriter, r *http.Request) {
	ctx, span := tel.Start(r.Context(), "Handler")
	defer span.End()
	vars := mux.Vars(r)
	dev := vars["id"]
	now := time.Now().UTC()
	iterStr := r.URL.Query().Get("i")
	changesFieldsStr := r.URL.Query().Get("fields[changes]")

	w.Header().Set("Content-Type", jsonapi.MediaType)

	// convert org id to int
	orgId, err := strconv.ParseInt(vars["org-id"], 10, 64)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("error converting org id")
		marshalError(w, r, err)
		return
	}

	// start database transaction
	tx, err := c.database.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
	if err != nil {
		tel.Log(ctx).WithError(err).Error("start db tx")
		marshalError(w, r, err)
		return
	}
	defer tx.Rollback(ctx)

	// fetch changes
	state, doc, errobjs := doGetChangeImpl(ctx, tx, rangeQuery{Start: iterStr, Length: numChanges}, orgId, &dev, changesFieldsStr, c.baseURL, now)
	if state != http.StatusOK {
		w.WriteHeader(state)
		jsonapi.MarshalErrors(w, errobjs)
		return
	}

	// end database transaction
	err = tx.Commit(ctx)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("commit tx")
		marshalError(w, r, err)
		return
	}

	// write out JSON
	marshalResult(ctx, doc, w)
}

func doGetLatestAppraisalFull(ctx context.Context, tx pgx.Tx, dev string) (int, jsonapi.Payloader, []*jsonapi.ErrorObject) {
	span := trace.SpanFromContext(ctx)

	devId, err := strconv.ParseInt(dev, 10, 64)
	if err != nil {
		// srsly why do we need string IDs, all the conversion could be done just when the requets handlers are first entered and inside the apisrv we just use ints
		// and never have to worry about error branches from boilerplate ID conversions
		tel.Log(ctx).WithError(err).Error("fetch appraisals")
		status, errobjs := buildError(ctx, err)
		return status, nil, errobjs
	}

	// fetch latest appraisal with issues (we always want to show issues on the web)
	apprs, err := appraisal.GetLatestAppraisalsByDeviceId(ctx, tx, devId, true, 1)

	// error handling
	if err != nil {
		tel.Log(ctx).WithError(err).Error("fetch appraisals")
		status, errobjs := buildError(ctx, err)
		return status, nil, errobjs
	}

	// convert single appraisal to JSON:API document
	// the client always awaits an array so we enforce a many payload here
	// an empty appraisal list is not a 404, this resource exists, it is just an empty list
	pdoc, err := jsonapi.Marshal(apprs)

	if err != nil {
		tel.Log(ctx).WithError(err).Error("serialize")
		status, errobjs := buildError(ctx, err)
		return status, nil, errobjs
	}

	// enforce manypayload
	if doc, ok := pdoc.(*jsonapi.ManyPayload); ok {
		return http.StatusOK, doc, nil
	}

	span.SetStatus(codes.Error, "Failed to cast jsonapi.Payloader to single payload doc")
	status, errobjs := buildError(ctx, errSerialize)
	return status, nil, errobjs
}

func (c *Router) doListAppraisalsOfDevice(w http.ResponseWriter, r *http.Request) {
	ctx, span := tel.Start(r.Context(), "Handler")
	defer span.End()
	vars := mux.Vars(r)
	dev := vars["id"]

	w.Header().Set("Content-Type", jsonapi.MediaType)

	// start database transaction
	tx, err := c.database.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
	if err != nil {
		tel.Log(ctx).WithError(err).Error("start db tx")
		marshalError(w, r, err)
		return
	}
	defer tx.Rollback(ctx)

	// fetch appraisals
	state, doc, errobjs := doGetLatestAppraisalFull(ctx, tx, dev)
	if state != http.StatusOK {
		w.WriteHeader(state)
		jsonapi.MarshalErrors(w, errobjs)
		return
	}

	// end database transaction
	err = tx.Commit(ctx)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("commit tx")
		marshalError(w, r, err)
		return
	}

	// write out JSON
	marshalResult(ctx, doc, w)
}

func (c *Router) doPatchDevice(w http.ResponseWriter, r *http.Request) {
	ctx, span := tel.Start(r.Context(), "Handler")
	defer span.End()
	vars := mux.Vars(r)
	now := time.Now().UTC()
	actor := vars["actor"]
	orgExt := vars["org-ext"]
	includeStr := r.URL.Query().Get("include")
	devicesFieldsStr := r.URL.Query().Get("fields[devices]")
	changesFieldsStr := r.URL.Query().Get("fields[changes]")
	appraisalsFieldsStr := r.URL.Query().Get("fields[appraisals]")

	w.Header().Set("Content-Type", jsonapi.MediaType)

	// convert org id to int
	orgId, err := strconv.ParseInt(vars["org-id"], 10, 64)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("error converting org id")
		marshalError(w, r, err)
		return
	}

	// convert dev id to int
	devId, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("error converting org id")
		marshalError(w, r, err)
		return
	}

	// parse patch
	var many jsonapi.ManyPayload
	err = json.NewDecoder(r.Body).Decode(&many)
	if err != nil {
		err = mapJsonClientRequestErrors(err)
		tel.Log(ctx).WithError(err).Error("parse")
		marshalError(w, r, newPayloadErr("device patch array", err))
		return
	}

	var attribs = make([]map[string]interface{}, len(many.Data))
	for i, d := range many.Data {
		attribs[i] = d.Attributes
	}
	buf, err := json.Marshal(attribs)
	if err != nil {
		err = mapJsonClientRequestErrors(err)
		tel.Log(ctx).WithError(err).Error("remarshal")
		marshalError(w, r, newPayloadErr("device patch array", err))
		return
	}
	var patches []*api.DevicePatch
	err = json.Unmarshal(buf, &patches)
	if err != nil {
		err = mapJsonClientRequestErrors(err)
		tel.Log(ctx).WithError(err).Error("reunmarshal")
		marshalError(w, r, newPayloadErr("device patch array", err))
		return
	}

	// start database transaction
	var doc jsonapi.Payloader
	var errobjs []*jsonapi.ErrorObject
	var state int
	err = database.QueryIsolatedRetry(ctx, c.database, func(ctx context.Context, tx pgx.Tx) error {
		checkQuota := false
		for idx, patch := range patches {
			if patch.State == nil {
				// Change tags
				if patch.Tags != nil {
					strTags := make([]string, len(patch.Tags))
					for i := range patch.Tags {
						strTags[i] = patch.Tags[i].Key
					}
					_, err := tag.Device(ctx, tx, devId, orgId, strTags)
					if err != nil {
						tel.Log(ctx).WithField("patch", idx).WithError(err).Error("patch tags")
						return err
					}
				}

				// Change policy
				if patch.Policy != nil {
					validPolicy := policy2.IsTrinary(patch.Policy.EndpointProtection) &&
						policy2.IsTrinary(patch.Policy.IntelTSC)
					if !validPolicy {
						err = errDevicePolicy
						tel.Log(ctx).WithField("policy", fmt.Sprintf("%#v", patch.Policy)).WithError(err).Error("check policy")
						return err
					}
					pol := policy2.New()
					pol.EndpointProtection = policy2.Trinary(patch.Policy.EndpointProtection)
					pol.IntelTSC = policy2.Trinary(patch.Policy.IntelTSC)
					err := device.Patch(ctx, tx, devId, orgId, nil, nil, pol, now, actor)
					if err != nil {
						tel.Log(ctx).WithField("patch", idx).WithError(err).Error("update policy")
						return err
					}
				}

				// Rename
				if patch.Name != nil && *patch.Name != "" {
					err := device.Patch(ctx, tx, devId, orgId, patch.Name, nil, nil, now, actor)
					if err != nil {
						tel.Log(ctx).WithField("patch", idx).WithError(err).Error("rename device")
						return err
					}
				}
			} else if *patch.State == api.StateRetired {
				// Retire device
				err = device.Retire(ctx, tx, devId, orgId, true, nil, now, actor)
				checkQuota = true
			} else {
				err = errors.New("invalid state value")
				tel.Log(ctx).WithField("patch", idx).WithError(err).Error("patch device")
				return err
			}

			if err != nil {
				tel.Log(ctx).WithError(err).Error("patch device")
				return err
			}
		}

		if checkQuota {
			// quota
			current, allowed, err := organization.Quota(ctx, tx, orgId)
			if err != nil {
				tel.Log(ctx).WithError(err).Error("fetch quota")
				return err
			}
			if current > allowed {
				tel.Log(ctx).
					WithFields(log.Fields{"current": current, "allowed": allowed}).
					Error("over quota")
				return err
			}

			// queue event
			_, err = event.BillingUpdate(ctx, tx, c.serviceName, map[string]int{orgExt: current}, now, now)
			if err != nil {
				tel.Log(ctx).WithError(err).Error("queue event")
				return err
			}
		}

		state, doc, errobjs = doGetDeviceImpl(ctx, tx, pointQuery{Id: devId}, orgId, includeStr, devicesFieldsStr, changesFieldsStr, appraisalsFieldsStr, c.baseURL, now)
		return nil
	})

	// write out JSON
	if err != nil {
		tel.Log(ctx).WithError(err).Error("read repeatable query")
		marshalError(w, r, err)
		return
	}
	w.WriteHeader(state)
	if state != http.StatusOK {
		jsonapi.MarshalErrors(w, errobjs)
	} else {
		marshalResult(ctx, doc, w)
	}
}

func (c *Router) doResurect(w http.ResponseWriter, r *http.Request) {
	ctx, span := tel.Start(r.Context(), "Handler")
	defer span.End()
	vars := mux.Vars(r)
	now := time.Now().UTC()
	actor := vars["actor"]
	orgExt := vars["org-ext"]
	includeStr := r.URL.Query().Get("include")
	devicesFieldsStr := r.URL.Query().Get("fields[devices]")
	changesFieldsStr := r.URL.Query().Get("fields[changes]")
	appraisalsFieldsStr := r.URL.Query().Get("fields[appraisals]")

	w.Header().Set("Content-Type", jsonapi.MediaType)

	// convert org id to int
	orgId, err := strconv.ParseInt(vars["org-id"], 10, 64)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("error converting org id")
		marshalError(w, r, err)
		return
	}

	// convert dev id to int
	devId, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("error converting org id")
		marshalError(w, r, err)
		return
	}

	var doc jsonapi.Payloader
	var errobjs []*jsonapi.ErrorObject
	var state int
	err = database.QueryIsolatedRetry(ctx, c.database, func(ctx context.Context, tx pgx.Tx) error {
		// create a new device
		err := device.Retire(ctx, tx, devId, orgId, false, nil, now, actor)
		if err != nil {
			tel.Log(ctx).WithError(err).Error("resurrect")
			return err
		}

		// quota
		current, allowed, err := organization.Quota(ctx, tx, orgId)
		if err != nil {
			tel.Log(ctx).WithError(err).Error("fetch quota")
			return err
		}
		if current > allowed {
			tel.Log(ctx).
				WithFields(log.Fields{"current": current, "allowed": allowed}).
				Error("over quota")
			return workflow.ErrQuotaExceeded
		}

		// queue event
		_, err = event.BillingUpdate(ctx, tx, c.serviceName, map[string]int{orgExt: current}, now, now)
		if err != nil {
			tel.Log(ctx).WithError(err).Error("queue event")
			return err
		}

		state, doc, errobjs = doGetDeviceImpl(ctx, tx, pointQuery{Id: devId}, orgId, includeStr, devicesFieldsStr, changesFieldsStr, appraisalsFieldsStr, c.baseURL, now)
		return nil
	})

	// write out JSON
	if err != nil {
		tel.Log(ctx).WithError(err).Error("read repeatable query")
		marshalError(w, r, err)
		return
	}
	w.WriteHeader(state)
	if state != http.StatusOK {
		jsonapi.MarshalErrors(w, errobjs)
	} else {
		marshalResult(ctx, doc, w)
	}
}

func (c *Router) doOverride(w http.ResponseWriter, r *http.Request) {
	ctx, span := tel.Start(r.Context(), "Handler")
	defer span.End()
	vars := mux.Vars(r)
	now := time.Now().UTC()
	actor := vars["actor"]
	includeStr := r.URL.Query().Get("include")
	devicesFieldsStr := r.URL.Query().Get("fields[devices]")
	changesFieldsStr := r.URL.Query().Get("fields[changes]")
	appraisalsFieldsStr := r.URL.Query().Get("fields[appraisals]")

	w.Header().Set("Content-Type", jsonapi.MediaType)

	// convert org id to int
	orgId, err := strconv.ParseInt(vars["org-id"], 10, 64)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("error converting org id")
		marshalError(w, r, err)
		return
	}

	// convert dev id to int
	devId, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("error converting org id")
		marshalError(w, r, err)
		return
	}

	// parse array
	var overrides []string
	err = json.NewDecoder(r.Body).Decode(&overrides)
	if err != nil {
		err = mapJsonClientRequestErrors(err)
		tel.Log(ctx).WithError(err).Error("parse")
		marshalError(w, r, newPayloadErr("override array", err))
		return
	}

	// update device override list and reattest against new overrides
	var doc jsonapi.Payloader
	var errobjs []*jsonapi.ErrorObject
	var state int
	err = database.QueryIsolatedRetry(ctx, c.database, func(ctx context.Context, tx pgx.Tx) error {
		_, err = workflow.Override(ctx, tx, c.store, devId, orgId, overrides, c.serviceName, now, actor)
		if err != nil {
			tel.Log(ctx).WithError(err).Error("override")
			return err
		}

		state, doc, errobjs = doGetDeviceImpl(ctx, tx, pointQuery{Id: devId}, orgId, includeStr, devicesFieldsStr, changesFieldsStr, appraisalsFieldsStr, c.baseURL, now)
		return nil
	})

	// write out JSON
	if err != nil {
		tel.Log(ctx).WithError(err).Error("read repeatable query")
		marshalError(w, r, err)
		return
	}
	w.WriteHeader(state)
	if state != http.StatusOK {
		jsonapi.MarshalErrors(w, errobjs)
	} else {
		marshalResult(ctx, doc, w)
	}
}

func parseEvidenceBody(ctx context.Context, body io.ReadCloser) (evidence *api.Evidence, err error) {
	ctx, span := tel.Start(ctx, "Parse evidence")
	defer span.End()

	// map errors on all return paths
	defer func() {
		err = mapJsonClientRequestErrors(err)
	}()

	var one jsonapi.OnePayload
	err = json.NewDecoder(body).Decode(&one)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("decode jsonapi")
		return
	}
	if one.Data == nil {
		err = errEvidence
		tel.Log(ctx).WithError(err).Error("decode jsonapi")
		return
	}
	buf, err := json.Marshal(one.Data.Attributes)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("reencode json")
		return
	}
	evidence = &api.Evidence{}
	err = json.Unmarshal(buf, evidence)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("redecode json")
		return
	}
	return
}

func parseEvidenceMultipart(ctx context.Context, r *http.Request) (*api.Evidence, map[string]multipart.File, error) {
	ctx, span := tel.Start(ctx, "Parse multipart evidence")
	defer span.End()

	// re-map multipart as hash map
	var evidence *api.Evidence
	hashBlobs := make(map[string]multipart.File)
	for _, fh := range r.MultipartForm.File {
		// reject multiple values for the same thing
		if len(fh) != 1 {
			return nil, nil, errEvidence
		}

		// find evidence json
		if fh[0].Filename == "evidencebody" && fh[0].Header.Get("Content-Type") == "application/json" {
			evBody, err := fh[0].Open()
			if err != nil {
				return nil, nil, err
			}

			evidence, err = parseEvidenceBody(ctx, evBody)
			if err != nil {
				return nil, nil, err
			}
		} else if len(fh[0].Filename) == 64 && fh[0].Header.Get("Content-Type") == "application/octet-stream" {
			blob, err := fh[0].Open()
			if err != nil {
				return nil, nil, err
			}
			hashBlobs[fh[0].Filename] = blob
		}
	}

	if evidence == nil {
		return nil, nil, errEvidence
	}

	return evidence, hashBlobs, nil
}

func parseEnrollmentBody(ctx context.Context, body io.ReadCloser) (enroll api.Enrollment, err error) {
	ctx, span := tel.Start(ctx, "Parse enrollment")
	defer span.End()

	// map errors on all return paths
	defer func() {
		err = mapJsonClientRequestErrors(err)
	}()

	var one jsonapi.OnePayload
	err = json.NewDecoder(body).Decode(&one)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("parse jsonapi")
		return
	}
	if one.Data == nil {
		err = errEnrollment
		tel.Log(ctx).WithError(err).Error("parse jsonapi")
		return
	}
	buf, err := json.Marshal(one.Data.Attributes)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("reencode json")
		return
	}
	err = json.Unmarshal(buf, &enroll)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("redecode json")
		return
	}
	return
}

func doGetChangeImpl(ctx context.Context, tx pgx.Tx, query interface{}, orgId int64, dev *string, changesFieldsStr string, baseURL string, now time.Time) (int, jsonapi.Payloader, []*jsonapi.ErrorObject) {
	span := trace.SpanFromContext(ctx)

	// parse fieldsets
	changesFields := parseFieldset(changesFieldsStr)

	// fetch change attributes. fills only indices of relationships
	var chs []*api.Change
	var next *string
	var err error
	var singleton bool
	var minResultSize int

	switch query := query.(type) {
	case pointQuery:
		singleton = true
		minResultSize = 1
		ch, err := change.Get(ctx, tx, query.Id, orgId)
		if err == nil {
			chs = []*api.Change{ch}
		}

	case setQuery:
		// change set
		singleton = false
		minResultSize = len(query.Set)
		chs, err = change.Set(ctx, tx, query.Set, orgId)

	case rangeQuery:
		// change list
		var i *string
		start := query.Start
		if start != "" {
			i = &start
		}
		singleton = false
		chs, next, err = change.List(ctx, tx, i, orgId, dev, query.Length)

	default:
		tel.Log(ctx).WithField("query", fmt.Sprintf("%#v", query)).Error("fetch changes")
		err = errInternalServerError
	}

	// error handling
	if err != nil {
		tel.Log(ctx).WithError(err).Error("fetch changes")
		status, errobjs := buildError(ctx, err)
		return status, nil, errobjs
	}
	if len(chs) < minResultSize {
		status, errobjs := buildError(ctx, database.ErrNotFound)
		return status, nil, errobjs
	}

	var included []*jsonapi.Node

	// convert to JSON:API document
	var pdoc jsonapi.Payloader
	if singleton {
		pdoc, err = jsonapi.Marshal(chs[0])
	} else {
		pdoc, err = jsonapi.Marshal(chs)
	}
	if err != nil {
		tel.Log(ctx).WithError(err).Error("serialize")
		status, errobjs := buildError(ctx, err)
		return status, nil, errobjs
	}

	// insert included
	if doc, ok := pdoc.(*jsonapi.OnePayload); ok && doc.Data != nil {
		filterFields(doc.Data, changesFields)
		doc.Included = included
		return http.StatusOK, doc, nil
	} else if doc, ok := pdoc.(*jsonapi.ManyPayload); ok {
		for i := range doc.Data {
			filterFields(doc.Data[i], changesFields)
		}
		if next != nil {
			if doc.Links == nil {
				doc.Links = &jsonapi.Links{}
			}
			(*doc.Links)["next"] = fmt.Sprintf(path.Join(baseURL, "changes?i=%s"), *next)
		}
		doc.Included = included
		return http.StatusOK, doc, nil
	} else {
		span.SetStatus(codes.Error, "Failed to cast jsonapi.Payloader to single payload doc")
		status, errobjs := buildError(ctx, errSerialize)
		return status, nil, errobjs
	}
}

func (c *Router) doListChanges(w http.ResponseWriter, r *http.Request) {
	ctx, span := tel.Start(r.Context(), "Handler")
	defer span.End()
	vars := mux.Vars(r)
	now := time.Now().UTC()
	iterStr := r.URL.Query().Get("i")
	changesFieldsStr := r.URL.Query().Get("fields[changes]")

	w.Header().Set("Content-Type", jsonapi.MediaType)

	// convert org id to int
	orgId, err := strconv.ParseInt(vars["org-id"], 10, 64)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("error converting org id")
		marshalError(w, r, err)
		return
	}

	// start database transaction
	tx, err := c.database.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
	if err != nil {
		tel.Log(ctx).WithError(err).Error("start db tx")
		marshalError(w, r, err)
		return
	}
	defer tx.Rollback(ctx)

	// fetch changes
	state, doc, errobjs := doGetChangeImpl(ctx, tx, rangeQuery{Start: iterStr, Length: numChanges}, orgId, nil, changesFieldsStr, c.baseURL, now)
	if state != http.StatusOK {
		w.WriteHeader(state)
		jsonapi.MarshalErrors(w, errobjs)
		return
	}

	// end database transaction
	err = tx.Commit(ctx)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("commit tx")
		marshalError(w, r, err)
		return
	}

	// write out JSON
	marshalResult(ctx, doc, w)
}

func (c *Router) doGetChange(w http.ResponseWriter, r *http.Request) {
	ctx, span := tel.Start(r.Context(), "Handler")
	defer span.End()
	vars := mux.Vars(r)
	now := time.Now().UTC()
	changesFieldsStr := r.URL.Query().Get("fields[changes]")

	w.Header().Set("Content-Type", jsonapi.MediaType)

	// convert org id to int
	orgId, err := strconv.ParseInt(vars["org-id"], 10, 64)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("error converting org id")
		marshalError(w, r, err)
		return
	}

	// convert dev id to int
	changeId, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("error converting org id")
		marshalError(w, r, err)
		return
	}

	// start database transaction
	tx, err := c.database.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
	if err != nil {
		tel.Log(ctx).WithError(err).Error("start db tx")
		marshalError(w, r, err)
		return
	}
	defer tx.Rollback(ctx)

	// fetch change
	state, doc, errobjs := doGetChangeImpl(ctx, tx, pointQuery{Id: changeId}, orgId, nil, changesFieldsStr, c.baseURL, now)
	if state != http.StatusOK {
		w.WriteHeader(state)
		jsonapi.MarshalErrors(w, errobjs)
		return
	}

	// end database transaction
	err = tx.Commit(ctx)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("commit tx")
		marshalError(w, r, err)
		return
	}

	// write out JSON
	marshalResult(ctx, doc, w)
}

func (c *Router) doGetTag(w http.ResponseWriter, r *http.Request) {
	ctx, span := tel.Start(r.Context(), "Handler")
	defer span.End()
	vars := mux.Vars(r)
	id := vars["id"]
	now := time.Now().UTC()
	tagFieldsStr := r.URL.Query().Get("fields[tags]")

	w.Header().Set("Content-Type", jsonapi.MediaType)

	// convert org id to int
	orgId, err := strconv.ParseInt(vars["org-id"], 10, 64)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("error converting org id")
		marshalError(w, r, err)
		return
	}

	// start database transaction
	tx, err := c.database.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
	if err != nil {
		tel.Log(ctx).WithError(err).Error("start db tx")
		marshalError(w, r, err)
		return
	}
	defer tx.Rollback(ctx)

	// fetch tag
	state, doc, errobjs := doGetTagImpl(ctx, tx, pointQueryStrId{Id: id}, orgId, tagFieldsStr, c.baseURL, now)
	if state != http.StatusOK {
		w.WriteHeader(state)
		jsonapi.MarshalErrors(w, errobjs)
		return
	}

	// end database transaction
	err = tx.Commit(ctx)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("commit tx")
		marshalError(w, r, err)
		return
	}

	// write out JSON
	marshalResult(ctx, doc, w)
}

func (c *Router) doListTags(w http.ResponseWriter, r *http.Request) {
	ctx, span := tel.Start(r.Context(), "Handler")
	defer span.End()
	vars := mux.Vars(r)
	now := time.Now().UTC()
	searchStr := r.URL.Query().Get("search")
	iterStr := r.URL.Query().Get("i")
	tagsFieldsStr := r.URL.Query().Get("fields[tags]")

	w.Header().Set("Content-Type", jsonapi.MediaType)

	// convert org id to int
	orgId, err := strconv.ParseInt(vars["org-id"], 10, 64)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("error converting org id")
		marshalError(w, r, err)
		return
	}

	// start database transaction
	tx, err := c.database.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
	if err != nil {
		tel.Log(ctx).WithError(err).Error("start db tx")
		marshalError(w, r, err)
		return
	}
	defer tx.Rollback(ctx)

	// fetch tags
	var q interface{}
	if searchStr != "" {
		q = textQuery{Fragment: searchStr}
	} else {
		q = rangeQuery{Start: iterStr, Length: numTags}
	}
	state, doc, errobjs := doGetTagImpl(ctx, tx, q, orgId, tagsFieldsStr, c.baseURL, now)
	if state != http.StatusOK {
		w.WriteHeader(state)
		jsonapi.MarshalErrors(w, errobjs)
		return
	}

	// end database transaction
	err = tx.Commit(ctx)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("commit tx")
		marshalError(w, r, err)
		return
	}

	// write out JSON
	marshalResult(ctx, doc, w)
}

func doGetTagImpl(ctx context.Context, tx pgx.Tx, query interface{}, orgId int64, tagFieldsStr string, baseURL string, now time.Time) (int, jsonapi.Payloader, []*jsonapi.ErrorObject) {
	span := trace.SpanFromContext(ctx)

	// parse fieldsets
	tagFieldsSet := parseFieldset(tagFieldsStr)
	tagFields := []string{}
	if tagFieldsSet != nil {
		for s := range *tagFieldsSet {
			tagFields = append(tagFields, s)
		}
	}

	// build the db query
	var err error
	var singleton bool
	var minResultSize int
	var st tag.Statement

	switch query := query.(type) {
	case pointQueryStrId:
		// singleton tag
		singleton = true
		minResultSize = 1
		st = tag.Fetch(ctx, tag.Point(query.Id, orgId))

	case textQuery:
		// full text search
		singleton = false
		st = tag.Fetch(ctx, tag.Text(query.Fragment, orgId))

	case setQueryStr:
		// tag set
		singleton = false
		minResultSize = len(query.Set)
		st = tag.Fetch(ctx, tag.Set(&orgId, query.Set...))

	case rangeQuery:
		// tag list
		var i *string
		start := query.Start
		if start != "" {
			i = &start
		}
		singleton = false
		st = tag.Fetch(ctx, tag.Range(i, orgId))

	default:
		tel.Log(ctx).WithField("query", fmt.Sprintf("%#v", query)).Error("build query")
		err = errInternalServerError
	}

	// fetch rows from the db
	var tagrows []*tag.Row
	if err == nil {
		st = st.Columns(tagFields...)
		tagrows, err = st.Do(tx)
	}

	// error handling
	if err != nil && err != database.ErrNotFound {
		tel.Log(ctx).WithError(err).Error("fetch tags")
		status, errobjs := buildError(ctx, err)
		return status, nil, errobjs
	}
	if len(tagrows) < minResultSize {
		status, errobjs := buildError(ctx, database.ErrNotFound)
		return status, nil, errobjs
	}

	var next *string

	// convert tag rows to API structures w/o attached resources
	tags := make([]*api.Tag, len(tagrows))
	for i, row := range tagrows {
		tag, err := tag.FromRow(row)
		if err != nil {
			tel.Log(ctx).WithError(err).Error("convert tag row")
			status, errobjs := buildError(ctx, err)
			return status, nil, errobjs
		}
		tags[i] = tag
		if next == nil || *next < tag.Id {
			next = &tag.Id
		}
	}

	// convert to JSON:API document
	var pdoc jsonapi.Payloader
	if singleton {
		pdoc, err = jsonapi.Marshal(tags[0])
	} else {
		pdoc, err = jsonapi.Marshal(tags)
	}
	if err != nil {
		tel.Log(ctx).WithError(err).Error("serialize")
		status, errobjs := buildError(ctx, err)
		return status, nil, errobjs
	}

	// insert included, filter tag fields
	if doc, ok := pdoc.(*jsonapi.OnePayload); ok {
		filterFields(doc.Data, tagFieldsSet)
		return http.StatusOK, doc, nil
	} else if doc, ok := pdoc.(*jsonapi.ManyPayload); ok {
		for i := range doc.Data {
			filterFields(doc.Data[i], tagFieldsSet)
		}
		if next != nil {
			if doc.Links == nil {
				doc.Links = &jsonapi.Links{}
			}
			(*doc.Links)["next"] = fmt.Sprintf(path.Join(baseURL, "tags?i=%s"), *next)
		}
		return http.StatusOK, doc, nil
	} else {
		span.SetStatus(codes.Error, "Failed to cast jsonapi.Payloader to single payload doc")
		status, errobjs := buildError(ctx, errSerialize)
		return status, nil, errobjs
	}
}

func (c *Router) doGetDashboard(w http.ResponseWriter, r *http.Request) {
	ctx, span := tel.Start(r.Context(), "Handler")
	defer span.End()
	vars := mux.Vars(r)

	w.Header().Set("Content-Type", jsonapi.MediaType)

	// convert org id to int
	orgId, err := strconv.ParseInt(vars["org-id"], 10, 64)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("error converting org id")
		marshalError(w, r, err)
		return
	}

	// rfc3339 timestamp for top risks timeframe; fallback to past week
	var risksFrom time.Time
	if risksFromStr := r.URL.Query().Get("risks_from"); len(risksFromStr) > 0 {
		risksFrom, err = time.Parse(time.RFC3339, risksFromStr)
		if err != nil {
			tel.Log(ctx).WithError(err).Error("risks timestamp")
			marshalError(w, r, errQueryStringWrong)
			return
		}
	} else {
		risksFrom = time.Now().UTC().Add(time.Hour * 24 * 7 * -1)
	}

	// start database transaction
	tx, err := c.database.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
	if err != nil {
		tel.Log(ctx).WithError(err).Error("start db tx")
		marshalError(w, r, err)
		return
	}
	defer tx.Rollback(ctx)

	// build dashboard document
	state, doc, errobjs := doGetDashboardImpl(ctx, tx, orgId, risksFrom)
	if state != http.StatusOK {
		w.WriteHeader(state)
		jsonapi.MarshalErrors(w, errobjs)
		return
	}

	// end database transaction
	err = tx.Commit(ctx)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("commit tx")
		marshalError(w, r, err)
		return
	}

	// write out JSON
	marshalResult(ctx, doc, w)
}

func doGetDashboardImpl(ctx context.Context, tx pgx.Tx, organizationId int64, risksFrom time.Time) (int, jsonapi.Payloader, []*jsonapi.ErrorObject) {
	span := trace.SpanFromContext(ctx)

	// get open incidents stats
	incidentStats, err := appraisal.GetIncidentStats(ctx, tx, organizationId)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("fetch incident stats")
		status, errobjs := buildError(ctx, err)
		return status, nil, errobjs
	}

	// get device state stats
	dashboardStats, err := appraisal.GetDashboardStats(ctx, tx, organizationId)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("fetch dashboard stats")
		status, errobjs := buildError(ctx, err)
		return status, nil, errobjs
	}

	// get top risks
	risks, err := appraisal.GetRisksToplist(ctx, tx, risksFrom, organizationId)
	if err != nil {
		if err != database.ErrNotFound {
			tel.Log(ctx).WithError(err).Error("fetch latest risks")
			status, errobjs := buildError(ctx, err)
			return status, nil, errobjs
		}

		// empty result
		risks = make([]appraisal.RowRisksByType, 0)
	}

	apiDeviceStats := &api.DeviceStats{}
	for _, s := range dashboardStats {
		switch s.DeviceState {
		case "unresponsive":
			apiDeviceStats.NumUnresponsive = s.NumDevices
		case "untrusted":
			apiDeviceStats.NumWithIncident = s.NumDevices
		case "atrisk":
			apiDeviceStats.NumAtRisk = s.NumDevices
		case "trusted":
			apiDeviceStats.NumTrusted = s.NumDevices
		}
	}

	// convert risks to API structures
	apiRisks := make([]*api.RiskStatsEntry, len(risks))
	for i, row := range risks {
		apiRisks[i] = row.ToApiType()
	}

	// plug in
	apiDashboard := &api.Dashboard{
		Id:               "1",
		IncidentCount:    incidentStats.IncidentCount,
		IncidentDevCount: incidentStats.DevicesAffected,
		DeviceStats:      apiDeviceStats,
		Risks:            apiRisks,
	}

	// convert to JSON:API document
	var pdoc jsonapi.Payloader
	pdoc, err = jsonapi.Marshal(apiDashboard)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("serialize")
		status, errobjs := buildError(ctx, err)
		return status, nil, errobjs
	}

	// enforce single payload
	if doc, ok := pdoc.(*jsonapi.OnePayload); ok {
		return http.StatusOK, doc, nil
	}

	span.SetStatus(codes.Error, "Failed to cast jsonapi.Payloader to single payload doc")
	status, errobjs := buildError(ctx, errSerialize)
	return status, nil, errobjs
}

func (c *Router) doListIncidents(w http.ResponseWriter, r *http.Request) {
	ctx, span := tel.Start(r.Context(), "Handler")
	defer span.End()
	vars := mux.Vars(r)

	w.Header().Set("Content-Type", jsonapi.MediaType)

	// convert org id to int
	orgId, err := strconv.ParseInt(vars["org-id"], 10, 64)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("error converting org id")
		marshalError(w, r, err)
		return
	}

	// start database transaction
	tx, err := c.database.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
	if err != nil {
		tel.Log(ctx).WithError(err).Error("start db tx")
		marshalError(w, r, err)
		return
	}
	defer tx.Rollback(ctx)

	// build incident list
	state, doc, errobjs := doListIncidentsImpl(ctx, tx, orgId)
	if state != http.StatusOK {
		w.WriteHeader(state)
		jsonapi.MarshalErrors(w, errobjs)
		return
	}

	// end database transaction
	err = tx.Commit(ctx)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("commit tx")
		marshalError(w, r, err)
		return
	}

	// write out JSON
	marshalResult(ctx, doc, w)
}

func doListIncidentsImpl(ctx context.Context, tx pgx.Tx, organizationId int64) (int, jsonapi.Payloader, []*jsonapi.ErrorObject) {
	span := trace.SpanFromContext(ctx)

	// get latest incidents
	incidents, err := appraisal.GetLatestIncidents(ctx, tx, organizationId)
	if err != nil {
		if err != database.ErrNotFound {
			tel.Log(ctx).WithError(err).Error("fetch latest incidents")
			status, errobjs := buildError(ctx, err)
			return status, nil, errobjs
		}

		// empty result
		incidents = make([]appraisal.RowLatestIncidents, 0)
	}

	// convert incidents to API structures
	apiIncidents := make([]*api.IncidentStatsEntry, len(incidents))
	for i, row := range incidents {
		apiIncidents[i] = row.ToApiType()
	}
	jsonApiIncidentsDoc := api.Incidents{Incidents: apiIncidents}

	// convert to JSON:API document
	var pdoc jsonapi.Payloader
	pdoc, err = jsonapi.Marshal(&jsonApiIncidentsDoc)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("serialize")
		status, errobjs := buildError(ctx, err)
		return status, nil, errobjs
	}

	// enforce one payload
	if doc, ok := pdoc.(*jsonapi.OnePayload); ok {
		return http.StatusOK, doc, nil
	}

	span.SetStatus(codes.Error, "Failed to cast jsonapi.Payloader to many payload doc")
	status, errobjs := buildError(ctx, errSerialize)
	return status, nil, errobjs
}

func (c *Router) doListRisks(w http.ResponseWriter, r *http.Request) {
	ctx, span := tel.Start(r.Context(), "Handler")
	defer span.End()
	vars := mux.Vars(r)

	w.Header().Set("Content-Type", jsonapi.MediaType)

	// convert org id to int
	orgId, err := strconv.ParseInt(vars["org-id"], 10, 64)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("error converting org id")
		marshalError(w, r, err)
		return
	}

	// start database transaction
	tx, err := c.database.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
	if err != nil {
		tel.Log(ctx).WithError(err).Error("start db tx")
		marshalError(w, r, err)
		return
	}
	defer tx.Rollback(ctx)

	// build incident list
	state, doc, errobjs := doListRisksImpl(ctx, tx, orgId)
	if state != http.StatusOK {
		w.WriteHeader(state)
		jsonapi.MarshalErrors(w, errobjs)
		return
	}

	// end database transaction
	err = tx.Commit(ctx)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("commit tx")
		marshalError(w, r, err)
		return
	}

	// write out JSON
	marshalResult(ctx, doc, w)
}

var since2019, _ = time.Parse("2006-01-02", "2019-01-02")

func doListRisksImpl(ctx context.Context, tx pgx.Tx, organizationId int64) (int, jsonapi.Payloader, []*jsonapi.ErrorObject) {
	span := trace.SpanFromContext(ctx)

	// get latest risks
	risks, err := appraisal.GetRisksToplist(ctx, tx, since2019, organizationId)
	if err != nil {
		if err != database.ErrNotFound {
			tel.Log(ctx).WithError(err).Error("fetch top risks")
			status, errobjs := buildError(ctx, err)
			return status, nil, errobjs
		}

		// empty result
		risks = make([]appraisal.RowRisksByType, 0)
	}

	// convert risks to API structures
	apiRisks := make([]*api.RiskStatsEntry, len(risks))
	for i, row := range risks {
		apiRisks[i] = row.ToApiType()
	}
	jsonApiRisksDoc := api.Risks{Risks: apiRisks}

	// convert to JSON:API document
	var pdoc jsonapi.Payloader
	pdoc, err = jsonapi.Marshal(&jsonApiRisksDoc)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("serialize")
		status, errobjs := buildError(ctx, err)
		return status, nil, errobjs
	}

	// enforce one payload
	if doc, ok := pdoc.(*jsonapi.OnePayload); ok {
		return http.StatusOK, doc, nil
	}

	span.SetStatus(codes.Error, "Failed to cast jsonapi.Payloader to many payload doc")
	status, errobjs := buildError(ctx, errSerialize)
	return status, nil, errobjs
}

func (c *Router) Populate(router *mux.Router) {
	// route debug path to default mux which is populated f.e. by builtin pprof
	router.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)
	router.PathPrefix("/debugv1").Handler(c.debug)

	router.Methods("GET").
		Path("/v2/configuration").
		HandlerFunc(c.doConfiguration)

	router.Methods("GET").
		Path("/v2/info").
		HandlerFunc(c.doInfo)

	router.Methods("GET").
		Path("/v2/healthy").
		HandlerFunc(c.doHealthy)

	router.Methods("GET").
		Path("/v2/ready").
		HandlerFunc(c.doReady)

	router.Methods("POST").
		Path("/v2/enroll").
		HandlerFunc(c.doEnroll)

	router.Methods("POST").
		Path("/v2/attest").
		HandlerFunc(c.doAttest)

		/*
			router.Methods("GET").
				Path("/v2/attest/nonce").
				HandlerFunc(c.doNonce)
		*/

	router.Methods("POST").
		Path("/v2/events").
		HandlerFunc(c.doEvents)

	// use a subrouter on remaining paths to facilitate authentication
	crudRouter := router.PathPrefix("/").
		Subrouter()
	crudRouter.Use(c.crudAuthenticationMiddleware)

	crudRouter.Methods("GET").
		Path("/v2/dashboard").
		HandlerFunc(c.doGetDashboard)

	crudRouter.Methods("GET").
		Path("/v2/devices").
		HandlerFunc(c.doListDevices)

	crudRouter.Methods("GET").
		Path("/v2/devices/{id:[0-9]+}").
		HandlerFunc(c.doGetDevice)

	crudRouter.Methods("GET").
		Path("/v2/devices/{id:[0-9]+}/changes").
		HandlerFunc(c.doListChangesOfDevice)

	crudRouter.Methods("GET").
		Path("/v2/devices/{id:[0-9]+}/appraisals").
		HandlerFunc(c.doListAppraisalsOfDevice)

	crudRouter.Methods("PATCH").
		Path("/v2/devices/{id:[0-9]+}").
		HandlerFunc(c.doPatchDevice)

	crudRouter.Methods("POST").
		Path("/v2/devices/{id:[0-9]+}/resurect").
		HandlerFunc(c.doResurect)

	crudRouter.Methods("POST").
		Path("/v2/devices/{id:[0-9]+}/override").
		HandlerFunc(c.doOverride)

	crudRouter.Methods("GET").
		Path("/v2/incidents").
		HandlerFunc(c.doListIncidents)

	crudRouter.Methods("GET").
		Path("/v2/risks").
		HandlerFunc(c.doListRisks)

	crudRouter.Methods("GET").
		Path("/v2/changes").
		HandlerFunc(c.doListChanges)

	crudRouter.Methods("GET").
		Path("/v2/changes/{id:[0-9]+}").
		HandlerFunc(c.doGetChange)

	crudRouter.Methods("GET").
		Path("/v2/tags").
		HandlerFunc(c.doListTags)

	crudRouter.Methods("GET").
		Path("/v2/tags/{id:[a-zA-Z0-9]{27}}").
		HandlerFunc(c.doGetTag)
}
