package immune

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net/url"
	"strconv"
	"time"

	"github.com/dop251/goja"
	"github.com/google/go-tpm/tpm2"
	"github.com/google/jsonapi"
	"github.com/gowebpki/jcs"
	"github.com/sirupsen/logrus"
	"go.k6.io/k6/js/modules"

	"github.com/immune-gmbh/agent/v3/pkg/api"
	"github.com/immune-gmbh/agent/v3/pkg/state"
	"github.com/immune-gmbh/agent/v3/pkg/tcg"
)

func init() {
	modules.Register("k6/x/immune", New())

	logrus.SetLevel(logrus.PanicLevel)
}

type (
	RootModule     struct{}
	ModuleInstance struct {
		vu    modules.VU
		agent *Agent
	}
)

var (
	_ modules.Instance = &ModuleInstance{}
	_ modules.Module   = &RootModule{}
)

func New() *RootModule {
	return &RootModule{}
}

func (*RootModule) NewModuleInstance(vu modules.VU) modules.Instance {
	return &ModuleInstance{
		vu:    vu,
		agent: &Agent{vu: vu},
	}
}

var (
	defaultEndpoint = "https://xxx.xxx.xxx/v2/"
)

type Agent struct {
	vu     modules.VU
	anchor tcg.TrustAnchor
	state  *state.State
	base   *url.URL
}

type CreateKeysArgs struct {
	NameHint        string `js:"nameHint"`
	EnrollmentToken string `js:"enrollmentToken"`
	Configuration   string `js:"configuration"`
}

func (a *Agent) CreateKeys(args CreateKeysArgs) string {
	var one jsonapi.OnePayload
	err := json.Unmarshal([]byte(args.Configuration), &one)
	if err != nil {
		panic(err)
	}
	buf, err := json.Marshal(one.Data.Attributes)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(buf, &a.state.Config)
	if err != nil {
		panic(err)
	}
	ekHandle, ekPub, err := a.anchor.GetEndorsementKey()
	if err != nil {
		panic(err)
	}
	defer ekHandle.Flush(a.anchor)
	a.state.EndorsementKey = api.PublicKey(ekPub)

	ekCert, err := a.anchor.ReadEKCertificate()
	if err != nil {
		a.state.EndorsementCertificate = nil
	} else {
		c := api.Certificate(*ekCert)
		a.state.EndorsementCertificate = &c
	}

	rootHandle, rootPub, err := a.anchor.CreateAndLoadRoot("", "", &a.state.Config.Root.Public)
	if err != nil {
		panic(err)
	}
	defer rootHandle.Flush(a.anchor)

	rootName, err := api.ComputeName(rootPub)
	if err != nil {
		panic(err)
	}
	a.state.Root.Name = rootName

	keyCerts := make(map[string]api.Key)
	a.state.Keys = make(map[string]state.DeviceKeyV3)
	for keyName, keyTmpl := range a.state.Config.Keys {
		keyAuth, err := tcg.GenerateAuthValue()
		if err != nil {
			panic(err)
		}
		key, priv, err := a.anchor.CreateAndCertifyDeviceKey(rootHandle, a.state.Root.Auth, keyTmpl, keyAuth)
		if err != nil {
			panic(err)
		}

		a.state.Keys[keyName] = state.DeviceKeyV3{
			Public:     key.Public,
			Private:    priv,
			Auth:       keyAuth,
			Credential: "",
		}
		keyCerts[keyName] = key
	}

	cookie, err := api.Cookie(rand.Reader)
	if err != nil {
		panic(err)
	}

	enroll := api.Enrollment{
		NameHint:               args.NameHint,
		Cookie:                 cookie,
		EndoresmentCertificate: a.state.EndorsementCertificate,
		EndoresmentKey:         a.state.EndorsementKey,
		Root:                   rootPub,
		Keys:                   keyCerts,
	}

	pdoc, err := jsonapi.Marshal(&enroll)
	if err != nil {
		panic(err)
	}
	doc, ok := pdoc.(*jsonapi.OnePayload)
	if !ok {
		panic(err)
	}
	doc.Data.Type = "enrollment"

	body, err := json.Marshal(doc)
	if err != nil {
		panic(err)
	}

	return string(body)
}

type ActivateCredentialsArgs struct {
	Credentials string `js:"credentials"`
}

func (a *Agent) ActivateCredentials(args ActivateCredentialsArgs) string {
	var many jsonapi.ManyPayload

	err := json.Unmarshal([]byte(args.Credentials), &many)
	if err != nil {
		panic(err)
	}

	var attribs []map[string]interface{} = make([]map[string]interface{}, len(many.Data))
	for i, d := range many.Data {
		attribs[i] = d.Attributes
	}
	buf, err := json.Marshal(attribs)
	if err != nil {
		panic(err)
	}
	var creds []*api.EncryptedCredential
	err = json.Unmarshal(buf, &creds)

	ekHandle, _, err := a.anchor.GetEndorsementKey()
	if err != nil {
		panic(err)
	}
	defer ekHandle.Flush(a.anchor)

	rootHandle, _, err := a.anchor.CreateAndLoadRoot("", "", &a.state.Config.Root.Public)
	if err != nil {
		panic(err)
	}
	defer rootHandle.Flush(a.anchor)

	keyCreds := make(map[string]string)
	for _, encCred := range creds {
		key, ok := a.state.Keys[encCred.Name]
		if !ok {
			panic("unknown key")
		}

		handle, err := a.anchor.LoadDeviceKey(rootHandle, a.state.Root.Auth, key.Public, key.Private)
		if err != nil {
			panic(err)
		}

		cred, err := a.anchor.ActivateDeviceKey(*encCred, "", key.Auth, handle, ekHandle, a.state)
		handle.Flush(a.anchor)
		if err != nil {
			panic(err)
		}

		keyCreds[encCred.Name] = cred
	}

	if len(keyCreds) != len(a.state.Keys) {
		if _, ok := keyCreds["aik"]; !ok {
			panic(fmt.Errorf("no aik credential"))
		}
	}

	for keyName, keyCred := range keyCreds {
		key := a.state.Keys[keyName]
		key.Credential = keyCred
		a.state.Keys[keyName] = key
	}

	return keyCreds["aik"]
}

type QuoteArgs struct {
	Firmware           string `js:"firmware"`
	FailureProbability string `js:"failureProbability"`
}

func (a *Agent) Quote(args QuoteArgs) string {
	var evidence api.Evidence
	err := json.Unmarshal([]byte(args.Firmware), &evidence)
	if err != nil {
		panic(err)
	}
	fwPropsJSON, err := json.Marshal(evidence.Firmware)
	if err != nil {
		panic(err)
	}

	// transform firmware info into json and crypto-safe canonical json representations
	fwPropsJCS, err := jcs.Transform(fwPropsJSON)
	if err != nil {
		panic(err)
	}
	fwPropsHash := sha256.Sum256(fwPropsJCS)

	// read selected PCRs
	pcrValues, err := a.anchor.PCRValues(tpm2.Algorithm(a.state.Config.PCRBank), a.state.Config.PCRs)
	if err != nil {
		panic(err)
	}
	quotedPCR := []int{}
	for k := range pcrValues {
		if i, err := strconv.ParseInt(k, 10, 32); err == nil {
			quotedPCR = append(quotedPCR, int(i))
		}
	}

	// read all PCRs
	//allPCRs, err := a.anchor.AllPCRValues()
	//if err != nil {
	//	panic(err)
	//}
	allPCRs := make(map[string]map[string]api.Buffer)
	allPCRs[fmt.Sprint(a.state.Config.PCRBank)] = pcrValues

	// load Root key
	rootHandle, rootPub, err := a.anchor.CreateAndLoadRoot("", a.state.Root.Auth, &a.state.Config.Root.Public)
	if err != nil {
		panic(err)
	}
	defer rootHandle.Flush(a.anchor)

	// make sure we're on the right TPM
	rootName, err := api.ComputeName(rootPub)
	if err != nil {
		panic(err)
	}

	// check the root name. this will change if the endorsement proof value is changed
	if !api.EqualNames(&rootName, &a.state.Root.Name) {
		panic(errors.New("root name changed"))
	}

	// load AIK
	aik, ok := a.state.Keys["aik"]
	if !ok {
		panic(errors.New("no-aik"))
	}
	aikHandle, err := a.anchor.LoadDeviceKey(rootHandle, a.state.Root.Auth, aik.Public, aik.Private)
	if err != nil {
		panic(err)
	}
	defer aikHandle.Flush(a.anchor)
	rootHandle.Flush(a.anchor)

	// convert used PCR banks to tpm2.Algorithm selection for quote
	var algs []tpm2.Algorithm
	for k, _ := range allPCRs {
		alg, err := strconv.ParseInt(k, 10, 16)
		if err != nil {
			panic(err)
		}
		algs = append(algs, tpm2.Algorithm(alg))
	}

	// generate quote
	quote, sig, err := a.anchor.Quote(aikHandle, aik.Auth, fwPropsHash[:], algs, quotedPCR)
	if err != nil || (sig.ECC == nil && sig.RSA == nil) {
		panic(err)
	}
	aikHandle.Flush(a.anchor)

	cookie, _ := api.Cookie(rand.Reader)
	evidence = api.Evidence{
		Type:      api.EvidenceType,
		Quote:     &quote,
		Signature: &sig,
		Algorithm: strconv.Itoa(int(a.state.Config.PCRBank)),
		PCRs:      pcrValues,
		AllPCRs:   allPCRs,
		Firmware:  evidence.Firmware,
		Cookie:    cookie,
	}

	// encode
	pdoc, err := jsonapi.Marshal(&evidence)
	if err != nil {
		panic(err)
	}
	doc, ok := pdoc.(*jsonapi.OnePayload)
	if !ok {
		panic(err)
	}
	doc.Data.Type = "evidence"

	evidenceJSON, err := json.Marshal(doc)
	if err != nil {
		panic(err)
	}

	return string(evidenceJSON)
}

func (a *Agent) ToString() string {
	return fmt.Sprintf("%#v", a)
}

func (mi *ModuleInstance) Exports() modules.Exports {
	return modules.Exports{
		Named: map[string]interface{}{"Agent": mi.NewAgent},
	}
}

type NewAgentArgs struct {
	Endpoint string
}

func generateCA() (ca string, key string) {
	// private key
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	derkey := x509.MarshalPKCS1PrivateKey(priv)
	key = string(pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: derkey,
	}))

	// ca
	tmpl := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "k6 Agent CA",
		},
		NotBefore: time.Now().Add(-10 * 60 * time.Second),
		NotAfter:  time.Now().Add(10 * 365 * 24 * time.Hour),

		SignatureAlgorithm:    x509.SHA256WithRSAPSS,
		KeyUsage:              x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	derca, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, &priv.PublicKey, priv)
	if err != nil {
		panic(err)
	}
	ca = string(pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: derca,
	}))

	return
}

func (mi *ModuleInstance) NewAgent(call goja.ConstructorCall) *goja.Object {
	anchor, err := tcg.NewSoftwareAnchor()
	if err != nil {
		panic(err)
	}

	// default arguments
	endpoint := defaultEndpoint
	ca, key := generateCA()

	var args map[string]any
	if len(call.Arguments) > 0 {
		args = call.Arguments[0].Export().(map[string]any)
	}
	state := state.NewState()
	if e, ok := args["endpoint"]; ok {
		if e, ok := e.(string); ok {
			endpoint = e
		}
	}
	if e, ok := args["key"]; ok {
		if e, ok := e.(string); ok {
			key = e
		}
	}
	if e, ok := args["ca"]; ok {
		if e, ok := e.(string); ok {
			ca = e
		}
	}

	// EK CA
	stubAnchor := anchor.(*tcg.SoftwareAnchor)
	cablk, _ := pem.Decode([]byte(ca))
	keyblk, _ := pem.Decode([]byte(key))
	if cablk == nil || cablk.Type != "CERTIFICATE" || keyblk == nil || keyblk.Type != "RSA PRIVATE KEY" {
		panic("invalid ca/key args")
	}
	keyparse, err := x509.ParsePKCS1PrivateKey(keyblk.Bytes)
	if err != nil {
		panic(err)
	}
	caparse, err := x509.ParseCertificate(cablk.Bytes)
	if err != nil {
		panic(err)
	}
	fmt.Println(caparse.Subject)
	err = stubAnchor.IssueEKCertificate("k6 agent", caparse, keyparse)
	if err != nil {
		panic(err)
	}

	// API endpoint
	base, err := url.Parse(endpoint)
	if err != nil {
		panic(err)
	}

	agent := &Agent{
		vu:     mi.vu,
		state:  state,
		anchor: anchor,
		base:   base,
	}
	rt := mi.vu.Runtime()

	return rt.ToValue(agent).ToObject(rt)
}
