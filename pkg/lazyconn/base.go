package lazyconn

import (
	"context"
	"log/slog"
	"sync"
)

type LazyClient[T any] struct {
	loadFunc func(ctx context.Context) (T, error)
}

func NewLazyClient[T any](initFunc func(ctx context.Context) (T, error)) LazyClient[T] {
	var (
		once   sync.Once
		client T
		err    error
	)
	loadFunc := func(ctx context.Context) (T, error) {
		once.Do(func() {
			client, err = initFunc(ctx)
		})
		return client, err
	}

	return LazyClient[T]{
		loadFunc: loadFunc,
	}
}

func (c LazyClient[T]) Load(ctx context.Context) (T, error) {
	return c.loadFunc(ctx)
}

func (c LazyClient[T]) Preload(ctx context.Context) {
	go func() {
		if _, err := c.loadFunc(ctx); err != nil {
			slog.Error(err.Error())
		}
	}()
}
