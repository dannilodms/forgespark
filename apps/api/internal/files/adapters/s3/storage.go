// Package s3 keeps the bytes of the files in an S3 storage, through the AWS
// SDK for Go v2 and only what the S3 API defines (Files in
// CODING_STANDARDS.md): Garage in development and staging, a provider in
// production. The browser sends the bytes itself, to a URL this adapter signs
// for the public host name of the storage; the API only reads the start of an
// object and removes one.
package s3

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/files/domain"
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
	// AccessKeyID and SecretAccessKey are the key the API signs with.
	AccessKeyID     string
	SecretAccessKey string
}

// Storage is the file storage of the files area.
type Storage struct {
	client *s3.Client
	signer *s3.PresignClient
	bucket string
}

// NewStorage returns the storage these settings name. It reaches nothing
// until it is used.
func NewStorage(settings Settings) *Storage {
	return &Storage{
		client: s3.New(optionsFor(settings, settings.URL)),
		signer: s3.NewPresignClient(s3.New(optionsFor(settings, settings.PublicURL))),
		bucket: settings.Bucket,
	}
}

// optionsFor gives the options of a client that reaches the storage at this
// address. The bucket is in the path of each address, never in the host
// name, so one host name serves it.
func optionsFor(settings Settings, endpoint string) s3.Options {
	return s3.Options{
		BaseEndpoint: aws.String(endpoint),
		UsePathStyle: true,
		Region:       settings.Region,
		Credentials:  credentials.NewStaticCredentialsProvider(settings.AccessKeyID, settings.SecretAccessKey, ""),
		HTTPClient:   awshttp.NewBuildableClient().WithTimeout(callTimeout),
	}
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
		return "", fmt.Errorf("sign the upload of an object: %w", err)
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
		return domain.StoredObject{}, false, fmt.Errorf("read the size of an object: %w", err)
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
		return nil, false, fmt.Errorf("read the first bytes of an object: %w", err)
	}
	defer object.Body.Close()
	first, err := io.ReadAll(io.LimitReader(object.Body, domain.FirstBytesRead))
	if err != nil {
		return nil, false, fmt.Errorf("read the first bytes of an object: %w", err)
	}
	return first, true, nil
}

// Remove removes the object under this key. The S3 API answers the removal
// of an object that does not exist as a success.
func (s *Storage) Remove(ctx context.Context, key domain.ObjectKey) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(string(key))})
	if err != nil {
		return fmt.Errorf("remove an object: %w", err)
	}
	return nil
}

// isNotFound tells whether the storage answered that there is no such object.
func isNotFound(err error) bool {
	var answered *awshttp.ResponseError
	return errors.As(err, &answered) && answered.HTTPStatusCode() == http.StatusNotFound
}
