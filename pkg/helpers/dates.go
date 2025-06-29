package helpers

import (
	"fmt"
	"time"
)

func TimeAsCalendarDateOnly(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

func ParseFlexibleDate(dateStr string) (time.Time, error) {
	// Common date formats to try
	formats := []string{
		"2006-01-02",           // YYYY-MM-DD
		"01/02/2006",           // MM/DD/YYYY
		"02/01/2006",           // DD/MM/YYYY
		"2006-01-02T15:04:05Z", // ISO 8601
		"2006-01-02 15:04:05",  // YYYY-MM-DD HH:MM:SS
		"Jan 2, 2006",          // Jan 2, 2006
		"January 2, 2006",      // January 2, 2006
		"02-Jan-2006",          // DD-MMM-YYYY
		"2006/01/02",           // YYYY/MM/DD
	}

	for _, format := range formats {
		if parsedDate, err := time.Parse(format, dateStr); err == nil {
			return parsedDate, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}
