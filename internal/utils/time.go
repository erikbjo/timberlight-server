package utils

import "time"

// IsEarlierThanToday checks if the given date is earlier than today.
func IsEarlierThanToday(date string) (bool, error) {
	parsed, err := time.Parse("2006-01-02", date)
	if err != nil {
		return false, err
	}

	today := time.Now()
	today = time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	return parsed.Before(today), nil
}
