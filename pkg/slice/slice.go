package slice

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
