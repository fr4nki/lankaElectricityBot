package helpers

import (
	"time"
)

func GetDateTimeRange(dayOffset int) (time.Time, time.Time, error) {
	now := time.Now()

	tz, tzError := GetCurrentTimeZone()
	if tzError != nil {
		return now, now, tzError
	}

	nowInTz := now.In(tz)
	nextInTz := now.In(tz).AddDate(0, 0, dayOffset)

	return nowInTz, nextInTz, nil
}
