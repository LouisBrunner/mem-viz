package utils

func MapSlice[T, U any](in []T, mapper func(in T) U) []U {
	out := make([]U, 0, len(in))
	for _, entry := range in {
		out = append(out, mapper(entry))
	}
	return out
}
