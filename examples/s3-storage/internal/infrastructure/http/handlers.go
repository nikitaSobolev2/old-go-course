package httpapi // HTTP-адаптер: REST поверх FileService

import (
	"errors"   // errors.Is для domain.ErrNotFound
	"io"       // копирование тела ответа из ReadCloser
	"net/http" // ResponseWriter, Request, коды статусов
	"strconv"  // форматирование Content-Length

	"github.com/go-chi/chi/v5" // роутер и URL-параметр {key}

	"github.com/example/go-examples/s3-storage/internal/application" // сервис загрузки/скачивания
	"github.com/example/go-examples/s3-storage/internal/domain"      // ErrNotFound
)

type Handlers struct {
	svc *application.FileService
}

func NewHandlers(svc *application.FileService) *Handlers {
	return &Handlers{svc: svc} // сохраняем сервис для хендлеров
}

func (h *Handlers) Routes() chi.Router {
	r := chi.NewRouter() // новый роутер chi
	r.Put("/objects/{key}", h.putObject) // загрузка: тело запроса = содержимое объекта
	r.Get("/objects/{key}", h.getObject) // скачивание: отдаём поток и длину
	return r
}

func (h *Handlers) putObject(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key") // извлекаем key из пути
	if key == "" { // защита от пустого сегмента
		http.Error(w, "missing key", http.StatusBadRequest) // 400
		return
	}
	size := r.ContentLength // размер из заголовка (если клиент передал)
	if size < 0 { // неизвестная длина — SDK может использовать chunked
		size = 0
	}
	ct := r.Header.Get("Content-Type") // тип содержимого для S3 metadata
	if ct == "" { // по умолчанию бинарный поток
		ct = "application/octet-stream"
	}
	if err := h.svc.Upload(r.Context(), key, r.Body, size, ct); err != nil { // читаем r.Body до EOF
		http.Error(w, err.Error(), http.StatusInternalServerError) // 500 при ошибке S3/валидации
		return
	}
	w.WriteHeader(http.StatusNoContent) // успешная загрузка без тела ответа
}

func (h *Handlers) getObject(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key") // ключ объекта из URL
	rc, n, err := h.svc.Download(r.Context(), key) // поток тела и размер
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) { // маппинг доменной ошибки на 404
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError) // прочие ошибки — 500
		return
	}
	defer func() { _ = rc.Close() }() // закрываем тело ответа S3 после копирования
	if n > 0 { // если размер известен — клиент может показать прогресс
		w.Header().Set("Content-Length", strconv.FormatInt(n, 10))
	}
	w.Header().Set("Content-Type", "application/octet-stream") // отдаём как произвольные байты
	_, _ = io.Copy(w, rc) // стримим в ответ клиенту
}
