# command names
GO=go
GOTEST=gotestsum

# build artifact
COMMAND=ops

# build-time parameters, values
RELEASE_ID?=$(shell git describe --tags)
GO_ENV=CGO_ENABLED=0 GOARCH=amd64
LDFLAGS=-X main.releaseId=$(RELEASE_ID) -extldflags "-static"

# suppress lots of legacy SCCS and RCS lookups
MAKEFLAGS += --no-builtin-rules 

.DEFAULT_GOAL:=all
.PHONY: all
all: main.go go.mod go.sum
	$(GO_ENV) $(GO) build -ldflags '$(LDFLAGS)' -gcflags="all=-N -l" -o $(COMMAND) main.go

deps: go.mod go.sum
	go mod download

.PHONY: clean 
clean:
	rm -f $(COMMAND)

# disable many builtin rules
.SUFFIXES:
