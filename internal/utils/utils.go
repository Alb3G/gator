package utils

import (
	"fmt"
	"strconv"
	"time"
)

func Now() time.Time {
	return time.Now().UTC()
}

func ParsePublishedDate(dateStr string) (time.Time, error) {
	layouts := []string{
		time.RFC1123Z, // "Mon, 02 Jan 2006 15:04:05 -0700"
		time.RFC1123,  // "Mon, 02 Jan 2006 15:04:05 MST"
		time.RFC3339,  // "2006-01-02T15:04:05Z07:00"
		time.RFC822,   // "02 Jan 06 15:04 MST"
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

func ParseLimit(args []string, defaultLimit int32) int32 {
	if len(args) < 2 {
		return defaultLimit
	}

	i, err := strconv.ParseInt(args[1], 10, 32)
	if err != nil || i <= 0 || i > 100 {
		return defaultLimit
	}

	return int32(i)
}
