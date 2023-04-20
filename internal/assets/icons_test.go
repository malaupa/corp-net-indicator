package assets_test

import (
	"testing"

	"de.telekom-mms.corp-net-indicator/internal/assets"
	"github.com/stretchr/testify/assert"
)

func TestGetIcon(t *testing.T) {
	assert := assert.New(t)

	// good cases
	assert.NotEmpty(assets.GetIcon(assets.ShieldOff))
	assert.NotEmpty(assets.GetIcon(assets.ShieldOn))
	assert.NotEmpty(assets.GetIcon(assets.Umbrella))
	assert.NotEmpty(assets.GetIcon(assets.Connect))
	assert.NotEmpty(assets.GetIcon(assets.Disconnect))
	assert.NotEmpty(assets.GetIcon(assets.Status))

	// bad case
	assert.Panics(func() {
		assets.GetIcon("not existing")
	})
}
