package helpers

import (
	"time"
)

func GetCurrentTimeZone() (*time.Location, error) {
	location := "Asia/Colombo"
	currentTZ, currentTZError := time.LoadLocation(location)

	if currentTZError != nil {
		return nil, currentTZError
	}

	return currentTZ, nil
}
