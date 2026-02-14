package utils

import "time"

func ParseTimeRFC3339(timeString string) (time.Time, error) {
	return time.Parse(time.RFC3339, timeString)
}

func ParseTimeRFC3339Ptr(timeString *string) (*time.Time, error) {
	if timeString == nil {
		return nil, nil
	}
	time, err := time.Parse(time.RFC3339, *timeString)
	if err != nil {
		return nil, err
	}
	return &time, nil
}

func FormatTimeRFC3339(t time.Time) string {
	return t.Format(time.RFC3339)
}

func FormatTimeRFC3339Ptr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	timeString := FormatTimeRFC3339(*t)
	return &timeString
}
