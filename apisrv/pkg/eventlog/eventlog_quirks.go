package eventlog

// inject3 appends three new events into the event log.
func inject3(e *EventLog, pcr int, data1, data2, data3 string) error {
	if err := inject(e, pcr, data1); err != nil {
		return err
	}
	if err := inject(e, pcr, data2); err != nil {
		return err
	}
	return inject(e, pcr, data3)
}

// inject2 appends two new events into the event log.
func inject2(e *EventLog, pcr int, data1, data2 string) error {
	if err := inject(e, pcr, data1); err != nil {
		return err
	}
	return inject(e, pcr, data2)
}

// inject appends a new event into the event log.
func inject(e *EventLog, pcr int, data string) error {
	evt := rawEvent{
		data:     []byte(data),
		index:    pcr,
		sequence: e.rawEvents[len(e.rawEvents)-1].sequence + 1,
	}
	for _, alg := range e.Algs {
		h := alg.CryptoHash().New()
		h.Write([]byte(data))
		evt.digests = append(evt.digests, digest{hash: alg.CryptoHash(), data: h.Sum(nil)})
	}
	e.rawEvents = append(e.rawEvents, evt)
	return nil
}

var eventlogWorkarounds = []elWorkaround{
	{
		id:          "EBS Invocation + Success",
		affectedPCR: 5,
		apply: func(e *EventLog) error {
			return inject2(e, 5, ebsInvocation, ebsSuccess)
		},
	},
	{
		id:          "EBS Invocation + Failure",
		affectedPCR: 5,
		apply: func(e *EventLog) error {
			return inject2(e, 5, ebsInvocation, ebsFailure)
		},
	},
	{
		id:          "EBS Invocation + Failure + Success",
		affectedPCR: 5,
		apply: func(e *EventLog) error {
			return inject3(e, 5, ebsInvocation, ebsFailure, ebsSuccess)
		},
	},
}
