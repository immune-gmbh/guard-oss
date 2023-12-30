package event

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"testing"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/assert"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/api"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/database"

	"github.com/immune-gmbh/guard/apisrv/v2/internal/mock"
)

const reportStr = `{"type": "report/1", "host": {"name": "Windows 10 RC2", "hostname": "example.com", "type": "windows", "cpu_vendor": "GenuineIntel"}, "smbios": {"manufacturer": "Lenovo", "product": "ThinkPad X230", "bios_vendor": "Lenovo", "bios_version": "1.0", "bios_release_date": "11/2/2011"}, "uefi":{"mode": "deployed", "secureboot": true, "platform_keys": [], "exchange_keys": [], "permitted_keys": [], "forbidden_keys": []}, "tpm": {"manufacturer": "IFX", "vendor_id": "Opti", "spec_version": "2", "eventlog": []}, "csme": {"variant": "consumer", "version": [1,2,3], "recovery_version": [1,2,3], "fitc_version": [1,2,3]}, "sgx": {"version": 2, "enabled": true, "flc": true, "kss": true, "enclave_size_64": 10000, "enclave_size_32": 1000, "epc": []}, "txt": {"ready": true}, "sev": {"enabled": true, "version": [1,18], "sme":true, "es": true, "vte": true, "snp": true, "vmpl": true, "guests": 3, "min_asid": 1}}`

func TestEvents(t *testing.T) {
	if testing.Short() {
		t.Skipf("skipping integration test")
	}

	ctx := context.Background()
	tests := map[string]func(*testing.T, *pgxpool.Pool){
		"BillingUpdate":            testBillingUpdate,
		"ExpiredAppraisal":         testExpiredAppraisal,
		"FailedAppraisal":          testFailedAppraisal,
		"ContinuedFailedAppraisal": testContinuedFailedAppraisal,
		"GenerateFixtures":         testGenerateFixtures,
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

func testBillingUpdate(t *testing.T, db *pgxpool.Pool) {
	ctx := context.Background()
	now := time.Now()
	devs := map[string]int{
		"d5d70fad-c11f-4609-8cab-17438305f0b0": 1337,
		"39f5930f-d9cf-49d4-9ad4-2054ae4dfea5": 1,
		"3fab3a3e-c286-47cf-a096-dfc899441c8e": 0,
		"f2561288-76e5-42cb-bea8-90c8d6265a63": 42,
		"ad89ec0a-d2a7-40a9-8297-d87fae349311": 23,
		"f8e47633-a3b5-4f7e-bcaa-3272fe51f7fa": 420,
		"1222fe6a-e762-41b9-8252-5af7a6547342": 999999,
	}

	row, err := BillingUpdate(ctx, db, "test", devs, now, now)
	assert.NoError(t, err)
	assert.NotNil(t, row)

	assert.Equal(t, jobType, row.Type)
	assert.Regexp(t, `billing-update/1\([0-9a-f]+\)`, row.Reference)

	_, err = BillingUpdate(ctx, db, "test", devs, now, now)
	assert.Error(t, err)

	_, err = BillingUpdate(ctx, db, "test", devs, now.Add(time.Hour), now)
	assert.NoError(t, err)
}

func testExpiredAppraisal(t *testing.T, db *pgxpool.Pool) {
	ctx := context.Background()
	now := time.Now()
	dev := api.Device{
		Id:     "112233",
		Cookie: "asasasas",
		Name:   "Device #1",
		State:  "outdated",
	}
	appr := api.Appraisal{
		Id:       "testlalala",
		Received: time.Now().UTC().Add(-10 * time.Hour * 24),
		Expires:  time.Now().UTC().Add(-1 * time.Hour * 24),
		Verdict: api.Verdict{
			Type:               api.VerdictType,
			Result:             api.Trusted,
			SupplyChain:        api.Trusted,
			Configuration:      api.Trusted,
			Firmware:           api.Trusted,
			Bootloader:         api.Trusted,
			OperatingSystem:    api.Trusted,
			EndpointProtection: api.Trusted,
		},
		Report: api.Report{},
	}
	err := json.Unmarshal([]byte(reportStr), &appr.Report)
	assert.NoError(t, err)

	row, err := AppraisalExpired(ctx, db, "test", "org", &dev, &appr, now)
	assert.NoError(t, err)
	assert.NotNil(t, row)

	assert.Equal(t, jobType, row.Type)
	assert.Equal(t, "appraisal-expired/1(testlalala)", row.Reference)

	_, err = AppraisalExpired(ctx, db, "test", "org", &dev, &appr, now)
	assert.Error(t, err)

	appr.Id = "332211"
	_, err = AppraisalExpired(ctx, db, "test", "org", &dev, &appr, now)
	assert.NoError(t, err)
}

func testFailedAppraisal(t *testing.T, db *pgxpool.Pool) {
	ctx := context.Background()
	now := time.Now()
	dev := api.Device{
		Id:     "112233",
		Cookie: "asasasas",
		Name:   "Device #1",
		State:  "vulnerable",
	}
	appr := api.Appraisal{
		Id:       "testlalala",
		Received: time.Now().UTC().Add(-10 * time.Hour * 24),
		Expires:  time.Now().UTC().Add(10 * time.Hour * 24),
		Verdict: api.Verdict{
			Type:               api.VerdictType,
			Result:             api.Vulnerable,
			SupplyChain:        api.Trusted,
			Configuration:      api.Trusted,
			Firmware:           api.Vulnerable,
			Bootloader:         api.Trusted,
			OperatingSystem:    api.Trusted,
			EndpointProtection: api.Trusted,
		},
		Report: api.Report{},
	}
	err := json.Unmarshal([]byte(reportStr), &appr.Report)
	assert.NoError(t, err)

	row, err := NewAppraisal(ctx, db, "test", "org", &dev, nil, &appr, now)
	assert.NoError(t, err)
	assert.NotNil(t, row)

	assert.Equal(t, jobType, row.Type)
	assert.Equal(t, "new-appraisal/1(testlalala)", row.Reference)

	_, err = NewAppraisal(ctx, db, "test", "org", &dev, nil, &appr, now)
	assert.Error(t, err)

	appr.Id = "332211"
	_, err = NewAppraisal(ctx, db, "test", "org", &dev, nil, &appr, now)
	assert.NoError(t, err)
}

func testContinuedFailedAppraisal(t *testing.T, db *pgxpool.Pool) {
	now := time.Now()
	ctx := context.Background()
	dev := api.Device{
		Id:     "112233",
		Cookie: "asasasas",
		Name:   "Device #1",
		State:  "vulnerable",
	}
	appr1 := api.Appraisal{
		Id:       "testlalala",
		Received: time.Now().UTC().Add(-10 * time.Hour * 24),
		Expires:  time.Now().UTC().Add(-5 * time.Hour * 24),
		Verdict: api.Verdict{
			Type:               api.VerdictType,
			Result:             api.Vulnerable,
			SupplyChain:        api.Trusted,
			Configuration:      api.Trusted,
			Firmware:           api.Vulnerable,
			Bootloader:         api.Trusted,
			OperatingSystem:    api.Trusted,
			EndpointProtection: api.Trusted,
		},
		Report: api.Report{},
	}
	appr2 := api.Appraisal{
		Id:       "testlololl",
		Received: time.Now().UTC().Add(-4 * time.Hour * 24),
		Expires:  time.Now().UTC().Add(10 * time.Hour * 24),
		Verdict: api.Verdict{
			Type:               api.VerdictType,
			Result:             api.Vulnerable,
			SupplyChain:        api.Trusted,
			Configuration:      api.Trusted,
			Firmware:           api.Vulnerable,
			Bootloader:         api.Trusted,
			OperatingSystem:    api.Trusted,
			EndpointProtection: api.Trusted,
		},
		Report: api.Report{},
	}
	err := json.Unmarshal([]byte(reportStr), &appr1.Report)
	assert.NoError(t, err)
	err = json.Unmarshal([]byte(reportStr), &appr2.Report)
	assert.NoError(t, err)

	row, err := NewAppraisal(ctx, db, "test", "org", &dev, &appr1, &appr2, now)
	assert.NoError(t, err)
	assert.NotNil(t, row)

	assert.Equal(t, jobType, row.Type)
	assert.Equal(t, "new-appraisal/1(testlololl)", row.Reference)

	_, err = NewAppraisal(ctx, db, "test", "org", &dev, &appr1, &appr2, now)
	assert.Error(t, err)

	appr2.Id = "332211"
	_, err = NewAppraisal(ctx, db, "test", "org", &dev, &appr1, &appr2, now)
	assert.NoError(t, err)
}

func testGenerateFixtures(t *testing.T, db *pgxpool.Pool) {
	ctx := context.Background()
	org := "test"
	dev1 := api.Device{
		Id:     "112233",
		Cookie: "asasasas",
		Name:   "Device #1",
		State:  api.StateVuln,
	}
	dev2 := api.Device{
		Id:     "112234",
		Cookie: "asasasas",
		Name:   "Device #2",
		State:  api.StateTrusted,
	}
	appr1 := api.Appraisal{
		Id:       "testlalala",
		Received: time.Now().UTC().Add(-10 * time.Hour * 24),
		Expires:  time.Now().UTC().Add(-5 * time.Hour * 24),
		Verdict: api.Verdict{
			Type:               api.VerdictType,
			Result:             api.Vulnerable,
			SupplyChain:        api.Trusted,
			Configuration:      api.Trusted,
			Firmware:           api.Vulnerable,
			Bootloader:         api.Trusted,
			OperatingSystem:    api.Trusted,
			EndpointProtection: api.Trusted,
		},
		Report: api.Report{},
	}
	appr2 := api.Appraisal{
		Id:       "testlololl",
		Received: time.Now().UTC().Add(-4 * time.Hour * 24),
		Expires:  time.Now().UTC().Add(10 * time.Hour * 24),
		Verdict: api.Verdict{
			Type:               api.VerdictType,
			Result:             api.Vulnerable,
			SupplyChain:        api.Trusted,
			Configuration:      api.Trusted,
			Firmware:           api.Vulnerable,
			Bootloader:         api.Trusted,
			OperatingSystem:    api.Trusted,
			EndpointProtection: api.Trusted,
		},
		Report: api.Report{},
	}
	appr3 := api.Appraisal{
		Id:       "testlololl",
		Received: time.Now().UTC().Add(-3 * time.Hour * 24),
		Expires:  time.Now().UTC().Add(13 * time.Hour * 24),
		Verdict: api.Verdict{
			Type:               api.VerdictType,
			Result:             api.Trusted,
			SupplyChain:        api.Trusted,
			Configuration:      api.Trusted,
			Firmware:           api.Trusted,
			Bootloader:         api.Trusted,
			OperatingSystem:    api.Trusted,
			EndpointProtection: api.Unsupported,
		},
		Report: api.Report{},
	}
	err := json.Unmarshal([]byte(reportStr), &appr1.Report)
	assert.NoError(t, err)
	err = json.Unmarshal([]byte(reportStr), &appr2.Report)
	assert.NoError(t, err)
	err = json.Unmarshal([]byte(reportStr), &appr3.Report)
	assert.NoError(t, err)

	ev := api.NewAppraisalEvent{
		Device:   dev1,
		Previous: &appr1,
		Next:     appr2,
	}
	ce := args(ctx, "localhost", &org, api.NewAppraisalEventType, ev)
	buf, err := json.Marshal(ce)
	assert.NoError(t, err)
	ioutil.WriteFile("../../test/new-appraisal-1.cloudevents.json", buf, 0644)

	ev = api.NewAppraisalEvent{
		Device: dev1,
		Next:   appr2,
	}
	ce = args(ctx, "localhost", &org, api.NewAppraisalEventType, ev)
	buf, err = json.Marshal(ce)
	assert.NoError(t, err)
	ioutil.WriteFile("../../test/new-appraisal-2.cloudevents.json", buf, 0644)

	ev = api.NewAppraisalEvent{
		Device: dev2,
		Next:   appr3,
	}
	ce = args(ctx, "localhost", &org, api.NewAppraisalEventType, ev)
	buf, err = json.Marshal(ce)
	assert.NoError(t, err)
	ioutil.WriteFile("../../test/new-appraisal-3.cloudevents.json", buf, 0644)
}
