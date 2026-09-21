// Package storage puts user-uploaded images into S3-compatible object storage.
package storage

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// ErrDisabled is returned when object storage is not configured.
var ErrDisabled = errors.New("object storage disabled")

// Kind is the folder an object goes to.
type Kind string

const (
	KindCover  Kind = "covers"
	KindIcon   Kind = "icons"
	KindLesson Kind = "lessons"
)

type Config struct {
	Endpoint  string // host:port of the S3 API
	Bucket    string
	AccessKey string
	SecretKey string
	UseSSL    bool
	PublicURL string // base URL the browser uses, e.g. https://cdn.example.com/golearn
}

type Store struct {
	client    *minio.Client
	bucket    string
	publicURL string

	mu    sync.Mutex
	ready bool // the bucket exists and is publicly readable
}

// LoadConfig reads the S3 settings; an empty endpoint means storage is disabled.
func LoadConfig() Config {
	return Config{
		Endpoint:  os.Getenv("S3_ENDPOINT"),
		Bucket:    envOr("S3_BUCKET", "golearn"),
		AccessKey: os.Getenv("S3_ACCESS_KEY"),
		SecretKey: os.Getenv("S3_SECRET_KEY"),
		UseSSL:    os.Getenv("S3_USE_SSL") == "true",
		PublicURL: strings.TrimSuffix(os.Getenv("S3_PUBLIC_URL"), "/"),
	}
}

// New builds the client for the configured bucket. It does not reach the server:
// the bucket is prepared on the first upload, so storage that starts after the
// server still works (Warm reports whether it is up already).
func New(cfg Config) (*Store, error) {
	if cfg.Endpoint == "" {
		return nil, ErrDisabled
	}
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("s3 client: %w", err)
	}

	public := cfg.PublicURL
	if public == "" {
		scheme := "http"
		if cfg.UseSSL {
			scheme = "https"
		}
		public = fmt.Sprintf("%s://%s/%s", scheme, cfg.Endpoint, cfg.Bucket)
	}

	return &Store{client: client, bucket: cfg.Bucket, publicURL: public}, nil
}

// Warm prepares the bucket ahead of the first upload; the caller only logs its error.
func (s *Store) Warm(ctx context.Context) error { return s.ensureReady(ctx) }

// ensureReady creates the bucket and its policy once, retrying on every call until it succeeds.
func (s *Store) ensureReady(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.ready {
		return nil
	}
	if err := s.ensureBucket(ctx); err != nil {
		return err
	}
	s.ready = true
	return nil
}

func (s *Store) ensureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("check bucket: %w", err)
	}
	if !exists {
		if err := s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{}); err != nil {
			return fmt.Errorf("create bucket: %w", err)
		}
	}
	return s.client.SetBucketPolicy(ctx, s.bucket, publicReadPolicy(s.bucket))
}

func publicReadPolicy(bucket string) string {
	return `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"AWS":["*"]},` +
		`"Action":["s3:GetObject"],"Resource":["arn:aws:s3:::` + bucket + `/*"]}]}`
}

// Put stores the image under a fresh key and returns its public URL. Keys are
// random, not content-addressed, so deleting one owner's image never breaks another's.
func (s *Store) Put(ctx context.Context, kind Kind, img Image) (string, error) {
	if err := s.ensureReady(ctx); err != nil {
		return "", err
	}
	name, err := randomName()
	if err != nil {
		return "", err
	}
	key := fmt.Sprintf("%s/%s%s", kind, name, img.Extension())

	opts := minio.PutObjectOptions{
		ContentType:  img.ContentType,
		CacheControl: "public, max-age=31536000, immutable",
	}
	// An SVG served inline could run scripts on the bucket origin; force a download instead.
	if img.ContentType == MimeSVG {
		opts.ContentDisposition = "attachment"
	}

	if _, err := s.client.PutObject(ctx, s.bucket, key, img.Reader(), int64(len(img.Data)), opts); err != nil {
		return "", fmt.Errorf("put object: %w", err)
	}
	return s.publicURL + "/" + key, nil
}

// Delete removes an object previously returned by Put; other URLs are ignored.
func (s *Store) Delete(ctx context.Context, publicURL string) error {
	key, ok := s.keyOf(publicURL)
	if !ok {
		return nil
	}
	if err := s.ensureReady(ctx); err != nil {
		return err
	}
	return s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{})
}

// Owns reports whether the URL points at this bucket.
func (s *Store) Owns(publicURL string) bool {
	_, ok := s.keyOf(publicURL)
	return ok
}

func (s *Store) keyOf(publicURL string) (string, bool) {
	prefix := s.publicURL + "/"
	if !strings.HasPrefix(publicURL, prefix) {
		return "", false
	}
	key, err := url.PathUnescape(strings.TrimPrefix(publicURL, prefix))
	if err != nil || key == "" {
		return "", false
	}
	return key, true
}

func randomName() (string, error) {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", fmt.Errorf("random name: %w", err)
	}
	return hex.EncodeToString(buf[:]), nil
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
