package slice

// Chunk returns a channel over consecutive sub-slices of up to n elements of slice.
// All but the last sub-slice will have size n.
// All sub-slices are clipped to have no capacity beyond the length.
// If s is empty, the sequence is empty: there is no empty slice in the sequence.
// Chunk panics if n is less than 1.
//
// Based on Go 1.23 source code
func Chunk[T any](slice []T, n uint64) <-chan []T {
	if n == 0 {
		panic("n can`t be less than 1")
	}

	channel := make(chan []T, 1)

	go func() {
		defer close(channel)
		for i := uint64(0); i < uint64(len(slice)); i += n {
			// Clamp the last chunk to the slice bound as necessary.
			end := min(n, uint64(len(slice[i:])))

			// Set the capacity of each chunk so that appending to a chunk does
			// not modify the original slice.
			channel <- slice[i : i+end : i+end]
		}
	}()

	return channel
}

func FromChannel[T any](channel <-chan T) []T {
	slice := make([]T, 0)
	for item := range channel {
		slice = append(slice, item)
	}

	return slice
}
