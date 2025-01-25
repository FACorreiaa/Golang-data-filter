package internal

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func ParseDateOrYear(raw string) (time.Time, error) {
	if strings.Contains(raw, "-") {
		t, err := time.Parse("2006-01-02", raw)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid date format %q: %v", raw, err)
		}
		return t, nil
	}

	yearInt, err := strconv.Atoi(raw)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid year %q: %v", raw, err)
	}

	//dateString := fmt.Sprintf("%04d-12-31", yearInt)
	return time.Date(yearInt, time.January, 1, 0, 0, 0, 0, time.UTC), nil
}
