package generator

import (
	"context"
	"sync"
)

type modifier[K comparable, T any] func(key K, value T) (K, T)
type valueModifier[T any] func(value T) T

func NewFromFunction[T any](generate func() (T, bool)) <-chan T {
	return NewFromFunctionWithContext[T](context.Background(), generate)
}

func NewFromFunctionWithContext[T any](ctx context.Context, generate func() (T, bool)) <-chan T {
	if generate == nil {
		panic("generate function cannot be nil")
	}

	channel := make(chan T, 1)

	go func() {
		defer close(channel)

		for {
			select {
			case <-ctx.Done():
				return
			default:
				value, ok := generate()
				if !ok {
					return
				}

				channel <- value
			}
		}
	}()

	return channel
}

type mapItem[K comparable, T any] struct {
	Key   K
	Value T
}

func NewFromMap[K comparable, T any](source map[K]T, modify modifier[K, T]) <-chan mapItem[K, T] {
	return NewFromMapWithContext[K, T](context.Background(), source, modify)
}

func NewFromMapWithContext[K comparable, T any](
	ctx context.Context,
	source map[K]T,
	modify modifier[K, T],
) <-chan mapItem[K, T] {
	channel := make(chan mapItem[K, T], 1)

	go func() {
		defer close(channel)

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

				channel <- mapItem[K, T]{key, value}
			}
		}
	}()

	return channel
}

func NewFromSyncMap[K comparable, T any](source *sync.Map, modify modifier[K, T]) <-chan mapItem[K, T] {
	return NewFromSyncMapWithContext[K, T](context.Background(), source, modify)
}

func NewFromSyncMapWithContext[K comparable, T any](
	ctx context.Context,
	source *sync.Map,
	modify modifier[K, T],
) <-chan mapItem[K, T] {
	channel := make(chan mapItem[K, T], 1)

	go func() {
		defer close(channel)

		source.Range(func(key, value any) bool {
			select {
			case <-ctx.Done():
				return false
			default:
				keyK, valueT := key.(K), value.(T)
				if modify != nil {
					var err error
					keyK, valueT = modify(keyK, valueT)
					if err != nil {
						return false
					}
				}

				channel <- mapItem[K, T]{keyK, valueT}
				return true
			}
		})
	}()

	return channel
}

func NewFromMapOnlyValue[K comparable, T any](source map[K]T, modify valueModifier[T]) <-chan T {
	return NewFromMapOnlyValueWithContext[K, T](context.Background(), source, modify)
}

func NewFromMapOnlyValueWithContext[K comparable, T any](
	ctx context.Context,
	source map[K]T,
	modify valueModifier[T],
) <-chan T {
	channel := make(chan T, 1)

	go func() {
		defer close(channel)

		for _, value := range source {
			select {
			case <-ctx.Done():
				return
			default:
				if modify != nil {
					var err error
					value = modify(value)
					if err != nil {
						return
					}
				}

				channel <- value
			}
		}
	}()

	return channel
}

func NewFromSyncMapOnlyValue[T any](source *sync.Map, modify valueModifier[T]) <-chan T {
	return NewFromSyncMapOnlyValueWithContext[T](context.Background(), source, modify)
}

func NewFromSyncMapOnlyValueWithContext[T any](
	ctx context.Context,
	source *sync.Map,
	modify valueModifier[T],
) <-chan T {
	channel := make(chan T, 1)

	go func() {
		defer close(channel)

		source.Range(func(_, value any) bool {
			select {
			case <-ctx.Done():
				return false
			default:
				valueT := value.(T)
				if modify != nil {
					var err error
					valueT = modify(valueT)
					if err != nil {
						return false
					}
				}

				channel <- valueT
				return true
			}
		})
	}()

	return channel
}
