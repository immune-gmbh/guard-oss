package evidence

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"testing"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/assert"

	"github.com/immune-gmbh/guard/apisrv/v2/internal/mock"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/api"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/check/baseline"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/check/policy"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/database"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/device"
)

func TestRow(t *testing.T) {
	if testing.Short() {
		t.Skipf("skipping integration test")
	}

	ctx := context.Background()
	tests := map[string]func(*testing.T, *pgxpool.Pool){
		"Persist":          testPersist,
		"ByRef":            testByRef,
		"IsReady":          testIsReady,
		"IsReadyAndLatest": testIsReadyAndLatest,
		// evidence.MostRecent()
		"MostRecentNone": testMostRecentNone,
		"GarbageCollect": testGC,
	}
	pgsqlC := mock.PostgresContainer(t, ctx)
	defer pgsqlC.Terminate(ctx)

	for name, fn := range tests {
		pgsqlC.Reset(t, ctx, database.MigrationFiles)

		t.Run(name, func(t *testing.T) {
			conn := pgsqlC.Connect(t, ctx)
			defer conn.Close()
			fn(t, conn)
		})
	}
}

func testByRef(t *testing.T, db *pgxpool.Pool) {
	ctx := context.Background()
	now := time.Now()
	values := Values{Type: ValuesType}
	aik := api.Name(api.GenerateName(rand.New(rand.NewSource(42))))
	bline := baseline.New()
	pol := policy.New()
	refa := "aaaaaaaaaaaaaaaaaaaaaaaaaaa"
	refb := "bbbbbbbbbbbbbbbbbbbbbbbbbbb"
	refc := "ccccccccccccccccccccccccccc"

	dev := device.DevAikRow{Id: 0, AIK: &device.KeysRow{QName: aik}}
	row, err := Persist(ctx, db, &values, bline, pol, &dev, refa, refb, refc, now)
	assert.NoError(t, err)

	rows, err := ByReference(ctx, db, refa)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(rows))
	assert.Equal(t, row.Id, rows[0].Id)

	rows, err = ByReference(ctx, db, refb)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(rows))
	assert.Equal(t, row.Id, rows[0].Id)

	rows, err = ByReference(ctx, db, refc)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(rows))
	assert.Equal(t, row.Id, rows[0].Id)

	row2, err := Persist(ctx, db, &values, bline, pol, &dev, refa, "", "", now)
	assert.NoError(t, err)

	rows, err = ByReference(ctx, db, refa)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(rows))
	assert.Contains(t, []string{row.Id, row2.Id}, rows[0].Id)
	assert.Contains(t, []string{row.Id, row2.Id}, rows[1].Id)
}

func testPersist(t *testing.T, db *pgxpool.Pool) {
	ctx := context.Background()
	now := time.Now()
	values := Values{Type: ValuesType}
	aik := api.Name(api.GenerateName(rand.New(rand.NewSource(42))))
	bline := baseline.New()
	pol := policy.New()
	ref := "aaaaaaaaaaaaaaaaaaaaaaaaaaa"
	dev := device.DevAikRow{Id: 0, AIK: &device.KeysRow{QName: aik}}
	row, err := Persist(ctx, db, &values, bline, pol, &dev, ref, ref, ref, now)
	assert.NoError(t, err)
	assert.NotEmpty(t, row.Id)
	assert.Equal(t, *row.BinarlyReference, ref)
	assert.Equal(t, *row.ImageReference, ref)
	assert.Equal(t, *row.IntelTSCReference, ref)
	assert.False(t, row.Baseline.IsNull())
	assert.False(t, row.Values.IsNull())
}

func testIsReady(t *testing.T, db *pgxpool.Pool) {
	ctx := context.Background()
	now := time.Now()
	values := Values{Type: ValuesType}
	aik := api.Name(api.GenerateName(rand.New(rand.NewSource(42))))
	bline := baseline.New()
	pol := policy.New()
	ref := "aaaaaaaaaaaaaaaaaaaaaaaaaaa"
	dev := device.DevAikRow{Id: 0, AIK: &device.KeysRow{QName: aik}}
	row, err := Persist(ctx, db, &values, bline, pol, &dev, ref, ref, ref, now)
	assert.NoError(t, err)
	ready, err := IsReadyAndLatest(ctx, db, row, time.Now().Add(-30*time.Second))
	assert.NoError(t, err)
	assert.False(t, ready)
}

func testIsReadyAndLatest(t *testing.T, db *pgxpool.Pool) {
	ctx := context.Background()
	now := time.Now()
	values := Values{Type: ValuesType}
	aik := api.Name(api.GenerateName(rand.New(rand.NewSource(42))))
	bline := baseline.New()
	pol := policy.New()
	maxAge := time.Now().Add(-30 * time.Second)

	ref1 := "aaaaaaaaaaaaaaaaaaaaaaaaaaa"
	dev := device.DevAikRow{Id: 0, AIK: &device.KeysRow{QName: aik}}
	row1, err := Persist(ctx, db, &values, bline, pol, &dev, ref1, ref1, ref1, now)
	assert.NoError(t, err)

	ready, err := IsReadyAndLatest(ctx, db, row1, maxAge)
	assert.NoError(t, err)
	assert.False(t, ready)

	_, err = db.Exec(ctx, `
    insert into v2.jobs (reference, type, next_run_at, scheduled_at, finished_at, successful)
    values ($1, 'something', 'now'::timestamptz, 'now'::timestamptz, 'now'::timestamptz, true)
  `, ref1)
	assert.NoError(t, err)

	ready, err = IsReadyAndLatest(ctx, db, row1, maxAge)
	assert.NoError(t, err)
	assert.True(t, ready)

	ref2 := "bbbbbbbbbbbbbbbbbbbbbbbbbbb"
	row2, err := Persist(ctx, db, &values, bline, pol, &dev, ref2, ref2, ref2, now)
	assert.NoError(t, err)

	ready, err = IsReadyAndLatest(ctx, db, row1, maxAge)
	assert.NoError(t, err)
	assert.True(t, ready)

	ready, err = IsReadyAndLatest(ctx, db, row2, maxAge)
	assert.NoError(t, err)
	assert.False(t, ready)

	_, err = db.Exec(ctx, `
    insert into v2.jobs (reference, type, next_run_at, scheduled_at, finished_at, successful)
    values ($1, 'something', 'now'::timestamptz, 'now'::timestamptz, 'now'::timestamptz, true)
   `, ref2)
	assert.NoError(t, err)

	ready, err = IsReadyAndLatest(ctx, db, row2, maxAge)
	assert.NoError(t, err)
	assert.True(t, ready)

	_, err = db.Exec(ctx, `
    insert into
      v2.organizations (id, external, devices, features, updated_at)
    values
      (100, 'ext', 100, array[]::v2.organizations_feature[], 'NOW');
    `)
	assert.NoError(t, err)

	_, err = db.Exec(ctx, `
    insert into
      v2.devices (id, hwid, fpr, name, baseline, retired, organization_id)
    values (
      100,
      E'\\x0022000b305c1823252de4490e639ec0c327f8d1da14e18634ef7f6c3b57b0ec33c1234a',
      E'\\x0022000b305c1823252de4490e639ec0c327f8d1da14e18634ef7f6c3b57b0ec33c1234a',
      'Test Device #1',
      '{"type": "baseline/3"}',
      false,
      100
    )
   `)
	assert.NoError(t, err)

	_, err = db.Exec(ctx, `
    insert into v2.appraisals (evidence_id, received_at, appraised_at, expires, verdict, report, device_id)
    values (
      $1::v2.ksuid,
      $2::timestamptz,
      $2::timestamptz,
      $2::timestamptz + '2 days',
      '{ "type": "verdict/2" }'::jsonb,
      '{ "type": "report/2" }'::jsonb,
      100
    )
    `, row2.Id, row2.ReceivedAt)
	assert.NoError(t, err)

	ready, err = IsReadyAndLatest(ctx, db, row1, maxAge)
	assert.NoError(t, err)
	assert.False(t, ready)
}

func testMostRecentNone(t *testing.T, db *pgxpool.Pool) {
	ctx := context.Background()
	rng := rand.New(rand.NewSource(42))
	_, err := db.Exec(ctx, `
    insert into v2.organizations (id, external, devices, features, updated_at)
      values (100, 'test', 1000, '{}', 'now'::timestamptz)
  `)
	assert.NoError(t, err)

	scenario := map[int]map[int]struct {
		InProgress bool
		Legacy     bool
	}{
		100: {},
		101: {
			101: {},
		},
		102: {
			201: {},
			202: {},
			203: {},
		},
		103: {
			301: {},
			302: {InProgress: true},
		},
		104: {
			401: {Legacy: true},
			402: {Legacy: true},
		},
	}

	for dev, apprs := range scenario {
		timeoff := time.Hour
		ek := api.Name(api.GenerateName(rng))
		root := api.Name(api.GenerateName(rng))
		aik := api.PublicKey(api.GeneratePublic(rng))
		name, err := api.ComputeName(aik)
		assert.NoError(t, err)

		_, err = db.Exec(ctx, `
      insert into v2.devices (id, hwid, fpr, name, baseline, policy, retired, organization_id)
      values ($1, $2::bytea, $3::bytea, 'Test device', '{"type":"lala"}', '{"type":"lala"}', true, 100)
    `, dev, ek, root)
		assert.NoError(t, err)

		_, err = db.Exec(ctx, `
      insert into v2.keys (id, public, name, fpr, credential, device_id)
      values ($1, $2::bytea, 'aik', $3::bytea, '{}', $1);
    `, dev, aik, api.Name(name))
		assert.NoError(t, err)

		apprids := []int{}
		for apprid := range apprs {
			apprids = append(apprids, apprid)
		}
		sort.Ints(apprids)

		for _, apprid := range apprids {
			sc := apprs[apprid]
			evid := fmt.Sprintf("%daaaaaaaaaaaaaaaaaaaaaaaa", apprid)
			now := time.Now().Add(-24*time.Hour + timeoff)
			exp := now.Add(24 * time.Hour)
			bline := baseline.New()
			pol := policy.New()
			rep := api.Report{Type: api.ReportType}
			vals := Values{Type: ValuesType}
			verdict := api.Verdict{Type: api.VerdictType}

			if !sc.Legacy {
				_, err = db.Exec(ctx, `
          insert into v2.evidence (id, received_at, signed_by, values, baseline, policy)
          select $1, $2, fpr, $3, $4, $5
          from v2.keys where device_id = $6
        `, evid, now, vals, bline, pol, dev)
				assert.NoError(t, err)
			} else {
				evid = ""
			}

			if !sc.InProgress {
				_, err = db.Exec(ctx, `
          insert into v2.appraisals (
            id, received_at, appraised_at, expires, verdict, evidence_id, report, key_id, device_id
          )
          select $1, $2, $2, $3, $4, nullif($5, ''), $6, id, $7
          from v2.keys where device_id = $7
        `, apprid, now, exp, verdict, evid, rep, dev)
				assert.NoError(t, err)
			}
			timeoff += time.Hour
		}
	}

	// no appraisal or evidence
	row, err := MostRecent(ctx, db, "100", int64(100))
	assert.Equal(t, database.ErrNotFound, err)
	assert.Nil(t, row)

	// one appraisal w/ evidence
	row, err = MostRecent(ctx, db, "101", int64(100))
	assert.NoError(t, err)
	assert.Equal(t, "101aaaaaaaaaaaaaaaaaaaaaaaa", row.Id)

	// three appraisals w/ evidence
	row, err = MostRecent(ctx, db, "102", int64(100))
	assert.NoError(t, err)
	assert.Equal(t, "203aaaaaaaaaaaaaaaaaaaaaaaa", row.Id)

	// two appraisals w/ evidence + one in progress
	row, err = MostRecent(ctx, db, "103", int64(100))
	assert.NoError(t, err)
	assert.Equal(t, "301aaaaaaaaaaaaaaaaaaaaaaaa", row.Id)

	// two legacy appraisals w/ inline evidence
	row, err = MostRecent(ctx, db, "104", int64(100))
	assert.Equal(t, database.ErrNotFound, err)
	assert.Nil(t, row)
}
