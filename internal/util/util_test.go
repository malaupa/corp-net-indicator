package util_test

import (
	"testing"

	"de.telekom-mms.corp-net-indicator/internal/util"
	"github.com/stretchr/testify/assert"
)

func TestFormatDate(t *testing.T) {
	assert.Equal(t, "12.12.2002 19:30:12", util.FormatDate(1039717812))
	assert.Equal(t, "-", util.FormatDate(0))
	assert.Equal(t, "01.01.1970 01:00:01", util.FormatDate(1))
}

func TestFormatValue(t *testing.T) {
	assert.Equal(t, "-", util.FormatValue(""))
	assert.Equal(t, "value", util.FormatValue("value"))
}
