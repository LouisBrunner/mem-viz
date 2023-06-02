package parse

func roundUp(x, y uint64) uint64 {
	return (x + y - 1) / y * y
}
