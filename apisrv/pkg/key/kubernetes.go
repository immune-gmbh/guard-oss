package key

import (
	"context"
	"encoding/base64"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	semconv "go.opentelemetry.io/collector/semconv/v1.9.0"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	tel "github.com/immune-gmbh/guard/apisrv/v2/pkg/telemetry"
)

const (
	issuerLabel         = "app.kubernetes.io/instance"
	keyAnnotationPrefix = "immu.ne/public-key-"
)

func parsePodAnnotations(ctx context.Context, pod corev1.Pod) []Key {
	iss := pod.Labels[issuerLabel]
	attr := map[string]string{
		semconv.AttributeK8SPodUID:  string(pod.UID),
		semconv.AttributeK8SPodName: pod.Name,
		"immune.issuer":             iss,
	}
	ctx, span := tel.Start(ctx, "Process pod", tel.WithAttributes{Attributes: attr})
	defer span.End()

	ret := make([]Key, 0)
	for annKey, annVal := range pod.Annotations {
		annFields := log.Fields{"annotation": annKey, "value": annVal}

		if !strings.HasPrefix(annKey, keyAnnotationPrefix) {
			tel.Log(ctx).WithFields(annFields).Trace("unknown prefix")
			continue
		}

		keyName := annKey[len(keyAnnotationPrefix):]
		annFields["key"] = keyName
		buf, err := base64.StdEncoding.DecodeString(annVal)
		if err != nil {
			tel.Log(ctx).WithError(err).WithFields(annFields).Error("decode")
			continue
		}

		key, err := NewKey(iss, buf)
		if err != nil {
			tel.Log(ctx).WithError(err).WithFields(annFields).Error("parse key")
			continue
		}

		annFields["kid"] = key.Kid
		annFields["issuer"] = key.Issuer
		tel.Log(ctx).WithFields(annFields).Trace("accept key")
		ret = append(ret, key)
	}

	return ret
}

func discoverKeys(cli *kubernetes.Clientset, ks *Set, ping chan bool, waitTime time.Duration, ctx context.Context, selector string) error {
	ctx, span := tel.Start(ctx, "Discover keys")
	defer span.End()

	opts := metav1.ListOptions{
		LabelSelector: selector,
	}
	pods, err := cli.CoreV1().Pods("").List(ctx, opts)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("enumerate pods")
		return err
	}

	// iterate all found pods
	newKeyset := []Key{}
	for _, pod := range pods.Items {
		newKeyset = append(newKeyset, parsePodAnnotations(ctx, pod)...)
	}

	// new key set different from existing key set -> replace existing
	if !ks.Equal(&newKeyset) {
		ks.Replace(&newKeyset)

		for _, k := range newKeyset {
			tel.Log(ctx).WithFields(log.Fields{"kid": k.Kid, "issuer": k.Issuer}).Info("new key")
		}
	} else {
		tel.Log(ctx).Trace("no change to key set")
	}
	return nil
}

func kubernetesWatcher(cli *kubernetes.Clientset, ks *Set, ping chan bool, waitTime time.Duration, ctx context.Context, selector string) {
	for {
		discoverKeys(cli, ks, ping, waitTime, ctx, selector)

		select {
		case <-time.After(waitTime):

		case <-ping:

		case <-ctx.Done():
			return
		}
	}
}

func Watcher(ks *Set, selector string) (cancel func(), ping chan bool, err error) {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		return
	}

	// creates the clientset
	cli, err := kubernetes.NewForConfig(config)
	if err != nil {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	ping = make(chan bool)
	go kubernetesWatcher(cli, ks, ping, time.Second*5, ctx, selector)

	return
}
