package report

import (
	"bytes"
	"context"
	"regexp"
	"strings"

	"github.com/digitalocean/go-smbios/smbios"
	"github.com/google/uuid"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/api"
	ev "github.com/immune-gmbh/guard/apisrv/v2/pkg/evidence"
	tel "github.com/immune-gmbh/guard/apisrv/v2/pkg/telemetry"
)

const (
	validHostnameRegex string = `^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])$`
)

var (
	validHostname = regexp.MustCompile(validHostnameRegex)
)

func newAnnotation(path string, id api.AnnotationID) api.Annotation {
	ann := api.NewAnnotation(id)
	ann.Path = path
	return *ann
}

func Compile(ctx context.Context, evidence *ev.Values) (*api.Report, error) {
	var report api.Report

	_, span := tel.Start(ctx, "Analyze firmware")
	defer span.End()

	report.Type = api.ReportType
	report.Annotations = make([]api.Annotation, 0)

	// host
	if err := doHost(&report, evidence); err != nil {
		return nil, err
	}
	// smbios
	if err := doSMBIOS(&report, evidence); err != nil {
		return nil, err
	}
	// nics
	if err := doNICs(ctx, &report, evidence); err != nil {
		return nil, err
	}
	// agent
	if err := doAgent(ctx, &report, evidence); err != nil {
		return nil, err
	}

	return &report, nil
}

func doHost(report *api.Report, evidence *ev.Values) error {
	var vals api.Host

	ann := make([]api.Annotation, 0)

	// name
	if evidence.OS.Release != "" {
		vals.OSName = evidence.OS.Release
	} else {
		ann = append(ann, newAnnotation("/host/name", api.AnnOSType))
	}

	// hostname
	if evidence.OS.Hostname == "" {
		ann = append(ann, newAnnotation("/host/hostname", api.AnnHostname))
	} else if !validHostname.Match([]byte(evidence.OS.Hostname)) {
		ann = append(ann, newAnnotation("/host/hostname", api.AnnHostname))
	} else {
		vals.Hostname = evidence.OS.Hostname
	}

	// type
	if strings.Contains(strings.ToLower(evidence.OS.Release), "windows") {
		vals.OSType = api.OSWindows
	} else if evidence.OS.Release != "" {
		vals.OSType = api.OSLinux
	} else {
		ann = append(ann, newAnnotation("/host/type", api.AnnOSType))
	}

	// cpu_vendor
	label, err := evidence.CPUVendorLabel()
	if err != nil {
		ann = append(ann, newAnnotation("/host/cpu_vendor", api.AnnCPUVendor))
	} else {
		vals.CPUVendor = label
	}

	report.Values.Host = vals
	report.Annotations = append(report.Annotations, ann...)
	return nil
}

func doSMBIOS(report *api.Report, evidence *ev.Values) error {
	var vals api.SMBIOS

	ann := make([]api.Annotation, 0)
	report.Values.SMBIOS = nil

	if evidence.SMBIOS.Error != api.NoError {
		// result has no effect
		//ann = append(ann, newAnnotation("/smbios", api.AnnNoSMBIOS))
		return nil
	}

	if len(evidence.SMBIOS.Data) == 0 {
		// result has no effect
		// ann = append(ann, newAnnotation("/smbios", api.AnnNoSMBIOS))
		return nil
	}

	// Decode SMBIOS structures from the stream.
	d := smbios.NewDecoder(bytes.NewReader([]byte(evidence.SMBIOS.Data)))
	ss, err := d.Decode()
	if err != nil {
		ann = append(ann, newAnnotation("/smbios", api.AnnInvalidSMBIOS))
	}

	foundType0 := false
	foundType1 := false

	for _, s := range ss {
		// SMBIOS 3.1.1 -- Table 6, BIOS Information
		if s.Header.Type == 0 && len(s.Strings) >= 3 {
			vals.BIOSVendor = s.Strings[0]
			vals.BIOSVersion = s.Strings[1]
			vals.BIOSReleaseDate = s.Strings[2]

			if foundType0 {
				ann = append(ann, newAnnotation("/smbios/bios_vendor", api.AnnSMBIOSType0Dup))
				ann = append(ann, newAnnotation("/smbios/bios_version", api.AnnSMBIOSType0Dup))
				ann = append(ann, newAnnotation("/smbios/bios_release_date", api.AnnSMBIOSType0Dup))
			} else {
				foundType0 = true
			}
		}

		// SMBIOS 3.1.1 -- Table 10, System Information
		if s.Header.Type == 1 && len(s.Strings) >= 4 && len(s.Formatted) >= 20 {
			vals.Manufacturer = s.Strings[0]
			vals.Product = s.Strings[1]
			vals.Serial = s.Strings[3]

			id, err := littleEndianUUID(s.Formatted[4:20])
			if err != nil {
				return err
			}
			vals.UUID = id.String()

			if foundType1 {
				ann = append(ann, newAnnotation("/smbios/manufacturer", api.AnnSMBIOSType1Dup))
				ann = append(ann, newAnnotation("/smbios/product", api.AnnSMBIOSType1Dup))
				ann = append(ann, newAnnotation("/smbios/serial", api.AnnSMBIOSType1Dup))
				ann = append(ann, newAnnotation("/smbios/uuid", api.AnnSMBIOSType1Dup))
			} else {
				foundType1 = true
			}
		}
	}

	if !foundType0 {
		ann = append(ann, newAnnotation("/smbios/bios_vendor", api.AnnSMBIOSType0Missing))
		ann = append(ann, newAnnotation("/smbios/bios_version", api.AnnSMBIOSType0Missing))
		ann = append(ann, newAnnotation("/smbios/bios_release_date", api.AnnSMBIOSType0Missing))
	}
	if !foundType1 {
		ann = append(ann, newAnnotation("/smbios/manufacturer", api.AnnSMBIOSType1Missing))
		ann = append(ann, newAnnotation("/smbios/product", api.AnnSMBIOSType1Missing))
	}

	// XXX: system uuid
	// XXX: cross check ME table with reported values
	// XXX: WBPT

	report.Values.SMBIOS = &vals
	report.Annotations = append(report.Annotations, ann...)
	return nil
}

func doNICs(ctx context.Context, report *api.Report, evidence *ev.Values) error {
	if evidence.NICs == nil {
		return nil
	}

	for _, nic := range evidence.NICs.List {
		// we get at minimum a MAC and everything else, including errors, is optional
		nic.Error = api.NoError
		report.Values.NICs = append(report.Values.NICs, nic)
	}

	return nil
}

func doAgent(ctx context.Context, report *api.Report, evidence *ev.Values) error {
	if evidence.Agent == nil {
		return nil
	}

	report.Values.AgentRelease = &evidence.Agent.Release
	return nil
}

// stop-gap for https://github.com/google/uuid/pull/75
func littleEndianUUID(u []byte) (uuid.UUID, error) {
	u[0], u[1], u[2], u[3] = u[3], u[2], u[1], u[0]
	u[4], u[5] = u[5], u[4]
	u[6], u[7] = u[7], u[6]
	return uuid.FromBytes(u)
}
