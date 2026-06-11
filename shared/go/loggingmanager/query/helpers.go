package query

// boolToInt converts a boolean flag to SQLite-compatible 0/1 integer.
func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

// boolPtrToIntPtr converts optional boolean flag to optional 0/1 integer.
func boolPtrToIntPtr(value *bool) *int {
	if value == nil {
		return nil
	}

	intValue := boolToInt(*value)
	return &intValue
}
