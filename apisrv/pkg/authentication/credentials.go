package authentication

// Subjects
//
// Service: tag:immu.ne,2021:service/<service name>
// User: tag:immu.ne,2021:user/<uuid>
// Organisation: tag:immu.ne,2021:organisation/<uuid>
// Device: tag:immu.ne,2021:device/<root key name as hex>
// Agent tag:immu.ne,2021:agent

// Tokens
// {
//   "kid": first 8 chars of the hex SHA256 hash of the key (DER encoding)
// }
// {
//   "sub": org, user or device
//   "act": if org {
//     "sub": agent, user, service
//   },
//   "iss": service
//   "exp": future,
//   "nbf": now
// }

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"math/rand"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/go-tpm/tpm2"
	log "github.com/sirupsen/logrus"

	api "github.com/immune-gmbh/guard/apisrv/v2/pkg/api"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/key"
	tel "github.com/immune-gmbh/guard/apisrv/v2/pkg/telemetry"
)

var (
	enrollCredentialValidityPeriod  = time.Hour * 24
	serviceCredentialValidityPeriod = time.Minute * 5
	userCredentialValidityPeriod    = time.Hour * 24 * 7

	tagOrganisationPrefix = "tag:immu.ne,2021:organisation/"
	tagServicePrefix      = "tag:immu.ne,2021:service/"
	tagDevicePrefix       = "tag:immu.ne,2021:device/"
	tagUserPrefix         = "tag:immu.ne,2021:user/"
	SubjectAgent          = "tag:immu.ne,2021:agent"
	SubjectAttestation    = "tag:immu.ne,2022:attestation-service"

	ErrFormat    = errors.New("token illformed")
	ErrExpiry    = errors.New("exp, iat or nbf wrong")
	ErrSignature = errors.New("signature wrong")
	ErrKey       = errors.New("no key")
	ErrClaims    = errors.New("sub or act wrong")
	ErrInternal  = errors.New("internal error")
)

var tokenSigningMethod = jwt.SigningMethodES256

type ActorClaims struct {
	Subject string `json:"sub,omitempty"`
	Name    string `json:"name,omitempty"`
	Role    string `json:"rol,omitempty"`
}

type Credential struct {
	jwt.StandardClaims
	Actor ActorClaims `json:"act,omitempty"`
}

type AuthenticatedUser struct {
	Actor                string
	OrganizationExternal string
}

func subjectOrganisation(org string) string {
	return tagOrganisationPrefix + org
}

func subjectUser(user string) string {
	return tagUserPrefix + user
}

func subjectService(serviceName string) string {
	return tagServicePrefix + serviceName
}

func subjectDevice(rootQName api.Name) (string, error) {
	buf, err := tpm2.Name(rootQName).Encode()
	if err != nil {
		return "", err
	}
	return tagDevicePrefix + hex.EncodeToString(buf), nil
}

func IssueCredential(subject string, actor *string, now time.Time, expiry time.Time, serviceName string, kid string, key *ecdsa.PrivateKey) (string, error) {
	var buf [32]byte
	if i, err := rand.Read(buf[:]); err != nil || i != len(buf) {
		return "", ErrInternal
	}
	jti := base64.StdEncoding.WithPadding(base64.NoPadding).EncodeToString(buf[:])
	var actorClaims ActorClaims
	if actor != nil {
		actorClaims = ActorClaims{
			Subject: *actor,
		}
	}
	var expiryClaim int64
	if !expiry.Equal(time.Time{}) {
		expiryClaim = expiry.Unix()
	}
	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodES256, Credential{
		StandardClaims: jwt.StandardClaims{
			Id:        jti,
			ExpiresAt: expiryClaim,
			NotBefore: now.Unix(),
			IssuedAt:  now.Unix(),
			Subject:   subject,
			Issuer:    subjectService(serviceName),
		},
		Actor: actorClaims,
	})
	tokenClaims.Header["kid"] = kid
	return tokenClaims.SignedString(key)
}

func IssueDeviceCredential(serviceName string, qname api.Name, now time.Time, kid string, key *ecdsa.PrivateKey) (string, error) {
	sub, err := subjectDevice(qname)
	if err != nil {
		return "", err
	}
	return IssueCredential(
		sub,
		nil,         // act
		now,         // nbf, iat
		time.Time{}, // exp
		serviceName, // iss
		kid, key)
}

func VerifyDeviceCredential(ctx context.Context, token string, keyset *key.Set) (*api.Name, error) {
	cred := Credential{}
	err := verifyCredential(ctx, token, keyset, &cred)
	if err != nil {
		return nil, err
	}

	if cred.NotBefore == 0 || cred.IssuedAt == 0 {
		return nil, ErrExpiry
	}

	if !strings.HasPrefix(cred.Subject, tagDevicePrefix) {
		return nil, ErrClaims
	}
	x, err := hex.DecodeString(strings.TrimPrefix(cred.Subject, tagDevicePrefix))
	if err != nil {
		return nil, ErrFormat
	}

	nam, err := tpm2.DecodeName(bytes.NewBuffer(x))
	if err != nil {
		return nil, ErrFormat
	}

	return (*api.Name)(nam), nil
}

func IssueServiceCredential(serviceName string, subjectOrg *string, now time.Time, kid string, key *ecdsa.PrivateKey) (string, error) {
	if subjectOrg != nil {
		act := subjectService(serviceName)
		return IssueCredential(
			subjectOrganisation(*subjectOrg),         // sub
			&act,                                     // act
			now,                                      // nbf
			now.Add(serviceCredentialValidityPeriod), // exp
			serviceName,                              // iss
			kid, key)
	} else {
		return IssueCredential(
			subjectService(serviceName),              // sub
			nil,                                      // act
			now,                                      // nbf
			now.Add(serviceCredentialValidityPeriod), // exp
			serviceName,                              // iss
			kid, key)
	}
}

func VerifyServiceCredential(ctx context.Context, token string, keyset *key.Set) (string, *string, error) {
	cred := Credential{}
	err := verifyCredential(ctx, token, keyset, &cred)
	if err != nil {
		return "", nil, err
	}

	if cred.IssuedAt == 0 || cred.ExpiresAt == 0 || cred.NotBefore == 0 {
		return "", nil, ErrExpiry
	}

	var svc string
	var org *string

	if strings.HasPrefix(cred.Subject, tagOrganisationPrefix) {
		sub := strings.TrimPrefix(cred.Subject, tagOrganisationPrefix)
		org = &sub

		if !strings.HasPrefix(cred.Actor.Subject, tagServicePrefix) {
			return "", nil, ErrClaims
		}
		svc = strings.TrimPrefix(cred.Actor.Subject, tagServicePrefix)
	} else if strings.HasPrefix(cred.Subject, tagServicePrefix) {
		svc = strings.TrimPrefix(cred.Subject, tagServicePrefix)
	} else {
		return "", nil, ErrClaims
	}

	if subjectService(svc) != cred.Issuer {
		return "", nil, ErrClaims
	}

	return svc, org, nil
}

func IssueEnrollmentCredential(serviceName string, org string, now time.Time, kid string, key *ecdsa.PrivateKey) (string, error) {
	return IssueCredential(
		subjectOrganisation(org),                // sub
		&SubjectAgent,                           // act
		now,                                     // nbf
		now.Add(enrollCredentialValidityPeriod), // exp
		serviceName,                             // iss
		kid, key)
}

func VerifyEnrollmentCredential(ctx context.Context, token string, keyset *key.Set) (string, error) {
	cred := Credential{}
	err := verifyCredential(ctx, token, keyset, &cred)
	if err != nil {
		return "", err
	}

	if cred.IssuedAt == 0 || cred.ExpiresAt == 0 || cred.NotBefore == 0 {
		return "", ErrExpiry
	}

	if !strings.HasPrefix(cred.Subject, tagOrganisationPrefix) {
		return "", ErrClaims
	}
	org := strings.TrimPrefix(cred.Subject, tagOrganisationPrefix)

	if cred.Actor.Subject != SubjectAgent {
		return "", ErrClaims
	}

	return org, nil
}

func IssueUserCredential(serviceName string, org string, user string, now time.Time, kid string, key *ecdsa.PrivateKey) (string, error) {
	act := subjectUser(user)
	return IssueCredential(
		subjectOrganisation(org), // sub
		&act,                     // act
		now,                      // nbf
		now.Add(time.Hour),       // exp
		serviceName,              // iss
		kid, key)
}

func VerifyUserCredential(ctx context.Context, token string, keyset *key.Set) (*AuthenticatedUser, error) {
	cred := Credential{}
	err := verifyCredential(ctx, token, keyset, &cred)
	if err != nil {
		return nil, err
	}

	if cred.IssuedAt == 0 || cred.ExpiresAt == 0 || cred.NotBefore == 0 {
		return nil, ErrExpiry
	}

	if !strings.HasPrefix(cred.Subject, tagOrganisationPrefix) {
		return nil, ErrClaims
	}
	org := strings.TrimPrefix(cred.Subject, tagOrganisationPrefix)

	if !strings.HasPrefix(cred.Actor.Subject, tagUserPrefix) {
		return nil, ErrClaims
	}

	return &AuthenticatedUser{Actor: cred.Actor.Subject, OrganizationExternal: org}, nil
}

func verifyCredential(ctx context.Context, str string, keyset *key.Set, claims *Credential) error {
	p := jwt.Parser{}
	token, _, err := p.ParseUnverified(str, claims)
	if err != nil {
		return ErrFormat
	}

	if token.Method != tokenSigningMethod {
		return ErrFormat
	}

	kidif, ok := token.Header["kid"]
	if !ok {
		tel.Log(ctx).Debug("no kid in header")
		return ErrKey
	}
	kid, ok := kidif.(string)
	if !ok {
		tel.Log(ctx).Debug("no kid isn't a string")
		return ErrKey
	}
	key, ok := keyset.Key(kid)
	if !ok {
		tel.Log(ctx).WithField("kid", kid).Debug("no key with that kid in keyset. see /v2/ready")
		return ErrKey
	}
	if !strings.HasPrefix(claims.Issuer, tagServicePrefix) {
		tel.Log(ctx).WithField("issuer", claims.Issuer).Debug("no key with that kid in keyset. see /v2/ready")
		return ErrKey
	}
	iss := strings.TrimPrefix(claims.Issuer, tagServicePrefix)
	if iss != key.Issuer {
		tel.Log(ctx).WithFields(log.Fields{"key": key.Issuer, "token": iss}).Debug("key issuer not matching token issuer")
		return ErrKey
	}
	_, err = p.Parse(str, func(*jwt.Token) (interface{}, error) { return &key.Key, nil })
	if err != nil {
		tel.Log(ctx).Debug("token signature wrong")
		return ErrSignature
	}
	return nil
}
