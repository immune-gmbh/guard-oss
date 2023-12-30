package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	_ "net/http/pprof"

	"github.com/jackc/pgx/v4"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/api"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/blob"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/configuration"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/database"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/event"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/inteltsc"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/key"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/queue"
	tel "github.com/immune-gmbh/guard/apisrv/v2/pkg/telemetry"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/web"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/workflow"
)

var releaseId string = "unknown"

func run() {
	var exit bool
	var keyset *key.Set

	exit = false
	keyset = key.NewSet()

	signalChan := notifyUnixSignals()

	ctx := context.Background()

	for !exit {
		var err error

		// sanity check
		if viper.GetInt("database.pool_size") < viper.GetInt("events.worker_pool") {
			log.Warnln("More event workers than database connections in pool")
		}

		// telemetry
		var flushTraceData func()
		if viper.IsSet("telemetry.jaeger") {
			flushTraceData, err = tel.Setup(
				viper.GetString("telemetry.service"), releaseId,
				tel.WithJaeger{Endpoint: viper.GetString("telemetry.jaeger")},
				tel.WithLog{Stdout: viper.GetBool("telemetry.console_log")})
			if err != nil {
				log.Fatalf("Telemetry setup failed: %s", err)
			}
		} else if viper.IsSet("telemetry.lightstep") {
			flushTraceData, err = tel.Setup(
				viper.GetString("telemetry.service"), releaseId,
				tel.WithOpenTelemtry{
					Endpoint: viper.GetString("telemetry.lightstep"),
					Token:    viper.GetString("telemetry.token"),
					TLS:      !viper.GetBool("telemetry.disable_tls")},
				tel.WithLog{Stdout: viper.GetBool("telemetry.console_log")})
			if err != nil {
				log.Fatalf("Telemetry setup failed: %s", err)
			}
		} else {
			log.Warnln("Telemetry disabled")
			_, err = tel.Setup(
				viper.GetString("telemetry.service"), releaseId,
				tel.WithLog{Stdout: viper.GetBool("telemetry.console_log")})
			if err != nil {
				log.Fatalf("Telemetry setup failed: %s", err)
			}
			flushTraceData = func() {}
		}

		// object storage
		store, err := blob.NewStorage(ctx,
			blob.WithBucket{Bucket: viper.GetString("bucket.name")},
			blob.WithEndpoint{
				Endpoint: viper.GetString("bucket.endpoint"),
				Region:   viper.GetString("bucket.region"),
			},
			blob.WithCredentials{
				Key:    viper.GetString("bucket.api_key"),
				Secret: viper.GetString("bucket.api_secret"),
			})
		if err != nil {
			log.Fatalf("Blob storage setup failed: %s", err)
		}

		// sql database
		database, err := database.Connect(
			viper.GetString("database.host"),
			uint16(viper.GetUint("database.port")),
			viper.GetString("database.user"),
			viper.GetString("database.password"),
			viper.GetString("database.name"),
			viper.GetString("database.certificate"),
			viper.GetString("telemetry.lightstep") != "",
			viper.GetBool("debug"),
			viper.GetInt("database.pool_size"))
		if err != nil {
			log.Fatalf("Database connection setup failed: %s", err)
		}

		// authentication key
		var authPrivKey *ecdsa.PrivateKey
		var authKey key.Key
		if viper.IsSet("keys.authentication") {
			authPrivKey, _, err = loadPrivateKey(viper.GetString("keys.authentication"))
			if err != nil {
				log.Fatalf("Loading authentication keypair failed: %s", err)
			}
		} else {
			log.Warnln("Authentication keypair not set. Generating a temporary one.")
			authPrivKey, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
			if err != nil {
				log.Fatalf("Generating temporary keypair failed: %s", err)
			}
		}
		if pkix, err := x509.MarshalPKIXPublicKey(&authPrivKey.PublicKey); err != nil {
			log.Fatalf("Generating temporary keypair failed: %s", err)
		} else {
			authKey, err = key.NewKey(viper.GetString("keys.service_name"), pkix)
			if err != nil {
				log.Fatalf("Generating temporary keypair failed: %s", err)
			}
		}

		// event queue
		eventp, err := event.NewProcessor(ctx,
			viper.GetString("events.receiver"),
			viper.GetString("keys.service_name"),
			event.WithCredentials{PrivateKey: authPrivKey, Kid: authKey.Kid})
		if err != nil {
			log.Fatalf("Event processor setup failed: %s\n", err)
		}

		// intel TSC API client
		tscsites := make([]inteltsc.Site, 0)
		err = viper.UnmarshalKey("inteltsc", &tscsites)
		if err != nil {
			log.Fatalf("Intel TSC API client setup failed: %s\n", err)
		}
		for i, site := range tscsites {
			if site.Password == "" {
				env, ok := os.LookupEnv(fmt.Sprintf("APISRV_INTELTSC_%s_PASSWORD", strings.ToUpper(site.Vendor)))
				if ok {
					tscsites[i].Password = env
				}
			}
			if site.ClientSecret == "" {
				env, ok := os.LookupEnv(fmt.Sprintf("APISRV_INTELTSC_%s_SECRET", strings.ToUpper(site.Vendor)))
				if ok {
					tscsites[i].ClientSecret = env
				}
			}
		}
		var tscp queue.Processor
		if len(tscsites) > 0 {
			tscp, err = inteltsc.NewProcessor(ctx, database, tscsites)
			if err != nil {
				log.Fatalf("Intel TSC API client setup failed: %s\n", err)
			}
		} else {
			log.Warn("Disable Intel TSC")
			workflow.DisableIntelTSC = true
		}

		// background job queue
		queue, err := queue.New(ctx, database,
			queue.WithWorkerPool{
				NumWorkers:   viper.GetInt("events.worker_pool"),
				PollInterval: time.Second,
			},
			queue.WithProcessor{Processor: eventp},
			queue.WithProcessor{Processor: binarlyp},
			queue.WithProcessor{Processor: tscp},
			queue.WithObserver{Fn: func(ctx context.Context, ty string, ref string) {
				workflow.Appraise(ctx, database, store, ref, viper.GetString("keys.service_name"), time.Now(), viper.GetDuration("queue.job_timeout"))
			}})
		if err != nil {
			log.Fatalf("Event queue setup failed: %s\n", err)
		}

		qctx, stopQueue := context.WithCancel(ctx)
		go func() {
			err := queue.Start(qctx)
			if err != nil {
				log.Fatalf("Event queue start failed: %s\n", err)
			}
		}()

		// appraisal signing key
		var caPrivKey *ecdsa.PrivateKey
		var caKey key.Key
		if viper.IsSet("keys.attestation") {
			caPrivKey, _, err = loadPrivateKey(viper.GetString("keys.attestation"))
			if err != nil {
				log.Fatalf("Loading appraisal signing keypair failed: %s", err)
			}
		} else {
			log.Warnln("Device appraisal signing keypair not set. Generating a temporary one.")
			caPrivKey, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
			if err != nil {
				log.Fatalf("Generating temporary keypair failed: %s", err)
			}
		}
		if pkix, err := x509.MarshalPKIXPublicKey(&caPrivKey.PublicKey); err != nil {
			log.Fatalf("Generating temporary keypair failed: %s", err)
		} else {
			caKey, err = key.NewKey(viper.GetString("keys.service_name"), pkix)
			if err != nil {
				log.Fatalf("Generating temporary keypair failed: %s", err)
			}
		}
		keys := []key.Key{caKey}

		// key discovery
		var stopKeyDiscovery func()
		var pingChan chan bool
		if viper.GetBool("keys.discover") {
			stopKeyDiscovery, pingChan, err = key.Watcher(keyset, viper.GetString("keys.label_selector"))
		} else {
			log.Warnln("Key discovery disabled")

			for iss, val := range viper.GetStringMap("keys.static") {
				if b64, ok := val.(string); ok {
					k, err := key.NewKeyBase64(iss, b64)

					if err != nil {
						log.Errorf("Failed to load key for %s: %s", iss, err)
						continue
					}
					keys = append(keys, k)
				} else {
					log.Errorf("Failed to load key for %s: not a string", iss)
				}
			}

			log.Infof("Using static keyset with %d keys", len(keys))
			keyset.Replace(&keys)
			stopKeyDiscovery, pingChan, err = key.Static(keyset)
		}
		if err != nil {
			log.Fatalf("Key discovery setup failed: %s", err)
		}

		// set configuration modification time to apisrv startup time
		// because a fixed time constant does not roll-back client-stored
		// configurations when an apisrv deployment was rolled-back to an
		// earlier git commit
		configuration.DefaultConfigurationModTime = time.Now()

		// web server
		api.BaseURL = viper.GetString("web.base_url")
		stopWebServer, err := web.Serve(ctx,
			viper.GetString("web.listen"),
			viper.GetString("keys.service_name"),
			viper.GetString("web.base_url"),
			viper.GetStringSlice("web.cors_origins"),
			viper.GetString("webapp.base_url"),
			viper.GetDuration("queue.job_timeout"),
			&caKey.Key, caPrivKey, keyset, pingChan, database, store)
		if err != nil {
			log.Fatalf("Starting web server failed: %s", err)
		}

		switch <-signalChan {
		case syscall.SIGTERM, syscall.SIGINT:
			// SIGTERM. Exit gracefully
			log.Infoln("SIGTERM. Exit gracefully")
			exit = true

		case syscall.SIGHUP:
			// SIGHUP. Restart all services
			log.Infoln("SIGHUP. Restarting all services")
		}

		stopQueue()
		stopWebServer()
		database.Close()
		stopKeyDiscovery()
		flushTraceData()
	}
}

func notifyUnixSignals() chan os.Signal {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)

	return c
}

func decodeCaKey(str string) (*ecdsa.PrivateKey, error) {
	buf, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return nil, err
	}

	ca, err := x509.ParseECPrivateKey(buf)
	if err != nil {
		return nil, err
	}

	if ca.Curve == elliptic.P256() {
		return ca, nil
	} else {
		return nil, fmt.Errorf("not a ECDSA key over NIST P-256")
	}
}

func main() {
	log.Infof("immune Guard apisrv %s\n", releaseId)

	viper.SetEnvPrefix("apisrv")

	viper.SetDefault("telemetry.service", "apisrv-v2")
	viper.SetDefault("telemetry.console_log", true)
	viper.BindEnv("telemetry.token", "APISRV_TELEMETRY_TOKEN")
	viper.BindEnv("bucket.api_secret", "APISRV_BUCKET_SECRET")

	viper.BindEnv("database.url", "APISRV_DATABASE_URL")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.name", "guard")
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.pool_size", "1")
	viper.BindEnv("database.password", "APISRV_DATABASE_PASSWORD")

	viper.SetDefault("web.listen", "0.0.0.0:8000")
	viper.SetDefault("web.schemaPath", "/var/immune/schemas")
	viper.SetDefault("web.base_url", "https://xxx.xxx.xxx/v2")
	viper.SetDefault("web.cors_origins", []string{"https://xxx.xxx.xxx"})

	viper.SetDefault("webapp.base_url", "https://xxx.xxx")

	viper.SetDefault("keys.service_name", "apisrv-v2")
	viper.BindEnv("keys.attestation", "APISRV_KEYS_ATTESTATION")
	viper.BindEnv("keys.authentication", "APISRV_KEYS_AUTHENTICATION")
	viper.BindEnv("keys.enrollment", "APISRV_KEYS_ENROLLMENT")
	viper.SetDefault("keys.discover", false)
	viper.SetDefault("keys.label_selector", "app.kubernetes.io/part-of=immune-guard,app.kubernetes.io/name")

	viper.SetDefault("events.receiver", "authsrv")
	viper.SetDefault("events.worker_pool", "1")

	viper.SetDefault("queue.job_timeout", "30s")

	viper.BindEnv("binarly.client_secret", "APISRV_BINARLY_CLIENT_SECRET")

	viper.SetConfigName("apisrv.yaml")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/immune")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Warn("No config file found. Using defaults.")
		} else {
			log.Fatalf("Fatal error config file '%s': %s", viper.ConfigFileUsed(), err)
		}
	} else {
		log.Infof("Configuration loaded from %s\n", viper.ConfigFileUsed())
	}

	// parse database url
	if viper.IsSet("database.url") {
		cfg, err := pgx.ParseConfig(viper.GetString("database.url"))
		if err != nil {
			log.Fatalf("parsing database URL: %s\n", err)
		}
		viper.Set("database.host", cfg.Host)
		viper.Set("database.port", cfg.Port)
		viper.Set("database.name", cfg.Database)
		viper.Set("database.user", cfg.User)
	}

	// load the database certificate
	if viper.IsSet("database.certificate_file") {
		if pemCert, err := ioutil.ReadFile(viper.GetString("database.certificate_file")); err != nil {
			log.Fatalf("Failed to read SQL database certificate: %s\n", err)
		} else if !viper.IsSet("database.certificate") {
			viper.Set("database.certificate", string(pemCert))
		}
	}

	run()
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
