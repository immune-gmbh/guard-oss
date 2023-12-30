package appraisal

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/google/go-tpm/tpm2"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"

	"github.com/immune-gmbh/guard/apisrv/v2/internal/mock"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/api"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/check"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/check/baseline"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/check/policy"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/database"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/device"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/evidence"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/issuesv1"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/organization"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/report"
)

var (
	goodVerdict = api.Verdict{
		Type:               api.VerdictType,
		Result:             api.Trusted,
		SupplyChain:        api.Trusted,
		Configuration:      api.Trusted,
		Firmware:           api.Trusted,
		Bootloader:         api.Trusted,
		OperatingSystem:    api.Trusted,
		EndpointProtection: api.Trusted,
	}
)

func parseEvidence(t *testing.T, f string) *evidence.Values {
	buf, err := ioutil.ReadFile(f)
	assert.NoError(t, err)

	var ev api.Evidence
	err = json.Unmarshal(buf, &ev)
	assert.NoError(t, err)

	if len(ev.AllPCRs) == 0 {
		ev.AllPCRs = map[string]map[string]api.Buffer{
			"11": ev.PCRs,
		}
	}

	val, err := evidence.WrapInsecure(&ev)
	assert.NoError(t, err)

	return val
}

func createWithEvidence(t *testing.T, ctx context.Context, tx pgx.Tx, p string) int64 {
	now := time.Now()
	ev := parseEvidence(t, p)

	buf, err := hex.DecodeString("0022000bf7f5c2feb339f9cc41f2fadd91c0d86ba0b0c9a05bd4e43cc78761693ccdc9a7")
	assert.NoError(t, err)
	name, err := tpm2.DecodeName(bytes.NewBuffer(buf))
	assert.NoError(t, err)
	dev, err := device.GetByFingerprint(ctx, tx, (*api.Name)(name))
	assert.NoError(t, err)

	evrow, err := evidence.Persist(ctx, tx, ev, baseline.New(), policy.New(), dev, "", "", "", now)
	assert.NoError(t, err)
	rep, err := report.Compile(ctx, ev)
	assert.NoError(t, err)
	id, err := Create(ctx, tx, evrow, rep, new(check.Result), dev.AIK, dev.Id, now, "test-actor")
	if !(assert.NoError(t, err) && assert.NotEmpty(t, id)) {
		t.FailNow()
	}
	return id
}

func storeAppraisalRow(t *testing.T, ctx context.Context, tx pgx.Tx, id int64, org, file string) {
	appr, err := GetAppraisalById(ctx, tx, id)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	fd, err := os.Create(file)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	err = json.NewEncoder(fd).Encode(appr)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	fd.Close()
}

func TestGenerateTestData(t *testing.T) {
	if testing.Short() {
		t.Skipf("skipping integration test")
	}

	ctx := context.Background()
	//XXX how about only instancing postgres once per package
	pgsqlC := mock.PostgresContainer(t, ctx)
	defer pgsqlC.Terminate(ctx)

	pgsqlC.ResetAndSeed(t, ctx, database.MigrationFiles, seedSql, 16)

	db := pgsqlC.Connect(t, ctx)
	defer db.Close()

	tx, err := db.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)

	buf, err := hex.DecodeString("0022000bf7f5c2feb339f9cc41f2fadd91c0d86ba0b0c9a05bd4e43cc78761693ccdc9a7")
	assert.NoError(t, err)
	name, err := tpm2.DecodeName(bytes.NewBuffer(buf))
	assert.NoError(t, err)
	devrow, err := device.GetByFingerprint(ctx, tx, (*api.Name)(name))
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	organizationExternal, err := organization.GetExternalById(ctx, tx, devrow.OrganizationId)
	assert.NoError(t, err)

	apprId := createWithEvidence(t, ctx, tx, "../../test/DESKTOP-MIMU51J-before.json")
	storeAppraisalRow(t, ctx, tx, apprId, *organizationExternal, "../../test/output/good.appraisal.json")

	apprId = createWithEvidence(t, ctx, tx, "../../test/DESKTOP-MIMU51J-no-secureboot.json")
	storeAppraisalRow(t, ctx, tx, apprId, *organizationExternal, "../../test/output/no-secureboot.appraisal.json")
}

func TestEmptyVerdict(t *testing.T) {
	res := check.Result{
		Issues: nil,
	}
	ver := determineVerdict(&res)
	myver := goodVerdict
	myver.SupplyChain = api.Unsupported
	myver.EndpointProtection = api.Unsupported
	assert.Equal(t, myver, ver)
}

func TestVerdictCategories(t *testing.T) {
	var iss issuesv1.TpmDummy
	iss.Common.Id = issuesv1.TpmDummyId
	iss.Common.Incident = issuesv1.TpmDummyIncident
	iss.Common.Aspect = issuesv1.TpmDummyAspect
	res := check.Result{
		Issues: []issuesv1.Issue{&iss},
	}
	ver := determineVerdict(&res)
	myver := goodVerdict
	myver.Result = api.Vulnerable
	myver.SupplyChain = api.Vulnerable
	myver.EndpointProtection = api.Unsupported
	assert.Equal(t, myver, ver)

	iss.Common.Id = issuesv1.UefiBootFailureId
	iss.Common.Aspect = issuesv1.UefiBootFailureAspect
	iss.Common.Incident = issuesv1.UefiBootFailureIncident
	ver = determineVerdict(&res)
	myver = goodVerdict
	myver.Result = api.Vulnerable
	myver.Firmware = api.Vulnerable
	myver.SupplyChain = api.Unsupported
	myver.EndpointProtection = api.Unsupported
	assert.Equal(t, myver, ver)

	iss.Common.Id = issuesv1.UefiGptChangedId
	iss.Common.Aspect = issuesv1.UefiGptChangedAspect
	iss.Common.Incident = issuesv1.UefiGptChangedIncident
	ver = determineVerdict(&res)
	myver = goodVerdict
	myver.Result = api.Vulnerable
	myver.Bootloader = api.Vulnerable
	myver.SupplyChain = api.Unsupported
	myver.EndpointProtection = api.Unsupported
	assert.Equal(t, myver, ver)

	iss.Common.Id = issuesv1.WindowsBootCounterReplayId
	iss.Common.Aspect = issuesv1.WindowsBootCounterReplayAspect
	iss.Common.Incident = issuesv1.WindowsBootCounterReplayIncident
	ver = determineVerdict(&res)
	myver = goodVerdict
	myver.Result = api.Vulnerable
	myver.OperatingSystem = api.Vulnerable
	myver.SupplyChain = api.Unsupported
	myver.EndpointProtection = api.Unsupported
	assert.Equal(t, myver, ver)
}

func TestNonFatalVerdict(t *testing.T) {
	var iss issuesv1.Binarly
	iss.Common.Id = issuesv1.BrlyIntelBssaDft
	iss.Common.Incident = issuesv1.BinarlyIncident
	iss.Common.Aspect = issuesv1.BinarlyAspect
	res := check.Result{
		Issues: []issuesv1.Issue{&iss},
	}
	ver := determineVerdict(&res)
	myver := goodVerdict
	myver.EndpointProtection = api.Unsupported
	myver.SupplyChain = api.Unsupported
	assert.Equal(t, myver, ver)
}

func TestVerdictDups(t *testing.T) {
	var iss issuesv1.UefiBootAppSet
	iss.Common.Id = issuesv1.UefiBootAppSetId
	iss.Common.Incident = issuesv1.UefiBootAppSetIncident
	iss.Common.Aspect = issuesv1.UefiBootAppSetAspect
	res := check.Result{
		Issues: []issuesv1.Issue{&iss, &iss},
	}
	ver := determineVerdict(&res)
	myver := goodVerdict
	myver.Result = api.Vulnerable
	myver.Bootloader = api.Vulnerable
	myver.SupplyChain = api.Unsupported
	myver.EndpointProtection = api.Unsupported
	assert.Equal(t, ver, myver)
}

func makeIssuesTestData() *check.Result {
	res := check.Result{
		Issues: []issuesv1.Issue{
			&issuesv1.GenericIssue{
				Common: issuesv1.Common{
					Id:       "test/issue-a",
					Aspect:   "firmware",
					Incident: false,
				},
			},
			&issuesv1.GenericIssue{
				Common: issuesv1.Common{
					Id:       "test/issue-b",
					Aspect:   "firmware",
					Incident: true,
				},
				Args: map[string]interface{}{"b-val": float64(23), "c-val": float64(42)},
			}},
	}
	return &res
}

func TestCreate(t *testing.T) {
	if testing.Short() {
		t.Skipf("skipping integration test")
	}

	ctx := context.Background()
	//XXX how about only instancing postgres once per package
	pgsqlC := mock.PostgresContainer(t, ctx)
	defer pgsqlC.Terminate(ctx)

	pgsqlC.ResetAndSeed(t, ctx, database.MigrationFiles, seedSql, 16)

	conn := pgsqlC.Connect(t, ctx)
	defer conn.Close()

	tx, err := conn.Begin(ctx)
	assert.NoError(t, err)
	defer tx.Rollback(ctx)

	now := time.Now()
	values := new(evidence.Values)
	values.Type = evidence.ValuesType
	aik := api.Name(api.GenerateName(rand.New(rand.NewSource(42))))
	bline := baseline.New()
	pol := policy.New()
	dev := device.DevAikRow{Id: 0, AIK: &device.KeysRow{QName: aik}}
	ev, err := evidence.Persist(ctx, tx, values, bline, pol, &dev, "", "", "", now)
	assert.NoError(t, err)
	key := device.KeysRow{Id: 100}
	rep := new(api.Report)
	rep.Type = api.ReportType
	id, err := Create(ctx, tx, ev, rep, makeIssuesTestData(), &key, 100, now, "test-actor")
	if assert.NoError(t, err) {
		assert.NotEmpty(t, id)
	}

	// shadily inline testing of fetch right here (as opposed to somewhere in fetch_test.go)
	issues, err := GetIssuesByAppraisal(ctx, tx, id)
	if assert.NoError(t, err) && assert.NotEmpty(t, issues) {
		assert.EqualValues(t, makeIssuesTestData().Issues, issues)
	}

	err = tx.Commit(ctx)
	assert.NoError(t, err)
}
