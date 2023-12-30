package event

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sort"
	"time"

	ce "github.com/cloudevents/sdk-go/v2"
	ceclient "github.com/cloudevents/sdk-go/v2/client"
	"github.com/georgysavva/scany/pgxscan"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/api"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/queue"
)

func args(ctx context.Context, serviceName string, subject *string, evtype string, evbody interface{}) interface{} {
	ev := ce.NewEvent()
	ev.SetType(evtype)
	ev.SetSource(serviceName)
	ev.SetTime(time.Now().UTC())
	if subject != nil {
		ev.SetSubject(*subject)
	}
	ev.SetData(ce.ApplicationJSON, evbody)

	return ceclient.DefaultIDToUUIDIfNotSet(ctx, ev)
}

func BillingUpdate(ctx context.Context, q pgxscan.Querier, sender string, currentDevicesCount map[string]int, date time.Time, now time.Time) (*queue.Row, error) {
	h := sha256.New()
	records := make([]api.BillingUpdateUsageRecord, len(currentDevicesCount))
	keys := make([]string, 0)

	h.Write([]byte(fmt.Sprintf("%d", date.UnixMilli())))
	for k := range currentDevicesCount {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for i, k := range keys {
		v := currentDevicesCount[k]
		h.Write([]byte(k))
		h.Write([]byte(fmt.Sprintf("%d", v)))
		records[i] = api.BillingUpdateUsageRecord{
			Organisation: k,
			Devices:      v,
		}
	}
	ev := api.BillingUpdateEvent{
		Timestamp: date,
		Records:   records,
	}
	ref := fmt.Sprintf("billing-update/1(%x)", h.Sum(nil))
	a := args(ctx, sender, nil, api.BillingUpdateEventType, ev)
	return queue.Enqueue(ctx, q, jobType, ref, a, now, now)
}

func NewAppraisal(ctx context.Context, q pgxscan.Querier, sender string, org string, dev *api.Device, prev *api.Appraisal, next *api.Appraisal, now time.Time) (*queue.Row, error) {
	ev := api.NewAppraisalEvent{
		Device:   *dev,
		Previous: prev,
		Next:     *next,
	}

	ref := fmt.Sprintf("new-appraisal/1(%s)", next.Id)
	a := args(ctx, sender, &org, api.NewAppraisalEventType, ev)
	return queue.Enqueue(ctx, q, jobType, ref, a, now, now)
}

func AppraisalExpired(ctx context.Context, q pgxscan.Querier, sender string, org string, dev *api.Device, prev *api.Appraisal, now time.Time) (*queue.Row, error) {
	ev := api.AppraisalExpiredEvent{
		Device:   *dev,
		Previous: *prev,
	}

	ref := fmt.Sprintf("appraisal-expired/1(%s)", prev.Id)
	a := args(ctx, sender, &org, api.AppraisalExpiredEventType, ev)
	return queue.Enqueue(ctx, q, jobType, ref, a, now, now)
}
