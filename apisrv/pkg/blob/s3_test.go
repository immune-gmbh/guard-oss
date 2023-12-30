package blob

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/klauspost/compress/zstd"
	"github.com/stretchr/testify/assert"

	"github.com/immune-gmbh/guard/apisrv/v2/internal/mock"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/database"
)

func TestS3(t *testing.T) {
	if testing.Short() {
		t.Skipf("skipping integration test")
	}

	ctx := context.Background()
	tests := map[string]func(*testing.T, *pgxpool.Pool, *Storage){
		"Upload":        testUpload,
		"UploadBatch":   testUploadBatch,
		"DownloadBatch": testDownloadBatch,
		"RaceCondition": testCollision,
	}
	pgsqlC := mock.PostgresContainer(t, ctx)
	defer pgsqlC.Terminate(ctx)
	minio := mock.MinioContainer(t, ctx)
	defer minio.Terminate(ctx)
	for name, fn := range tests {
		pgsqlC.Reset(t, ctx, database.MigrationFiles)

		t.Run(name, func(t *testing.T) {
			conn := pgsqlC.Connect(t, ctx)
			store, err := NewStorage(ctx,
				WithBucket{Bucket: minio.Bucket},
				WithEndpoint{
					Endpoint: minio.Endpoint,
					Region:   "us-east-1",
				},
				WithCredentials{
					Key:    minio.Key,
					Secret: minio.Secret,
				},
				WithTestMode{Enable: true})
			assert.NoError(t, err)

			defer conn.Close()
			fn(t, conn, store)
		})
	}
}

func testUpload(t *testing.T, db *pgxpool.Pool, store *Storage) {
	ctx := context.Background()
	now := time.Now()
	b := []byte{1, 2, 3}
	h := sha256.Sum256(b)
	row, err := store.Upload(ctx, db, BIOS, b, now)
	assert.NoError(t, err)
	assert.Equal(t, now, row.CreatedAt)
	assert.True(t, strings.HasPrefix(row.Filename(), fmt.Sprintf("%x-", h)))

	row2, err := store.Fetch(ctx, db, BIOS, h)
	assert.NoError(t, err)
	assert.Equal(t, row.Id, row2.Id)
	assert.Equal(t, row.Snowflake, row2.Snowflake)
	assert.Equal(t, row.Digest, row2.Digest)
	assert.Equal(t, row.Namespace, row2.Namespace)
	assert.WithinDuration(t, row.CreatedAt, row2.CreatedAt, time.Second)
	meta, err := row.Metadata()
	meta2, err := row2.Metadata()
	assert.Equal(t, meta, meta2)

	// upload same
	row3, err := store.Upload(ctx, db, BIOS, b, now)
	assert.NoError(t, err)
	assert.Equal(t, row3.Id, row.Id)

	// download file
	buf := new(bytes.Buffer)
	err = store.Download(ctx, row, buf)
	assert.NoError(t, err)
	assert.Equal(t, buf.Bytes(), b)
}

func testUploadBatch(t *testing.T, db *pgxpool.Pool, store *Storage) {
	ctx := context.Background()
	now := time.Now()
	upbatch := make([]*UploadJob, 10)
	for i := range upbatch {
		var job UploadJob
		job.Namespace = WindowsExecutable
		buf := new(bytes.Buffer)

		if i%2 != 0 {
			wr, err := zstd.NewWriter(buf)
			assert.NoError(t, err)
			_, err = io.CopyN(wr, rand.Reader, 10)
			assert.NoError(t, err)
			err = wr.Close()
			assert.NoError(t, err)
			job.CompressedContents = buf.Bytes()
			assert.NotEmpty(t, job.CompressedContents)
		} else {
			_, err := io.CopyN(buf, rand.Reader, 10)
			assert.NoError(t, err)
			job.Contents = buf.Bytes()
		}

		upbatch[i] = &job
	}
	store.UploadBatch(ctx, db, upbatch, now)
	for _, job := range upbatch {
		assert.NoError(t, job.Error)
	}
	store.UploadBatch(ctx, db, upbatch, now)
	for _, job := range upbatch {
		assert.NoError(t, job.Error)
	}

	// download all
	downbatch := make([]*DownloadJob, 10)
	for i := range downbatch {
		var job DownloadJob

		job.Namespace = WindowsExecutable
		d, err := hex.DecodeString(upbatch[i].Row.Digest)
		assert.NoError(t, err)
		copy(job.Digest[:], d)
		downbatch[i] = &job
	}

	err := store.DownloadBatch(ctx, db, downbatch)
	assert.NoError(t, err)
	for i, job := range downbatch {
		assert.NoError(t, job.Error)
		if len(upbatch[i].Contents) > 0 {
			assert.Equal(t, job.Blob, upbatch[i].Contents)
		} else {
			dec, err := zstd.NewReader(nil)
			assert.NoError(t, err)
			data, err := dec.DecodeAll(upbatch[i].CompressedContents, nil)
			assert.NoError(t, err)
			assert.Equal(t, job.Blob, data)
		}
	}

	// dups
	downbatch = make([]*DownloadJob, 10)
	for i := range downbatch {
		var job DownloadJob

		job.Namespace = WindowsExecutable
		d, err := hex.DecodeString(upbatch[1].Row.Digest)
		assert.NoError(t, err)
		copy(job.Digest[:], d)
		downbatch[i] = &job
	}
	err = store.DownloadBatch(ctx, db, downbatch)
	assert.NoError(t, err)
	for _, job := range downbatch {
		assert.NoError(t, job.Error)
		assert.NotEmpty(t, job.Blob)
		if len(upbatch[1].Contents) > 0 {
			assert.Equal(t, upbatch[1].Contents, job.Blob)
		} else {
			dec, err := zstd.NewReader(nil)
			assert.NoError(t, err)
			data, err := dec.DecodeAll(upbatch[1].CompressedContents, nil)
			assert.NoError(t, err)
			assert.Equal(t, data, job.Blob)
		}
	}

	// both empty
	upbatch = make([]*UploadJob, 10)
	for i := range upbatch {
		var job UploadJob
		job.Namespace = WindowsExecutable
		upbatch[i] = &job
	}
	store.UploadBatch(ctx, db, upbatch, now)
	for _, job := range upbatch {
		assert.NoError(t, job.Error)
	}
	downbatch = make([]*DownloadJob, 10)
	for i := range downbatch {
		var job DownloadJob

		job.Namespace = WindowsExecutable
		d, err := hex.DecodeString(upbatch[1].Row.Digest)
		assert.NoError(t, err)
		copy(job.Digest[:], d)
		downbatch[i] = &job
	}
	err = store.DownloadBatch(ctx, db, downbatch)
	assert.NoError(t, err)
	for _, job := range downbatch {
		assert.NoError(t, job.Error)
		assert.Empty(t, job.Blob)
	}
}

func testDownloadBatch(t *testing.T, db *pgxpool.Pool, store *Storage) {
	ctx := context.Background()

	// empty batch
	err := store.DownloadBatch(ctx, db, nil)
	assert.NoError(t, err)

	// non existent
	downbatch := make([]*DownloadJob, 10)
	for i := range downbatch {
		var job DownloadJob

		job.Namespace = WindowsExecutable
		_, err = io.ReadFull(rand.Reader, job.Digest[:])
		assert.NoError(t, err)
		downbatch[i] = &job
	}
	err = store.DownloadBatch(ctx, db, downbatch)
	assert.Error(t, NotFoundErr, err)

	// wrong namespace
	now := time.Now()
	upbatch := make([]*UploadJob, 10)
	for i := range upbatch {
		var job UploadJob
		job.Namespace = WindowsExecutable
		buf := new(bytes.Buffer)

		if i%2 != 0 {
			wr, err := zstd.NewWriter(buf)
			assert.NoError(t, err)
			_, err = io.CopyN(wr, rand.Reader, 10)
			assert.NoError(t, err)
			err = wr.Close()
			assert.NoError(t, err)
			job.CompressedContents = buf.Bytes()
			assert.NotEmpty(t, job.CompressedContents)
		} else {
			_, err := io.CopyN(buf, rand.Reader, 10)
			assert.NoError(t, err)
			job.Contents = buf.Bytes()
		}

		upbatch[i] = &job
	}
	store.UploadBatch(ctx, db, upbatch, now)
	for _, job := range upbatch {
		assert.NoError(t, job.Error)
	}
	downbatch = make([]*DownloadJob, 10)
	for i := range downbatch {
		var job DownloadJob

		if i%2 != 0 {
			job.Namespace = WindowsExecutable
		} else {
			job.Namespace = BIOS
		}
		d, err := hex.DecodeString(upbatch[i].Row.Digest)
		assert.NoError(t, err)
		copy(job.Digest[:], d)
		downbatch[i] = &job
	}
	err = store.DownloadBatch(ctx, db, downbatch)
	assert.NoError(t, err)
	for i, job := range downbatch {
		if i%2 != 0 {
			assert.NoError(t, job.Error)
			assert.NotEmpty(t, job.Blob)
		} else {
			assert.Error(t, NotFoundErr, job.Error)
			assert.Empty(t, job.Blob)
		}
	}
}

type storageFunctionsMock struct {
	store *Storage
	db    *pgxpool.Pool
}

func (st *storageFunctionsMock) uploadFile(ctx context.Context, row *Row, contents []byte, bucket string, client *s3.Client) error {
	fn := st.store.functions
	st.store.functions = new(storageFunctionsImpl)
	st.store.Upload(ctx, st.db, row.Namespace, contents, time.Now())
	err := new(storageFunctionsImpl).uploadFile(ctx, row, contents, bucket, client)
	st.store.functions = fn
	return err
}

func (st *storageFunctionsMock) ping(ctx context.Context, bucket string, client *s3.Client) error {
	return nil
}

func testCollision(t *testing.T, db *pgxpool.Pool, store *Storage) {
	ctx := context.Background()
	now := time.Now()
	b := []byte{1, 2, 3}
	h := sha256.Sum256(b)

	// simulate race-condition
	store.functions = &storageFunctionsMock{store, db}
	row1, err := store.Upload(ctx, db, BIOS, b, now)
	assert.NoError(t, err)

	var rows []Row
	err = pgxscan.Select(ctx, db, &rows, "select * from v2.blobs")
	assert.NoError(t, err)
	assert.Len(t, rows, 2)
	assert.NotEqual(t, rows[0].Snowflake, rows[1].Snowflake)
	assert.NotEqual(t, rows[0].Id, rows[1].Id)
	assert.Equal(t, rows[0].Digest, rows[1].Digest)
	assert.True(t, rows[0].Id == row1.Id || rows[1].Id == row1.Id)

	row2, err := store.Fetch(ctx, db, BIOS, h)
	assert.NoError(t, err)
	assert.True(t, rows[0].Id == row2.Id || rows[1].Id == row2.Id)
}
