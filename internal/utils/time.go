package utils

import (
	"log"
	"strconv"
	"time"
)

func GetDuration(seconds string) (time.Duration, error) {
	secondsInt, err := strconv.Atoi(seconds)
	if err != nil {
		log.Printf("invalid value for %s: %v\n", seconds, err)
		return 0, err
	}
	return time.Duration(secondsInt) * time.Second, nil
}
