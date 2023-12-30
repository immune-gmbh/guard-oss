package mock

import (
	"context"
	"embed"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/go-pg/migrations/v8"
	"github.com/go-pg/pg/v10"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type Postgres struct {
	Container     testcontainers.Container
	UserRole      string
	UserPassword  string
	UserDatabase  string
	AdminRole     string
	AdminPassword string
	AdminDatabase string
}

func (i *Postgres) Reset(t assert.TestingT, ctx context.Context, files embed.FS) {
	conn := i.ConnectAdmin(t, ctx)
	defer conn.Close()

	// drop db
	_, err := conn.Exec(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s", i.UserDatabase))
	assert.NoError(t, err)
	_, err = conn.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s", i.UserDatabase))
	assert.NoError(t, err)

	// reset role
	flag := -1
	err = pgxscan.Get(ctx, conn, &flag, "SELECT count(*) FROM pg_roles WHERE rolname = $1", i.UserRole)
	assert.NoError(t, err)
	if flag == 0 {
		_, err = conn.Exec(ctx, fmt.Sprintf("CREATE ROLE %s WITH LOGIN PASSWORD '%s'", i.UserRole, i.UserPassword))
		assert.NoError(t, err)
	}

	// migrate
	i.migrate(t, ctx, files, -1)
}

func (i *Postgres) ResetAndSeed(t assert.TestingT, ctx context.Context, files embed.FS, seed string, seedSchemaVer int) {
	conn := i.ConnectAdmin(t, ctx)
	defer conn.Close()

	// drop db
	_, err := conn.Exec(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s", i.UserDatabase))
	assert.NoError(t, err)
	_, err = conn.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s", i.UserDatabase))
	assert.NoError(t, err)

	// reset role
	flag := -1
	err = pgxscan.Get(ctx, conn, &flag, "SELECT count(*) FROM pg_roles WHERE rolname = $1", i.UserRole)
	assert.NoError(t, err)
	if flag == 0 {
		_, err = conn.Exec(ctx, fmt.Sprintf("CREATE ROLE %s WITH LOGIN PASSWORD '%s'", i.UserRole, i.UserPassword))
		assert.NoError(t, err)
	}
	conn.Close()
	conn = i.Connect(t, ctx)

	// migrate
	i.migrate(t, ctx, files, seedSchemaVer)
	_, err = conn.Exec(ctx, seed)
	conn.Close()
	if !assert.NoError(t, err) {
		assert.FailNow(t, "!!! FAILED TO SEED DATABASE !!!")
	}
	i.migrate(t, ctx, files, -1)
}

func (i *Postgres) Terminate(ctx context.Context) {
	i.Container.Terminate(ctx)
}

func (i *Postgres) connect(t assert.TestingT, ctx context.Context, role string, pwd string, db string) *pgxpool.Pool {
	pghost, err := i.Container.Host(ctx)
	assert.NoError(t, err)
	pgport, err := i.Container.MappedPort(ctx, "5432/tcp")
	assert.NoError(t, err)

	connOpts, err := pgxpool.ParseConfig("")
	assert.NoError(t, err)
	connOpts.ConnConfig.Host = pghost
	connOpts.ConnConfig.Port = uint16(pgport.Int())
	connOpts.ConnConfig.User = role
	connOpts.ConnConfig.Password = pwd
	connOpts.ConnConfig.Database = db
	connOpts.LazyConnect = true
	connOpts.MaxConns = 5
	connOpts.HealthCheckPeriod = time.Second * 30
	connOpts.MaxConnLifetime = time.Minute

	pool, err := pgxpool.ConnectConfig(ctx, connOpts)
	assert.NoError(t, err)
	return pool
}

func (i *Postgres) DumpLogs(t assert.TestingT, ctx context.Context) {
	logs, err := i.Container.Logs(ctx)
	assert.NoError(t, err)
	io.Copy(os.Stderr, logs)
}

func (i *Postgres) ConnectionCmd(t assert.TestingT, ctx context.Context) string {
	pghost, err := i.Container.Host(ctx)
	assert.NoError(t, err)
	pgport, err := i.Container.MappedPort(ctx, "5432/tcp")
	assert.NoError(t, err)

	ret := fmt.Sprintf("Database access: psql -h %s -p %d -U %s\n", pghost, pgport.Int(), i.UserRole)
	ret += fmt.Sprintf("Database password: %s\n", i.UserPassword)

	return ret
}

func (i *Postgres) Connect(t assert.TestingT, ctx context.Context) *pgxpool.Pool {
	return i.connect(t, ctx, i.UserRole, i.UserPassword, i.UserDatabase)
}

func (i *Postgres) ConnectAdmin(t assert.TestingT, ctx context.Context) *pgxpool.Pool {
	return i.connect(t, ctx, i.AdminRole, i.AdminPassword, i.AdminDatabase)
}

func (i *Postgres) ConnectForMigration(t assert.TestingT, ctx context.Context) *pgxpool.Pool {
	return i.connect(t, ctx, i.AdminRole, i.AdminPassword, i.UserDatabase)
}

func (i *Postgres) migrate(t assert.TestingT, ctx context.Context, files embed.FS, target int) {
	// load the embedded SQL migrations
	co := migrations.NewCollection().DisableSQLAutodiscover(true)
	err := co.DiscoverSQLMigrationsFromFilesystem(http.FS(files), "/")
	assert.NoError(t, err)

	pghost, err := i.Container.Host(ctx)
	assert.NoError(t, err)
	pgport, err := i.Container.MappedPort(ctx, "5432/tcp")
	assert.NoError(t, err)
	conn := pg.Connect(&pg.Options{
		Addr:     fmt.Sprintf("%s:%d", pghost, pgport.Int()),
		User:     i.AdminRole,
		Password: i.AdminPassword,
		Database: i.UserDatabase,
	})
	defer conn.Close()

	var tgt int64 = 0
	for _, m := range co.Migrations() {
		if m.Version > tgt {
			tgt = m.Version
		}
	}

	// init DB
	_, _, err = co.Run(conn, "init")
	assert.NoError(t, err)
	// migrate DB
	if target > 0 {
		_, _, err = co.Run(conn, "up", fmt.Sprintf("%d", target))
	} else {
		_, _, err = co.Run(conn, "up")
	}
	assert.NoError(t, err)
}

func PostgresContainer(t assert.TestingT, ctx context.Context) *Postgres {
	user := "apisrv"
	pwd := "blub"
	req := testcontainers.ContainerRequest{
		Image:        "postgres:14",
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor:   wait.ForLog("database system is ready to accept connections"),
		Env: map[string]string{
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "postgres",
		},
	}
	pgsqlC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	assert.NoError(t, err)

	i := Postgres{
		Container:     pgsqlC,
		UserRole:      user,
		UserPassword:  pwd,
		UserDatabase:  "apisrv",
		AdminRole:     "postgres",
		AdminPassword: "postgres",
		AdminDatabase: "postgres",
	}

	var conn *pgxpool.Pool
	now := time.Now()
	for {
		conn = i.ConnectAdmin(t, ctx)
		_, err := conn.Exec(ctx, "select 1")
		if err == nil {
			break
		}
		if time.Since(now) > 5*time.Second {
			t.Errorf("database container failed to become ready")
			break
		}
		time.Sleep(200 * time.Millisecond)
	}

	return &i
}
