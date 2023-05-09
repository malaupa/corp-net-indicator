package util_test

import (
	"testing"
	"time"

	"com.telekom-mms.corp-net-indicator/internal/test"
	"com.telekom-mms.corp-net-indicator/internal/util"
	"github.com/stretchr/testify/assert"
)

func TestFormatDate(t *testing.T) {
	assert.Equal(t, time.Unix(1039717812, 0).Local().Format(util.DATE_TIME_FORMAT), util.FormatDate(test.Pointer(int64(1039717812))))
	assert.Equal(t, "-", util.FormatDate(test.Pointer(int64(0))))
	assert.Equal(t, time.Unix(1, 0).Local().Format(util.DATE_TIME_FORMAT), util.FormatDate(test.Pointer(int64(1))))
}

func TestFormatValue(t *testing.T) {
	assert.Equal(t, "-", util.FormatValue(test.Pointer("")))
	assert.Equal(t, "value", util.FormatValue(test.Pointer("value")))
}
