package web

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/api"
	cred "github.com/immune-gmbh/guard/apisrv/v2/pkg/authentication"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/key"
	tel "github.com/immune-gmbh/guard/apisrv/v2/pkg/telemetry"
)

var (
	ErrAuthentication = errors.New("authentication")
)

func authenticateEnroll(ctx context.Context, keyset *key.Set, r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")

	if !strings.HasPrefix(authHeader, "Bearer ") {
		tel.Log(ctx).WithField("token", authHeader).Error("Authorization header does not start with Bearer")
		return "", ErrAuthentication
	}

	return cred.VerifyEnrollmentCredential(ctx, strings.TrimPrefix(authHeader, "Bearer "), keyset)
}

func authenticateAttest(ctx context.Context, keyset *key.Set, r *http.Request) (*api.Name, string, error) {
	authHeader := r.Header.Get("Authorization")

	if !strings.HasPrefix(authHeader, "Bearer ") {
		tel.Log(ctx).WithField("token", authHeader).Error("Authorization header does not start with Bearer")
		return nil, "", ErrAuthentication
	}

	name, err := cred.VerifyDeviceCredential(ctx, strings.TrimPrefix(authHeader, "Bearer "), keyset)
	if err != nil {
		tel.Log(ctx).WithError(err).WithField("token", authHeader).WithField("keyset", keyset).
			Error("Verification failed")
	}
	return name, "agent", err
}

func authenticateEvent(ctx context.Context, keyset *key.Set, r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")

	if !strings.HasPrefix(authHeader, "Bearer ") {
		tel.Log(ctx).WithField("token", authHeader).Error("Authorization header does not start with Bearer")
		return "", ErrAuthentication
	}

	act, _, err := cred.VerifyServiceCredential(ctx, strings.TrimPrefix(authHeader, "Bearer "), keyset)
	if err != nil {
		tel.Log(ctx).WithError(err).WithField("token", authHeader).WithField("keyset", keyset).
			Error("Verification failed")
	}
	return act, err
}

func authenticateCrud(ctx context.Context, keyset *key.Set, r *http.Request) (*cred.AuthenticatedUser, error) {
	authHeader := r.Header.Get("Authorization")

	if !strings.HasPrefix(authHeader, "Bearer ") {
		tel.Log(ctx).WithField("token", authHeader).Error("Authorization header does not start with Bearer")
		return nil, ErrAuthentication
	}

	return cred.VerifyUserCredential(ctx, strings.TrimPrefix(authHeader, "Bearer "), keyset)
}
