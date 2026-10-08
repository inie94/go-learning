package middleware

import (
	"errors"
	"log/slog"

	"github.com/ogen-go/ogen/middleware"
)

// ─── Recovery ───────────────────────────────────────────────────────────────

// Recovery перехватывает паники в нижележащих обработчиках,
// логирует их и возвращает ошибку, которую ogen превратит в 500.
func Recovery(logger *slog.Logger) middleware.Middleware {
	return func(
		req middleware.Request,
		next func(req middleware.Request) (middleware.Response, error),
	) (middleware.Response, error) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("panic recovered",
					slog.Any("panic", r),
					slog.String("operation", req.OperationName),
					slog.String("operation_id", req.OperationID),
				)
			}
		}()

		resp, err := next(req)
		if err != nil {
			// Если нижележащий обработчик вернул ошибку,
			// оборачиваем её в internal‑ошибку, чтобы клиент
			// получил 500, а не детали реализации.
			return resp, errors.New("internal server error")
		}
		return resp, nil
	}
}
