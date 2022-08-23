package source

func toPtr[T any](a T) *T {
	return &a
}
