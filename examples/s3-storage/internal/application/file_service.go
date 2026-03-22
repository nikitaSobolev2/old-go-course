package application // use case: загрузка и скачивание файлов через доменный порт

import (
	"context" // отмена и таймауты при вызовах хранилища
	"errors"  // конструктор ErrEmptyKey
	"io"      // поток тела при загрузке/скачивании
	"strings" // TrimSpace для ключа

	"github.com/example/go-examples/s3-storage/internal/domain" // интерфейс ObjectStorage
)

// ErrEmptyKey is returned when object key is blank.
var ErrEmptyKey = errors.New("empty key") // ошибка валидации: пустой ключ после trim

type FileService struct {
	store domain.ObjectStorage
}

func NewFileService(store domain.ObjectStorage) *FileService {
	return &FileService{store: store} // внедряем реализацию хранилища (S3)
}

func (s *FileService) Upload(ctx context.Context, key string, body io.Reader, size int64, contentType string) error {
	key = strings.TrimSpace(key) // убираем пробелы по краям
	if key == "" { // ключ обязателен
		return ErrEmptyKey
	}
	return s.store.Put(ctx, key, body, size, contentType) // делегируем инфраструктуре
}

func (s *FileService) Download(ctx context.Context, key string) (io.ReadCloser, int64, error) {
	return s.store.Get(ctx, strings.TrimSpace(key)) // чтение объекта и размер для Content-Length
}
