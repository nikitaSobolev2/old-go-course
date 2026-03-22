package storage // реализация domain.ObjectStorage через AWS SDK v2 S3

import (
	"context" // отмена Put/Get
	"errors"  // errors.As для разбора HTTP-ошибки S3
	"io"      // поток тела Put/GetObject

	"github.com/aws/aws-sdk-go-v2/aws"                        // String/Int64 для указателей полей запроса
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http" // ResponseError со статусом HTTP
	"github.com/aws/aws-sdk-go-v2/service/s3"                 // клиент PutObject/GetObject

	"github.com/example/go-examples/s3-storage/internal/domain" // ErrNotFound для 404
)

type S3Storage struct {
	client *s3.Client
	bucket string
}

func NewS3Storage(client *s3.Client, bucket string) *S3Storage {
	return &S3Storage{client: client, bucket: bucket} // клиент уже настроен на endpoint и credentials
}

func (s *S3Storage) Put(ctx context.Context, key string, body io.Reader, size int64, contentType string) error {
	in := &s3.PutObjectInput{ // параметры загрузки объекта
		Bucket:      aws.String(s.bucket), // имя бакета
		Key:         aws.String(key), // ключ объекта в бакете
		Body:        body, // поток данных
		ContentType: aws.String(contentType), // MIME для объекта
	}
	if size > 0 { // при известной длине — помогаем SDK выставить Content-Length
		in.ContentLength = aws.Int64(size)
	}
	_, err := s.client.PutObject(ctx, in) // выполняем запрос к S3/MinIO
	return err
}

func (s *S3Storage) Get(ctx context.Context, key string) (io.ReadCloser, int64, error) {
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{ // скачивание по bucket+key
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var re *awshttp.ResponseError // тип ошибки с HTTP-статусом
		if errors.As(err, &re) && re.HTTPStatusCode() == 404 { // объект не найден
			return nil, 0, domain.ErrNotFound // унифицируем для application/HTTP
		}
		return nil, 0, err // сетевая или иная ошибка SDK
	}
	return out.Body, *out.ContentLength, nil // тело и длина для заголовков ответа
}
