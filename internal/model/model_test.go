package model_test

import (
	"testing"

	"com.telekom-mms.corp-net-indicator/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestWrite(t *testing.T) {
	ctx := model.NewContext()
	values := ctx.Write(func(values *model.ContextValues) {
		assert.Equal(t, &model.ContextValues{}, values)
		values.Connected = true
	})
	assert.Equal(t, model.ContextValues{Connected: true}, values)
}

func TestRead(t *testing.T) {
	ctx := model.NewContext()
	ctx.Write(func(values *model.ContextValues) {
		assert.Equal(t, &model.ContextValues{}, values)
		values.Connected = true
	})
	assert.Equal(t, model.ContextValues{Connected: true}, ctx.Read())
}
