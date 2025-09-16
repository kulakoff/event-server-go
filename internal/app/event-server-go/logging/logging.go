package logging

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
)

func NewLogger() *slog.Logger {
	handler := &CallerHandler{
		Handler: slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug, // Уровень логирования
		}),
	}
	return slog.New(handler)
}

type CallerHandler struct {
	slog.Handler
}

func (h *CallerHandler) Handle(ctx context.Context, r slog.Record) error {
	// Получаем информацию о вызове
	pc, file, line, ok := runtime.Caller(4) // 4 уровня стека, чтобы дойти до места вызова logger.Debug и т.д.
	if ok {
		// Имя файла (только базовое имя, без полного пути)
		fileName := filepath.Base(file)
		// Имя пакета можно извлечь из пути, но для простоты используем fileName
		caller := slog.String("caller", fmt.Sprintf("%s:%d", fileName, line))

		// Добавляем атрибут в запись логов
		r.AddAttrs(caller)
	}
	return h.Handler.Handle(ctx, r)
}
