package middleware

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/ogen-go/ogen/middleware"
)

func Logger(logger *slog.Logger) middleware.Middleware {
	return func(
		req middleware.Request,
		next middleware.Next,
	) (middleware.Response, error) {
		start := time.Now()

		log := logger.With(
			slog.String("operation", req.OperationName),
			slog.String("operation_id", req.OperationID),
			slog.String("method", req.Raw.Method),
			slog.String("path", req.Raw.URL.Path),
			slog.String("remote_addr", req.Raw.RemoteAddr),
		)

		log.Info("handling request")

		resp, err := next(req)
		duration := time.Since(start)

		// Сначала пробуем достать статус из ответа.
		status, ok := statusFromResponse(resp)

		if err != nil {
			// Ответа нет/статус не извлёкся — берём из ошибки.
			if !ok {
				status = statusFromError(err)
			}

			log.Error("request failed",
				slog.String("error", err.Error()),
				slog.Int("status_code", status),
				slog.Duration("duration", duration),
			)
			return resp, err
		}

		// Успешный путь: статус обязательно должен быть.
		if !ok {
			status = http.StatusOK
		}

		log.LogAttrs(req.Raw.Context(), slog.LevelInfo, "request completed",
			slog.Int("status_code", status),
			slog.Duration("duration", duration),
		)
		return resp, nil
	}
}

// statusFromResponse пытается извлечь статус-код из ответа ogen.
// Сгенерированные типы ответов реализуют интерфейс GetStatusCode() int.
func statusFromResponse(resp middleware.Response) (int, bool) {
	if resp.Type == nil {
		return 0, false
	}
	if s, ok := resp.Type.(interface{ GetStatusCode() int }); ok {
		return s.GetStatusCode(), true
	}
	return 0, false
}

// statusFromError пытается извлечь статус-код из ошибки ogen.
// Сгенерированные типы ошибок реализуют интерфейс StatusCode() int
// (см. ogen.ErrorStatusCode). Если интерфейс не реализован,
// возвращаем 500 по умолчанию.
func statusFromError(err error) int {
	var withStatus interface{ StatusCode() int }
	if errors.As(err, &withStatus) {
		return withStatus.StatusCode()
	}
	return http.StatusInternalServerError
}