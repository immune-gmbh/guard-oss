package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"runtime"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	log "github.com/sirupsen/logrus"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/database"
)

var releaseId string = "unknown"

func main() {
	os.Exit(run())
}

func run() int {
	ctx := context.Background()

	var (
		dbHost       string
		dbPort       int
		dbName       string
		dbUser       string
		dbAdmin      string
		dbDefault    string
		dbUrl        string
		certPath     string
		doWait       bool
		doCreateDb   bool
		doCreateRole bool
	)

	// introduce ourselves
	log.Printf("immune Guard apisrv2 -- SQL migration tool " + releaseId + " (" + runtime.GOARCH + ")")

	flag.StringVar(&dbAdmin, "database-admin", "postgres", "Name of the database administrator role. Needs to have ALTER/DROP/CREATE TABLE privileges on the database. Needs CREATE DATABASE privileges if -create-db is given and CREATE ROLE if -create-role is. Password is set via the CONN_PASSWORD environment variable.")
	flag.StringVar(&dbDefault, "database-default", "postgres", "Name of the default database used when connecting as administrator.")
	flag.StringVar(&dbUser, "database-user", "apisrv", "Name of the database role to create if -create-role is given.")
	flag.StringVar(&dbName, "database-name", "apisrv", "Name of the database to create when -create-db is given. Password is set via the ROLE_PASSWORD environment variable.")
	flag.StringVar(&dbUrl, "database-url", "", "Database connection URL. Replaces other database arguments except -database-user.")
	flag.StringVar(&dbHost, "database-host", "localhost", "Hostname of the PostgreSQL database server.")
	flag.IntVar(&dbPort, "database-port", 5432, "Port of the PostgreSQL database server.")

	flag.StringVar(&certPath, "database-cert", "", "Path to the CA certificate of the server. Connection will require SSL if set.")

	flag.BoolVar(&doWait, "wait", false, "Wait for the database to come online.")
	flag.BoolVar(&doCreateDb, "create-database", false, "Create the databse if it does not exist.")
	flag.BoolVar(&doCreateRole, "create-role", false, "Create a 'guard' role for the API server if it does not exist.")
	flag.Parse()

	createRolePasswd := os.Getenv("USER_PASSWORD")
	dbPasswd := os.Getenv("ADMIN_PASSWORD")
	if dbPasswd == "" {
		dbPasswd = `""`
	}

	// -certificate
	dbTLS := ""
	if _, err := ioutil.ReadFile(certPath); err == nil {
		dbTLS = fmt.Sprintf(` sslmode=verify-full sslrootcert=%s`, certPath)
	}

	// Database URL
	if dbUrl != "" {
		cfg, err := pgx.ParseConfig(dbUrl)
		if err != nil {
			log.Fatalf("parsing database URL: %s\n", err)
		}
		dbHost = cfg.Host
		dbPort = int(cfg.Port)
		dbDefault = cfg.Database
		dbAdmin = cfg.User
	}

	// Database connection
	connOpts, err := pgxpool.ParseConfig(fmt.Sprintf(
		`user=%s password=%s host=%s port=%d dbname=%s%s`,
		dbAdmin, dbPasswd, dbHost, dbPort, dbDefault, dbTLS))
	if err != nil {
		log.Error(err)
		return 1
	}
	connOpts.LazyConnect = true

	// -wait
	if doWait {
		log.Infof("waiting for %s", connOpts.ConnConfig.Host)
		waitForDatabase(connOpts)
	}

	// -create-db, -create-role
	if doCreateDb || doCreateRole {
		conn, err := pgxpool.ConnectConfig(context.Background(), connOpts)
		if err != nil {
			log.Error(err)
			return 1
		}

		if doCreateDb {
			if err := database.EnsureDatabaseExists(ctx, conn, dbName); err != nil {
				log.Error(err)
				conn.Close()
				return 1
			}
		}

		if doCreateRole {
			if err := database.EnsureDatabaseRoleExists(ctx, conn, dbUser, createRolePasswd); err != nil {
				log.Error(err)
				conn.Close()
				return 1
			}
		}

		conn.Close()
	}

	// migrate
	connOpts.ConnConfig.Database = dbName
	conn, err := pgxpool.ConnectConfig(context.Background(), connOpts)
	if err != nil {
		log.Error(err)
		return 1
	}
	err = database.RunDatabaseMigrations(ctx, conn)
	conn.Close()
	if err != nil {
		log.Error(err)
		return 1
	}

	return 0
}

func waitForDatabase(opts *pgxpool.Config) {
	for {
		// ConnectConfig should not return network errors when connecting lazily
		// however, we just collect errors of ConnectConfig and Ping (which opens
		// the actual connection) just to be safe
		conn, err := pgxpool.ConnectConfig(context.Background(), opts)
		if err == nil {
			err = conn.Ping(context.Background())
			conn.Close()
			if err == nil {
				return
			}
		}

		// test if this is a network error
		var netErr net.Error
		noLog := false
		if errors.As(err, &netErr) {
			// timeouts are expected and should be ok
			if netErr.Timeout() {
				noLog = true
			} else {
				// opErrors are read errors or refused connections, which
				// can happen when postgre is not started or starting up
				_, ok := netErr.(*net.OpError)
				noLog = ok
			}
		}

		// if this is not a network connect error (which is expected) then log the error
		// this might under some circumstances log one read error from within the pgconn library
		// which is when we happen to try to connect during a state where sockets are inited
		// that error however is hard to sort out and should be ignored
		if !noLog {
			log.Error(err)
		}

		time.Sleep(time.Second * 1)
	}
}
