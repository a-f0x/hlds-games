package common

func StringOrNil(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
