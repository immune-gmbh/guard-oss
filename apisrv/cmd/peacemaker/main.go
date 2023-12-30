package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"flag"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"time"

	ce "github.com/cloudevents/sdk-go/v2"
	ceclient "github.com/cloudevents/sdk-go/v2/client"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
	log "github.com/sirupsen/logrus"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/api"
	cred "github.com/immune-gmbh/guard/apisrv/v2/pkg/authentication"
)

var (
	flagTarget              string
	flagInitialWait         time.Duration
	flagWait                time.Duration
	flagCompressRevokations bool
	flagExpireAppraisals    bool
	flagReportUsage         bool
	flagPrivateKey          string
	flagHelp                bool

	targetUrl  *url.URL
	privateKey *ecdsa.PrivateKey
	kid        string

	serviceName  string = "peacemaker"
	maximalTries int    = 3
)

func main() {
	flag.StringVar(&flagTarget, "target", "http://localhost:3000/v2/events", "Target service URL")
	flag.StringVar(&flagPrivateKey, "key", "token.key", "Path to private authentication key.")
	flag.DurationVar(&flagInitialWait, "initial-wait", time.Duration(10)*time.Second, "Initial wait.")
	flag.DurationVar(&flagWait, "wait", time.Duration(10)*time.Second, "Wait between trys.")
	flag.BoolVar(&flagCompressRevokations, "revokations", true, "Gabage collect revokations.")
	flag.BoolVar(&flagExpireAppraisals, "expire", true, "Check for expired Appraisals.")
	flag.BoolVar(&flagReportUsage, "usage", false, "Send a usage update to authsrv.")
	flag.BoolVar(&flagHelp, "help", false, "Display this help.")
	flag.Parse()

	if flagHelp {
		flag.Usage()
		os.Exit(0)
	}

	var err error
	targetUrl, err = url.Parse(flagTarget)
	if err != nil {
		panic(err)
	}
	privateKey, _, err = loadPrivateKey(flagPrivateKey)
	if err != nil {
		panic(err)
	}

	os.Exit(run())
}

func run() int {
	data := api.HeartbeatEvent{
		ExpireAppraisals:    flagExpireAppraisals,
		CompressRevocations: flagCompressRevokations,
		ReportUsage:         flagReportUsage,
	}

	log.Infof("Sending Heartbeat to %s", targetUrl)
	log.Infof("ExpireAppraisals    = %t", flagExpireAppraisals)
	log.Infof("CompressRevocations = %t", flagCompressRevokations)
	log.Infof("ReportUsage         = %t", flagReportUsage)

	if flagInitialWait > 0 {
		log.Infof("Waiting %s...", flagInitialWait)
		time.Sleep(flagInitialWait)
	}

	backoff := 10 * time.Second
	for try := 1; true; try += 1 {
		ctx, cancel := context.WithTimeout(context.Background(), backoff)
		defer cancel()

		if send(ctx, data) {
			log.Infof("Heartbeat sent on %d. try", try)
			return 0
		}
		if try > maximalTries {
			break
		}
		backoff <<= 1
		if backoff > time.Second*60 {
			backoff = time.Second * 60
		}
		log.Warnf("%d. try failed. Waiting %s and setting timeout to %s", try, flagWait, backoff)
		time.Sleep(flagWait)
	}

	log.Error("Sending heartbeat failed too often. Giving up.")
	return 1
}

func send(ctx context.Context, data api.HeartbeatEvent) bool {
	ev := ce.NewEvent()
	ev.SetSource(serviceName)
	ev.SetType(api.HeartbeatEventType)
	ev.SetTime(time.Now().UTC())
	ev.SetData(ce.ApplicationJSON, data)

	tok, err := cred.IssueServiceCredential(serviceName, nil, time.Now().UTC(), kid, privateKey)
	if err != nil {
		panic(err)
	}
	client, err := ce.NewClientHTTP(
		cehttp.WithTarget(targetUrl.String()),
		cehttp.WithHeader("Authorization", fmt.Sprintf("Bearer %s", tok)),
	)
	if err != nil {
		panic(err)
	}

	result := client.Send(ctx, ceclient.DefaultIDToUUIDIfNotSet(ctx, ev))
	if !ce.IsACK(result) {
		log.Error(result)
		return false
	}
	return true
}

func loadPrivateKey(path string) (*ecdsa.PrivateKey, string, error) {
	b64, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, "", err
	}
	buf, err := base64.StdEncoding.DecodeString(string(b64))
	if err != nil {
		return nil, "", err
	}

	// must be PKCS#8, DER encoded
	ec, err := x509.ParseECPrivateKey(buf)
	if err != nil {
		return nil, "", err
	}

	// must be a ECDSA key over NIST P-256
	if ec.Curve != elliptic.P256() {
		return nil, "", fmt.Errorf("not a ECDSA key on NIST P-256")
	}

	// This code must be kept in sync with key_discovery.NewKey()

	// Encode public and compute kid
	buf2, err := x509.MarshalPKIXPublicKey(&ec.PublicKey)
	if err != nil {
		return nil, "", err
	}

	cksum := sha256.Sum256(buf2)
	kid := hex.EncodeToString(cksum[0:8])

	return ec, kid, nil
}
