package tools

func IsString(v interface{}) bool {
	switch v.(type) {
	case string:
		return true
	}

	return false
}

func IsFloat64(v interface{}) bool {
	switch v.(type) {
	case float64:
		return true
	}

	return false
}