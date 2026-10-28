package clipboard

// Fallback clipboard implementation for headless and non-GUI Windows environments.
func ReadAll() (string, error) {
	return "", nil
}

func WriteAll(text string) error {
	return nil
}

var Unsupported = false
