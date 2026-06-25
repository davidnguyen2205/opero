// Package storage provides object storage for uploaded media (e.g. check-in
// photos), backed by an S3-compatible service (MinIO in local/dev).
//
// It is platform infrastructure, not a domain module: it exposes a small
// interface so handlers can depend on the behavior and tests can fake it.
package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// ObjectStore stores opaque objects under a key and produces a retrievable URL.
//
// Tenancy note: object keys are namespaced by tenant by the caller (see the
// media handler). This package never opens a tenant DB and never resolves a
// tenant itself — it only stores bytes under the key it is given.
type ObjectStore interface {
	// Upload writes size bytes read from r under key with the given contentType.
	Upload(ctx context.Context, key, contentType string, size int64, r io.Reader) error
	// URL returns a URL that can be used to retrieve the object at key.
	URL(ctx context.Context, key string) (string, error)
}

// Config holds the settings needed to reach the object store.
type Config struct {
	Endpoint  string // host:port of the S3 API (no scheme)
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool

	// PublicURL, if set, is used as the base for returned object URLs instead
	// of presigning. Empty means presigned GET URLs are returned.
	PublicURL string

	// PresignExpiry is the validity window for presigned GET URLs.
	PresignExpiry time.Duration
}

// MinioStore is the MinIO/S3 implementation of ObjectStore.
type MinioStore struct {
	client  *minio.Client
	bucket  string
	pubURL  string
	expires time.Duration
}

// NewMinioStore constructs a MinioStore from cfg. It does not create the bucket
// (that is handled by the compose init job / provisioning); it only opens a
// client.
func NewMinioStore(cfg Config) (*MinioStore, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("storage: new minio client: %w", err)
	}
	expires := cfg.PresignExpiry
	if expires <= 0 {
		expires = 7 * 24 * time.Hour
	}
	return &MinioStore{
		client:  client,
		bucket:  cfg.Bucket,
		pubURL:  cfg.PublicURL,
		expires: expires,
	}, nil
}

// Upload implements ObjectStore.
func (s *MinioStore) Upload(ctx context.Context, key, contentType string, size int64, r io.Reader) error {
	_, err := s.client.PutObject(ctx, s.bucket, key, r, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("storage: put object: %w", err)
	}
	return nil
}

// URL implements ObjectStore. If a public base URL is configured it returns
// "<public>/<bucket>/<key>"; otherwise it returns a presigned GET URL valid for
// the configured expiry.
func (s *MinioStore) URL(ctx context.Context, key string) (string, error) {
	if s.pubURL != "" {
		base, err := url.Parse(s.pubURL)
		if err != nil {
			return "", fmt.Errorf("storage: parse public url: %w", err)
		}
		base.Path = joinPath(base.Path, s.bucket, key)
		return base.String(), nil
	}
	u, err := s.client.PresignedGetObject(ctx, s.bucket, key, s.expires, nil)
	if err != nil {
		return "", fmt.Errorf("storage: presign get: %w", err)
	}
	return u.String(), nil
}

func joinPath(parts ...string) string {
	out := ""
	for _, p := range parts {
		for len(p) > 0 && p[0] == '/' {
			p = p[1:]
		}
		for len(p) > 0 && p[len(p)-1] == '/' {
			p = p[:len(p)-1]
		}
		if p == "" {
			continue
		}
		out += "/" + p
	}
	return out
}
