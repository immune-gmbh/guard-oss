package database

import (
	"context"
	"crypto/tls"
	"embed"
	"fmt"
	"net/http"

	"github.com/go-pg/migrations/v8"
	"github.com/go-pg/pg/v10"
	"github.com/jackc/pgx/v4/pgxpool"
	log "github.com/sirupsen/logrus"

	tel "github.com/immune-gmbh/guard/apisrv/v2/pkg/telemetry"
)

//go:embed *tx.up.sql *tx.down.sql
var MigrationFiles embed.FS

func connect(pool *pgxpool.Pool) *pg.DB {
	var tlsConfig *tls.Config

	cfg := pool.Config().ConnConfig
	if cfg.TLSConfig != nil && cfg.TLSConfig.RootCAs != nil {
		tlsConfig = cfg.TLSConfig
	}

	return pg.Connect(&pg.Options{
		Addr:      fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		User:      cfg.User,
		Password:  cfg.Password,
		Database:  cfg.Database,
		TLSConfig: tlsConfig,
	})
}

// Make sure <dbName> database exists
func EnsureDatabaseExists(ctx context.Context, pool *pgxpool.Pool, database string) error {
	conn := connect(pool)
	defer conn.Close()

	flag := -1
	_, err := conn.QueryOne(pg.Scan(&flag), "SELECT count(*) FROM pg_database WHERE datname = ?", database)
	if err != nil {
		return err
	}

	if flag == 0 {
		log.Printf("No '%s' database, creating one\n", database)

		_, err = conn.Exec("CREATE DATABASE ?", pg.Ident(database))
		if err != nil {
			return err
		}
	}
	return nil
}

func EnsureSchemaExists(ctx context.Context, pool *pgxpool.Pool, schema string) error {
	conn := connect(pool)
	defer conn.Close()

	flag := -1
	_, err := conn.QueryOne(pg.Scan(&flag), "SELECT count(*) FROM information_schema.schemata WHERE schema_name = ?", schema)
	if err != nil {
		return err
	}

	if flag == 0 {
		log.Printf("No '%s' schema, creating one\n", schema)
		_, err = conn.Exec("CREATE SCHEMA ?", pg.Ident(schema))
		if err != nil {
			return err
		}
	}

	return nil
}

// Make sure all roles exist
func EnsureDatabaseRoleExists(ctx context.Context, pool *pgxpool.Pool, role string, passwd string) error {
	conn := connect(pool)
	defer conn.Close()

	flag := -1
	_, err := conn.Query(pg.Scan(&flag), "SELECT count(*) FROM pg_roles WHERE rolname = ?", role)
	if err != nil {
		return err
	}

	if flag == 0 {
		log.Printf("No '%s' database role, creating one\n", role)

		_, err := conn.Exec("CREATE ROLE ? WITH LOGIN PASSWORD ?", pg.Ident(role), passwd)
		if err != nil {
			return err
		}
	}

	return nil
}

func RunDatabaseMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	return RunDatabaseMigrationsToTarget(ctx, pool, 0)
}

func ReverseDatabaseMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	// load the embedded SQL migrations
	co := migrations.NewCollection().DisableSQLAutodiscover(true)
	err := co.DiscoverSQLMigrationsFromFilesystem(http.FS(MigrationFiles), "/")
	if err != nil {
		return err
	}

	conn := connect(pool)
	defer conn.Close()

	// reset DB
	_, _, err = co.Run(conn, "reset")
	return err
}

func RunDatabaseMigrationsToTarget(ctx context.Context, pool *pgxpool.Pool, target int) error {
	// load the embedded SQL migrations
	co := migrations.NewCollection().DisableSQLAutodiscover(true)
	err := co.DiscoverSQLMigrationsFromFilesystem(http.FS(MigrationFiles), "/")
	if err != nil {
		return err
	}

	num := len(co.Migrations())
	switch num {
	case 0:
		log.Error("No migrations compiled in")
		return fmt.Errorf("no migrations")
	case 1:
		log.Info("Loaded 1 migration")
	default:
		log.Infof("Loaded %d migrations", num)
	}

	conn := connect(pool)
	defer conn.Close()

	// Gather target and current schema version
	cur, err := co.Version(conn)
	if err != nil {
		cur = -1
	}

	var tgt int64 = 0
	for _, m := range co.Migrations() {
		if m.Version > tgt {
			tgt = m.Version
		}
	}

	if cur > tgt {
		return fmt.Errorf("database schema is too new. Got %d, understand %d", cur, tgt)
	}

	if cur < tgt {
		log.Infof("Database schema version is %d, need %d. Running migrations.", cur, tgt)
	}

	// init DB
	_, _, err = co.Run(conn, "init")
	if err != nil {
		tel.Log(ctx).WithError(err).Error("migrations init")
		return err
	}

	// migrate DB
	if target > 0 {
		_, _, err = co.Run(conn, "up", fmt.Sprintf("%d", target))
	} else {
		_, _, err = co.Run(conn, "up")
	}
	if err != nil {
		tel.Log(ctx).WithError(err).WithField("target", target).Error("migrations up")
		return err
	}

	return nil
}
