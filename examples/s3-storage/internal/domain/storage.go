package domain // доменный порт хранилища объектов без привязки к AWS

import (
	"context" // контекст для отмены операций Put/Get
	"errors"  // ErrNotFound
	"io"      // потоки чтения тела объекта
)

var ErrNotFound = errors.New("object not found") // объект отсутствует в хранилище (маппится с 404)

// ObjectStorage is a domain port (implemented in infrastructure).
type ObjectStorage interface {
	Put(ctx context.Context, key string, body io.Reader, size int64, contentType string) error
	Get(ctx context.Context, key string) (io.ReadCloser, int64, error)
}
