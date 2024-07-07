package generator

import (
	"context"
	"sync"
)

type modifier[K comparable, T any] func(key K, value T) (K, T)

func New[T any](generate func() (T, bool)) <-chan T {
	return NewWithContext[T](context.Background(), generate)
}

func NewWithContext[T any](ctx context.Context, generate func() (T, bool)) <-chan T {
	if generate == nil {
		panic("generate function cannot be nil")
	}

	ch := make(chan T, 1)

	go func() {
		defer close(ch)

		for {
			select {
			case <-ctx.Done():
				return
			default:
				value, ok := generate()
				if !ok {
					return
				}

				ch <- value
			}
		}
	}()

	return ch
}

func NewFromMap[K comparable, T any](source map[K]T, modify modifier[K, T]) (<-chan K, <-chan T) {
	return NewFromMapWithContext[K, T](context.Background(), source, modify)
}

func NewFromMapWithContext[K comparable, T any](
	ctx context.Context,
	source map[K]T,
	modify modifier[K, T],
) (<-chan K, <-chan T) {
	chK := make(chan K, 1)
	chT := make(chan T, 1)

	go func() {
		defer close(chK)
		defer close(chT)

		for key, value := range source {
			select {
			case <-ctx.Done():
				return
			default:
				if modify != nil {
					var err error
					key, value = modify(key, value)
					if err != nil {
						return
					}
				}

				chK <- key
				chT <- value
			}
		}
	}()

	return chK, chT
}

func NewFromSyncMap[K comparable, T any](source *sync.Map, modify modifier[K, T]) (<-chan K, <-chan T) {
	return NewFromSyncMapWithContext[K, T](context.Background(), source, modify)
}

func NewFromSyncMapWithContext[K comparable, T any](
	ctx context.Context,
	source *sync.Map,
	modify modifier[K, T],
) (<-chan K, <-chan T) {
	chK := make(chan K, 1)
	chT := make(chan T, 1)

	go func() {
		defer close(chK)
		defer close(chT)

		source.Range(func(key, value any) bool {
			select {
			case <-ctx.Done():
				return false
			default:
				keyK, valueT := key.(K), value.(T)
				if modify != nil {
					var err error
					key, value = modify(keyK, valueT)
					if err != nil {
						return false
					}
				}

				chK <- keyK
				chT <- valueT
				return true
			}
		})
	}()

	return chK, chT
}
