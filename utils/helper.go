package utils

func Find[T comparable](slice []T, item T) int {
	for i := range slice {
		if slice[i] == item {
			return i
		}
	}
	return -1
}
