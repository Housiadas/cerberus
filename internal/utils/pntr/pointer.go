package pntr

// DerefStr returns the string value or empty string if nil.
func DerefStr(s *string) string {
	if s == nil {
		return ""
	}

	return *s
}
