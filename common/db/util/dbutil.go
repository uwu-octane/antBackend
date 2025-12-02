package commonutil

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type FallbackHook func(ctx context.Context, conn sqlx.SqlConn) error

type Selector struct {
	First                       sqlx.SqlConn
	Backup                      sqlx.SqlConn
	ReadFromReplica             bool
	FallbackToMasterOnReadError bool
	OnFallback                  FallbackHook
}

func NewSelector(first sqlx.SqlConn, backup sqlx.SqlConn, readFromReplica bool, fallbackToMasterOnReadError bool, onFallback FallbackHook) *Selector {
	return &Selector{
		First:                       first,
		Backup:                      backup,
		ReadFromReplica:             readFromReplica,
		FallbackToMasterOnReadError: fallbackToMasterOnReadError,
		OnFallback:                  onFallback,
	}
}

func (s *Selector) WithFallback(hook FallbackHook) *Selector {
	s.OnFallback = hook
	return s
}

func (s *Selector) Do(ctx context.Context, fn func(ctx context.Context, conn sqlx.SqlConn) error) error {
	first, backup := s.First, s.Backup
	if !s.ReadFromReplica {
		first, backup = backup, first
	}
	if err := fn(ctx, first); err != nil {
		if !s.FallbackToMasterOnReadError {
			return err
		}
		if s.OnFallback != nil {
			if fallbackErr := s.OnFallback(ctx, backup); fallbackErr != nil {
				// Log fallback hook error but continue with fallback
				_ = fallbackErr
			}
		}
		return fn(ctx, backup)
	}
	return nil
}
