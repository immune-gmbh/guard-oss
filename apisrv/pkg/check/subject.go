package check

import (
	"context"
	"encoding/hex"

	acert "github.com/google/go-attestation/attributecert"
	"github.com/klauspost/compress/zstd"
	log "github.com/sirupsen/logrus"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/api"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/binarly"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/blob"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/check/baseline"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/check/policy"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/eventlog"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/evidence"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/inteltsc"
	tel "github.com/immune-gmbh/guard/apisrv/v2/pkg/telemetry"
)

type Subject struct {
	Policy           *policy.Values
	Baseline         *baseline.Values
	BaselineModified bool

	Values               *evidence.Values
	Image                *blob.Row
	BootEventLogIdx      int
	CurrentEventLogIdx   int
	EventLogs            []*eventlog.EventLog
	IMALog               []eventlog.TPMEvent
	WindowsLogs          []*eventlog.WinEvents
	AntiMalwareProcesses map[string][]byte
	EarlyLaunchDrivers   map[string][]byte
	Boot                 evidence.Boot
	BinarlyReport        *binarly.Report
	IntelTSCData         *inteltsc.Data
	PlatformCertificates []*acert.AttributeCertificate
	FwupdDevices         map[string]FwupdDevice
	BootApps             map[string][]byte
}

type FwupdDevice struct {
	Name          string
	Version       string
	VersionFormat int
	Releases      []FwupdRelease
}

type FwupdRelease struct {
	Version string
	Flags   uint64
}

type WithBinarly struct {
	Report *binarly.Report
}

type WithIntelTSC struct {
	Data         *inteltsc.Data
	Certificates []*acert.AttributeCertificate
}

// map hex digest -> uncompressed contents
type WithBlobs struct {
	Blobs map[string][]byte
}

func unzstd(ctx context.Context, buf []byte) ([]byte, error) {
	dec, err := zstd.NewReader(nil)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("create zstd decoder")
		return nil, err
	}
	defer dec.Close()

	ret, err := dec.DecodeAll(buf, nil)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("decompress")
		return nil, err
	}

	return ret, nil
}

// parse eventlogs into decoded event logs and the cold boot log into a windows boot log and a boot state
func parseEventLogs(ctx context.Context, eventLogs []api.HashBlob, boot *evidence.Boot) (eventLogsOut []*eventlog.EventLog, winBootLogs []*eventlog.WinEvents, bootEventLogIdx, curEventLogIdx int, err error) {
	var (
		log1   []eventlog.TPMEvent
		log256 []eventlog.TPMEvent
	)

	for i, logBlob := range eventLogs {
		if len(logBlob.Data) > 0 {
			var eventLog *eventlog.EventLog

			// parse binary log into event log structure
			eventLog, err = eventlog.ParseEventLog(logBlob.Data)
			if err != nil {
				tel.Log(ctx).WithError(err).Infof("error parsing eventlog %d/%d", i, len(eventLogs))

				// parse what is possible and skip the remainder
				continue
			}
			eventLogsOut = append(eventLogsOut, eventLog)

			// parse raw SHA256 events
			rawlog256 := eventLog.Events(eventlog.HashSHA256)
			log256, err = eventlog.ParseEvents(rawlog256)
			if err != nil {
				tel.Log(ctx).WithError(err).Infof("error parsing sha256 eventlogs %d/%d", i, len(eventLogs))
			}

			// parse raw SHA1 events
			rawlog1 := eventLog.Events(eventlog.HashSHA1)
			log1, err = eventlog.ParseEvents(rawlog1)
			if err != nil {
				tel.Log(ctx).WithError(err).Infof("error parsing sha256 eventlogs %d/%d", i, len(eventLogs))
			}

			// compute windows boot state
			var winlog *eventlog.WinEvents
			winlog, err = eventlog.ParseWinEvents(rawlog256)
			if err != nil {
				tel.Log(ctx).WithError(err).Infof("parse windows rawlog256 %d/%d", i, len(eventLogs))
				winlog, err = eventlog.ParseWinEvents(rawlog1)
				if err != nil {
					tel.Log(ctx).WithError(err).Infof("parse windows rawlog1 %d/%d", i, len(eventLogs))
					winlog = nil
				}
			}
			var winColdBoot bool
			if winlog != nil {
				winBootLogs = append(winBootLogs, winlog)
				winColdBoot = winlog.ColdBoot
			}

			// if we only have one log it is either linux and thus a cold boot log or
			// it is windows and we assume it is the cold boot log (technically the agent
			// may only be able to access the most recent log which might be a hibernate/resume log)
			if len(eventLogsOut) == 1 || winColdBoot {
				// compute boot state from potentially manipulated events. manipulation
				// check is separate from parsing it.
				for _, ev := range log256 {
					err = boot.Consume(ctx, ev)
					if err != nil {
						// an error here might indicate manipulated events
						return
					}
				}
				for _, ev := range log1 {
					err = boot.Consume(ctx, ev)
					if err != nil {
						// an error here might indicate manipulated events
						return
					}
				}

				bootEventLogIdx = i
			}

			curEventLogIdx = i
		}
	}

	return
}

func NewSubject(ctx context.Context, values *evidence.Values, bline *baseline.Values, pol *policy.Values, opts ...interface{}) (*Subject, error) {
	var (
		eventLogs       []*eventlog.EventLog
		winBootLogs     []*eventlog.WinEvents
		curEventLogIdx  int
		bootEventLogIdx int
		boot            *evidence.Boot
		report          *binarly.Report
		data            *inteltsc.Data
		certs           []*acert.AttributeCertificate
		blobs           map[string][]byte
	)

	for _, opt := range opts {
		switch opt := opt.(type) {
		case WithBinarly:
			report = opt.Report

		case WithIntelTSC:
			data = opt.Data
			certs = opt.Certificates

		case WithBlobs:
			blobs = opt.Blobs

		default:
			tel.Log(ctx).WithField("option", opt).Error("unknown option")
		}
	}

	// load SMBIOS
	if len(values.SMBIOS.Data) == 0 && len(values.SMBIOS.Sha256) > 0 {
		digest := hex.EncodeToString(values.SMBIOS.Sha256)
		data, ok := blobs[digest]
		if !ok {
			tel.Log(ctx).WithFields(log.Fields{"digest": digest}).Error("smbios blob missing")
		} else {
			values.SMBIOS.Data = data
		}
	}

	// load ACPI tables
	for k := range values.ACPI.Blobs {
		// skip if this was stored inline or if we have no hash
		if len(values.ACPI.Blobs[k].Data) > 0 || len(values.ACPI.Blobs[k].Sha256) == 0 {
			continue
		}
		digest := hex.EncodeToString(values.ACPI.Blobs[k].Sha256)
		data, ok := blobs[digest]
		if !ok {
			tel.Log(ctx).WithFields(log.Fields{"acpi": k, "digest": digest}).Error("blob missing")
			continue
		}
		tmp := values.ACPI.Blobs[k]
		tmp.Data = data
		values.ACPI.Blobs[k] = tmp
	}

	// load TXT
	if len(values.TXTPublicSpace.Data) == 0 && len(values.TXTPublicSpace.Sha256) > 0 {
		digest := hex.EncodeToString(values.TXTPublicSpace.Sha256)
		data, ok := blobs[digest]
		if !ok {
			tel.Log(ctx).WithFields(log.Fields{"digest": digest}).Error("txt blob missing")
		} else {
			values.TXTPublicSpace.Data = data
		}
	}

	// load eventlogs
	for i := range values.TPM2EventLogs {
		// skip if this was stored inline or if we have no hash
		if len(values.TPM2EventLogs[i].Data) > 0 || len(values.TPM2EventLogs[i].Sha256) == 0 {
			continue
		}
		digest := hex.EncodeToString(values.TPM2EventLogs[i].Sha256)
		data, ok := blobs[digest]
		if !ok {
			tel.Log(ctx).WithFields(log.Fields{"eventlog": i, "digest": digest}).Error("blob missing")
			continue
		}
		values.TPM2EventLogs[i].Data = data
	}

	// parse event log
	if boot == nil {
		boot = evidence.EmptyBoot()
		if len(values.TPM2EventLogs) > 0 {
			var err error
			eventLogs, winBootLogs, bootEventLogIdx, curEventLogIdx, err = parseEventLogs(ctx, values.TPM2EventLogs, boot)
			if err != nil {
				if err == evidence.ErrPayload {
					tel.Log(ctx).WithError(err).Infof("eventlog manipulated")
					return nil, err
				}
				tel.Log(ctx).WithError(err).Error("error parsing eventlogs")
				return nil, err
			}
		}
	}

	// linux ima log
	var imaLog []eventlog.TPMEvent
	if values.IMALog != nil && len(values.IMALog.Data) > 0 && values.IMALog.Error == "" {
		logdata, err := unzstd(ctx, []byte(values.IMALog.Data))
		if err != nil {
			tel.Log(ctx).WithError(err).Error("decompress ima log")
			return nil, err
		}

		// log verification is done in check
		imaLog, err = eventlog.ParseIMA(ctx, logdata)
		if err != nil {
			return nil, err
		}

		for _, event := range imaLog {
			if err := boot.Consume(ctx, event); err != nil {
				tel.Log(ctx).WithError(err).WithField("event", event).Error("consume event")
				return nil, err
			}
		}
	}

	// load windows sys and exe files
	antiMalware := make(map[string][]byte)
	for path, rawdigest := range values.AntiMalwareProcesses {
		digest := hex.EncodeToString(rawdigest)
		data, ok := blobs[digest]
		if !ok {
			tel.Log(ctx).WithFields(log.Fields{"path": path, "digest": digest}).Error("blob missing")
			continue
		}
		antiMalware[path] = data
	}
	earlyLaunch := make(map[string][]byte)
	for path, rawdigest := range values.EarlyLaunchDrivers {
		digest := hex.EncodeToString(rawdigest)
		data, ok := blobs[digest]
		if !ok {
			tel.Log(ctx).WithFields(log.Fields{"path": path, "digest": digest}).Error("blob missing")
			continue
		}
		earlyLaunch[path] = data
	}

	// transform fwupd update data
	fwupdDevs := make(map[string]FwupdDevice)
	if values.Devices != nil {
		for releaseDeviceId, releases := range values.Devices.Releases {
			if len(releases) == 0 {
				continue
			}

			// get device for update
			var releaseDevice api.FWUPdDevice
			for i, dev := range values.Devices.Topology {
				tmp, ok := dev["DeviceId"]
				if !ok {
					tel.Log(ctx).Error("no device id, index=", i)
					continue
				}

				devId, ok := tmp.(string)
				if !ok {
					tel.Log(ctx).Error("device id not a string, index=", i)
					continue
				}

				if devId == releaseDeviceId {
					releaseDevice = dev
				}
			}

			if releaseDevice == nil {
				tel.Log(ctx).Error("release for non-existent device ", releaseDeviceId)
				continue
			}

			fwupDev, ok := fwupdDevs[releaseDeviceId]
			if !ok {
				fwupDev = FwupdDevice{}
				fwupDev.Name, _ = releaseDevice["Name"].(string)
				fwupDev.Version, _ = releaseDevice["Version"].(string)
				floatf, _ := releaseDevice["VersionFormat"].(float64)
				fwupDev.VersionFormat = int(floatf)
			}

			for _, release := range releases {
				rv, _ := release["Version"].(string)
				if len(rv) > 0 {
					flags, _ := release["TrustFlags"].(float64)
					fwupDev.Releases = append(fwupDev.Releases, FwupdRelease{rv, uint64(flags)})
				}
			}

			fwupdDevs[releaseDeviceId] = fwupDev
		}
	}

	// load uefi boot apps
	bootApps := make(map[string][]byte)
	for path, rawdigest := range values.BootApps {
		digest := hex.EncodeToString(rawdigest)
		data, ok := blobs[digest]
		if !ok {
			tel.Log(ctx).WithFields(log.Fields{"path": path, "digest": digest}).Error("blob missing")
			continue
		}
		bootApps[path] = data
	}

	subject := Subject{
		Policy:           pol,
		Baseline:         bline,
		BaselineModified: false,

		Values:               values,
		Image:                nil,
		BootEventLogIdx:      bootEventLogIdx,
		CurrentEventLogIdx:   curEventLogIdx,
		EventLogs:            eventLogs,
		IMALog:               imaLog,
		Boot:                 *boot,
		WindowsLogs:          winBootLogs,
		AntiMalwareProcesses: antiMalware,
		EarlyLaunchDrivers:   earlyLaunch,
		BinarlyReport:        report,
		IntelTSCData:         data,
		PlatformCertificates: certs,
		FwupdDevices:         fwupdDevs,
		BootApps:             bootApps,
	}

	return &subject, nil
}
