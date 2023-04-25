package util_test

import (
	"testing"

	"de.telekom-mms.corp-net-indicator/internal/test"
	"de.telekom-mms.corp-net-indicator/internal/util"
	"github.com/stretchr/testify/assert"
)

func TestFormatDate(t *testing.T) {
	assert.Equal(t, "12.12.2002 19:30:12", util.FormatDate(test.Pointer(int64(1039717812))))
	assert.Equal(t, "-", util.FormatDate(test.Pointer(int64(0))))
	assert.Equal(t, "01.01.1970 01:00:01", util.FormatDate(test.Pointer(int64(1))))
}

func TestFormatValue(t *testing.T) {
	assert.Equal(t, "-", util.FormatValue(test.Pointer("")))
	assert.Equal(t, "value", util.FormatValue(test.Pointer("value")))
}
