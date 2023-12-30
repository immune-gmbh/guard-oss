package evidence

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"

	"github.com/digitalocean/go-smbios/smbios"
	"github.com/google/go-tpm/tpm2"
	"github.com/google/uuid"
	"github.com/gowebpki/jcs"
	"github.com/klauspost/compress/zstd"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"

	"github.com/immune-gmbh/agent/v3/pkg/typevisit"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/api"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/eventlog"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/intelme"
	tel "github.com/immune-gmbh/guard/apisrv/v2/pkg/telemetry"
)

var (
	ErrNotFound            = errors.New("not-found")
	ErrinvalidValue        = errors.New("invalid")
	ErrNoResponse          = errors.New("no-resp")
	ErrUnknownEvidenceType = errors.New("unknown evidence type")
)

const ValuesTypeLegacy = "evidence/1"
const ValuesTypeV1 = "values/1"
const ValuesTypeV2 = "values/2"
const ValuesType = "values/3"

var HashBlobVisitor *typevisit.TypeVisitorTree

func init() {
	// construct a type visitor tree for re-use
	tvt, err := typevisit.New(&Values{}, api.HashBlob{}, "")
	if err != nil {
		panic(err)
	}
	HashBlobVisitor = tvt
}

type Values struct {
	Type            string                       `json:"type"`
	UEFIVariables   []api.UEFIVariable           `json:"uefi,omitempty"`
	MSRs            []api.MSR                    `json:"msrs,omitempty"`
	CPUIDLeafs      []api.CPUIDLeaf              `json:"cpuid,omitempty"`
	SEV             []api.SEVCommand             `json:"sev,omitempty"`
	ME              []api.MEClientCommands       `json:"me,omitempty"`
	TPM2Properties  []api.TPM2Property           `json:"tpm2_properties,omitempty"`
	TPM2NVRAM       []api.TPM2NVIndex            `json:"tpm2_nvram,omitempty"`
	PCR             map[string]map[string]string `json:"pcr"`
	PCIConfigSpaces []api.PCIConfigSpace         `json:"pci,omitempty"`
	ACPI            api.ACPITables               `json:"acpi"`
	SMBIOS          api.HashBlob                 `json:"smbios"`
	TXTPublicSpace  api.HashBlob                 `json:"txt"`
	VTdRegisterSet  api.HashBlob                 `json:"vtd"`
	TPM2EventLogs   []api.HashBlob               `json:"event_logs,omitempty"`
	PCPQuoteKeys    map[string]api.Buffer        `json:"PCPQuoteKeys,omitempty"` // windows only
	MACAddresses    api.MACAddresses             `json:"mac"`
	OS              api.OS                       `json:"os"`
	NICs            *api.NICList                 `json:"nic,omitempty"`
	Memory          api.Memory                   `json:"memory"`
	Agent           *api.Agent                   `json:"agent,omitempty"`
	IMALog          *api.ErrorBuffer             `json:"ima_log,omitempty"`
	ESET            *api.ESETConfig              `json:"eset,omitempty"`
	Devices         *api.Devices                 `json:"devices,omitempty"`

	// path -> digest in WindowsExecutable namespace
	AntiMalwareProcesses map[string]api.Buffer `json:"antimalware_processes,omitempty"`
	EarlyLaunchDrivers   map[string]api.Buffer `json:"early_launch_drivers,omitempty"`

	// path -> digest
	BootApps map[string]api.Buffer `json:"boot_apps,omitempty"`
}

type valuesV2 struct {
	Type            string                       `json:"type"`
	UEFIVariables   []api.UEFIVariable           `json:"uefi,omitempty"`
	MSRs            []api.MSR                    `json:"msrs,omitempty"`
	CPUIDLeafs      []api.CPUIDLeaf              `json:"cpuid,omitempty"`
	SEV             []api.SEVCommand             `json:"sev,omitempty"`
	ME              []api.MEClientCommands       `json:"me,omitempty"`
	TPM2Properties  []api.TPM2Property           `json:"tpm2_properties,omitempty"`
	TPM2NVRAM       []api.TPM2NVIndex            `json:"tpm2_nvram,omitempty"`
	PCR             map[string]map[string]string `json:"pcr"`
	PCIConfigSpaces []api.PCIConfigSpace         `json:"pci,omitempty"`
	ACPI            api.ACPITablesV1             `json:"acpi"`
	SMBIOS          api.HashBlob                 `json:"smbios"`
	TXTPublicSpace  api.HashBlob                 `json:"txt"`
	VTdRegisterSet  api.HashBlob                 `json:"vtd"`
	TPM2EventLogZ   *api.ErrorBuffer             `json:"event_log_z"`
	PCPQuoteKeys    map[string]api.Buffer        `json:"PCPQuoteKeys,omitempty"` // windows only
	MACAddresses    api.MACAddresses             `json:"mac"`
	OS              api.OS                       `json:"os"`
	NICs            *api.NICList                 `json:"nic,omitempty"`
	Memory          api.Memory                   `json:"memory"`
	Agent           *api.Agent                   `json:"agent,omitempty"`
	IMALog          *api.ErrorBuffer             `json:"ima_log,omitempty"`
	ESET            *api.ESETConfig              `json:"eset,omitempty"`
	Devices         *api.Devices                 `json:"devices,omitempty"`

	// path -> digest in WindowsExecutable namespace
	AntiMalwareProcesses map[string]api.Buffer `json:"antimalware_processes,omitempty"`
	EarlyLaunchDrivers   map[string]api.Buffer `json:"early_launch_drivers,omitempty"`

	// path -> digest
	BootApps map[string]api.Buffer `json:"boot_apps,omitempty"`
}

type valuesV1 struct {
	Type            string                       `json:"type"`
	UEFIVariables   []api.UEFIVariable           `json:"uefi,omitempty"`
	MSRs            []api.MSR                    `json:"msrs,omitempty"`
	CPUIDLeafs      []api.CPUIDLeaf              `json:"cpuid,omitempty"`
	SEV             []api.SEVCommand             `json:"sev,omitempty"`
	ME              []api.MEClientCommands       `json:"me,omitempty"`
	TPM2Properties  []api.TPM2Property           `json:"tpm2_properties,omitempty"`
	TPM2NVRAM       []api.TPM2NVIndex            `json:"tpm2_nvram,omitempty"`
	PCR             map[string]map[string]string `json:"pcr"`
	PCIConfigSpaces []api.PCIConfigSpace         `json:"pci,omitempty"`
	ACPI            api.ACPITablesV1             `json:"acpi"`
	SMBIOS          api.ErrorBuffer              `json:"smbios"`
	TXTPublicSpace  api.ErrorBuffer              `json:"txt"`
	VTdRegisterSet  api.ErrorBuffer              `json:"vtd"`
	TPM2EventLog    api.ErrorBuffer              `json:"event_log"`
	MACAddresses    api.MACAddresses             `json:"mac"`
	OS              api.OS                       `json:"os"`
	NICs            *api.NICList                 `json:"nic,omitempty"`
	Memory          api.Memory                   `json:"memory"`
	Agent           *api.Agent                   `json:"agent,omitempty"`
}

// migrateACPIBlobs migrates the ACPI tables from ErrorBuffer to HashBlob (applies to valuesV1 and V2).
// all code that works on the tables uses the blobs instead of the tables now,
// so this func must be called whenever a values struct is loaded, that is during
// wrapinsecure but also in FromRow.
func migrateACPIBlobs(tables *api.ACPITablesV1) api.ACPITables {
	if tables == nil {
		return api.ACPITables{}
	}

	newTables := api.ACPITables{Error: tables.Error}
	newTables.Blobs = make(map[string]api.HashBlob, len(tables.Tables))

	// there might be some DB entries that use tables and some that use blobs member
	// it is a remnant from a short time where the blobs member was introduced in V2 values
	// and the tables member was migrated to blobs when it was saved to DB
	// there should not be any entry that has both maps filled, so we can just iterate over both
	for i, d := range tables.Tables {
		newTables.Blobs[i] = api.HashBlob{Data: d}
	}

	for i, d := range tables.Blobs {
		newTables.Blobs[i] = d
	}
	return newTables
}

// unpackHashBlobs unpacks all packed, inline hashblobs
func unpackHashBlobs(values *Values) error {
	decoder, err := zstd.NewReader(nil)
	if err != nil {
		return err
	}
	defer decoder.Close()

	HashBlobVisitor.Visit(values, func(v reflect.Value, opts typevisit.FieldOpts) {
		// we need a special treatment for maps because there are no pointers to map elements
		// and that would prevent us from being able to plug-back the values
		if v.Kind() == reflect.Map {
			mi := v.MapRange()
			for mi.Next() {
				hb := mi.Value().Interface().(api.HashBlob)
				if len(hb.Data) == 0 && len(hb.ZData) > 0 {
					// decompression errors are silently ignored
					hb.Data, err = decoder.DecodeAll(hb.ZData, make([]byte, 0, len(hb.ZData)))
					if err == nil {
						v.SetMapIndex(mi.Key(), reflect.ValueOf(hb))
					}
				}
			}
		} else {
			hb := v.Addr().Interface().(*api.HashBlob)
			if len(hb.Data) == 0 && len(hb.ZData) > 0 {
				// decompression errors are silently ignored
				hb.Data, err = decoder.DecodeAll(hb.ZData, make([]byte, 0, len(hb.ZData)))
				if err != nil {
					hb.Data = nil
				}
			}

		}
	})

	return nil
}

// FromRow converts a row from the database into a Values struct.
// It also migrates the data from the legacy format to the new format.
// When evidence is incoming from the API it is first converted to a values type
// using WrapInsecure and then stored in the database. The persist evidence function
// that stores the evidence in the database will return the just stored row and use
// this function. This means that this function is the single point all evidence runs
// through before being analyzed and that makes this function perfect for migrating
// the data from any legacy format to the current format.
func FromRow(row *Row) (*Values, error) {
	if row.Values.IsNull() {
		return nil, nil
	}

	switch row.Values.Type() {
	case ValuesTypeLegacy:
		var evidence api.Evidence
		if err := row.Values.Decode(&evidence); err != nil {
			return nil, err
		}
		val, err := WrapInsecure(&evidence)
		return val, err

	case ValuesTypeV1:
		var values valuesV1
		err := row.Values.Decode(&values)
		if err != nil {
			return nil, err
		}

		// pass the single eventlog as unpacked HashBlob data without hash
		var eventLogs []api.HashBlob
		if len(values.TPM2EventLog.Data) > 0 || values.TPM2EventLog.Error != "" {
			eventLogs = []api.HashBlob{{Data: values.TPM2EventLog.Data, Error: values.TPM2EventLog.Error}}
		}

		return &Values{
			Type:            ValuesType,
			UEFIVariables:   values.UEFIVariables,
			MSRs:            values.MSRs,
			CPUIDLeafs:      values.CPUIDLeafs,
			SEV:             values.SEV,
			ME:              values.ME,
			TPM2Properties:  values.TPM2Properties,
			TPM2NVRAM:       values.TPM2NVRAM,
			PCR:             values.PCR,
			PCIConfigSpaces: values.PCIConfigSpaces,
			ACPI:            migrateACPIBlobs(&values.ACPI),
			SMBIOS:          api.HashBlob{Data: values.SMBIOS.Data, Error: values.SMBIOS.Error},
			TXTPublicSpace:  api.HashBlob{Data: values.TXTPublicSpace.Data, Error: values.TXTPublicSpace.Error},
			VTdRegisterSet:  api.HashBlob{Data: values.VTdRegisterSet.Data, Error: values.VTdRegisterSet.Error},
			TPM2EventLogs:   eventLogs,
			MACAddresses:    values.MACAddresses,
			OS:              values.OS,
			NICs:            values.NICs,
			Memory:          values.Memory,
			Agent:           values.Agent,
		}, err

	case ValuesTypeV2:
		var values valuesV2
		err := row.Values.Decode(&values)
		if err != nil {
			return nil, err
		}

		// unpack multiple eventlogs and pass them as unpacked HashBlob data without hash
		var eventLogs []api.HashBlob
		if values.TPM2EventLogZ != nil {
			if len(values.TPM2EventLogZ.Data) > 0 {
				tmp, err := eventlog.UnpackTPM2EventLogZ(values.TPM2EventLogZ.Data)
				if err != nil {
					return nil, err
				}

				eventLogs = make([]api.HashBlob, len(tmp))
				for i, e := range tmp {
					eventLogs[i] = api.HashBlob{Data: e, Error: values.TPM2EventLogZ.Error}
				}
			} else if values.TPM2EventLogZ.Error != "" {
				eventLogs = []api.HashBlob{{Error: values.TPM2EventLogZ.Error}}
			}
		}

		valuesNew := Values{
			Type:                 ValuesType,
			UEFIVariables:        values.UEFIVariables,
			MSRs:                 values.MSRs,
			CPUIDLeafs:           values.CPUIDLeafs,
			SEV:                  values.SEV,
			ME:                   values.ME,
			TPM2Properties:       values.TPM2Properties,
			TPM2NVRAM:            values.TPM2NVRAM,
			PCR:                  values.PCR,
			PCIConfigSpaces:      values.PCIConfigSpaces,
			ACPI:                 migrateACPIBlobs(&values.ACPI),
			SMBIOS:               values.SMBIOS,
			TXTPublicSpace:       values.TXTPublicSpace,
			VTdRegisterSet:       values.VTdRegisterSet,
			TPM2EventLogs:        eventLogs,
			MACAddresses:         values.MACAddresses,
			OS:                   values.OS,
			NICs:                 values.NICs,
			Memory:               values.Memory,
			Agent:                values.Agent,
			IMALog:               values.IMALog,
			ESET:                 values.ESET,
			Devices:              values.Devices,
			AntiMalwareProcesses: values.AntiMalwareProcesses,
			EarlyLaunchDrivers:   values.EarlyLaunchDrivers,
			BootApps:             values.BootApps,
		}

		// unpack all packed, inline hashblobs
		err = unpackHashBlobs(&valuesNew)
		return &valuesNew, err

	case ValuesType:
		var values Values
		err := row.Values.Decode(&values)
		if err != nil {
			return nil, err
		}

		// unpack all packed, inline hashblobs
		err = unpackHashBlobs(&values)
		return &values, err

	default:
		return nil, ErrUnknownEvidenceType
	}
}

// WrapInsecure wraps evidence we get via the API into a Values struct.
// It is also called by the test code to load test data from stored evidence json files.
func WrapInsecure(in *api.Evidence) (*Values, error) {
	pcr := make(map[string]map[string]string)
	for bank, v := range in.AllPCRs {
		pcr[bank] = make(map[string]string)
		for k, v := range v {
			pcr[bank][k] = hex.EncodeToString(v)
		}
	}
	bank := in.Algorithm
	if bank == "" {
		bank = fmt.Sprint(int(tpm2.AlgSHA256))
	}
	if len(pcr[bank]) == 0 {
		pcr[bank] = make(map[string]string)
		for k, v := range in.PCRs {
			pcr[bank][k] = hex.EncodeToString(v)
		}
	}

	// handle two old eventlog formats; this is complex code and we should just drop the legacy stuff at some point
	var eventLogs []api.HashBlob

	if len(in.Firmware.TPM2EventLogs) > 0 {
		eventLogs = in.Firmware.TPM2EventLogs
	} else if (in.Firmware.TPM2EventLogZ == nil || (len(in.Firmware.TPM2EventLogZ.Data) == 0 && in.Firmware.TPM2EventLogZ.Error == "")) &&
		(len(in.Firmware.TPM2EventLog.Data) > 0 || in.Firmware.TPM2EventLog.Error != "") {
		// this is the case when we have a single eventlog in the old format
		eventLogs = []api.HashBlob{{Data: in.Firmware.TPM2EventLog.Data, Error: in.Firmware.TPM2EventLog.Error}}
	} else if in.Firmware.TPM2EventLogZ != nil {
		// this is the case when we have multiple eventlogs in the packed format
		if len(in.Firmware.TPM2EventLogZ.Data) > 0 {
			tmp, err := eventlog.UnpackTPM2EventLogZ(in.Firmware.TPM2EventLogZ.Data)
			if err != nil {
				return nil, err
			}

			eventLogs = make([]api.HashBlob, len(tmp))
			for i, e := range tmp {
				eventLogs[i] = api.HashBlob{Data: e, Error: in.Firmware.TPM2EventLogZ.Error}
			}
		} else if in.Firmware.TPM2EventLogZ.Error != "" {
			eventLogs = []api.HashBlob{{Error: in.Firmware.TPM2EventLogZ.Error}}
		}
	}

	vals := Values{
		Type:            ValuesType,
		UEFIVariables:   in.Firmware.UEFIVariables,
		MSRs:            in.Firmware.MSRs,
		CPUIDLeafs:      in.Firmware.CPUIDLeafs,
		SEV:             in.Firmware.SEV,
		ME:              in.Firmware.ME,
		TPM2Properties:  in.Firmware.TPM2Properties,
		TPM2NVRAM:       in.Firmware.TPM2NVRAM,
		PCR:             pcr,
		PCIConfigSpaces: in.Firmware.PCIConfigSpaces,
		ACPI:            migrateACPIBlobs(&in.Firmware.ACPI),
		SMBIOS:          in.Firmware.SMBIOS,
		TXTPublicSpace:  in.Firmware.TXTPublicSpace,
		VTdRegisterSet:  in.Firmware.VTdRegisterSet,
		TPM2EventLogs:   eventLogs,
		PCPQuoteKeys:    in.Firmware.PCPQuoteKeys,
		MACAddresses:    in.Firmware.MACAddresses,
		OS:              in.Firmware.OS,
		NICs:            in.Firmware.NICs,
		Memory:          in.Firmware.Memory,
		Agent:           in.Firmware.Agent,
		IMALog:          in.Firmware.IMALog,
		Devices:         in.Firmware.Devices,
	}

	if in.Firmware.EPPInfo != nil {
		vals.AntiMalwareProcesses = make(map[string]api.Buffer)
		for path, hb := range in.Firmware.EPPInfo.AntimalwareProcesses {
			vals.AntiMalwareProcesses[path] = hb.Sha256
		}
		vals.EarlyLaunchDrivers = make(map[string]api.Buffer)
		for path, hb := range in.Firmware.EPPInfo.EarlyLaunchDrivers {
			vals.EarlyLaunchDrivers[path] = hb.Sha256
		}
		vals.ESET = in.Firmware.EPPInfo.ESET
	}

	if in.Firmware.BootApps != nil && len(in.Firmware.BootApps.Images) > 0 {
		vals.BootApps = make(map[string]api.Buffer)
		for path, hb := range in.Firmware.BootApps.Images {
			vals.BootApps[path] = hb.Sha256
		}
	}

	return &vals, nil
}

func (c *Values) CPUIDLeaf(eax, ecx uint32) (uint32, uint32, uint32, uint32, error) {
	for _, leaf := range c.CPUIDLeafs {
		if leaf.LeafEAX == eax && leaf.LeafECX == ecx {
			if leaf.EAX != nil && leaf.EBX != nil && leaf.ECX != nil && leaf.EDX != nil {
				return *leaf.EAX, *leaf.EBX, *leaf.ECX, *leaf.EDX, nil
			} else {
				return 0, 0, 0, 0, errors.New(string(leaf.Error))
			}
		}
	}

	return 0, 0, 0, 0, ErrNotFound
}

func (c *Values) HasTPM() bool {
	return c.HasNVValues() || c.HasTPMProperties()
}

func (c *Values) HasNVValues() bool {
	for _, p := range c.TPM2NVRAM {
		if (p.Public != nil || p.Value != nil) && p.Error == "" {
			return true
		}
	}

	return false
}

func (c *Values) NVPublic(index uint32) (*api.NVPublic, error) {
	for _, m := range c.TPM2NVRAM {
		if m.Index == index {
			if m.Public != nil {
				return m.Public, nil
			} else {
				return nil, errors.New(string(m.Error))
			}
		}
	}

	return nil, ErrNotFound
}

func (c *Values) NVValue(index uint32) (*api.Buffer, error) {
	for _, m := range c.TPM2NVRAM {
		if m.Index == index {
			if m.Value != nil {
				return m.Value, nil
			} else {
				return nil, errors.New(string(m.Error))
			}
		}
	}

	return nil, ErrNotFound
}

func (c *Values) RangeReserved(start uint64, end uint64) (bool, error) {
	if start > end {
		return false, ErrinvalidValue
	}

	if c.Memory.Error != api.NoError {
		return false, errors.New(string(c.Memory.Error))
	}

	for _, m := range c.Memory.Values {
		left := end <= m.Start
		right := start > m.Start+m.Bytes

		if !left && !right && m.Reserved {
			return true, nil
		}
	}

	return false, nil
}

func (c *Values) ACPITable(name string) (*api.Buffer, error) {
	if c.ACPI.Error != "" {
		return nil, errors.New(string(c.ACPI.Error))
	}
	if m, ok := c.ACPI.Blobs[name]; ok {
		return &m.Data, nil
	}

	return nil, ErrNotFound
}

func (c *Values) MachineSpecReg(msr uint32) (uint64, error) {
	for _, m := range c.MSRs {
		if m.MSR != msr {
			continue
		}
		if len(m.Values) == 0 {
			break
		}
		if m.Error != "" {
			return 0, errors.New(string(m.Error))
		}
		prev := m.Values[0]
		for _, v := range m.Values {
			if prev != v {
				return 0, errors.New("divergent MSR across cores")
			}
		}

		return prev, nil
	}

	return 0, ErrNotFound
}

func (c *Values) HasTPMProperties() bool {
	for _, p := range c.TPM2Properties {
		if p.Value != nil && p.Error == "" {
			return true
		}
	}

	return false
}

func (c *Values) TPMProperty(prop tpm2.TPMProp) (uint32, error) {
	for _, m := range c.TPM2Properties {
		if m.Property == uint32(prop) {
			if m.Value != nil {
				return *m.Value, nil
			} else {
				return 0, errors.New(string(m.Error))
			}
		}
	}

	return 0, ErrNotFound
}

func (c *Values) TPMVendorId() (string, error) {
	venbuf := new(bytes.Buffer)
	norm := transform.NewWriter(venbuf, normalizer())
	vidProps := []tpm2.TPMProp{tpm2.VendorString1, tpm2.VendorString2, tpm2.VendorString3, tpm2.VendorString4}
	err := writeTPMProperties(norm, binary.BigEndian, vidProps, c)
	if err != nil {
		return "", err
	}

	norm.Close()
	return venbuf.String(), nil
}

func (c *Values) SEVResponse(cmd uint32) (*api.Buffer, error) {
	for _, m := range c.SEV {
		if m.Command == cmd {
			if m.Response != nil {
				return m.Response, nil
			} else {
				return nil, errors.New(string(m.Error))
			}
		}
	}

	return nil, ErrNotFound
}

func (c *Values) HasMEIResponses() bool {
	for _, c := range c.ME {
		if c.Error == "" {
			for _, cc := range c.Commands {
				if cc.Error == "" && len(cc.Response) > 0 {
					return true
				}
			}
		}
	}

	return false
}

func (c *Values) MEIResponse(client *uuid.UUID, cmd []byte) (*api.Buffer, error) {
	for _, c := range c.ME {

		if c.GUID != nil {
			a := [16]byte(*c.GUID)
			b := [16]byte(*client)
			if bytes.Equal(a[:], b[:]) {
				for _, m := range c.Commands {
					if bytes.Equal(m.Command, api.Buffer(cmd)) {
						if m.Response != nil {
							return &m.Response, nil
						} else if len(m.Error) > 0 {
							return nil, errors.New(string(m.Error))
						} else {
							return nil, ErrNoResponse
						}
					}
				}
			}
		}
	}

	return nil, ErrNotFound
}

func (c *Values) PCIConfigSpace(bus, device uint16, function uint8) (*api.Buffer, error) {
	for _, cfgSpc := range c.PCIConfigSpaces {
		if cfgSpc.Bus == bus && cfgSpc.Device == device && cfgSpc.Function == function {
			if len(cfgSpc.Value) > 0 {
				return &cfgSpc.Value, nil
			} else {
				return nil, ErrNotFound
			}
		}
	}

	return nil, ErrNotFound
}

func (c *Values) PCIVendorID(cfgSpc *api.Buffer) uint16 {
	if len(*cfgSpc) >= 2 {
		return binary.LittleEndian.Uint16([]byte(*cfgSpc)[0:2])
	} else {
		return 0
	}
}

func (c *Values) PCIDeviceID(cfgSpc *api.Buffer) uint16 {
	if len(*cfgSpc) >= 4 {
		return binary.LittleEndian.Uint16([]byte(*cfgSpc)[2:4])
	} else {
		return 0
	}
}

func (c *Values) HasUEFIVariables() bool {
	notImpl := true
	hasVars := false
	for _, v := range c.UEFIVariables {
		hasVars = hasVars || v.Error == ""

		if v.Error != "" {
			notImpl = notImpl && (v.Error == api.NotImplemented)
		}
	}

	return hasVars || !notImpl
}

func (c *Values) UEFIVariable(name string) (*api.Buffer, error) {
	guid := api.EFIGlobalVariable
	if ok := api.EFIImageSecurityDatabases[name]; ok {
		guid = api.EFIImageSecurityDatabase
	}
	return c.UEFIVendorVariable(guid, name)
}

func (c *Values) UEFIVendorVariable(vendor uuid.UUID, name string) (*api.Buffer, error) {
	for _, v := range c.UEFIVariables {
		if v.Vendor == vendor.String() && v.Name == name {
			if v.Error != "" {
				return nil, errors.New(string(v.Error))
			} else {
				return v.Value, nil
			}
		}
	}

	return nil, ErrNotFound
}

func (c *Values) CSMEVersions(ctx context.Context) ([]int, []int, []int, error) {
	buf, err := c.MEIResponse(&intelme.CSME_MKHIGuid, intelme.EncodeGetFirmwareVersion())
	if err != nil {
		return nil, nil, nil, err
	}

	fwVer, fitc, err := intelme.DecodeGetFirmwareVersion(ctx, []byte(*buf))
	if err != nil {
		return nil, nil, nil, err
	}

	// Active, FITC and Recovery version
	verary := []int{
		int(fwVer.MajorCode),
		int(fwVer.MinorCode),
		int(fwVer.HotFixCode),
		int(fwVer.BuildNumberCode),
	}
	var fitcary []int
	if fitc != nil {
		fitcary = []int{
			int(fitc.MajorFITC),
			int(fitc.MinorFITC),
			int(fitc.HotFixFITC),
			int(fitc.BuildNumberFITC),
		}
	}
	recary := []int{
		int(fwVer.MajorRecovery),
		int(fwVer.MinorRecovery),
		int(fwVer.HotFixRecovery),
		int(fwVer.BuildNumberRecovery),
	}

	return verary, recary, fitcary, nil
}

func (c *Values) SMBIOSPlatformVendor(ctx context.Context) (string, error) {
	d := smbios.NewDecoder(bytes.NewReader([]byte(c.SMBIOS.Data)))
	ss, err := d.Decode()
	if err != nil {
		return "", err
	}

	for _, s := range ss {
		// SMBIOS 3.1.1 -- Table 6, BIOS Information
		if s.Header.Type == 0 {
			// BIOS version, release date
			return s.Strings[0], nil
		}
	}

	return "", err
}

func (c *Values) SMBIOSPlatformVersion(ctx context.Context) (string, string, error) {
	d := smbios.NewDecoder(bytes.NewReader([]byte(c.SMBIOS.Data)))
	ss, err := d.Decode()
	if err != nil {
		return "", "", err
	}

	for _, s := range ss {
		// SMBIOS 3.1.1 -- Table 6, BIOS Information
		if s.Header.Type == 0 {
			// BIOS version, release date
			return s.Strings[1], s.Strings[2], nil
		}
	}

	return "", "", err
}

func EvidenceHashes(ctx context.Context, ev *api.Evidence) ([][]byte, error) {
	_, span := tel.Start(ctx, "evidence.EvidenceHashes")
	defer span.End()

	hashes := make([][]byte, 2)

	// compute fw properties hash w/o ima log
	imaLog := ev.Firmware.IMALog
	ev.Firmware.IMALog = nil
	fwPropsJSON, err := json.Marshal(ev.Firmware)
	ev.Firmware.IMALog = imaLog

	if err != nil {
		return nil, err
	}
	fwPropsJCS, err := jcs.Transform(fwPropsJSON)
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(fwPropsJCS)
	hashes[0] = append([]byte{}, sum[:]...)

	// compute fw properties hash w/ ima log
	fwPropsJSON, err = json.Marshal(ev.Firmware)
	if err != nil {
		return nil, err
	}
	fwPropsJCS, err = jcs.Transform(fwPropsJSON)
	if err != nil {
		return nil, err
	}
	sum = sha256.Sum256(fwPropsJCS)
	hashes[1] = append([]byte{}, sum[:]...)

	return hashes, nil
}

type nullTerminated struct {
	terminated bool
}

func (t nullTerminated) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	if len(dst) < len(src) {
		return 0, 0, transform.ErrShortDst
	}
	if t.terminated {
		return 0, 0, nil
	}
	for i, b := range src {
		if b == 0 {
			return i, i, nil
		}
		dst[i] = b
	}
	return len(src), len(src), nil
}

func (t nullTerminated) Reset() {
}

func normalizer() transform.Transformer {
	isOk := func(r rune) bool {
		return r < 32 || r >= 127
	}
	// The isOk filter is such that there is no need to chain to norm.NFC
	return transform.Chain(norm.NFKD, nullTerminated{}, runes.Remove(runes.Predicate(isOk)))
}

func writeTPMProperties(wr io.Writer, order binary.ByteOrder, tps []tpm2.TPMProp, values *Values) error {
	for _, tp := range tps {
		prop, err := values.TPMProperty(tp)
		if err != nil {
			return err
		}
		err = binary.Write(wr, order, prop)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Values) CPUVendorLabel() (api.CPUVendor, error) {
	_, ebx, ecx, edx, err := c.CPUIDLeaf(0, 0)
	if err != nil {
		return "", err
	} else {
		vendor := new(bytes.Buffer)
		norm := transform.NewWriter(vendor, normalizer())
		binary.Write(norm, binary.LittleEndian, ebx)
		binary.Write(norm, binary.LittleEndian, edx)
		binary.Write(norm, binary.LittleEndian, ecx)
		norm.Close()

		return api.CPUVendor(vendor.String()), nil
	}
}

func (c *Values) IsAMD64() (bool, error) {
	// LME & LMA
	mask := uint64(0b101 << 8)
	msr, err := c.MachineSpecReg(api.MSREFER)
	return (msr & mask) == mask, err
}

func (c *Values) PlatformSerial() (string, string, error) {
	d := smbios.NewDecoder(bytes.NewReader([]byte(c.SMBIOS.Data)))
	ss, err := d.Decode()
	if err != nil {
		return "", "", err
	}

	for _, s := range ss {
		// SMBIOS 3.1.1 -- Table 10, System Information
		if s.Header.Type == 1 {
			if len(s.Strings) < 4 {
				return "", "", ErrNotFound
			}

			// Manufacturer, serial
			return s.Strings[0], s.Strings[3], nil
		}
	}

	return "", "", ErrNotFound
}
