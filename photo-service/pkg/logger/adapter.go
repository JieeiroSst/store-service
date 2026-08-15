package logger

import (
	"github.com/JIeeiroSst/photo-service/internal/application/ports"
	"go.uber.org/zap"
)

type zapAdapter struct {
	l *zap.SugaredLogger
}

func NewPort(l *zap.Logger) ports.Logger {
	return &zapAdapter{l: l.Sugar()}
}

func (a *zapAdapter) Debug(msg string, kv ...any) { a.l.Debugw(msg, kv...) }
func (a *zapAdapter) Info(msg string, kv ...any)  { a.l.Infow(msg, kv...) }
func (a *zapAdapter) Warn(msg string, kv ...any)  { a.l.Warnw(msg, kv...) }
func (a *zapAdapter) Error(msg string, kv ...any) { a.l.Errorw(msg, kv...) }
