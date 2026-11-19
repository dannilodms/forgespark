// Package s3 keeps the bytes of the files in an S3 storage, through the AWS
// SDK for Go v2 and only what the S3 API defines (Files in
// CODING_STANDARDS.md): Garage in development and staging, a provider in
// production. The browser sends the bytes itself, and reads them, through URLs
// this adapter signs for the public host name of the storage; the API only
// reads the start of an object, copies one inside the storage and removes
// one, and the worker lists the bucket to clean it up.
package s3

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go/tracing/smithyoteltracing"
	"go.opentelemetry.io/otel/trace"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/files/domain"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/config"
)

// callTimeout bounds each call to the storage, below the drain deadline of
// the shutdown.
const callTimeout = 10 * time.Second

// Settings say where the storage is and how to sign for it.
type Settings struct {
	// URL is where the API reaches the storage.
	URL string
	// PublicURL is where the browser reaches it. A signed URL is signed for
	// this host name, which the signature covers.
	PublicURL string
	// Region is the region the storage signs with.
	Region string
	// Bucket is the one private bucket of the files.
	Bucket string
	// AccessKeyID and SecretAccessKey are the key the API signs with. The
	// secret prints as a placeholder, and only the SDK receives it.
	AccessKeyID     string
	SecretAccessKey config.Secret
	// TracerProvider traces each call to the storage. Nil traces nothing.
	TracerProvider trace.TracerProvider
}

// Storage is the file storage of the files area.
type Storage struct {
	client *s3.Client
	signer *s3.PresignClient
	bucket string
	// listPageSize is how many keys a page of the listing asks for. Zero
	// leaves it to the storage, which gives a thousand at most.
	listPageSize int32
}

// NewStorage returns the storage these settings name. It reaches nothing
// until it is used.
func NewStorage(settings Settings) *Storage {
	return &Storage{
		client:       s3.New(optionsFor(settings, settings.URL)),
		signer:       s3.NewPresignClient(s3.New(optionsFor(settings, settings.PublicURL))),
		bucket:       settings.Bucket,
		listPageSize: 0,
	}
}

// optionsFor gives the options of a client that reaches the storage at this
// address. The bucket is in the path of each address, never in the host
// name, so one host name serves it.
func optionsFor(settings Settings, endpoint string) s3.Options {
	options := s3.Options{
		BaseEndpoint: aws.String(endpoint),
		UsePathStyle: true,
		Region:       settings.Region,
		Credentials:  credentials.NewStaticCredentialsProvider(settings.AccessKeyID, settings.SecretAccessKey.Reveal(), ""),
		HTTPClient:   awshttp.NewBuildableClient().WithTimeout(callTimeout),
	}
	if settings.TracerProvider != nil {
		// The spans of the SDK carry the method, the endpoint, the status and
		// the error of a call: never a signature, nor the key of an object.
		options.TracerProvider = keylessTracing{TracerProvider: smithyoteltracing.Adapt(settings.TracerProvider)}
	}
	return options
}

// SignUpload signs a PUT of the object under this key, valid for this long.
// The type and the size declared are signed with it, so the storage refuses a
// request that sends another Content-Type or another Content-Length.
func (s *Storage) SignUpload(ctx context.Context, key domain.ObjectKey, declared domain.DeclaredImage, validity time.Duration) (string, error) {
	signed, err := s.signer.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(string(key)),
		ContentType:   aws.String(string(declared.MediaType())),
		ContentLength: aws.Int64(declared.Size()),
	}, s3.WithPresignExpires(validity))
	if err != nil {
		return "", failed("sign the upload of an object", err)
	}
	return signed.URL, nil
}

// SignDownload signs a GET of the object under this key, valid for this long,
// which the browser sends to the public host name of the storage: an image
// shows from there, and its bytes never pass through the API.
func (s *Storage) SignDownload(ctx context.Context, key domain.ObjectKey, validity time.Duration) (string, error) {
	signed, err := s.signer.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(string(key)),
	}, s3.WithPresignExpires(validity))
	if err != nil {
		return "", failed("sign the download of an object", err)
	}
	return signed.URL, nil
}

// Inspect reads the size of the object under this key and its first bytes,
// at most domain.FirstBytesRead of them: never the whole object.
func (s *Storage) Inspect(ctx context.Context, key domain.ObjectKey) (domain.StoredObject, bool, error) {
	head, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(string(key))})
	if isNotFound(err) {
		return domain.StoredObject{}, false, nil
	}
	if err != nil {
		return domain.StoredObject{}, false, failed("read the size of an object", err)
	}
	size := aws.ToInt64(head.ContentLength)
	if size == 0 {
		return domain.StoredObject{Size: 0, FirstBytes: []byte{}}, true, nil
	}
	first, found, err := s.firstBytes(ctx, key)
	if !found || err != nil {
		return domain.StoredObject{}, found, err
	}
	return domain.StoredObject{Size: size, FirstBytes: first}, true, nil
}

func (s *Storage) firstBytes(ctx context.Context, key domain.ObjectKey) ([]byte, bool, error) {
	object, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(string(key)),
		Range:  aws.String(fmt.Sprintf("bytes=0-%d", domain.FirstBytesRead-1)),
	})
	// Removed since it was measured.
	if isNotFound(err) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, failed("read the first bytes of an object", err)
	}
	defer object.Body.Close()
	first, err := io.ReadAll(io.LimitReader(object.Body, domain.FirstBytesRead))
	if err != nil {
		return nil, false, failed("read the first bytes of an object", err)
	}
	return first, true, nil
}

// Copy copies the object under one key to another with CopyObject, inside
// the storage: the bytes never reach the API. A URL signed for the first key
// reaches the first key only.
func (s *Storage) Copy(ctx context.Context, from, to domain.ObjectKey) error {
	_, err := s.client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(s.bucket),
		Key:        aws.String(string(to)),
		CopySource: aws.String(url.PathEscape(s.bucket) + "/" + url.PathEscape(string(from))),
	})
	if err != nil {
		return failed("copy an object", err)
	}
	return nil
}

// Remove removes the object under this key. The S3 API answers the removal
// of an object that does not exist as a success.
func (s *Storage) Remove(ctx context.Context, key domain.ObjectKey) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(string(key))})
	if err != nil {
		return failed("remove an object", err)
	}
	return nil
}

// List lists the bucket with ListObjectsV2 and calls each with the objects of
// every page, in the order of their keys, each with the time it was last
// written. It stops at the first failure, of the storage or of each.
func (s *Storage) List(ctx context.Context, each func(objects []domain.ListedObject) error) error {
	pages := s3.NewListObjectsV2Paginator(s.client, &s3.ListObjectsV2Input{Bucket: aws.String(s.bucket)},
		func(options *s3.ListObjectsV2PaginatorOptions) { options.Limit = s.listPageSize })
	var err error
	for err == nil && pages.HasMorePages() {
		err = nextPage(ctx, pages, each)
	}
	return err
}

func nextPage(ctx context.Context, pages *s3.ListObjectsV2Paginator, each func(objects []domain.ListedObject) error) error {
	page, err := pages.NextPage(ctx)
	if err != nil {
		return failed("list the objects of the bucket", err)
	}
	objects := make([]domain.ListedObject, 0, len(page.Contents))
	for _, object := range page.Contents {
		objects = append(objects, domain.ListedObject{
			Key:          domain.ObjectKey(aws.ToString(object.Key)),
			LastModified: aws.ToTime(object.LastModified).UTC(),
		})
	}
	return each(objects)
}

// isNotFound tells whether the storage answered that there is no such object.
func isNotFound(err error) bool {
	var answered *awshttp.ResponseError
	return errors.As(err, &answered) && answered.HTTPStatusCode() == http.StatusNotFound
}
