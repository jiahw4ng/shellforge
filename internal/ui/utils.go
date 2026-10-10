package ui

// DimensionWithFallback returns fallback when value is not a usable dimension.
func DimensionWithFallback(value, fallback int) int {
	if value <= 0 {
		return fallback
	}
	return value
}
