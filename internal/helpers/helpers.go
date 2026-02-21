package helpers

func NullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}
