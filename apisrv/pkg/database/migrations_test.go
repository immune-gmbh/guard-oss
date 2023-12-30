package database

import (
	"context"
	_ "embed"
	"testing"

	"github.com/immune-gmbh/guard/apisrv/v2/internal/mock"
	"github.com/stretchr/testify/assert"
)

//go:embed seed_8.sql
var seed8Sql string

func TestMigrations(t *testing.T) {
	if testing.Short() {
		t.Skipf("skipping integration test")
	}

	ctx := context.Background()
	pgsqlC := mock.PostgresContainer(t, ctx)
	defer pgsqlC.Terminate(ctx)

	// Create database
	conn := pgsqlC.ConnectAdmin(t, ctx)
	err := EnsureDatabaseExists(ctx, conn, pgsqlC.UserDatabase)
	assert.NoError(t, err)

	// Create database roles
	err = EnsureDatabaseRoleExists(ctx, conn, pgsqlC.UserRole, pgsqlC.UserPassword)
	conn.Close()
	assert.NoError(t, err)

	// Run migrations to 8
	conn = pgsqlC.ConnectForMigration(t, ctx)
	err = RunDatabaseMigrationsToTarget(ctx, conn, 8)
	conn.Close()
	assert.NoError(t, err)

	// Seed database
	conn = pgsqlC.Connect(t, ctx)
	_, err = conn.Exec(ctx, seed8Sql)
	conn.Close()
	assert.NoError(t, err)

	// Run remaining migrations
	conn = pgsqlC.ConnectForMigration(t, ctx)
	err = RunDatabaseMigrationsToTarget(ctx, conn, 0)
	assert.NoError(t, err)

	// Migrate down
	err = ReverseDatabaseMigrations(ctx, conn)
	conn.Close()
	assert.NoError(t, err)
}

func TestRunDatabaseMigrations(t *testing.T) {
	if testing.Short() {
		t.Skipf("skipping integration test")
	}

	ctx := context.Background()
	pgsqlC := mock.PostgresContainer(t, ctx)
	defer pgsqlC.Terminate(ctx)

	conn := pgsqlC.ConnectAdmin(t, ctx)
	_, err := conn.Exec(ctx, "CREATE DATABASE "+pgsqlC.UserDatabase)
	assert.NoError(t, err)
	_, err = conn.Exec(ctx, "CREATE ROLE "+pgsqlC.UserRole)
	assert.NoError(t, err)

	err = RunDatabaseMigrations(ctx, conn)
	assert.NoError(t, err)
}

func TestRunDatabaseMigrationsTwice(t *testing.T) {
	if testing.Short() {
		t.Skipf("skipping integration test")
	}

	ctx := context.Background()
	pgsqlC := mock.PostgresContainer(t, ctx)
	defer pgsqlC.Terminate(ctx)

	conn := pgsqlC.ConnectAdmin(t, ctx)
	_, err := conn.Exec(ctx, "CREATE DATABASE "+pgsqlC.UserDatabase)
	assert.NoError(t, err)
	_, err = conn.Exec(ctx, "CREATE ROLE "+pgsqlC.UserRole)
	assert.NoError(t, err)

	err = RunDatabaseMigrations(ctx, conn)
	assert.NoError(t, err)

	err = RunDatabaseMigrations(ctx, conn)
	assert.NoError(t, err)
	conn.Close()
}
