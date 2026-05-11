package query

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func boolPtrToIntPtr(value *bool) *int {
	if value == nil {
		return nil
	}

	intValue := boolToInt(*value)
	return &intValue
}
