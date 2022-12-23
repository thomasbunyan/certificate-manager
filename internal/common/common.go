package common

// Ptr generically returns a pointer to the provided value.
func Ptr[T any](v T) *T {
	return &v
}

// Val generically returns the value of a provided pointer, or its zero value if the pointer is `nil`.
func Val[T any](v *T) T {
	if v != nil {
		return *v
	}
	return *new(T)
}
