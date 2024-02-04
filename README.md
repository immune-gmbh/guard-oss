immune Guard
============

The following documents how to install, configure and run immune Guard.

## Architecture

immune Guard is a set of web services that receive telemetry data from the
immune Guard Agent application. The data includes firmware versions and
configuration. immune Guard sends out alerts if vulnerabilities of insecure
settings are detected.

An installation of immune Guard consists of three services: the attestation and
authentication services as well as a server serving the web frontend.

**Attestation service (apisrv)**: Receives telemetry data and manages devices
and their security policy.

**Authentication service (authsrv)**: Handles user authentication and sends
alerts.

**Web frontend (appsrv)**: Serves the web frontend

**Operator tool (ops)**: Command line tool for various administrative tasks.

Please find it [here](https://github.com/immune-gmbh/agent).

## Prerequisites

* Kubernetes 1.21.11 with kubelet
* Kustomize 4.4.1
* Postgres 12 database with `hstore`, `intarray` and `pgcrypto` extensions
* S3 compatible storage (tested with Minio and Digitalocean Spaces)

**Optional**

* Mailgun API keys
* Google OAuth API keys
* Github OAuth API keys

## Installation

The immune Guard installation consists of a set of Kubernetes resources. Before
deploying these need to be filled in with various installation specific values.
This section guides you through what files need to be modified. The
configuration files have comments where values need to be replaced.

### Postgres

The Postgres database needs to have to following roles and database set up.

| Service  | Role name  | Database name |
| -        | -          | -             |
| apisrv   | apisrv     | *any*         |
| authsrv  | *any*      | *any*         |

```bash
pqsl '<POSTGRES CONNECTION URL>' <<EOF
create role apisrv login password '<APISRV PASSWORD>';
create role authsrv login password '<AUTHSRV PASSWORD>';
create database apisrv;
create database authsrv;
EOF
```

Names and credentials for both Postgres new roles as well as the
superuser role used for migrations need to be put into the following
two config files.

 - [ ] `overlays/production/database-config.yaml`
 - [ ] `overlays/production/database-secret.yaml`

### API service

The API service needs to configured with the URLs of webapp and API
endpoints, Intel TSC and S3 bucket credentials.

 - [ ] `overlays/production/apisrv/apisrv.yaml`

The Ingress resource needs to be configured with the right Ingress Class and
hostname.

 - [ ] `overlays/production/apisrv/kustomization.yaml`

All secret credentials are supplied via environment variables read from the
following ConfigMap.

 - [ ] `overlays/production/apisrv/env-secret.yaml`

The API service needs two pairs of signing keys. One for authenticating itself
at other services and one for certifying attestation keys on devices. These key
pairs are created with the `ops` tool. `ops key new NAME` will generate a new
key pair and write public and private keys to `NAME.pub` and `NAME.key`
respectively.

```bash
./ops keys new apisrv-auth
./ops keys new apisrv-enroll
```

The private keys are injected as files from a Secret.

```bash
kubectl create secret generic apisrv-v2-keys \
  --from-file=token.key=apisrv-auth.key \
  --from-file=device-ca.key=apisrv-enroll.key
```

The public keys are configured as annotations on the API service Deployment.

 - [ ] `overlays/production/apisrv/deployment-config.yaml`

### Authentication service

The authentication service needs to configured with the URLs of webapp and API
endpoints, Google/Github OAuth and Mailgun credentials.

 - [ ] `overlays/production/authsrv/settings.yml`

The Ingress resource needs to be configured with the right Ingress Class and
hostname.

 - [ ] `overlays/production/authsrv/kustomization.yaml`

The authentication service needs a pair of signing keys for authenticating itself
to other services as well as signing authentication tokens issued to users.
These key pairs are created with the `ops` tool as well.

```bash
./ops keys new authsrv-auth
```

The private keys are injected as files from a Secret.

```bash
kubectl create secret generic authsrv-v2-keys \
  --from-file=token.key=authsrv-auth.key
```

The public keys are configured as annotations on the authentication service
Deployment.

 - [ ] `overlays/production/authsrv/deployment-config.yaml`

### Web frontend

The web frontend is served by a separate container. It needs to know the URLs
of the frontend and the API.

 - [ ] `overlays/production/appsrv/env.production`

The Ingress resource needs to be configured with the right Ingress Class and
hostname.

 - [ ] `overlays/production/appsrv/kustomization.yaml`

## Image Registry

For pulling the container images from the private registry we need to configure
an Image Pull Secret.

```
kubectl create secret docker-registry ghcr-secret --docker-server=https://ghcr.io --docker-username=<USERNAME> --docker-password=<GITHUB TOKEN>
```

### Deploy

Now all Kubernetes resources are configured and can be deployed.

```bash
kustomize build overlay/production | kubectl apply -
```

## Notes

None of the services monitor configuration files. Changes are only picked up
after to pod has been restarted.

## Monitoring

The API and authentication service deployments use `prometheus.io/*`
annotations for scraping with Prometheus's Kubernetes auto discovery feature.
Distributed tracing is supported via OTLP (option is called lightstep in
apisrv.yaml).

## Support

Just provide issues on Github directly, same goes for PR's.
