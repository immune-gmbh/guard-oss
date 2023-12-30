package authentication

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/jackc/pgx/v4/pgxpool"
)

func IsRevoked(ctx context.Context, pool *pgxpool.Pool, token string) (bool, error) {
	claims := Credential{}
	p := jwt.Parser{}
	raw, _, err := p.ParseUnverified(token, &claims)
	if err != nil {
		return false, err
	}

	var kid *string
	if str, ok := raw.Header["kid"].(string); ok {
		kid = &str
	}

	rows, err := pool.Query(ctx, `
		SELECT 1 FROM v2.revokations
		WHERE tid = $1 OR ($2::text IS NOT NULL AND kid = $2)
	`, claims.Id, kid)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	return rows.Next(), nil
}

func RevokeToken(ctx context.Context, pool *pgxpool.Pool, jti string, expiresAt time.Time) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO v2.revokations (id, tid, kid, expiry)
		VALUES (DEFAULT, $1, NULL, $2)
	`, jti, expiresAt)

	return err
}

func RevokeKey(ctx context.Context, pool *pgxpool.Pool, kid string, expiresAt time.Time) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO v2.revokations (id, tid, kid, expiry)
		VALUES (DEFAULT, NULL, $1, $2)
	`, kid, expiresAt)

	return err
}

func GarbageCollectRevokations(ctx context.Context, pool *pgxpool.Pool, now time.Time) error {
	_, err := pool.Exec(ctx, `
		DELETE FROM v2.revokations WHERE expiry < $1
	`, now)
	return err
}
