package mock

import (
	"context"
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/minio/madmin-go/v2"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type Minio struct {
	Container testcontainers.Container
	Endpoint  string
	Key       string
	Secret    string
	Bucket    string
}

func MinioContainer(t *testing.T, ctx context.Context) *Minio {
	user := "blah"
	pwd := "blubblub"
	rng := make([]byte, 8)
	rand.Read(rng)
	req := testcontainers.ContainerRequest{
		Image:        "minio/minio:latest",
		ExposedPorts: []string{"9000/tcp"},
		Mounts: testcontainers.Mounts(
			testcontainers.VolumeMount(fmt.Sprintf("data-%x", rng), "/data"),
		),
		WaitingFor: wait.ForHTTP("/minio/health/live").WithPort("9000/tcp"),
		Env: map[string]string{
			"MINIO_ROOT_USER":     user,
			"MINIO_ROOT_PASSWORD": pwd,
		},
		Cmd:        []string{"server", "/data"},
		AutoRemove: true,
	}
	minioC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	assert.NoError(t, err)

	s3host, err := minioC.Host(ctx)
	assert.NoError(t, err)
	s3port, err := minioC.MappedPort(ctx, "9000/tcp")
	assert.NoError(t, err)

	i := Minio{
		Container: minioC,
		Endpoint:  fmt.Sprintf("http://%s:%d", s3host, s3port.Int()),
		Key:       "test",
		Secret:    "testtest",
		Bucket:    "default",
	}

	// add user
	mdmClnt, err := madmin.New(fmt.Sprintf("%s:%d", s3host, s3port.Int()), user, pwd, false)
	assert.NoError(t, err)
	err = mdmClnt.AddUser(ctx, i.Key, i.Secret)
	assert.NoError(t, err)
	err = mdmClnt.SetPolicy(ctx, "readwrite", "test", false)
	assert.NoError(t, err)

	// add bucket
	minioClient, err := minio.New(fmt.Sprintf("%s:%d", s3host, s3port.Int()), &minio.Options{
		Creds: credentials.NewStaticV4(i.Key, i.Secret, ""),
	})
	assert.NoError(t, err)
	err = minioClient.MakeBucket(ctx, i.Bucket, minio.MakeBucketOptions{Region: "us-east-1"})
	assert.NoError(t, err)
	found, err := minioClient.BucketExists(ctx, i.Bucket)
	assert.NoError(t, err)
	assert.True(t, found)

	return &i
}

func (i *Minio) Terminate(ctx context.Context) {
	i.Container.Terminate(ctx)
}
