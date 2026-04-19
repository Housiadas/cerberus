package mapper

// MapSlice applies fn to each element of src and returns the results.
func MapSlice[S, D any](src []S, fn func(S) D) []D {
	out := make([]D, len(src))
	for i, s := range src {
		out[i] = fn(s)
	}

	return out
}
