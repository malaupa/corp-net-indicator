package util

import "time"

const DATE_TIME_FORMAT = "02.01.2006 15:04:05"

func FormatDate(t int64) string {
	return time.Unix(t, 0).Local().Format(DATE_TIME_FORMAT)
}
