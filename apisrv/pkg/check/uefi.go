package check

import (
	"bytes"
	"context"
	"crypto"
	"crypto/sha256"
	cryptocert "crypto/x509"
	_ "embed"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"reflect"
	"sort"
	"strings"
	"time"

	efi "github.com/foxboron/go-uefi/efi/signature"
	"github.com/linuxboot/fiano/pkg/uefi"
	"github.com/saferwall/pe"
	"golang.org/x/text/encoding/unicode"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/api"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/check/baseline"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/eventlog"
	ev "github.com/immune-gmbh/guard/apisrv/v2/pkg/evidence"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/issuesv1"
	tel "github.com/immune-gmbh/guard/apisrv/v2/pkg/telemetry"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/windows"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/x509"
)

// https://uefi.org/revocationlistfile
// Last update: 11th Mar 2022
var (
	//go:embed dbx-amd64
	dbxAmd64Raw []byte
	//go:embed dbx-x86
	dbxX86Raw []byte
	//go:embed dbx-arm64
	dbxArm64Raw []byte

	dbxAmd64 map[string]bool
	dbxX86   map[string]bool
	dbxArm64 map[string]bool
)

func init() {
	readDbx := func(raw []byte) map[string]bool {
		rd := bytes.NewBuffer(raw)
		_, err := efi.ReadEFIVariableAuthencation2(rd)
		if err != nil {
			panic(err)
		}
		lst, err := ioutil.ReadAll(rd)
		if err != nil {
			panic(err)
		}
		_, hashes, err := eventlog.ParseEfiSignatureList(lst)
		if err != nil {
			panic(err)
		}
		ret := make(map[string]bool, len(hashes))
		for _, h := range hashes {
			ret[fmt.Sprintf("%x", h)] = true
		}
		return ret
	}

	dbxAmd64 = readDbx(dbxAmd64Raw)
	dbxX86 = readDbx(dbxX86Raw)
	dbxArm64 = readDbx(dbxArm64Raw)
}

type uefiBootConfig struct{}

func (uefiBootConfig) String() string {
	return "UEFI boot config"
}

func (uefiBootConfig) Verify(ctx context.Context, subj *Subject) issuesv1.Issue {
	if subj.Boot.IsEmpty || subj.Baseline.BootVariables == nil {
		return nil
	}

	changed, added, removed := fullDiffSets(subj.Baseline.BootVariables, subj.Boot.BootVariables)
	if len(changed) == 0 && len(added) == 0 && len(removed) == 0 {
		return nil
	}

	var iss issuesv1.UefiBootOrder
	iss.Common.Id = issuesv1.UefiBootOrderId
	iss.Common.Aspect = issuesv1.UefiBootOrderAspect
	iss.Common.Incident = true

	// sort keys
	keys := make([]string, len(subj.Baseline.BootVariables))
	i := 0
	for k := range subj.Baseline.BootVariables {
		keys[i] = k
		i++
	}
	sort.Strings(keys)

	for _, key := range keys {
		var before, after *baseline.Hash
		if w, ok := subj.Baseline.BootVariables[key]; ok {
			before = &w
		}
		if w, ok := subj.Boot.BootVariables[key]; ok {
			after = &w
		}

		var v issuesv1.UefiBootOrderVariable
		v.Name = key
		v.Before, v.After = baseline.BeforeAfter(before, after)
		iss.Args.Variables = append(iss.Args.Variables, v)
	}

	return &iss
}

func (uefiBootConfig) Update(ctx context.Context, overrides []string, subj *Subject) {
	allowChange := hasIssue(overrides, issuesv1.UefiBootOrderId)

	for k, v := range subj.Boot.BootVariables {
		if subj.Baseline.BootVariables == nil {
			subj.Baseline.BootVariables = make(map[string]baseline.Hash)
		}
		vv, ok := subj.Baseline.BootVariables[k]
		if !ok || allowChange {
			subj.Baseline.BootVariables[k] = v
			subj.BaselineModified = true
		} else {
			if vv.UnionWith(&v) {
				subj.BaselineModified = true
				subj.Baseline.BootVariables[k] = vv
			}
		}
	}
}

type uefiBootApp struct{}

func (uefiBootApp) String() string {
	return "UEFI boot app"
}

// getPEVerifiedAuthentihashSignerCert returns the signer certificate of a PE file if it is signed and the signature is valid.
func getPEVerifiedAuthentihashSignerCert(pe *pe.File) (*cryptocert.Certificate, error) {
	if !pe.IsSigned || pe.Certificates.Content.Verify() != nil {
		return nil, nil
	}

	// check if signed hash matches computed one; pe.Authentihash()
	// the loop using this function already did pe.Authentihash() and it is called again here -> inefficient
	ok, err := windows.CheckPEAuthentiHashSha256(&pe.Certificates.Content, pe.Authentihash())
	if err != nil || !ok {
		return nil, err
	}

	return pe.Certificates.Content.GetOnlySigner(), nil
}

type findImageState struct {
	sha1   map[[20]byte]*pe.File
	sha256 map[[32]byte]*pe.File
	keys   []string
	i      int
}

// findPEImageByHash finds a PE image by its hash in the boot app images. It returns nil if the image is not found.
// The function caches results in state. It calculates PE image hashes until the right one is found and skips the rest.
func findPEImageByHash(ctx context.Context, state findImageState, measuredAppHash baseline.Hash, bootAppImages map[string][]byte) *pe.File {
	if len(bootAppImages) == 0 {
		return nil
	}

	if state.sha256 != nil {
		if f, ok := state.sha256[*measuredAppHash.Sha256]; ok {
			return f
		}
	} else {
		state.sha256 = make(map[[32]byte]*pe.File)
	}
	if state.sha1 != nil {
		if f, ok := state.sha1[*measuredAppHash.Sha1]; ok {
			return f
		}
	} else {
		state.sha1 = make(map[[20]byte]*pe.File)
	}

	if state.keys == nil {
		state.keys = make([]string, 0, len(bootAppImages))
		for k := range bootAppImages {
			state.keys = append(state.keys, k)
		}
		sort.Strings(state.keys)
	}

	sha256 := crypto.SHA256.New()
	sha1 := crypto.SHA1.New()
	for ; state.i < len(state.keys); state.i++ {
		f, err := windows.Parse(bootAppImages[state.keys[state.i]])
		if err != nil {
			tel.Log(ctx).WithError(err).WithField("hash", measuredAppHash.String()).Error("failed to parse boot app")
			continue
		}

		if !f.IsSigned {
			continue
		}

		sha1.Reset()
		sha256.Reset()
		r := f.AuthentihashExt(sha1, sha256)

		if (measuredAppHash.Sha256 != nil && bytes.Equal(measuredAppHash.Sha256[:], r[1])) ||
			(measuredAppHash.Sha1 != nil && bytes.Equal(measuredAppHash.Sha1[:], r[0])) {
			return f
		}
	}

	return nil
}

// findVerfiableCertByHash finds a PE image by its hash in the boot app images and returns the signer certificate if the image is signed and the signature is valid.
func findVerfiableCertByHash(ctx context.Context, searchState findImageState, measuredAppHash baseline.Hash, bootAppImages map[string][]byte) *cryptocert.Certificate {
	f := findPEImageByHash(ctx, searchState, measuredAppHash, bootAppImages)
	if f == nil {
		return nil
	}

	cert, err := getPEVerifiedAuthentihashSignerCert(f)
	if err != nil {
		return nil
	}

	return cert
}

func (uefiBootApp) Verify(ctx context.Context, subj *Subject) issuesv1.Issue {
	if subj.Boot.IsEmpty || subj.Baseline.BootApplications == nil {
		return nil
	}

	// if we get past this point then boot apps have changed
	changed, added, removed := fullDiffSetsBootApp(subj.Baseline.BootApplications, subj.Boot.BootApplications)
	if len(changed) == 0 && len(added) == 0 && len(removed) == 0 {
		return nil
	}

	// parse certificates in uefi boot apps
	// there are certificate chains but currently we only use the leaf certificate and
	// throw away the remaining ones; the leaf certificate will have been be pinned to a boot-app
	// tofu-style and is used to auto-accept peroperly signed changes to the boot app.
	// in the future we might pin the complete chain and use it to accept changes to the
	// leaf certificates.
	var changed2 bool
	var searchState findImageState
	for _, k := range changed {
		cert := findVerfiableCertByHash(ctx, searchState, subj.Boot.BootApplications[k], subj.BootApps)
		if cert == nil {
			changed2 = true
			continue
		}

		// check the fingerprint and if it matches ignore the change
		fp := sha256.Sum256(cert.Raw)
		if v := subj.Baseline.BootApplications[k].PinnedCertificateFingerprint; v == nil || !bytes.Equal(fp[:], v[:]) {
			changed2 = true
		}
	}

	// re-verify changes after checking the certificates
	if !changed2 && len(added) == 0 && len(removed) == 0 {
		return nil
	}

	// construct an issue skeleton to be filled with data
	var iss issuesv1.UefiBootAppSet
	iss.Common.Id = issuesv1.UefiBootAppSetId
	iss.Common.Aspect = issuesv1.UefiBootAppSetAspect
	iss.Common.Incident = true

	// insert all baseline boot entries and sort them
	var keys []string
	for k := range subj.Baseline.BootApplications {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// merge insert and deduplicate new boot applications
	// this only works when keys array has been sorted before
	for k := range subj.Boot.BootApplications {
		i := sort.SearchStrings(keys, k)
		if i < len(keys) {
			if keys[i] != k {
				keys = append(keys[:i+1], keys[i:]...)
				keys[i] = k
			}
		} else {
			keys = append(keys, k)
		}
	}

	// convert the set of new boot apps to issue data and attach to issue
	for _, key := range keys {
		var before, after *baseline.Hash
		if w, ok := subj.Baseline.BootApplications[key]; ok {
			before = &w.Hash
		}
		if w, ok := subj.Boot.BootApplications[key]; ok {
			after = &w
		}

		var v issuesv1.UefiBootAppSetApp
		v.Path = key
		v.Before, v.After = baseline.BeforeAfter(before, after)
		iss.Args.Apps = append(iss.Args.Apps, v)
	}

	return &iss
}

// hashing and parsing of PE boot apps is done here and in Verify() above,
// which is inefficient; in the future we need to refactor the check engine to
// allow for a single pass over the boot apps. what's keeping us from doing this
// is the fact that when the subject is first created we can't return issues
// like when some certificate chain is broken. but also at that step we haven't
// parsed the eventlog enough to know what boot apps we want to analyze. not all of
// them are relevant for the baseline, as there may be tons of unused boot apps on
// any EFI partition.
func (uefiBootApp) Update(ctx context.Context, overrides []string, subj *Subject) {
	allowChange := hasIssue(overrides, issuesv1.UefiBootAppSetId)
	change := false

	// accept all boot apps as the new baseline; this is the case when the device is new or when overriding
	if allowChange || subj.Baseline.BootApplications == nil {
		subj.Baseline.BootApplications = make(map[string]baseline.BootAppMeasurement)

		// if the boot app is signed, verify the signature and pin the leaf certificate to the boot app in baseline
		// when something doesn't check out we just ignore the boot app cert and no pinning happens
		var searchState findImageState
		for k, v := range subj.Boot.BootApplications {
			m := baseline.BootAppMeasurement{Hash: v}
			cert := findVerfiableCertByHash(ctx, searchState, v, subj.BootApps)
			if cert != nil {
				// add the leaf certificate fingerprint to the baseline
				sum := sha256.Sum256(cert.Raw)
				m.PinnedCertificateFingerprint = &sum
			}

			subj.Baseline.BootApplications[k] = m
		}
		change = true
	} else {
		// this adds new hashes, f.e. the device had sha1 only and now sha256 has been enabled so we have both now
		for k := range subj.Boot.BootApplications {
			if fw, ok := subj.Baseline.BootApplications[k]; ok {
				change = fw.Hash.UnionWith(&fw.Hash) || change
				subj.Baseline.BootApplications[k] = fw
			}
		}

		// check if boot apps have changed and if they have, see if the pinned certificate matches and can verify
		// the new content; if it does, auto-accept the change
		// also pin certificates for unchanged boot apps that had not been pinned before
		changed, _, _ := fullDiffSetsBootApp(subj.Baseline.BootApplications, subj.Boot.BootApplications)
		sort.Strings(changed)
		var searchState findImageState
		for k, v := range subj.Baseline.BootApplications {
			measuredHash := subj.Boot.BootApplications[k]
			cert := findVerfiableCertByHash(ctx, searchState, measuredHash, subj.BootApps)
			if cert == nil {
				continue
			}

			pinnedCertHash := v.PinnedCertificateFingerprint
			if pinnedCertHash != nil {
				// handle changed boot apps with pinned certs here
				if sort.SearchStrings(changed, k) >= len(changed) {
					continue
				}

				// check the fingerprint and if it matches, update the hash
				incomingCertHash := sha256.Sum256(cert.Raw)
				if bytes.Equal(incomingCertHash[:], pinnedCertHash[:]) {
					v.Hash = measuredHash
					subj.Baseline.BootApplications[k] = v
					change = true
				}
			} else {
				// handle unchanged boot apps without pinned certs here
				incomingCertHash := sha256.Sum256(cert.Raw)
				v.PinnedCertificateFingerprint = &incomingCertHash
				subj.Baseline.BootApplications[k] = v
				change = true
			}
		}
	}

	subj.BaselineModified = subj.BaselineModified || change
}

type uefiPartitionTable struct{}

func (uefiPartitionTable) String() string {
	return "GPT disk"
}

func (uefiPartitionTable) Verify(ctx context.Context, subj *Subject) issuesv1.Issue {
	if subj.Boot.IsEmpty {
		return nil
	}

	if subj.Baseline.GPT.IntersectsWith(&subj.Boot.GPT) {
		return nil
	}

	var iss issuesv1.UefiGptChanged
	iss.Common.Id = issuesv1.UefiGptChangedId
	iss.Common.Aspect = issuesv1.UefiGptChangedAspect
	iss.Common.Incident = true
	iss.Args.Guid = subj.Boot.PartitionTableHeader.DiskGUID.String()
	iss.Args.Partitions = make([]issuesv1.UefiGptChangedPartition, len(subj.Boot.Partitions))

	utf16 := unicode.UTF16(unicode.BigEndian, unicode.UseBOM)
	for i, part := range subj.Boot.Partitions {
		buf := bytes.NewBuffer(nil)
		for _, w := range part.PartitionName {
			buf.WriteByte(byte(w >> 8))
			buf.WriteByte(byte(w))
		}
		rawname, err := ioutil.ReadAll(utf16.NewDecoder().Reader(buf))
		if err != nil {
			rawname = nil
		}
		name := strings.Trim(strings.ToValidUTF8(string(rawname), "?"), "\x00")

		iss.Args.Partitions[i].Guid = part.PartitionGUID.String()
		iss.Args.Partitions[i].Name = name
		iss.Args.Partitions[i].Start = fmt.Sprintf("%x", part.FirstLBA)
		iss.Args.Partitions[i].End = fmt.Sprintf("%x", part.LastLBA)
	}

	return &iss
}

func (uefiPartitionTable) Update(ctx context.Context, overrides []string, subj *Subject) {
	allowChange := hasIssue(overrides, issuesv1.UefiGptChangedId)
	change := false

	if allowChange {
		change = change || !reflect.DeepEqual(subj.Baseline.GPT, subj.Boot.GPT)
		subj.Baseline.GPT = subj.Boot.GPT
	} else {
		change = subj.Baseline.GPT.UnionWith(&subj.Boot.GPT) || change
	}

	subj.BaselineModified = subj.BaselineModified || change
}

type uefiSecureBootDisabled struct{}

func (uefiSecureBootDisabled) String() string {
	return "UEFI Secure Boot variable check"
}

func (uefiSecureBootDisabled) Verify(ctx context.Context, subj *Subject) issuesv1.Issue {
	if subj.Boot.IsEmpty {
		return nil
	}

	// Secure Boot on -> off
	nowSB := subj.Boot.SecureBoot != nil && *subj.Boot.SecureBoot == 1
	if !subj.Baseline.SecureBootEnabled || nowSB {
		return nil
	}

	var iss issuesv1.UefiSecureBootVariables
	iss.Common.Id = issuesv1.UefiSecureBootVariablesId
	iss.Common.Aspect = issuesv1.UefiSecureBootVariablesAspect
	iss.Common.Incident = true

	if subj.Boot.SecureBoot != nil {
		iss.Args.SecureBoot = fmt.Sprint(*subj.Boot.SecureBoot)
	}
	if subj.Boot.AuditMode != nil {
		iss.Args.AuditMode = fmt.Sprint(*subj.Boot.AuditMode)
	}
	if subj.Boot.DeployedMode != nil {
		iss.Args.DeployedMode = fmt.Sprint(*subj.Boot.DeployedMode)
	}
	if subj.Boot.SetupMode != nil {
		iss.Args.SetupMode = fmt.Sprint(*subj.Boot.SetupMode)
	}

	return &iss
}

func (uefiSecureBootDisabled) Update(ctx context.Context, overrides []string, subj *Subject) {
	allowOff := hasIssue(overrides, issuesv1.UefiSecureBootVariablesId)
	secureBoot := subj.Boot.SecureBoot != nil && *subj.Boot.SecureBoot == 1

	if allowOff || (secureBoot && !subj.Baseline.SecureBootEnabled) {
		subj.BaselineModified = subj.Baseline.SecureBootEnabled != secureBoot
		subj.Baseline.SecureBootEnabled = secureBoot
	}
}

type uefiSecureBootKeys struct{}

func (uefiSecureBootKeys) String() string {
	return "UEFI Secure Boot keys unmodified"
}

func (uefiSecureBootKeys) Verify(ctx context.Context, subj *Subject) issuesv1.Issue {
	if subj.Boot.IsEmpty {
		return nil
	}

	conv := func(c *x509.Certificate) *issuesv1.UefiSecureBootKeysCertificate {
		if c == nil {
			return nil
		}
		fpr := sha256.Sum256(c.RawTBSCertificate)
		return &issuesv1.UefiSecureBootKeysCertificate{
			Fpr:       hex.EncodeToString(fpr[:]),
			Issuer:    c.Issuer.String(),
			Subject:   c.Subject.String(),
			NotBefore: c.NotBefore.Format(time.RFC3339),
			NotAfter:  c.NotAfter.Format(time.RFC3339),
		}
	}

	var iss issuesv1.UefiSecureBootKeys
	iss.Common.Id = issuesv1.UefiSecureBootKeysId
	iss.Common.Aspect = issuesv1.UefiSecureBootKeysAspect
	iss.Common.Incident = true
	iss.Args.Pk = conv(subj.Boot.PKParsed)
	iss.Args.Kek = make([]*issuesv1.UefiSecureBootKeysCertificate, len(subj.Boot.KEKParsed))

	for i, c := range subj.Boot.KEKParsed {
		iss.Args.Kek[i] = conv(&c)
	}

	// PK or KEK changed
	if !subj.Baseline.PK.IntersectsWith(&subj.Boot.PK) {
		return &iss
	}
	if !subj.Baseline.KEK.IntersectsWith(&subj.Boot.KEK) {
		return &iss
	}

	return nil
}

func (uefiSecureBootKeys) Update(ctx context.Context, overrides []string, subj *Subject) {
	allowKeys := hasIssue(overrides, issuesv1.UefiSecureBootKeysId)
	change := false

	if allowKeys {
		change = change || !reflect.DeepEqual(subj.Baseline.PK, subj.Boot.PK)
		subj.Baseline.PK = subj.Boot.PK
	} else {
		change = subj.Baseline.PK.UnionWith(&subj.Boot.PK) || change
	}
	if allowKeys {
		change = change || !reflect.DeepEqual(subj.Baseline.KEK, subj.Boot.KEK)
		subj.Baseline.KEK = subj.Boot.KEK
	} else {
		change = subj.Baseline.KEK.UnionWith(&subj.Boot.KEK) || change
	}

	subj.BaselineModified = subj.BaselineModified || change
}

type uefiDbx struct{}

func (uefiDbx) String() string {
	return "UEFI dbx unmodified"
}

func (uefiDbx) Verify(ctx context.Context, subj *Subject) issuesv1.Issue {
	if subj.Boot.IsEmpty {
		return nil
	}

	var iss issuesv1.UefiSecureBootDbx
	iss.Common.Id = issuesv1.UefiSecureBootDbxId
	iss.Common.Aspect = issuesv1.UefiSecureBootDbxAspect
	iss.Common.Incident = true

	// entries removed from dbx
	if subj.Baseline.DBXContents != nil && subj.Boot.DbxContents == nil {
		for fpr := range subj.Baseline.DBXContents {
			iss.Args.Fprs = append(iss.Args.Fprs, fpr)
		}
		return &iss
	}
	for k := range subj.Baseline.DBXContents {
		if _, ok := subj.Boot.DbxContents[k]; !ok {
			for fpr := range subj.Baseline.DBXContents {
				iss.Args.Fprs = append(iss.Args.Fprs, fpr)
			}
		}
		if len(iss.Args.Fprs) > 0 {
			return &iss
		}
	}

	return nil
}

func (uefiDbx) Update(ctx context.Context, overrides []string, subj *Subject) {
	allowDbx := hasIssue(overrides, issuesv1.UefiSecureBootDbxId)
	change := false

	if allowDbx || (subj.Baseline.DBXContents == nil && len(subj.Boot.DbxContents) > 0) {
		subj.Baseline.DBXContents = make(map[string]bool)
		change = true
	}
	for k := range subj.Boot.DbxContents {
		_, ok := subj.Baseline.DBXContents[k]
		subj.Baseline.DBXContents[k] = true
		change = change || !ok
	}

	subj.BaselineModified = subj.BaselineModified || change
}

type uefiExitBootServices struct{}

func (uefiExitBootServices) String() string {
	return "Exit Boot Services"
}

func (uefiExitBootServices) Verify(ctx context.Context, subj *Subject) issuesv1.Issue {
	if subj.Boot.IsEmpty {
		return nil
	}

	if subj.Boot.IsLenovo {
		// Lenovo UEFI does not seem to measure ExitBootServices event (X1 Carbon Gen9)
		return nil
	}
	if subj.Boot.ExitBootServices != ev.ExitBootServicesDone && !subj.Baseline.AllowMissingExitBootServices {
		var iss issuesv1.UefiNoExitBootSrv
		iss.Common.Id = issuesv1.UefiNoExitBootSrvId
		iss.Common.Aspect = issuesv1.UefiNoExitBootSrvAspect
		iss.Common.Incident = issuesv1.UefiNoExitBootSrvIncident
		iss.Args.Entered = subj.Boot.ExitBootServices != ev.PreExitBootServices
		return &iss
	}

	return nil
}

func (uefiExitBootServices) Update(ctx context.Context, overrides []string, subj *Subject) {
	allow := hasIssue(overrides, issuesv1.UefiNoExitBootSrvId)

	if allow && !subj.Baseline.AllowMissingExitBootServices {
		subj.Baseline.AllowMissingExitBootServices = true
		subj.BaselineModified = true
	}
}

type uefiSeparators struct{}

func (uefiSeparators) String() string {
	return "Separators"
}

func (uefiSeparators) Verify(ctx context.Context, subj *Subject) issuesv1.Issue {
	if subjectHasDummyTPM(subj) {
		return nil
	}

	var iss issuesv1.UefiBootFailure
	iss.Common.Id = issuesv1.UefiBootFailureId
	iss.Common.Aspect = issuesv1.UefiBootFailureAspect
	iss.Common.Incident = true

	for i := 0; i < 7; i += 1 {
		seps, ok := subj.Boot.Separators[i]
		errord := !subj.Boot.IsEmpty && !ok
		errord = errord || (ok && len(seps.SHA1) > 0 && !bytes.Equal(seps.SHA1, []byte{0, 0, 0, 0}))
		errord = errord || (ok && len(seps.SHA256) > 0 && !bytes.Equal(seps.SHA256, []byte{0, 0, 0, 0}))
		whitelisted := sort.SearchInts(subj.Baseline.AllowBootFailure, i) < len(subj.Baseline.AllowBootFailure)
		if errord && !whitelisted {
			iss.Args.Pcr0 = fmt.Sprintf("%x", subj.Boot.Separators[0].SHA256)
			iss.Args.Pcr1 = fmt.Sprintf("%x", subj.Boot.Separators[1].SHA256)
			iss.Args.Pcr2 = fmt.Sprintf("%x", subj.Boot.Separators[2].SHA256)
			iss.Args.Pcr3 = fmt.Sprintf("%x", subj.Boot.Separators[3].SHA256)
			iss.Args.Pcr4 = fmt.Sprintf("%x", subj.Boot.Separators[4].SHA256)
			iss.Args.Pcr5 = fmt.Sprintf("%x", subj.Boot.Separators[5].SHA256)
			iss.Args.Pcr6 = fmt.Sprintf("%x", subj.Boot.Separators[6].SHA256)
			iss.Args.Pcr7 = fmt.Sprintf("%x", subj.Boot.Separators[7].SHA256)
			return &iss
		}
	}
	return nil
}

func (uefiSeparators) Update(ctx context.Context, overrides []string, subj *Subject) {
	allow := hasIssue(overrides, issuesv1.UefiBootFailureId)
	change := false

	if !allow {
		return
	}
	for i := 0; i < 7; i += 1 {
		if subj.Baseline.AllowBootFailure == nil {
			subj.Baseline.AllowBootFailure = []int{i}
			change = true
		} else if sort.SearchInts(subj.Baseline.AllowBootFailure, i) >= len(subj.Baseline.AllowBootFailure) {
			subj.Baseline.AllowBootFailure = append(subj.Baseline.AllowBootFailure, i)
			change = true
		}
	}

	subj.BaselineModified = subj.BaselineModified || change
}

type uefiOfficialDbx struct{}

func (uefiOfficialDbx) String() string {
	return "Official UEFI dbx"
}

func missingRevokations(dbx map[string]bool, amd64 bool) map[string]bool {
	misses := make(map[string]bool)

	if amd64 {
		for fpr := range dbxAmd64 {
			if hit, ok := dbx[fpr]; !ok || !hit {
				misses[fpr] = true
			}
		}
	} else {
		for fpr := range dbxX86 {
			if hit, ok := dbx[fpr]; !ok || !hit {
				misses[fpr] = true
			}
		}
	}

	return misses
}

func (uefiOfficialDbx) Verify(ctx context.Context, subj *Subject) issuesv1.Issue {
	// dbx misses elements from the official list
	cpu, err := subj.Values.CPUVendorLabel()
	if err != nil {
		return nil
	}
	is64, err := subj.Values.IsAMD64()
	if err != nil {
		return nil
	}
	switch cpu {
	case api.IntelCPU:
		fallthrough
	case api.AMDCPU:
		var iss issuesv1.UefiOfficialDbx
		iss.Common.Id = issuesv1.UefiOfficialDbxId
		iss.Common.Aspect = issuesv1.UefiOfficialDbxAspect
		iss.Common.Incident = issuesv1.UefiOfficialDbxIncident
		misses := missingRevokations(subj.Boot.DbxContents, is64)

		for fpr := range misses {
			if sort.SearchStrings(subj.Baseline.RevokedKeyWhitelist, fpr) >= len(subj.Baseline.RevokedKeyWhitelist) {
				iss.Args.Fprs = append(iss.Args.Fprs, fpr)
			}
		}
		if len(iss.Args.Fprs) > 0 {
			return &iss
		}

	default:
		// XXX: ARM64
	}

	return nil
}

func (uefiOfficialDbx) Update(ctx context.Context, overrides []string, subj *Subject) {
	allow := hasIssue(overrides, issuesv1.UefiOfficialDbxId)

	if !allow {
		return
	}
	cpu, err := subj.Values.CPUVendorLabel()
	if err != nil {
		return
	}
	is64, err := subj.Values.IsAMD64()
	if err != nil {
		return
	}
	change := false
	switch cpu {
	case api.IntelCPU:
		fallthrough
	case api.AMDCPU:
		misses := missingRevokations(subj.Boot.DbxContents, is64)
		for fpr := range misses {
			if subj.Baseline.RevokedKeyWhitelist == nil {
				subj.Baseline.RevokedKeyWhitelist = make([]string, 0)
				change = true
			}

			if sort.SearchStrings(subj.Baseline.RevokedKeyWhitelist, fpr) >= len(subj.Baseline.RevokedKeyWhitelist) {
				change = true
				subj.Baseline.RevokedKeyWhitelist = append(subj.Baseline.RevokedKeyWhitelist, fpr)
				sort.Strings(subj.Baseline.RevokedKeyWhitelist)
			}
		}

	default:
		// XXX: ARM64
	}

	subj.BaselineModified = subj.BaselineModified || change
}

type uefiEmbeddedFirmware struct{}

func (uefiEmbeddedFirmware) String() string {
	return "UEFI embedded firmware"
}

func fullDiffSets(before map[string]baseline.Hash, after map[string]baseline.Hash) ([]string, []string, []string) {
	added := []string{}
	removed := []string{}
	changed := []string{}

	keys := make(map[string]bool)
	for k := range before {
		keys[k] = true
	}
	for k := range after {
		keys[k] = true
	}

	for k := range keys {
		b, beforeOk := before[k]
		a, afterOk := after[k]

		if beforeOk && afterOk {
			if !b.IntersectsWith(&a) {
				changed = append(changed, k)
			}
		} else if beforeOk && !afterOk {
			removed = append(removed, k)
		} else if !beforeOk && afterOk {
			added = append(added, k)
		}
	}

	return changed, added, removed
}

// XXX unify the functions somehow by using a generic parameter and a callback to handle the type
func fullDiffSetsBootApp(before map[string]baseline.BootAppMeasurement, after map[string]baseline.Hash) ([]string, []string, []string) {
	added := []string{}
	removed := []string{}
	changed := []string{}

	keys := make(map[string]bool)
	for k := range before {
		keys[k] = true
	}
	for k := range after {
		keys[k] = true
	}

	for k := range keys {
		b, beforeOk := before[k]
		a, afterOk := after[k]

		if beforeOk && afterOk {
			if !b.Hash.IntersectsWith(&a) {
				changed = append(changed, k)
			}
		} else if beforeOk && !afterOk {
			removed = append(removed, k)
		} else if !beforeOk && afterOk {
			added = append(added, k)
		}
	}

	return changed, added, removed
}

type visitor struct {
	indent int
}

func (v *visitor) Run(f uefi.Firmware) error {
	return v.Visit(f)
}

func (v *visitor) Visit(f uefi.Firmware) error {
	pad := ""
	for i := 0; i < v.indent; i += 1 {
		pad += "  "
	}

	switch f := f.(type) {

	case *uefi.File:
		fmt.Printf("%sfile: %s\n", pad, f.Type)
		v.indent += 1
		err := f.ApplyChildren(v)
		v.indent -= 1
		return err

	case *uefi.Section:
		fmt.Printf("%ssection: '%s' %s\n", pad, f.String(), f.Type)
		v.indent += 1
		err := f.ApplyChildren(v)
		v.indent -= 1
		return err

	default:
		fmt.Printf("%sunknown\n", pad)
		v.indent += 1
		err := f.ApplyChildren(v)
		v.indent -= 1
		return err
	}
}

func (uefiEmbeddedFirmware) Verify(ctx context.Context, subj *Subject) issuesv1.Issue {
	if subj.Boot.IsEmpty || subj.Baseline.EmbeddedFirmware == nil {
		return nil
	}

	//for k, v := range baseline.OptionROMs {
	//	fmt.Printf("baseline: %s: %s\n", k, v.String())
	//}
	//decoder, err := zstd.NewReader(bytes.NewBuffer(evidence.Firmware.Flash.Data))
	//if err != nil {
	//	tel.Log(ctx).WithError(err).Error("create decompressor")
	//	return nil
	//}
	//image, err := ioutil.ReadAll(decoder)
	//if err != nil {
	//	tel.Log(ctx).WithError(err).Error("decompress image")
	//	return nil
	//}
	//fmt.Println("image ", len(image))
	//for k, v := range boot.EmbeddedFirmware {
	//	fmt.Printf("boot: %s: %#v\n", k, v)
	//	addr, err := strconv.ParseUint(k, 16, 32)
	//	if err != nil {
	//		tel.Log(ctx).WithError(err).Error("decode base address")
	//		continue
	//	}
	//	fw := image[v.Offset : v.Offset+v.Length]
	//	if !v.Hash.CompareDigest(fw) {
	//		fmt.Println("hash wrong")
	//	} else {
	//		fmt.Println("hash good")
	//	}
	//	fv, err := uefi.NewFirmwareVolume(fw, uint64(v.Offset), false)
	//	if err != nil {
	//		tel.Log(ctx).WithError(err).Error("decode fv")
	//		continue
	//	}
	//	fmt.Println(fv.String())
	//	fv.Apply(&visitor{})
	//}

	update := UEFIUpdated(ctx, subj.Baseline, subj.Values)
	changed, _, _ := fullDiffSets(subj.Baseline.EmbeddedFirmware, subj.Boot.EmbeddedFirmware)
	if len(changed) > 0 && !update {
		tel.Log(ctx).WithField("changed", changed).Info("embedded fw")
		var iss issuesv1.UefiOptionRomSet
		iss.Common.Id = issuesv1.UefiOptionRomSetId
		iss.Common.Aspect = issuesv1.UefiOptionRomSetAspect
		iss.Common.Incident = true

		for _, f := range changed {
			b := subj.Baseline.EmbeddedFirmware[f]
			a := subj.Boot.EmbeddedFirmware[f]
			before, after := baseline.BeforeAfter(&b, &a)
			iss.Args.Devices = append(iss.Args.Devices, issuesv1.UefiOptionRomSetDevice{
				Address: f,
				Before:  before,
				After:   after,
				Name:    "",
				Vendor:  "",
			})
		}
		return &iss
	}

	return nil
}

func (uefiEmbeddedFirmware) Update(ctx context.Context, overrides []string, subj *Subject) {
	allowChange := hasIssue(overrides, issuesv1.UefiOptionRomSetId) || UEFIUpdated(ctx, subj.Baseline, subj.Values)
	change := false

	if allowChange || subj.Baseline.EmbeddedFirmware == nil {
		subj.Baseline.EmbeddedFirmware = make(map[string]baseline.Hash)
		for k, v := range subj.Boot.EmbeddedFirmware {
			subj.Baseline.EmbeddedFirmware[k] = v
		}
		change = true
	} else {
		for k, v := range subj.Boot.EmbeddedFirmware {
			if fw, ok := subj.Baseline.EmbeddedFirmware[k]; ok {
				change = fw.UnionWith(&v) || change
				subj.Baseline.EmbeddedFirmware[k] = fw
			}
		}
	}

	subj.BaselineModified = subj.BaselineModified || change
}
