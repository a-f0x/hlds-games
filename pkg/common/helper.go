package common

func StringPtr(value string) *string {
	return &value
}

func StringOrNil(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
