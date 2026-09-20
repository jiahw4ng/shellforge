package ui

func DimensionWithFallback(value, fallback int) int {
	if value <= 0 {
		return fallback
	}
	return value
}
