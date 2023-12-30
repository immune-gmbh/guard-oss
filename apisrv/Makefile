# command names
GO=go
GOTEST=gotestsum

# build artifacts, sources
COMMANDS=apisrv migration ops
GOFILES=$(wildcard pkg/**/*.go) pkg/debugv1/debugv1.resolvers.go pkg/issuesv1/generated.go $(wildcard internal/**/*.go)

# build-time parameters, values
RELEASE_ID?=$(shell git describe --tags)
GO_ENV=CGO_ENABLED=0 GOARCH=amd64 GIT_TERMINAL_PROMPT=1 GOPRIVATE="github.com/binarly-io/*"
LDFLAGS=-X main.releaseId=$(RELEASE_ID) -extldflags "-static"

# suppress lots of legacy SCCS and RCS lookups
MAKEFLAGS += --no-builtin-rules 

.DEFAULT_GOAL:=all
.PHONY: all
all: $(COMMANDS)

deps: go.mod go.sum
	$(GO_ENV) go mod download
	$(GO_ENV) go install gotest.tools/gotestsum@latest
	$(GO_ENV) go install github.com/google/go-licenses

pkg/issuesv1/generated.go: api/issuesv1.schema.yaml internal/jsonschema/ cmd/jsonschema/main.go
	$(GO_ENV) go run cmd/jsonschema/main.go $< $@

pkg/debugv1/debugv1.resolvers.go: api/debugv1.graphql cmd/graphql/main.go gqlgen.yml
	$(GO_ENV) go run cmd/graphql/main.go generate ./...

%: cmd/%/main.go $(GOFILES) go.sum go.mod
	$(GO_ENV) $(GO) build -ldflags '$(LDFLAGS)' -gcflags="all=-N -l" -o $@ $<

.PHONY: clean 
clean:
	rm -f $(COMMANDS)

.PHONY: test
test: deps
	GOMAXPROCS=1 $(GOTEST) -- -v -coverprofile=coverage.txt -covermode=atomic ./pkg/*

.PHONY: license-check 
license-check: deps
	$(foreach var,$(COMMANDS),go-licenses --logtostderr check --ignore github.com/immune-gmbh/guard --exclude-restricted ./cmd/$(var);) 

# disable many builtin rules
.SUFFIXES:
