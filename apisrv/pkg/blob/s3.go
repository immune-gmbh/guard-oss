package blob

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	awscred "github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3ty "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/klauspost/compress/zstd"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/database"
	tel "github.com/immune-gmbh/guard/apisrv/v2/pkg/telemetry"
)

const (
	defaultBucket   = "default"
	defaultRegion   = "fra1"
	defaultEndpoint = "https://fra1.digitaloceanspaces.com"
)

var (
	NotFoundErr         = errors.New("not found")
	UnknownMetadataErr  = errors.New("unknown metadata type")
	BucketPermissionErr = errors.New("insufficient bucket permissions")
)

// Content-addressable storage for blobs backed by an S3 compatible service.
type Storage struct {
	key       string
	secret    string
	client    *s3.Client
	Config    aws.Config
	Bucket    string
	functions storageFunctions
}

// some functions are moved into an interface to allow mocking
type storageFunctions interface {
	ping(context.Context, string, *s3.Client) error
	uploadFile(context.Context, *Row, []byte, string, *s3.Client) error
}

// concrete impl of above
type storageFunctionsImpl struct{}

type WithCredentials struct {
	Key    string
	Secret string
}

type WithEndpoint struct {
	Endpoint string
	Region   string
}

type WithBucket struct {
	Bucket string
}

type WithTestMode struct {
	Enable bool
}

// Creates a new Storage. Possible options are With*
func NewStorage(ctx context.Context, opts ...interface{}) (*Storage, error) {
	var key, secret, endpoint, region string
	var testmode bool
	st := new(Storage)

	st.Bucket = defaultBucket
	st.functions = &storageFunctionsImpl{}
	endpoint = defaultEndpoint
	region = defaultRegion

	for _, opt := range opts {
		switch opt := opt.(type) {
		case WithCredentials:
			if secret != "" || key != "" {
				tel.Log(ctx).Warn("blob storage credentials reset")
			}
			key = opt.Key
			secret = opt.Secret

		case WithEndpoint:
			region = opt.Region
			endpoint = opt.Endpoint

		case WithTestMode:
			testmode = opt.Enable

		case WithBucket:
			st.Bucket = opt.Bucket

		default:
			tel.Log(ctx).WithField("option", opt).Error("skipping blob storage option")
		}
	}

	config, err := awscfg.LoadDefaultConfig(ctx,
		awscfg.WithRegion(region),
		awscfg.WithCredentialsProvider(awscred.StaticCredentialsProvider{
			Value: aws.Credentials{
				AccessKeyID: key, SecretAccessKey: secret,
				Source: "apisrv.yaml",
			}}),
		awscfg.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{URL: endpoint}, nil
			})))
	if err != nil {
		tel.Log(ctx).WithError(err).Error("load aws default config")
		return nil, err
	}
	st.Config = config.Copy()
	// telemetry
	otelaws.AppendMiddlewares(&config.APIOptions)
	st.client = s3.NewFromConfig(config, func(opts *s3.Options) {
		opts.UsePathStyle = testmode
	})

	return st, nil
}

func (*storageFunctionsImpl) uploadFile(ctx context.Context, row *Row, contents []byte, bucket string, client *s3.Client) error {
	object := s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(row.Filename()),
		Body:   bytes.NewReader(contents),
		ACL:    s3ty.ObjectCannedACLPrivate,
		Metadata: map[string]string{
			"x-amz-meta-imn-digest": row.Digest,
		},
	}
	_, err := client.PutObject(ctx, &object)
	return err
}

func (*storageFunctionsImpl) ping(ctx context.Context, bucket string, client *s3.Client) error {
	input := s3.GetBucketAclInput{
		Bucket: aws.String(bucket),
	}
	output, err := client.GetBucketAcl(ctx, &input)
	if err != nil {
		return err
	}
	var read, write bool
	for _, grant := range output.Grants {
		write = write || grant.Permission == s3ty.PermissionWrite
		write = write || grant.Permission == s3ty.PermissionFullControl
		read = read || grant.Permission == s3ty.PermissionRead
		read = read || grant.Permission == s3ty.PermissionFullControl
	}

	if !read || !write {
		tel.Log(ctx).WithFields(map[string]interface{}{"read": read, "write": write}).Error("bucket permissions")
		return BucketPermissionErr
	} else {
		return nil
	}
}

// Trys to retrieve the metadata for a blob with SHA-256 hash `digest`. Blobs
// are sorted into namespaces. Deduplication only happens inside namespaces.
// May returns NotFoundErr.
func (st *Storage) Fetch(ctx context.Context, q pgxscan.Querier, namespace string, digest [32]byte) (*Row, error) {
	rows := make([]Row, 0)
	err := pgxscan.Select(ctx, q, &rows, `
		select * from v2.blobs where namespace = $1 and digest = $2
`, namespace, hex.EncodeToString(digest[:]))
	if err != nil {
		return nil, err
	} else if len(rows) == 0 {
		return nil, NotFoundErr
	} else if len(rows) == 1 {
		return &rows[0], nil
	} else {
		tel.Log(ctx).WithField("rows", len(rows)).Warn("fetch returned more than one blob")
		return &rows[0], nil
	}
}

// Writes a new blob into the store, retuning its metadata row. In case a blob
// with the same contents already exist in the specific namespace, the function
// *may* returns its row instead of creating a new one. The function contains a
// intended race-condition between uploading the blob's contents to S3 and
// inserting the metadata row. To prevent collisions each blob contains a
// random `snowflake` that is guaranteed to be unique. Blobs with the same
// contents but different snowflakes could be cleaned up by a garbage
// collection job.
func (st *Storage) Upload(ctx context.Context, q pgxscan.Querier, namespace string, contents []byte, now time.Time) (*Row, error) {
	ctx, span := tel.Start(ctx, "blob.Storage.Upload")
	defer span.End()

	chksm := sha256.Sum256(contents)
	row, err := st.Fetch(ctx, q, namespace, chksm)
	if err == nil {
		return row, nil
	} else if err != NotFoundErr {
		return nil, err
	}

	metadata := Metadata{Type: MetadataType}
	rawMetadata, err := metadata.ToRow()
	if err != nil {
		return nil, err
	}
	snowflake := make([]byte, 4)
	_, err = rand.Reader.Read(snowflake)
	if err != nil {
		return nil, err
	}
	row = &Row{
		Digest:      hex.EncodeToString(chksm[:]),
		Snowflake:   base64.RawURLEncoding.EncodeToString(snowflake),
		Namespace:   namespace,
		RawMetadata: rawMetadata,
		CreatedAt:   now,
	}

	err = st.functions.uploadFile(ctx, row, contents, st.Bucket, st.client)
	if err != nil {
		return nil, err
	}

	err = pgxscan.Get(ctx, q, &row.Id, `
		insert into v2.blobs (digest, snowflake, namespace, metadata, created_at)
		values ($1, $2, $3, $4, $5)
		returning id
	`, row.Digest, row.Snowflake, row.Namespace, row.RawMetadata, row.CreatedAt)
	if err != nil {
		return nil, err
	}

	return row, nil
}

// Fetch the contents of the blob pointed to by the metadata row and write it
// into `wr`.
func (st *Storage) Download(ctx context.Context, row *Row, wr io.Writer) error {
	input := &s3.GetObjectInput{
		Bucket: aws.String(st.Bucket),
		Key:    aws.String(row.Filename()),
	}
	result, err := st.client.GetObject(ctx, input)
	if err != nil {
		return err
	}
	_, err = io.Copy(wr, result.Body)
	if err != nil {
		return err
	}

	return nil
}

type DownloadJob struct {
	// filled by caller
	Namespace string
	Digest    [32]byte
	// filled by callee
	Error error
	Blob  []byte
}

func (st *Storage) DownloadBatch(ctx context.Context, q pgxscan.Querier, batch []*DownloadJob) error {
	if len(batch) == 0 {
		return nil
	}

	var wg sync.WaitGroup

	var str []string
	var args []interface{}
	for i := range batch {
		str = append(str, fmt.Sprintf("($%d::text, $%d::text)", i*2+1, i*2+2))
		args = append(args, batch[i].Namespace, fmt.Sprintf("%x", batch[i].Digest[:]))
	}
	sql := fmt.Sprint("select * from v2.blobs where (namespace, digest) in (", strings.Join(str, ","), ")")
	rows := make([]Row, 0)
	err := pgxscan.Select(ctx, q, &rows, sql, args...)
	if err != nil {
		return database.Error(err)
	}
	var rowmap map[string]*Row = make(map[string]*Row)
	for _, row := range rows {
		row := row
		rowmap[fmt.Sprintf("%s-%s", row.Namespace, row.Digest)] = &row
	}

	for i := range batch {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx, span := tel.Start(ctx, "Batch Download")
			defer span.End()

			job := batch[i]
			wr := new(bytes.Buffer)
			row, ok := rowmap[fmt.Sprintf("%s-%x", job.Namespace, job.Digest)]
			if !ok {
				job.Error = NotFoundErr
			} else {
				job.Error = st.Download(ctx, row, wr)
				job.Blob = wr.Bytes()
			}
		}()
	}

	wg.Wait()
	return nil
}

type UploadJob struct {
	// filled by caller
	Contents           []byte
	CompressedContents []byte
	Namespace          string
	// filled by callee
	Error error
	Row   *Row
}

func (st *Storage) UploadBatch(ctx context.Context, pool *pgxpool.Pool, batch []*UploadJob, now time.Time) {
	var wg sync.WaitGroup

	for i := range batch {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx, span := tel.Start(ctx, "Batch Upload")
			defer span.End()

			job := batch[i]
			if len(job.Contents) == 0 && len(job.CompressedContents) > 0 {
				var dec *zstd.Decoder
				dec, job.Error = zstd.NewReader(bytes.NewBuffer(job.CompressedContents))
				if job.Error != nil {
					tel.Log(ctx).WithError(job.Error).Error("create zstd decoder")
					return
				}
				buf := bytes.NewBuffer(nil)
				_, job.Error = io.Copy(buf, dec)
				if job.Error != nil {
					tel.Log(ctx).WithError(job.Error).Error("decompress data")
					return
				}
				defer dec.Close()
				job.Contents = buf.Bytes()
			}

			job.Row, job.Error = st.Upload(ctx, pool, job.Namespace, job.Contents, now)
		}()
	}

	wg.Wait()
	return
}

func (st *Storage) Ping(ctx context.Context) error {
	return st.functions.ping(ctx, st.Bucket, st.client)
}
