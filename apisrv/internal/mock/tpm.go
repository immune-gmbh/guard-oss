package mock

import (
	"context"
	"fmt"
	"io"
	"net"
	"testing"

	"github.com/google/go-tpm/tpm2"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type TPM struct {
	Container testcontainers.Container
}

func TPMContainer(t *testing.T, ctx context.Context) *TPM {
	req := testcontainers.ContainerRequest{
		Image:        "ghcr.io/immune-gmbh/tpm2-simulator:r2",
		ExposedPorts: []string{"2322/tcp"},
		WaitingFor:   wait.ForListeningPort("2322/tcp"),
		Env:          map[string]string{},
	}
	tpmC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	assert.NoError(t, err)

	i := TPM{
		Container: tpmC,
	}

	return &i
}

func (i *TPM) Connect(t *testing.T, ctx context.Context) io.ReadWriteCloser {
	tpmhost, err := i.Container.Host(ctx)
	assert.NoError(t, err)
	tpmport, err := i.Container.MappedPort(ctx, "2322/tcp")
	assert.NoError(t, err)
	sock, err := net.Dial("tcp", fmt.Sprintf("%s:%d", tpmhost, tpmport.Int()))
	assert.NoError(t, err)
	conn := io.ReadWriteCloser(sock)
	_ = tpm2.Startup(conn, tpm2.StartupClear)

	return conn
}

func (i *TPM) URL(t *testing.T, ctx context.Context) string {
	tpmhost, err := i.Container.Host(ctx)
	assert.NoError(t, err)
	tpmport, err := i.Container.MappedPort(ctx, "2322/tcp")
	assert.NoError(t, err)
	return fmt.Sprintf("net://%s:%d", tpmhost, tpmport.Int())
}

func (i *TPM) Terminate(ctx context.Context) {
	i.Container.Terminate(ctx)
}
