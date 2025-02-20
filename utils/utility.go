package bank_integration_utils

import "time"

func ParseDateTime(dateTimeStr string) (time.Time, error) {
	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04:05", // Without timezone
	}

	var t time.Time
	var err error
	for _, layout := range layouts {
		t, err = time.Parse(layout, dateTimeStr)
		if err == nil {
			return t, nil
		}
	}
	return t, err
}
