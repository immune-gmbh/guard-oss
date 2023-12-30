# syntax = docker/dockerfile:1-experimental

FROM golang:1.20-alpine AS go-builder

# Set necessary environmet variables needed for our image
ENV GO111MODULE=on

# set to the output of "git describe --tags"
ARG RELEASE_ID

# Move to working directory /build
WORKDIR /build

RUN apk --update add make git

# Copy and download dependency using go mod
COPY apisrv/Makefile apisrv/go.sum apisrv/go.mod apisrv/gqlgen.yml apisrv/
WORKDIR /build/apisrv
RUN make deps

# Build the application
COPY apisrv/cmd/ cmd/
COPY apisrv/pkg/ pkg/
COPY apisrv/api/ api/
COPY apisrv/internal/ internal/
#RUN --mount=type=cache,target=/root/.cache/go-build make migration apisrv
RUN make migration apisrv

# Build a small image
FROM alpine:latest
RUN apk --update add bash

COPY --from=go-builder /build/apisrv/apisrv /
COPY --from=go-builder /build/apisrv/migration /
COPY apisrv/build/entrypoint.bash /

ENTRYPOINT [ "bash", "/entrypoint.bash" ]
