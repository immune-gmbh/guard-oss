package evidence

import (
	"bytes"
	"context"
	"crypto"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/klauspost/compress/zstd"
	log "github.com/sirupsen/logrus"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/api"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/check/baseline"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/eventlog"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/intelme"
	tel "github.com/immune-gmbh/guard/apisrv/v2/pkg/telemetry"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/x509"
)

// ExitBootServices
const (
	PreExitBootServices     = 0
	ExitBootServicesRunning = 1
	ExitBootServicesDone    = 2
)

var (
	bootVarRegexp = regexp.MustCompile(`^Boot\d\d\d\d$`)
	ErrPayload    = errors.New("event payload manipulated")
)

type Separator struct {
	SHA1   []byte
	SHA256 []byte
}

type Boot struct {
	IsEmpty bool

	BootGuardIBB    baseline.Hash
	BootGuardStatus []byte

	Setup            baseline.Hash
	POST             baseline.Hash
	BootVariables    map[string]baseline.Hash
	BootOrder        baseline.Hash
	EmbeddedFirmware map[string]baseline.Hash
	IsLenovo         bool
	IsDell           bool

	// Intel CSME
	CSMEInfo               *intelme.FirmwareInfoEvent
	AMTConfig              *intelme.FirmwareConfigEvent
	CSMEComponentVersions  map[uint8]intelme.ManifestVersionPayload
	CSMEComponentHash      map[uint8][]byte
	CSMESecurityParameters *intelme.SecurityParametersPayload
	CSMEOperationMode      *intelme.OperationMode

	Separators       map[int]Separator
	ExitBootServices int

	BootApplications     map[string]baseline.Hash
	GPT                  baseline.Hash
	PartitionTableHeader *eventlog.EFIPartitionTableHeader
	Partitions           []eventlog.EFIPartition

	// UEFI Secure Boot
	SecureBoot   *byte
	AuditMode    *byte
	DeployedMode *byte
	SetupMode    *byte
	PK           baseline.Hash
	PKParsed     *x509.Certificate
	KEK          baseline.Hash
	KEKParsed    []x509.Certificate
	Db           baseline.Hash
	DbxContents  map[string]bool

	// shim
	MokList  baseline.Hash
	MokListX baseline.Hash

	// GRUB
	LinuxFile    string
	LinuxDigest  baseline.Hash
	LinuxCommand []string
	InitrdFile   string
	InitrdDigest baseline.Hash

	// IMA
	Files         map[string]baseline.Hash
	BootAggregate baseline.Hash
	KexecCmdline  baseline.Hash
}

func EmptyBoot() *Boot {
	return &Boot{
		IsEmpty:               true,
		BootVariables:         make(map[string]baseline.Hash),
		DbxContents:           make(map[string]bool),
		EmbeddedFirmware:      make(map[string]baseline.Hash),
		Separators:            make(map[int]Separator),
		BootApplications:      make(map[string]baseline.Hash),
		CSMEComponentHash:     make(map[uint8][]byte),
		CSMEComponentVersions: make(map[uint8]intelme.ManifestVersionPayload),
		Files:                 make(map[string]baseline.Hash),
	}
}

// used in testing
func BootFromEvidence(ctx context.Context, evidence *api.Evidence) (*Boot, error) {
	boot := EmptyBoot()
	// extract tpm 2.0 log
	if evidence.Firmware.TPM2EventLogZ != nil && len(evidence.Firmware.TPM2EventLogZ.Data) > 0 {
		logs, err := eventlog.UnpackTPM2EventLogZ(evidence.Firmware.TPM2EventLogZ.Data)
		if err != nil {
			return nil, err
		}
		evidence.Firmware.TPM2EventLog.Data = logs[0]
	}
	log, err := eventlog.ParseEventLog([]byte(evidence.Firmware.TPM2EventLog.Data))
	if err != nil {
		return nil, err
	}

	// verify log
	var pcrs []eventlog.PCR
	for alg, bank := range evidence.AllPCRs {
		var hash crypto.Hash
		switch alg {
		case "4":
			hash = crypto.SHA1
		case "11":
			hash = crypto.SHA256
		default:
			return nil, baseline.ErrUnknownBank
		}

		if pcrs, err = eventlog.ConvertPCRAPIBufferMapToEventLogPCRs(bank, hash); err != nil {
			return nil, err
		}
	}
	l, err := log.Verify(pcrs)
	if err != nil {
		return nil, err
	}

	// parse events
	events, err := eventlog.ParseEvents(l)
	if err != nil {
		return nil, err
	}

	// linux ima log
	if evidence.Firmware.IMALog != nil && len(evidence.Firmware.IMALog.Data) > 0 && evidence.Firmware.IMALog.Error == "" {
		d, err := zstd.NewReader(nil)
		if err != nil {
			tel.Log(ctx).WithError(err).Error("create decoder")
			return nil, err
		}

		logdata, err := d.DecodeAll([]byte(evidence.Firmware.IMALog.Data), nil)
		if err != nil {
			tel.Log(ctx).WithError(err).Error("decompress ima log")
			return nil, err
		}

		imalog, err := eventlog.ParseIMA(ctx, logdata)
		if err != nil {
			tel.Log(ctx).WithError(err).Error("parse ima log")
			return nil, err
		}
		events = append(events, imalog...)

		// verify log
		for algo, bank := range evidence.AllPCRs {
			var h hash.Hash
			switch algo {
			case "4":
				h = sha1.New()
			case "11":
				h = sha256.New()
			default:
				tel.Log(ctx).WithField("bank", algo).Error("unknown hash algo")
			}

			strbank := make(map[string]string)
			for k, v := range bank {
				strbank[k] = hex.EncodeToString(v)
			}
			_, err := eventlog.VerifyIMA(imalog, strbank, h)
			if err != nil {
				tel.Log(ctx).WithError(err).Error("verify ima log")
				return nil, err
			}
		}
	}

	for _, event := range events {
		if err := boot.Consume(ctx, event); err != nil {
			tel.Log(ctx).WithError(err).WithField("event", event).Error("consume event")
			return nil, err
		}
	}

	return boot, nil
}

func (boot *Boot) Consume(ctx context.Context, event eventlog.TPMEvent) error {
	boot.IsEmpty = false
	switch event := event.(type) {
	case eventlog.NoActionEvent:
	case eventlog.SeparatorEvent:
		return boot.processSeparator(ctx, event)
	case eventlog.PrebootCertEvent:
	case eventlog.EventTagEvent:
	case eventlog.PostEvent:
	case eventlog.CRTMContentEvent:
		return boot.processCRTMContent(ctx, event)
	case eventlog.CRTMEvent:
		return boot.processCRTMEvent(ctx, event)
	case eventlog.MicrocodeEvent:
	case eventlog.PlatformConfigFlagsEvent:
	case eventlog.TableOfDevicesEvent:
	case eventlog.CompactHashEvent:
		return boot.processDellConfig(ctx, event)
	case eventlog.IPLEvent:
		err := boot.processGrubEvent(ctx, event)
		if err != nil {
			return err
		}
		return boot.processShimEvent(ctx, event)
	case eventlog.IPLPartitionEvent:
	case eventlog.NonHostCodeEvent:
	case eventlog.NonHostConfigEvent:
	case eventlog.NonHostInfoEvent:
		return boot.processCSME(ctx, event)

	case intelme.ManifestVersionEvent:
		return boot.processCSMEManifestVersion(ctx, event)
	case intelme.ExtendManifestEvent:
		return boot.processCSMEExtendManifest(ctx, event)
	case intelme.InitializeManifestEvent:
		return boot.processCSMEInitializeManifest(ctx, event)
	case intelme.OperationModeEvent:
		return boot.processCSMEOperationMode(ctx, event)
	case intelme.SecurityParametersEvent:
		return boot.processCSMESecurityParameters(ctx, event)

	case eventlog.OmitBootDeviceEventsEvent:
	case eventlog.UEFIVariableEvent:
	case eventlog.UEFIVariableDriverConfigEvent:
		err := boot.processUEFISecureBootVariable(ctx, event)
		if err != nil {
			return err
		}
		err = boot.processUEFIConfig(ctx, event)
		if err != nil {
			return err
		}
		return boot.processLenovoConfig(ctx, event)

	case eventlog.UEFIBootVariableEvent:
		return boot.processUEFIBootVariable(ctx, event)
	case eventlog.UEFIVariableAuthorityEvent:
	case eventlog.UEFIImageLoadEvent:
	case eventlog.UEFIBootServicesApplicationEvent:
		return boot.processBootApplications(ctx, event)
	case eventlog.UEFIBootServicesDriverEvent:
	case eventlog.UEFIRuntimeServicesDriverEvent:
	case eventlog.UEFIActionEvent:
		return boot.processExitBootServices(ctx, event)
	case eventlog.UEFIGPTEvent:
		return boot.processGPT(ctx, event)
	case eventlog.UEFIPlatformFirmwareBlobEvent:
		return boot.processOptionROM(ctx, event)
	case eventlog.UEFIHandoffTableEvent:
	case eventlog.OptionROMConfigEvent:
	case eventlog.MicrosoftBootEvent:
	case eventlog.ActionEvent:
	case eventlog.ImaNgEvent:
		return boot.processIMAEvent(ctx, event)
	default:
	}
	return nil
}

func (boot *Boot) pastSeperator(event eventlog.TPMEvent) bool {
	raw := event.RawEvent()
	if sep, ok := boot.Separators[raw.Index]; ok {
		switch raw.Alg {
		case eventlog.HashSHA1:
			return len(sep.SHA1) > 0
		case eventlog.HashSHA256:
			return len(sep.SHA256) > 0
		default:
			return false
		}
	} else {
		return false
	}
}

func (boot *Boot) processIMAEvent(ctx context.Context, event eventlog.ImaNgEvent) error {
	// XXX: check template hash
	raw := event.RawEvent()
	if !bytes.Equal(raw.Digest, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}) {
		h := sha1.New()
		h.Write(event.Data)
		tmpl := h.Sum(nil)
		if len(raw.Digest) > 0 && !bytes.Equal(raw.Digest, tmpl) {
			tel.Log(ctx).WithFields(log.Fields{"log": raw.Digest, "computed": tmpl}).Error("verify ima event")
			return ErrPayload
		}
	}

	switch event.Path {
	case "boot_aggregate":
		h, err := baseline.NewHash(event.FileDigest)
		if err != nil {
			return err
		}
		boot.BootAggregate.UnionWith(&h)
	case "kexec-cmdline":
		// XXX: Hash of the kxec boot args
	default:
		if strings.HasPrefix(event.Path, "/") {
			// files
			h, ok := boot.Files[event.Path]
			if ok {
				h.ReplaceWith(baseline.EventDigest(event))
				boot.Files[event.Path] = h
			} else {
				boot.Files[event.Path] = *baseline.EventDigest(event)
			}
		}
	}

	return nil
}

func (boot *Boot) processCRTMEvent(ctx context.Context, event eventlog.CRTMEvent) error {
	if event.Index != 0 {
		return nil
	}
	if boot.pastSeperator(event) {
		return nil
	}
	// XXX: does not work on Dells SHA-1 bank
	//raw := event.RawEvent()
	//if err := raw.DigestEquals([]byte(raw.Data)); err != nil {
	//	fmt.Println(err)
	//	return ErrPayload
	//}

	//if !boot.BootGuardIBB.IsUnset() {
	//	boot.BootGuardStatus = event.Data
	//}

	return nil
}

func (boot *Boot) processCRTMContent(ctx context.Context, event eventlog.CRTMContentEvent) error {
	if event.Index != 0 {
		return nil
	}
	if boot.pastSeperator(event) {
		return nil
	}
	switch event.Message {
	case "Boot Guard Measured S-CRTM\000":
		boot.BootGuardIBB.UnionWith(baseline.EventDigest(event))
	default:
	}

	return nil
}

func (boot *Boot) processUEFIBootVariable(ctx context.Context, event eventlog.UEFIBootVariableEvent) error {
	if event.Index != 1 {
		return nil
	}
	if boot.pastSeperator(event) {
		return nil
	}
	if event.VariableGUID.String() != api.EFIGlobalVariable.String() {
		return nil
	}
	if !bootVarRegexp.MatchString(event.VariableName) {
		if _, ok := boot.BootVariables[event.VariableName]; ok {
			return nil
		}
		val := baseline.Hash{}
		if v, ok := boot.BootVariables[event.VariableName]; ok {
			val = v
		}
		val.UnionWith(baseline.EventDigest(event))
		boot.BootVariables[event.VariableName] = val
	} else if event.VariableName == "BootOrder" {
		boot.BootOrder.UnionWith(baseline.EventDigest(event))
	}
	return nil
}

func (boot *Boot) processOptionROM(ctx context.Context, event eventlog.UEFIPlatformFirmwareBlobEvent) error {
	if event.Index != 0 {
		return nil
	}
	if boot.pastSeperator(event) {
		return nil
	}
	if event.BlobBase < 0xff00_0000 || event.BlobBase+event.BlobLength >= 0x1_0000_0000 {
		return nil
	}
	addr := fmt.Sprintf("%x", event.BlobBase)
	hash := baseline.Hash{}
	if fw, ok := boot.EmbeddedFirmware[addr]; ok {
		hash = fw
	}
	hash.UnionWith(baseline.EventDigest(event))
	boot.EmbeddedFirmware[addr] = hash

	return nil
}

func (boot *Boot) processCSME(ctx context.Context, event eventlog.NonHostInfoEvent) error {
	if boot.pastSeperator(event) {
		return nil
	}
	raw := event.RawEvent()

	switch event.Index {
	case 0:
		fallthrough
	case 2:
		if info, err := intelme.ParseInfoEvent(event.Data); err == nil {
			if len(raw.Digest) > 0 && raw.DigestEquals(raw.Data) != nil {
				return ErrPayload
			}
			boot.CSMEInfo = info
		} else if csmeEvents, err := intelme.ParseMeasurmentEvent(event); err == nil {
			var alg eventlog.HashAlg
			switch csmeEvents.ERHashAlgorithm {
			case intelme.SHA1Algorithm:
				alg = eventlog.HashSHA1
			case intelme.SHA256Algorithm:
				alg = eventlog.HashSHA256
			case intelme.SHA384Algorithm:
				alg = eventlog.HashSHA384
			}
			er := intelme.ReplayER(alg, false, csmeEvents.Events)

			if len(raw.Digest) > 0 && raw.DigestEquals(er) != nil {
				return ErrPayload
			}
			for _, event := range csmeEvents.Events {
				err = boot.Consume(ctx, event)

				if err != nil {
					return err
				}
			}
		}
	case 1:
		fallthrough
	case 3:
		if cfg, err := intelme.ParseConfigEvent(event.Data); err == nil {
			if len(raw.Digest) > 0 && raw.DigestEquals(raw.Data) != nil {
				return ErrPayload
			}
			boot.AMTConfig = cfg
		}
	}
	return nil
}

func (boot *Boot) processCSMEManifestVersion(ctx context.Context, event intelme.ManifestVersionEvent) error {
	// pre separator
	if boot.pastSeperator(event) {
		return nil
	}
	if event.Index != 0 && event.Index != 2 {
		return nil
	}

	if _, ok := boot.CSMEComponentVersions[event.MeasuredEntityID]; ok {
		return nil
	}
	boot.CSMEComponentVersions[event.MeasuredEntityID] = event.ManifestVersion
	return nil
}

func (boot *Boot) processCSMESecurityParameters(ctx context.Context, event intelme.SecurityParametersEvent) error {
	// pre separator
	if boot.pastSeperator(event) {
		return nil
	}
	if event.Index != 0 && event.Index != 2 {
		return nil
	}
	if boot.CSMESecurityParameters != nil {
		return nil
	}
	boot.CSMESecurityParameters = &event.SecurityParameters
	return nil
}

func (boot *Boot) processCSMEOperationMode(ctx context.Context, event intelme.OperationModeEvent) error {
	// pre separator
	if boot.pastSeperator(event) {
		return nil
	}
	if event.Index != 0 && event.Index != 2 {
		return nil
	}
	if boot.CSMEOperationMode != nil {
		return nil
	}
	boot.CSMEOperationMode = &event.OperationMode
	return nil
}

func (boot *Boot) processCSMEExtendManifest(ctx context.Context, event intelme.ExtendManifestEvent) error {
	// pre separator
	if boot.pastSeperator(event) {
		return nil
	}
	if event.Index != 0 && event.Index != 2 {
		return nil
	}

	if _, ok := boot.CSMEComponentHash[event.MeasuredEntityID]; ok {
		return nil
	}
	boot.CSMEComponentHash[event.MeasuredEntityID] = event.Data
	return nil
}

func (boot *Boot) processCSMEInitializeManifest(ctx context.Context, event intelme.InitializeManifestEvent) error {
	// pre separator
	if boot.pastSeperator(event) {
		return nil
	}
	if event.Index != 0 && event.Index != 2 {
		return nil
	}

	if _, ok := boot.CSMEComponentHash[event.MeasuredEntityID]; ok {
		return nil
	}
	boot.CSMEComponentHash[event.MeasuredEntityID] = event.Data
	return nil
}

func (boot *Boot) processSeparator(ctx context.Context, event eventlog.SeparatorEvent) error {
	if event.Index > 7 {
		return nil
	}
	if boot.pastSeperator(event) {
		return nil
	}
	raw := event.RawEvent()
	if len(raw.Digest) > 0 && raw.DigestEquals(raw.Data) != nil {
		return ErrPayload
	}
	if !bytes.Equal(event.Data, []byte{0, 0, 0, 0}) {
		return nil
	}
	sep := boot.Separators[event.Index]
	switch raw.Alg {
	case eventlog.HashSHA1:
		sep.SHA1 = raw.Data
	case eventlog.HashSHA256:
		sep.SHA256 = raw.Data
	}
	boot.Separators[event.Index] = sep
	return nil
}

func (boot *Boot) processBootApplications(ctx context.Context, event eventlog.UEFIBootServicesApplicationEvent) error {
	if event.Index != 4 {
		return nil
	}
	if !boot.pastSeperator(event) {
		return nil
	}
	hash := baseline.Hash{}
	if app, ok := boot.BootApplications[event.DevicePath]; ok {
		hash = app
	}
	hash.UnionWith(baseline.EventDigest(event))
	boot.BootApplications[event.DevicePath] = hash
	return nil
}

func (boot *Boot) processLenovoConfig(ctx context.Context, event eventlog.UEFIVariableDriverConfigEvent) error {
	if event.Index != 1 {
		return nil
	}
	if boot.pastSeperator(event) {
		return nil
	}

	name := "LenovoSecurityConfig"
	guid := uuid.MustParse("a2c1808f-0d4f-4cc9-a619-d1e641d39d49")
	if event.VariableGUID.String() == guid.String() && event.VariableName == name {
		boot.IsLenovo = true
	}

	return nil
}

func (boot *Boot) processDellConfig(ctx context.Context, event eventlog.CompactHashEvent) error {
	if event.Index != 1 {
		return nil
	}
	if boot.pastSeperator(event) {
		return nil
	}
	magic1 := []byte("Dell Configuration Information 1")
	magic2 := []byte("Dell Configuration Information 2")

	if bytes.Equal(event.Data, magic1) || bytes.Equal(event.Data, magic2) {
		boot.IsDell = true
	}

	return nil
}

func (boot *Boot) processGPT(ctx context.Context, event eventlog.UEFIGPTEvent) error {
	if event.Index != 5 {
		return nil
	}
	// past separator
	if !boot.pastSeperator(event) {
		return nil
	}
	raw := event.RawEvent()
	if len(raw.Digest) > 0 && raw.DigestEquals(raw.Data) != nil {
		return ErrPayload
	}
	boot.GPT.UnionWith(baseline.EventDigest(event))
	boot.Partitions = event.Partitions
	boot.PartitionTableHeader = &event.UEFIPartitionHeader

	return nil
}

func (boot *Boot) processExitBootServices(ctx context.Context, event eventlog.UEFIActionEvent) error {
	if event.Index != 5 {
		return nil
	}
	// past separator
	if !boot.pastSeperator(event) {
		return nil
	}
	raw := event.RawEvent()
	if len(raw.Digest) > 0 && raw.DigestEquals(raw.Data) != nil {
		return ErrPayload
	}

	switch boot.ExitBootServices {
	case PreExitBootServices:
		if event.Message == "Exit Boot Services Invocation" {
			boot.ExitBootServices = ExitBootServicesRunning
		}
	case ExitBootServicesRunning:
		if event.Message == "Exit Boot Services Returned with Success" {
			boot.ExitBootServices = ExitBootServicesDone
		}
	}

	return nil
}

func (boot *Boot) processUEFISecureBootVariable(ctx context.Context, event eventlog.UEFIVariableDriverConfigEvent) error {
	if event.Index != 7 {
		return nil
	}
	// pre separator
	if boot.pastSeperator(event) {
		return nil
	}
	raw := event.RawEvent()
	if len(raw.Digest) > 0 && raw.DigestEquals(raw.Data) != nil {
		return ErrPayload
	}
	if event.VariableGUID.String() == api.EFIImageSecurityDatabase.String() {
		switch event.VariableName {
		case "db":
			boot.Db.UnionWith(baseline.EventDigest(event))
		case "dbx":
			if c, h, err := eventlog.ParseEfiSignatureList(event.VariableData); err == nil {
				boot.DbxContents = make(map[string]bool)
				for _, cert := range c {
					fpr := sha256.Sum256(cert.RawTBSCertificate)
					boot.DbxContents[hex.EncodeToString(fpr[:])] = true
				}
				for _, hash := range h {
					boot.DbxContents[hex.EncodeToString(hash)] = true
				}
			}
		default:
			return nil
		}
	}
	if event.VariableGUID.String() == api.EFIGlobalVariable.String() {
		switch event.VariableName {
		case "SecureBoot":
			if len(event.VariableData) == 1 {
				boot.SecureBoot = &event.VariableData[0]
			}
		case "AuditMode":
			if len(event.VariableData) == 1 {
				boot.AuditMode = &event.VariableData[0]
			}
		case "DeployedMode":
			if len(event.VariableData) == 1 {
				boot.DeployedMode = &event.VariableData[0]
			}
		case "SetupMode":
			if len(event.VariableData) == 1 {
				boot.SetupMode = &event.VariableData[0]
			}
		case "PK":
			if pk, err := x509.ParseCertificate(event.VariableData); err == nil && boot.PKParsed == nil {
				boot.PKParsed = pk
			}
			boot.PK.UnionWith(baseline.EventDigest(event))
		case "KEK":
			if c, _, err := eventlog.ParseEfiSignatureList(event.VariableData); err == nil {
				boot.KEKParsed = append(boot.KEKParsed, c...)
			}
			boot.KEK.UnionWith(baseline.EventDigest(event))
		default:
			return nil
		}
	}

	return nil
}

func (boot *Boot) processUEFIConfig(ctx context.Context, event eventlog.UEFIVariableDriverConfigEvent) error {
	if event.Index != 1 {
		return nil
	}
	// pre separator
	if boot.pastSeperator(event) {
		return nil
	}
	raw := event.RawEvent()
	if len(raw.Digest) > 0 && raw.DigestEquals(raw.Data) != nil {
		return ErrPayload
	}
	if event.VariableGUID.String() != "ec87d643-eba4-4bb5-a1e5-3f3e36b20da9" {
		return nil
	}
	if event.VariableName != "Setup" {
		return nil
	}
	boot.Setup.UnionWith(baseline.EventDigest(event))

	return nil
}

func (boot *Boot) processGrubEvent(ctx context.Context, event eventlog.IPLEvent) error {
	msg := strings.TrimRight(event.Message, "\000")
	switch event.Index {
	case 8:
		cmd := strings.Split(msg, " ")
		if len(cmd) < 2 {
			return nil
		}
		switch cmd[0] {
		case "grub_cmd:":
			switch cmd[1] {
			case "linux":
				if boot.LinuxFile == "" && len(cmd) > 2 {
					boot.LinuxFile = cmd[2]
				}
			case "initrd":
				if boot.InitrdFile == "" && len(cmd) > 2 {
					boot.InitrdFile = cmd[2]
				}
			}
		case "kernel_cmdline:":
			if boot.LinuxCommand == nil {
				boot.LinuxCommand = cmd[3:]
			}
		}
	case 9:
		if boot.LinuxFile != "" && boot.LinuxFile == msg {
			boot.LinuxDigest.UnionWith(baseline.EventDigest(event))
		}
		if boot.InitrdFile != "" && boot.InitrdFile == msg {
			boot.InitrdDigest.UnionWith(baseline.EventDigest(event))
		}
	}
	return nil
}

func (boot *Boot) processShimEvent(ctx context.Context, event eventlog.IPLEvent) error {
	if event.Index != 14 {
		return nil
	}

	switch event.Message {
	case "MokList\000":
		boot.MokList.UnionWith(baseline.EventDigest(event))
	case "MokListX\000":
		boot.MokListX.UnionWith(baseline.EventDigest(event))
	}

	return nil
}

func PrintEvent(a eventlog.TPMEvent) {
	fmt.Printf("%d [%x]: ", a.RawEvent().Index, a.RawEvent().Digest)
	switch a := a.(type) {
	case eventlog.NoActionEvent:
		fmt.Println("NoAction event")
	case eventlog.SeparatorEvent:
		fmt.Printf("separator event: %x\n", a.Data)
	case eventlog.PrebootCertEvent:
		fmt.Println("Preboot event", a)
	case eventlog.EventTagEvent:
		fmt.Println("tag event", a)
	case eventlog.PostEvent:
		var pev eventlog.UEFIPlatformFirmwareBlobEvent
		if err := eventlog.ParseUefiPlatformFirmwareBlobEvent(a.RawEvent().Data, &pev); err == nil {
			fmt.Printf("Post event: %x:%x\n", pev.BlobBase, pev.BlobLength)
		} else {
			fmt.Printf("Post event: %s %x\n", a.Message, a.Message)
		}
	case eventlog.CRTMContentEvent:
		fmt.Println("crtm content event: ", a.Message)
	case eventlog.CRTMEvent:
		fmt.Printf("crtm event: %s, %x\n", a.Message, a.Data)
	case eventlog.MicrocodeEvent:
		fmt.Printf("microcode event: %s, %x\n", a.Message, a.Data)
	case eventlog.PlatformConfigFlagsEvent:
		fmt.Println("platform config event")
	case eventlog.TableOfDevicesEvent:
		fmt.Println("device toc event")
	case eventlog.CompactHashEvent:
		fmt.Printf("compact hashevent: %x %s\n", a.RawEvent().Data, a.RawEvent().Data)
	case eventlog.IPLEvent:
		fmt.Println("ipl event: ", a.Message)
	case eventlog.IPLPartitionEvent:
		fmt.Println("ipl partition event")
	case eventlog.NonHostCodeEvent:
		fmt.Printf("non host code event: %x\n", a.RawEvent().Data)
	case eventlog.NonHostConfigEvent:
		if config, err := intelme.ParseConfigEvent(a.RawEvent().Data); err == nil {
			fmt.Printf("csme config event: %#v\n", config)
		} else {
			fmt.Printf("non host config event: %x\n", a.RawEvent().Data)
		}
	case eventlog.NonHostInfoEvent:
		if info, err := intelme.ParseInfoEvent(a.RawEvent().Data); err == nil {
			fmt.Printf("csme info event: %#v\n", info)
		} else if measure, err := intelme.ParseMeasurmentEvent(a); err == nil {
			fmt.Printf("csme measurement event: %#v\n", measure)
			for _, e := range measure.Events {
				PrintEvent(e)
			}
		} else {
			fmt.Printf("non-host info event: '%x'\n", a.RawEvent().Data)
		}

	case intelme.Event:
		fmt.Println("csme event")
	case intelme.ExtendManifestEvent:
		fmt.Println("csme extend event")
	case intelme.InitializeManifestEvent:
		fmt.Println("csme init event")
	case intelme.ManifestVersionEvent:
		fmt.Println("csme ver event")
	case intelme.SecurityParametersEvent:
		fmt.Println("csme sec event")
	case intelme.OEMEnabledCapabilitiesEvent:
		fmt.Println("csme caps event")
	case intelme.OperationModeEvent:
		fmt.Println("csme mode event")
	case intelme.SKUInformationEvent:
		fmt.Println("csme sku event")

	case eventlog.OmitBootDeviceEventsEvent:
		fmt.Println("omit boot event")
	case eventlog.UEFIVariableEvent:
		fmt.Println("uefi vars event")
	case eventlog.UEFIVariableDriverConfigEvent:
		fmt.Println("uefi var driver event: ", a.VariableName, a.VariableGUID.String())
		//fmt.Println(pev.VariableData)
	case eventlog.UEFIBootVariableEvent:
		fmt.Printf("uefi boot var event. name: %s contents: %x path: %s\n", a.VariableName, a.VariableData, a.DevicePath)
	case eventlog.UEFIVariableAuthorityEvent:
		fmt.Println("uefi authority event: ", a.VariableName)
	case eventlog.UEFIImageLoadEvent:
		fmt.Println("uefi image load event")
	case eventlog.UEFIBootServicesApplicationEvent:
		fmt.Println("uefi boot app event: ", a.DevicePath)
	case eventlog.UEFIBootServicesDriverEvent:
		fmt.Println("uefi boot driver event: ", a.DevicePath)
	case eventlog.UEFIRuntimeServicesDriverEvent:
		fmt.Println("uefi runtime driver event: ", a.DevicePath)
	case eventlog.UEFIActionEvent:
		fmt.Println("uefi action event: ", a.Message)
	case eventlog.UEFIGPTEvent:
		fmt.Printf("gpt event: %d partitions\n", len(a.Partitions))
	case eventlog.UEFIPlatformFirmwareBlobEvent:
		// if BlobBase is < 0xfec0_0000 it's likely a PCI Option ROM
		fmt.Printf("uefi platform fw event: %x:%x\n", a.BlobBase, a.BlobLength)
	case eventlog.UEFIHandoffTableEvent:
		fmt.Println("uefi handoff event")
		for _, tbl := range a.Tables {
			fmt.Printf("Table %x %s\n", tbl.VendorTable, tbl.VendorGUID.String())
		}
	case eventlog.OptionROMConfigEvent:
		fmt.Println("opt rom event")
	case eventlog.MicrosoftBootEvent:
		fmt.Println("ms windows event")
		//ev := a.(eventlog.MicrosoftBootEvent)
		//for _, b := range ev.Events {
		//	switch b.(type) {
		//	case eventlog.MicrosoftStringEvent:
		//		fmt.Println("ms stirng event")
		//	case eventlog.MicrosoftRevocationEvent:
		//		fmt.Println("ms recv event")
		//	case eventlog.MicrosoftDataEvent:
		//		fmt.Println("ms data event")
		//	default:
		//		fmt.Println("unknown ms event", a.RawEvent().Type)
		//	}
		//}
	case eventlog.ActionEvent:
		fmt.Println("action event", a.Message)

	case eventlog.ImaNgEvent:
		fmt.Println("ima event", a)

	default:
		fmt.Println("unknown event", a.RawEvent().Type)
	}
}
